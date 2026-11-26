package identity

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var (
	phonePattern = regexp.MustCompile(`^[0-9]+$`)
	otpPattern   = regexp.MustCompile(`^[0-9]{6}$`)
	regions      = map[string]struct{}{"+86": {}, "+852": {}}
	locales      = map[string]struct{}{"zh-Hans": {}, "zh-Hant": {}}
	purposes     = map[string]struct{}{"register": {}, "login": {}}
)

// Options configures the identity service.
// Zero durations and counts use the documented defaults.
// Code, when set, replaces the random code generator. Production leaves it nil.
// Store persists accounts and sessions. Nil keeps those records in process memory.
type Options struct {
	Sender           Sender
	Clock            Clock
	Code             func() (string, error)
	Store            Store
	SendCooldown     time.Duration
	SendBudget       int
	SendWindow       time.Duration
	SendIPBudget     int
	VerifyBudget     int
	VerifyWindow     time.Duration
	VerifyIPBudget   int
	OTPTTL           time.Duration
	OTPMaxAttempts   int
	AccessTTL        time.Duration
	RefreshTTL       time.Duration
	IdempotencyTTL   time.Duration
	NicknameMaxRunes int
	AnswerMaxBytes   int
	// RetainCode keeps the current code in process memory for a staging reveal
	// route. Production and every other environment leave it false. The code
	// is still not written to logs or to the learner send response.
	RetainCode bool
}

// Service is the learner identity API.
type Service struct {
	mu          sync.Mutex
	pepper      []byte
	sender      Sender
	clock       Clock
	code        func() (string, error)
	cfg         limits
	accounts    Store
	otps        map[string]*otpRecord
	idem        map[string]*idemRecord
	sendPhone   *bucket
	sendIP      *bucket
	verifyPhone *bucket
	verifyIP    *bucket
	retainCode  bool
}

type limits struct {
	sendCooldown     time.Duration
	sendBudget       int
	sendWindow       time.Duration
	sendIPBudget     int
	verifyBudget     int
	verifyWindow     time.Duration
	verifyIPBudget   int
	otpTTL           time.Duration
	otpMaxAttempts   int
	accessTTL        time.Duration
	refreshTTL       time.Duration
	idempotencyTTL   time.Duration
	nicknameMaxRunes int
	answerMaxBytes   int
}

type user struct {
	ID            string
	PhoneRegion   string
	PhoneNational string
	Locale        string
	RegionPref    string
	Answers       map[string]string
	HasFirstPack  bool
}

type otpRecord struct {
	Hash     []byte
	Expires  time.Time
	Attempts int
	plain    []byte
}

type session struct {
	ID            string
	UserID        string
	AccessHash    [32]byte
	RefreshHash   [32]byte
	AccessExpiry  time.Time
	RefreshExpiry time.Time
	Revoked       bool
}

type idemRecord struct {
	Hash    [32]byte
	Status  int
	Body    []byte
	ErrCode string
	ErrText string
	Details map[string]any
	Expires time.Time
}

type bucket struct {
	stamps map[string][]time.Time
	ready  map[string]time.Time
}

// Result is an HTTP result. Err is set for contract failures.
type Result struct {
	Status int
	Body   []byte
	Err    *Error
}

type otpSendBody struct {
	Region  string  `json:"region_code"`
	Phone   string  `json:"phone_national"`
	Purpose string  `json:"purpose"`
	Locale  *string `json:"locale"`
}

type otpVerifyBody struct {
	Region  string  `json:"region_code"`
	Phone   string  `json:"phone_national"`
	Purpose string  `json:"purpose"`
	Code    string  `json:"code"`
	Locale  *string `json:"locale"`
}

type refreshBody struct {
	RefreshToken string `json:"refresh_token"`
}

type revokeBody struct {
	RefreshToken string `json:"refresh_token"`
}

type onboardingBody struct {
	StepID   string  `json:"step_id"`
	Nickname *string `json:"nickname"`
	Answer   *string `json:"answer"`
}

type profileBody struct {
	Locale *string `json:"locale"`
	Region *string `json:"region_code"`
}

// New builds an identity service. A nil Store keeps accounts and sessions in process memory.
func New(opts Options) (*Service, error) {
	pepper := make([]byte, 32)
	if _, err := rand.Read(pepper); err != nil {
		return nil, err
	}
	sender := opts.Sender
	if sender == nil {
		sender = DiscardSender{}
	}
	clock := opts.Clock
	if clock == nil {
		clock = systemClock{}
	}
	code := opts.Code
	if code == nil {
		code = randomCode
	}
	accounts := opts.Store
	if accounts == nil {
		accounts = newMemoryStore()
	}
	return &Service{
		pepper:      pepper,
		sender:      sender,
		clock:       clock,
		code:        code,
		cfg:         limitsFrom(opts),
		accounts:    accounts,
		otps:        map[string]*otpRecord{},
		idem:        map[string]*idemRecord{},
		sendPhone:   newBucket(),
		sendIP:      newBucket(),
		verifyPhone: newBucket(),
		verifyIP:    newBucket(),
		retainCode:  opts.RetainCode,
	}, nil
}

func limitsFrom(opts Options) limits {
	return limits{
		sendCooldown:     orDuration(opts.SendCooldown, DefaultSendCooldown),
		sendBudget:       orInt(opts.SendBudget, DefaultSendBudget),
		sendWindow:       orDuration(opts.SendWindow, DefaultSendWindow),
		sendIPBudget:     orInt(opts.SendIPBudget, DefaultSendIPBudget),
		verifyBudget:     orInt(opts.VerifyBudget, DefaultVerifyBudget),
		verifyWindow:     orDuration(opts.VerifyWindow, DefaultVerifyWindow),
		verifyIPBudget:   orInt(opts.VerifyIPBudget, DefaultVerifyIPBudget),
		otpTTL:           orDuration(opts.OTPTTL, DefaultOTPTTL),
		otpMaxAttempts:   orInt(opts.OTPMaxAttempts, DefaultOTPMaxAttempts),
		accessTTL:        orDuration(opts.AccessTTL, DefaultAccessTTL),
		refreshTTL:       orDuration(opts.RefreshTTL, DefaultRefreshTTL),
		idempotencyTTL:   orDuration(opts.IdempotencyTTL, DefaultIdempotencyTTL),
		nicknameMaxRunes: orInt(opts.NicknameMaxRunes, DefaultNicknameMaxRunes),
		answerMaxBytes:   orInt(opts.AnswerMaxBytes, DefaultAnswerMaxBytes),
	}
}

func newBucket() *bucket {
	return &bucket{stamps: map[string][]time.Time{}, ready: map[string]time.Time{}}
}

// Send accepts a raw OTP send body. key is the Idempotency-Key.
func (s *Service) Send(ctx context.Context, key string, raw []byte, remote string) Result {
	if !validIdempotencyKey(key) {
		return Result{Err: codeErr("VALIDATION_MISSING_FIELD")}
	}
	canon, err := canonicalJSON(raw)
	if err != nil {
		return Result{Err: codeErr("VALIDATION_MISSING_FIELD")}
	}
	hash := sha256.Sum256(canon)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sweepIdem()
	if result, found, mismatch := s.lookupIdem(key, hash); mismatch {
		return Result{Err: codeErr("IDEMPOTENCY_KEY_MISMATCH")}
	} else if found {
		return result
	}
	var body otpSendBody
	if decErr := decodeObject(raw, &body); decErr != nil {
		return s.remember(key, hash, Result{Err: decErr})
	}
	if fieldErr := validateContact(body.Region, body.Phone, body.Purpose, body.Locale); fieldErr != nil {
		return s.remember(key, hash, Result{Err: fieldErr})
	}
	if gate := s.accountGate(ctx, body.Region, body.Phone, body.Purpose); gate != nil {
		return s.remember(key, hash, Result{Err: gate})
	}
	phoneKey := phoneKey(body.Region, body.Phone)
	ip := clientIP(remote)
	if limited, wait := s.sendLimited(phoneKey, ip); limited {
		return s.remember(key, hash, Result{Err: rateErr(wait)})
	}
	code, codeErrs := s.code()
	if codeErrs != nil || !otpPattern.MatchString(code) {
		return Result{Err: codeErr("INTERNAL")}
	}
	locale := ""
	if body.Locale != nil {
		locale = *body.Locale
	}
	if sendErr := s.sender.Send(ctx, SMS{
		RegionCode:    body.Region,
		PhoneNational: body.Phone,
		Purpose:       body.Purpose,
		Code:          code,
		Locale:        locale,
	}); sendErr != nil {
		return Result{Err: codeErr("INTERNAL")}
	}
	now := s.clock.Now()
	record := &otpRecord{
		Hash:    s.hashCode(phoneKey, body.Purpose, code),
		Expires: now.Add(s.cfg.otpTTL),
	}
	if s.retainCode {
		record.plain = append([]byte(nil), code...)
	}
	otpID := otpKey(phoneKey, body.Purpose)
	s.dropOTP(otpID)
	s.otps[otpID] = record
	s.mark(s.sendPhone, phoneKey, s.cfg.sendWindow)
	s.mark(s.sendIP, ip, s.cfg.sendWindow)
	s.sendPhone.ready[phoneKey] = now.Add(s.cfg.sendCooldown)
	payload, mErr := json.Marshal(sendView{
		RetryAfter: int64(s.cfg.sendCooldown / time.Second),
		Hint:       destinationHint(body.Region, body.Phone),
	})
	if mErr != nil {
		return Result{Err: codeErr("INTERNAL")}
	}
	return s.remember(key, hash, Result{Status: 200, Body: payload})
}

// Verify checks a code and issues a session. The access token is not an input.
func (s *Service) Verify(ctx context.Context, raw []byte, remote string) Result {
	var body otpVerifyBody
	if decErr := decodeObject(raw, &body); decErr != nil {
		return Result{Err: decErr}
	}
	if fieldErr := validateContact(body.Region, body.Phone, body.Purpose, body.Locale); fieldErr != nil {
		return Result{Err: fieldErr}
	}
	if body.Code == "" {
		return Result{Err: codeErr("VALIDATION_MISSING_FIELD")}
	}
	if !otpPattern.MatchString(body.Code) {
		return Result{Err: codeErr("VALIDATION_INVALID_OTP_FORMAT")}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if gate := s.accountGate(ctx, body.Region, body.Phone, body.Purpose); gate != nil {
		return Result{Err: gate}
	}
	phoneKey := phoneKey(body.Region, body.Phone)
	ip := clientIP(remote)
	if limited, wait := s.verifyLimited(phoneKey, ip); limited {
		return Result{Err: rateErr(wait)}
	}
	s.mark(s.verifyPhone, phoneKey, s.cfg.verifyWindow)
	s.mark(s.verifyIP, ip, s.cfg.verifyWindow)
	otpID := otpKey(phoneKey, body.Purpose)
	record := s.otps[otpID]
	now := s.clock.Now()
	if record == nil || !now.Before(record.Expires) {
		s.dropOTP(otpID)
		return Result{Err: codeErr("AUTH_OTP_EXPIRED")}
	}
	if record.Attempts >= s.cfg.otpMaxAttempts {
		return Result{Err: codeErr("AUTH_OTP_MAX_ATTEMPTS")}
	}
	if !hmac.Equal(record.Hash, s.hashCode(phoneKey, body.Purpose, body.Code)) {
		record.Attempts++
		if record.Attempts >= s.cfg.otpMaxAttempts {
			return Result{Err: codeErr("AUTH_OTP_MAX_ATTEMPTS")}
		}
		return Result{Err: codeErr("AUTH_OTP_INVALID")}
	}
	return s.commitVerify(ctx, body, phoneKey)
}

func (s *Service) commitVerify(ctx context.Context, body otpVerifyBody, phoneKey string) Result {
	locale := defaultLocale
	if body.Locale != nil {
		locale = *body.Locale
	}
	var account *user
	created := false
	if body.Purpose == "register" {
		account = &user{
			ID:            newID("usr_"),
			PhoneRegion:   body.Region,
			PhoneNational: body.Phone,
			Locale:        locale,
			RegionPref:    body.Region,
			Answers:       map[string]string{},
		}
		if err := s.accounts.Insert(ctx, account); err != nil {
			if errors.Is(err, errAccountExists) {
				return Result{Err: codeErr("AUTH_ACCOUNT_EXISTS")}
			}
			return Result{Err: codeErr("INTERNAL")}
		}
		created = true
	} else {
		found, err := s.accounts.FindByPhone(ctx, body.Region, body.Phone)
		if err != nil {
			return Result{Err: codeErr("INTERNAL")}
		}
		if found == nil {
			return Result{Err: codeErr("AUTH_ACCOUNT_NOT_FOUND")}
		}
		account = found
		if body.Locale != nil {
			account.Locale = locale
		}
		account.RegionPref = body.Region
		if err := s.accounts.SaveProfile(ctx, account.ID, account.Locale, account.RegionPref); err != nil {
			return Result{Err: codeErr("INTERNAL")}
		}
	}
	issued, err := s.issue(ctx, account, created)
	if err != nil {
		if created {
			_ = s.accounts.DeleteAccount(ctx, account.ID)
		}
		return Result{Err: codeErr("INTERNAL")}
	}
	s.dropOTP(otpKey(phoneKey, body.Purpose))
	return issued
}

// dropOTP removes a challenge and zeroes any retained plaintext. The caller holds s.mu.
func (s *Service) dropOTP(key string) {
	if rec := s.otps[key]; rec != nil && len(rec.plain) > 0 {
		clear(rec.plain)
	}
	delete(s.otps, key)
}

// revealCode returns the current code when this process was started with RetainCode.
// It does not log the code and does not remove the challenge.
func (s *Service) revealCode(raw []byte) (string, bool) {
	if s == nil || !s.retainCode {
		return "", false
	}
	var body otpSendBody
	if decodeObject(raw, &body) != nil {
		return "", false
	}
	if validateContact(body.Region, body.Phone, body.Purpose, body.Locale) != nil {
		return "", false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	otpID := otpKey(phoneKey(body.Region, body.Phone), body.Purpose)
	record := s.otps[otpID]
	if record == nil || !s.clock.Now().Before(record.Expires) {
		s.dropOTP(otpID)
		return "", false
	}
	code := string(record.plain)
	if !otpPattern.MatchString(code) {
		return "", false
	}
	return code, true
}

// Refresh rotates a refresh token. It does not read the access token.
func (s *Service) Refresh(ctx context.Context, raw []byte) Result {
	var body refreshBody
	if decErr := decodeObject(raw, &body); decErr != nil {
		return Result{Err: decErr}
	}
	if strings.TrimSpace(body.RefreshToken) == "" {
		return Result{Err: codeErr("VALIDATION_MISSING_FIELD")}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, err := s.sessionByHash(ctx, false, body.RefreshToken)
	if err != nil {
		return Result{Err: codeErr("INTERNAL")}
	}
	now := s.clock.Now()
	if sess == nil || sess.Revoked || !now.Before(sess.RefreshExpiry) {
		return Result{Err: codeErr("AUTH_EXPIRED")}
	}
	account, err := s.accounts.FindByID(ctx, sess.UserID)
	if err != nil {
		return Result{Err: codeErr("INTERNAL")}
	}
	if account == nil {
		return Result{Err: codeErr("AUTH_EXPIRED")}
	}
	if err := s.dropTokens(ctx, sess); err != nil {
		return Result{Err: codeErr("INTERNAL")}
	}
	issued, err := s.issue(ctx, account, false)
	if err != nil {
		return Result{Err: codeErr("INTERNAL")}
	}
	return issued
}

// Revoke drops the session for an access token, a refresh token, or both.
// An unknown token still succeeds so logout can be repeated.
func (s *Service) Revoke(ctx context.Context, accessToken string, raw []byte) Result {
	if strings.HasPrefix(accessToken, prefixInternal) {
		return Result{Err: codeErr("AUTH_WRONG_REALM")}
	}
	var refresh string
	if len(strings.TrimSpace(string(raw))) > 0 {
		var body revokeBody
		if decErr := decodeObject(raw, &body); decErr != nil {
			return Result{Err: decErr}
		}
		refresh = body.RefreshToken
		if strings.HasPrefix(refresh, prefixInternal) {
			return Result{Err: codeErr("AUTH_WRONG_REALM")}
		}
	}
	if accessToken == "" && strings.TrimSpace(refresh) == "" {
		return Result{Err: codeErr("AUTH_REQUIRED")}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if accessToken != "" {
		sess, err := s.sessionByHash(ctx, true, accessToken)
		if err != nil {
			return Result{Err: codeErr("INTERNAL")}
		}
		if sess != nil {
			if err := s.dropTokens(ctx, sess); err != nil {
				return Result{Err: codeErr("INTERNAL")}
			}
		}
	}
	if strings.TrimSpace(refresh) != "" {
		sess, err := s.sessionByHash(ctx, false, refresh)
		if err != nil {
			return Result{Err: codeErr("INTERNAL")}
		}
		if sess != nil {
			if err := s.dropTokens(ctx, sess); err != nil {
				return Result{Err: codeErr("INTERNAL")}
			}
		}
	}
	return Result{Status: 204}
}

// Onboarding returns the 7-step state for an access token.
func (s *Service) Onboarding(ctx context.Context, accessToken string) Result {
	s.mu.Lock()
	defer s.mu.Unlock()
	account, err := s.authenticate(ctx, accessToken)
	if err != nil {
		return Result{Err: err}
	}
	return s.marshal(s.onboardingView(account, false))
}

// WriteOnboarding stores the current step or rewrites a finished step.
func (s *Service) WriteOnboarding(ctx context.Context, accessToken string, raw []byte) Result {
	s.mu.Lock()
	defer s.mu.Unlock()
	account, err := s.authenticate(ctx, accessToken)
	if err != nil {
		return Result{Err: err}
	}
	var body onboardingBody
	if decErr := decodeObject(raw, &body); decErr != nil {
		return Result{Err: decErr}
	}
	if strings.TrimSpace(body.StepID) == "" {
		return Result{Err: codeErr("VALIDATION_MISSING_FIELD")}
	}
	if stepIndex(body.StepID) < 0 {
		return Result{Err: codeErr("VALIDATION_INVALID_STEP")}
	}
	if orderErr := s.order(account, body.StepID); orderErr != nil {
		return Result{Err: orderErr}
	}
	if body.StepID == "nickname" {
		if body.Nickname == nil {
			return Result{Err: codeErr("VALIDATION_MISSING_FIELD")}
		}
		trimmed := strings.TrimSpace(*body.Nickname)
		if nickErr := validateNickname(trimmed, s.cfg.nicknameMaxRunes); nickErr != nil {
			return Result{Err: nickErr}
		}
		if saveErr := s.saveAnswer(ctx, account, "nickname", trimmed); saveErr != nil {
			return Result{Err: saveErr}
		}
	} else {
		if body.Answer == nil {
			return Result{Err: codeErr("VALIDATION_MISSING_FIELD")}
		}
		trimmed := strings.TrimSpace(*body.Answer)
		if answerErr := validateAnswer(trimmed, s.cfg.answerMaxBytes); answerErr != nil {
			return Result{Err: answerErr}
		}
		if saveErr := s.saveAnswer(ctx, account, body.StepID, trimmed); saveErr != nil {
			return Result{Err: saveErr}
		}
	}
	return s.marshal(s.onboardingView(account, false))
}

// Profile returns locale and region preference.
func (s *Service) Profile(ctx context.Context, accessToken string) Result {
	s.mu.Lock()
	defer s.mu.Unlock()
	account, err := s.authenticate(ctx, accessToken)
	if err != nil {
		return Result{Err: err}
	}
	return s.marshal(profileView{Locale: account.Locale, Region: account.RegionPref})
}

// UpdateProfile changes locale, region preference, or both. It does not change the bound phone.
func (s *Service) UpdateProfile(ctx context.Context, accessToken string, raw []byte) Result {
	s.mu.Lock()
	defer s.mu.Unlock()
	account, err := s.authenticate(ctx, accessToken)
	if err != nil {
		return Result{Err: err}
	}
	var body profileBody
	if decErr := decodeObject(raw, &body); decErr != nil {
		return Result{Err: decErr}
	}
	if body.Locale == nil && body.Region == nil {
		return Result{Err: codeErr("VALIDATION_MISSING_FIELD")}
	}
	if body.Locale != nil {
		if _, ok := locales[*body.Locale]; !ok {
			return Result{Err: codeErr("VALIDATION_UNSUPPORTED_LOCALE")}
		}
	}
	if body.Region != nil {
		if _, ok := regions[*body.Region]; !ok {
			return Result{Err: codeErr("VALIDATION_UNSUPPORTED_REGION")}
		}
	}
	if body.Locale != nil {
		account.Locale = *body.Locale
	}
	if body.Region != nil {
		account.RegionPref = *body.Region
	}
	if err := s.accounts.SaveProfile(ctx, account.ID, account.Locale, account.RegionPref); err != nil {
		return Result{Err: codeErr("INTERNAL")}
	}
	return s.marshal(profileView{Locale: account.Locale, Region: account.RegionPref})
}

// AccountSnapshot is the learner summary embedded by other domains.
// IsNewUser is false here. That flag is true only on the verify response that created the account.
type AccountSnapshot struct {
	UserID              string
	IsNewUser           bool
	OnboardingCompleted bool
	HasFirstPack        bool
	NextStep            string
}

// AccountFromAccess authenticates a learner access token and returns the current summary.
// NextStep still follows H3. H3 is an implementation assumption, not a confirmed product rule.
func (s *Service) AccountFromAccess(ctx context.Context, accessToken string) (AccountSnapshot, *Error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	account, err := s.authenticate(ctx, accessToken)
	if err != nil {
		return AccountSnapshot{}, err
	}
	return snapshotOf(account), nil
}

// CompleteFirstPack records that the first-pack step is done.
// It never clears the flag. It is not an HTTP route.
func (s *Service) CompleteFirstPack(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	account, err := s.accounts.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if account == nil {
		return errors.New("account missing")
	}
	if account.HasFirstPack {
		return nil
	}
	return s.accounts.SaveFirstPack(ctx, userID, true)
}

func snapshotOf(account *user) AccountSnapshot {
	return AccountSnapshot{
		UserID:              account.ID,
		IsNewUser:           false,
		OnboardingCompleted: firstIncomplete(account) < 0,
		HasFirstPack:        account.HasFirstPack,
		NextStep:            nextStepOf(account),
	}
}

// TestingSetFirstPack sets the knowledge-domain flag used by next_step.
// It is not an HTTP endpoint. H3 is an implementation assumption, not a confirmed product rule.
func (s *Service) TestingSetFirstPack(ctx context.Context, userID string, has bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	account, err := s.accounts.FindByID(ctx, userID)
	if err != nil || account == nil {
		return false
	}
	if err := s.accounts.SaveFirstPack(ctx, userID, has); err != nil {
		return false
	}
	return true
}

// TestingInsertUser creates an account without an OTP. Tests use it for the verify race.
func (s *Service) TestingInsertUser(ctx context.Context, region, national string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	account := &user{
		ID:            newID("usr_"),
		PhoneRegion:   region,
		PhoneNational: national,
		Locale:        defaultLocale,
		RegionPref:    region,
		Answers:       map[string]string{},
	}
	if err := s.accounts.Insert(ctx, account); err != nil {
		return ""
	}
	return account.ID
}

func (s *Service) authenticate(ctx context.Context, token string) (*user, *Error) {
	if token == "" {
		return nil, codeErr("AUTH_REQUIRED")
	}
	if strings.HasPrefix(token, prefixInternal) {
		return nil, codeErr("AUTH_WRONG_REALM")
	}
	sess, err := s.sessionByHash(ctx, true, token)
	if err != nil {
		return nil, codeErr("INTERNAL")
	}
	if sess == nil || sess.Revoked || !s.clock.Now().Before(sess.AccessExpiry) {
		return nil, codeErr("AUTH_EXPIRED")
	}
	account, err := s.accounts.FindByID(ctx, sess.UserID)
	if err != nil {
		return nil, codeErr("INTERNAL")
	}
	if account == nil {
		return nil, codeErr("AUTH_EXPIRED")
	}
	return account, nil
}

func (s *Service) saveAnswer(ctx context.Context, account *user, step, answer string) *Error {
	if account.Answers == nil {
		account.Answers = map[string]string{}
	}
	if err := s.accounts.SaveAnswer(ctx, account.ID, step, answer); err != nil {
		return codeErr("INTERNAL")
	}
	account.Answers[step] = answer
	return nil
}

func (s *Service) issue(ctx context.Context, account *user, isNew bool) (Result, error) {
	access, accessHash, err := newToken(prefixAccess)
	if err != nil {
		return Result{}, err
	}
	refresh, refreshHash, err := newToken(prefixRefresh)
	if err != nil {
		return Result{}, err
	}
	now := s.clock.Now()
	sess := &session{
		ID:            newID("ses_"),
		UserID:        account.ID,
		AccessHash:    accessHash,
		RefreshHash:   refreshHash,
		AccessExpiry:  now.Add(s.cfg.accessTTL),
		RefreshExpiry: now.Add(s.cfg.refreshTTL),
	}
	if err := s.accounts.InsertSession(ctx, sess); err != nil {
		return Result{}, err
	}
	secs := int64(s.cfg.accessTTL / time.Second)
	if secs < 1 {
		secs = 1
	}
	payload, err := json.Marshal(sessionView{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    secs,
		Account:      accountViewOf(account, isNew),
		NextStep:     nextStepOf(account),
		Profile:      profileView{Locale: account.Locale, Region: account.RegionPref},
	})
	if err != nil {
		_ = s.accounts.DeleteSession(ctx, sess.ID)
		return Result{}, err
	}
	return Result{Status: 200, Body: payload}, nil
}

func (s *Service) dropTokens(ctx context.Context, sess *session) error {
	if err := s.accounts.DeleteSession(ctx, sess.ID); err != nil {
		return err
	}
	sess.Revoked = true
	return nil
}

func (s *Service) sessionByHash(ctx context.Context, access bool, token string) (*session, error) {
	sum := sha256.Sum256([]byte(token))
	if access {
		return s.accounts.FindSessionByAccess(ctx, sum)
	}
	return s.accounts.FindSessionByRefresh(ctx, sum)
}

func (s *Service) accountGate(ctx context.Context, region, phone, purpose string) *Error {
	account, err := s.accounts.FindByPhone(ctx, region, phone)
	if err != nil {
		return codeErr("INTERNAL")
	}
	exists := account != nil
	if purpose == "register" && exists {
		return codeErr("AUTH_ACCOUNT_EXISTS")
	}
	if purpose == "login" && !exists {
		return codeErr("AUTH_ACCOUNT_NOT_FOUND")
	}
	return nil
}

func (s *Service) hashCode(phoneKey, purpose, code string) []byte {
	mac := hmac.New(sha256.New, s.pepper)
	_, _ = mac.Write([]byte(phoneKey))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(purpose))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(code))
	return mac.Sum(nil)
}

func (s *Service) lookupIdem(key string, hash [32]byte) (Result, bool, bool) {
	rec, ok := s.idem[key]
	if !ok {
		return Result{}, false, false
	}
	if !s.clock.Now().Before(rec.Expires) {
		delete(s.idem, key)
		return Result{}, false, false
	}
	if rec.Hash != hash {
		return Result{}, false, true
	}
	if rec.ErrCode != "" {
		return Result{Err: &Error{Code: rec.ErrCode, Message: rec.ErrText, Details: rec.Details}}, true, false
	}
	return Result{Status: rec.Status, Body: rec.Body}, true, false
}

func (s *Service) remember(key string, hash [32]byte, result Result) Result {
	if result.Err != nil && result.Err.Code == "INTERNAL" {
		return result
	}
	rec := &idemRecord{Hash: hash, Status: result.Status, Expires: s.clock.Now().Add(s.cfg.idempotencyTTL)}
	if result.Err != nil {
		rec.ErrCode = result.Err.Code
		rec.ErrText = result.Err.Message
		rec.Details = result.Err.Details
		rec.Status = statusOf(result.Err.Code)
	} else {
		rec.Body = append([]byte(nil), result.Body...)
	}
	s.idem[key] = rec
	return result
}

func (s *Service) sweepIdem() {
	if len(s.idem) < 256 {
		return
	}
	now := s.clock.Now()
	for key, rec := range s.idem {
		if !now.Before(rec.Expires) {
			delete(s.idem, key)
		}
	}
}

func (s *Service) sendLimited(phoneKey, ip string) (bool, int) {
	phoneBlocked, phoneWait := s.over(s.sendPhone, phoneKey, s.cfg.sendBudget, s.cfg.sendWindow, true)
	ipBlocked, ipWait := s.over(s.sendIP, ip, s.cfg.sendIPBudget, s.cfg.sendWindow, false)
	return mergeLimit(phoneBlocked, phoneWait, ipBlocked, ipWait)
}

func (s *Service) verifyLimited(phoneKey, ip string) (bool, int) {
	phoneBlocked, phoneWait := s.over(s.verifyPhone, phoneKey, s.cfg.verifyBudget, s.cfg.verifyWindow, false)
	ipBlocked, ipWait := s.over(s.verifyIP, ip, s.cfg.verifyIPBudget, s.cfg.verifyWindow, false)
	return mergeLimit(phoneBlocked, phoneWait, ipBlocked, ipWait)
}

func (s *Service) over(b *bucket, key string, budget int, window time.Duration, cooldown bool) (bool, int) {
	now := s.clock.Now()
	b.stamps[key] = prune(b.stamps[key], now, window)
	wait := 0
	blocked := false
	if len(b.stamps[key]) >= budget {
		blocked = true
		wait = ceilSeconds(b.stamps[key][0].Add(window).Sub(now))
	}
	if cooldown {
		if until := b.ready[key]; until.After(now) {
			blocked = true
			w := ceilSeconds(until.Sub(now))
			if w > wait {
				wait = w
			}
		}
	}
	return blocked, wait
}

func (s *Service) mark(b *bucket, key string, window time.Duration) {
	now := s.clock.Now()
	b.stamps[key] = append(prune(b.stamps[key], now, window), now)
}

func (s *Service) marshal(body any) Result {
	payload, err := json.Marshal(body)
	if err != nil {
		return Result{Err: codeErr("INTERNAL")}
	}
	return Result{Status: 200, Body: payload}
}

func (s *Service) order(account *user, step string) *Error {
	target := stepIndex(step)
	next := firstIncomplete(account)
	if next >= 0 && target > next {
		return codeErr("VALIDATION_STEP_OUT_OF_ORDER")
	}
	return nil
}

func (s *Service) onboardingView(account *user, isNew bool) onboardingView {
	next := firstIncomplete(account)
	view := onboardingView{
		CurrentIndex: 7,
		Completed:    next < 0,
		Steps:        make([]stepView, 0, len(stepOrder)),
		Account:      accountViewOf(account, isNew),
		NextStep:     nextStepOf(account),
	}
	if next >= 0 {
		view.CurrentIndex = int64(next + 1)
		id := stepOrder[next]
		view.CurrentStepID = &id
	}
	if answer, ok := account.Answers["nickname"]; ok {
		view.Nickname = &answer
	}
	for i, id := range stepOrder {
		step := stepView{StepID: id, Index: int64(i + 1), Status: "pending"}
		if answer, ok := account.Answers[id]; ok {
			step.Status = "completed"
			step.Answer = &answer
		}
		view.Steps = append(view.Steps, step)
	}
	return view
}

func mergeLimit(leftOn bool, left int, rightOn bool, right int) (bool, int) {
	if !leftOn && !rightOn {
		return false, 0
	}
	wait := left
	if right > wait {
		wait = right
	}
	if wait < 1 {
		wait = 1
	}
	return true, wait
}

func validateContact(region, phone, purpose string, locale *string) *Error {
	if strings.TrimSpace(region) == "" || strings.TrimSpace(phone) == "" || strings.TrimSpace(purpose) == "" {
		return codeErr("VALIDATION_MISSING_FIELD")
	}
	if _, ok := regions[region]; !ok {
		return codeErr("VALIDATION_UNSUPPORTED_REGION")
	}
	if !phonePattern.MatchString(phone) || len(phone) > DefaultPhoneMaxDigits {
		return codeErr("VALIDATION_INVALID_PHONE")
	}
	if _, ok := purposes[purpose]; !ok {
		return codeErr("VALIDATION_INVALID_PURPOSE")
	}
	if locale != nil {
		if _, ok := locales[*locale]; !ok {
			return codeErr("VALIDATION_UNSUPPORTED_LOCALE")
		}
	}
	return nil
}

func validateNickname(value string, maxRunes int) *Error {
	if value == "" || utf8.RuneCountInString(value) > maxRunes {
		return codeErr("VALIDATION_NICKNAME")
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return codeErr("VALIDATION_NICKNAME")
		}
	}
	return nil
}

func validateAnswer(value string, maxBytes int) *Error {
	if value == "" || len(value) > maxBytes {
		return codeErr("VALIDATION_INVALID_STEP_ANSWER")
	}
	return nil
}

func decodeObject(raw []byte, dest any) *Error {
	if len(raw) > maxBodyBytes {
		return codeErr("VALIDATION_MISSING_FIELD")
	}
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dest); err != nil {
		return codeErr("VALIDATION_MISSING_FIELD")
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return codeErr("VALIDATION_MISSING_FIELD")
	}
	return nil
}

func canonicalJSON(raw []byte) ([]byte, error) {
	if len(bytesTrim(raw)) == 0 || len(raw) > maxBodyBytes {
		return nil, errors.New("body")
	}
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	var value any
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, err
	}
	if _, ok := value.(map[string]any); !ok {
		return nil, errors.New("object")
	}
	return json.Marshal(value)
}

func bytesTrim(raw []byte) string {
	return strings.TrimSpace(string(raw))
}

func validIdempotencyKey(key string) bool {
	if key == "" || len(key) > DefaultIdempotencyKey {
		return false
	}
	for _, r := range key {
		if r < 0x21 || r > 0x7e {
			return false
		}
	}
	return true
}

func phoneKey(region, national string) string { return region + "|" + national }

func otpKey(phoneKey, purpose string) string { return phoneKey + "|" + purpose }

func stepIndex(id string) int {
	for i, step := range stepOrder {
		if step == id {
			return i
		}
	}
	return -1
}

func firstIncomplete(account *user) int {
	for i, step := range stepOrder {
		if _, ok := account.Answers[step]; !ok {
			return i
		}
	}
	return -1
}

// nextStepOf applies H3: onboarding first, then the first knowledge pack, then home.
// H3 is an implementation assumption, not a confirmed product rule.
func nextStepOf(account *user) string {
	if firstIncomplete(account) >= 0 {
		return "NEED_ONBOARDING"
	}
	if !account.HasFirstPack {
		return "NEED_FIRST_PACK"
	}
	return "READY_HOME"
}

func accountViewOf(account *user, isNew bool) accountView {
	return accountView{
		UserID:     account.ID,
		IsNew:      isNew,
		Onboarding: firstIncomplete(account) < 0,
		FirstPack:  account.HasFirstPack,
	}
}

func clientIP(remote string) string {
	host, _, err := net.SplitHostPort(remote)
	if err != nil || host == "" {
		if remote == "" {
			return "unknown"
		}
		return remote
	}
	return host
}

func destinationHint(region, national string) string {
	if len(national) <= 2 {
		return region + " " + strings.Repeat("*", len(national))
	}
	return region + " " + strings.Repeat("*", len(national)-2) + national[len(national)-2:]
}

func randomCode() (string, error) {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	n := (int(buf[0])<<24 | int(buf[1])<<16 | int(buf[2])<<8 | int(buf[3])) & 0x7fffffff
	n %= 1000000
	digits := []byte("000000")
	for i := 5; i >= 0; i-- {
		digits[i] = byte('0' + n%10)
		n /= 10
	}
	return string(digits), nil
}

func newToken(prefix string) (string, [32]byte, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", [32]byte{}, err
	}
	token := prefix + base64.RawURLEncoding.EncodeToString(buf)
	return token, sha256.Sum256([]byte(token)), nil
}

func newID(prefix string) string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return prefix + "unknown"
	}
	return prefix + hex.EncodeToString(buf)
}

func prune(stamps []time.Time, now time.Time, window time.Duration) []time.Time {
	cutoff := now.Add(-window)
	i := 0
	for i < len(stamps) && !stamps[i].After(cutoff) {
		i++
	}
	return stamps[i:]
}

func ceilSeconds(d time.Duration) int {
	if d <= 0 {
		return 0
	}
	secs := int((d + time.Second - 1) / time.Second)
	if secs < 1 {
		return 1
	}
	return secs
}

func orDuration(value, fallback time.Duration) time.Duration {
	if value <= 0 {
		return fallback
	}
	return value
}

func orInt(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
