package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	qlutchapi "github.com/zh7551069/qlutch-projects/packages/contracts/gen/go"
	"github.com/zh7551069/qlutch-projects/services/api/internal/identity"
)

const testCode = "246810"

func TestIdentityRegisterLoginOnboardingAndProfile(t *testing.T) {
	h := newIdentity(t)
	phone := "13800010001"
	sent := h.send(t, "reg-1", "+86", phone, "register", nil)
	if sent.StatusCode != http.StatusOK || sent.Body.Retry_after_seconds != int64(identity.DefaultSendCooldown/time.Second) {
		t.Fatalf("send = %+v raw %s", sent, sent.Raw)
	}
	hint := ""
	if sent.Body.Destination_hint != nil {
		hint = *sent.Body.Destination_hint
	}
	if strings.Contains(hint, phone) || strings.Contains(string(sent.Raw), testCode) {
		t.Fatalf("send leaked phone or code: %s", sent.Raw)
	}
	if len(h.sender.Codes()) != 1 || h.sender.Codes()[0] != testCode {
		t.Fatalf("codes = %#v", h.sender.Codes())
	}

	created := h.verify(t, "+86", phone, "register", testCode, localePtr(qlutchapi.LocaleZhHant))
	if created.StatusCode != http.StatusOK || !created.Body.Account.Is_new_user || created.Body.Next_step != qlutchapi.NextStepNEEDONBOARDING {
		t.Fatalf("verify = %+v raw %s", created.StatusCode, created.Raw)
	}
	if created.Body.Profile.Locale != qlutchapi.LocaleZhHant || created.Body.Profile.Region_code != qlutchapi.RegionCodeValue86 {
		t.Fatalf("profile = %+v", created.Body.Profile)
	}
	if created.Body.Expires_in != int64(identity.DefaultAccessTTL/time.Second) || created.Body.Access_token == "" || created.Body.Refresh_token == "" {
		t.Fatalf("session = %+v", created.Body)
	}
	access := created.Body.Access_token

	again := h.send(t, "reg-2", "+86", phone, "register", nil)
	if code := mustCode(t, again.StatusCode, again.Raw, again.RequestID); code != "AUTH_ACCOUNT_EXISTS" || again.StatusCode != http.StatusConflict {
		t.Fatalf("exists = %d %s", again.StatusCode, code)
	}
	if len(h.sender.Codes()) != 1 {
		t.Fatal("account exists must not send sms")
	}

	missing := h.send(t, "login-missing", "+86", "13800019999", "login", nil)
	if code := mustCode(t, missing.StatusCode, missing.Raw, missing.RequestID); code != "AUTH_ACCOUNT_NOT_FOUND" || missing.StatusCode != http.StatusNotFound {
		t.Fatalf("missing = %d %s", missing.StatusCode, code)
	}

	state := h.onboarding(t, access)
	if state.Body.Current_step_id == nil || *state.Body.Current_step_id != "nickname" || state.Body.Current_index != 1 || state.Body.Completed {
		t.Fatalf("start = %+v", state.Body)
	}
	if code := h.putStep(t, access, "exam_track", "", "language"); code != "VALIDATION_STEP_OUT_OF_ORDER" {
		t.Fatalf("skip = %s", code)
	}
	if code := h.putStep(t, access, "nickname", "  ", ""); code != "VALIDATION_NICKNAME" {
		t.Fatalf("blank nickname = %s", code)
	}
	if code := h.putStep(t, access, "nickname", strings.Repeat("明", identity.DefaultNicknameMaxRunes+1), ""); code != "VALIDATION_NICKNAME" {
		t.Fatalf("long nickname = %s", code)
	}
	h.putOK(t, access, "nickname", "小明", "")
	if code := h.putStep(t, access, "nope", "", "x"); code != "VALIDATION_INVALID_STEP" {
		t.Fatalf("unknown step = %s", code)
	}
	empty := ""
	emptyAnswer, err := h.client.PutIdentityOnboarding(context.Background(), qlutchapi.OnboardingWriteRequest{
		Step_id: qlutchapi.StepIdAge,
		Answer:  &empty,
	}, qlutchapi.CallMeta{Bearer: access})
	if err != nil {
		t.Fatal(err)
	}
	if code := mustCode(t, emptyAnswer.StatusCode, emptyAnswer.Raw, emptyAnswer.RequestID); code != "VALIDATION_INVALID_STEP_ANSWER" {
		t.Fatalf("empty answer = %s", code)
	}
	if code := h.putStep(t, access, "identity", "", "candidate"); code != "VALIDATION_STEP_OUT_OF_ORDER" {
		t.Fatalf("skip after nickname = %s", code)
	}
	h.putOK(t, access, "age", "", "18-24")
	resumed := h.onboarding(t, access)
	if resumed.Body.Current_step_id == nil || *resumed.Body.Current_step_id != "study_or_work" || resumed.Body.Current_index != 3 {
		t.Fatalf("resume = %+v", resumed.Body)
	}
	rewritten := h.putOK(t, access, "nickname", "小红", "")
	if rewritten.Body.Nickname == nil || *rewritten.Body.Nickname != "小红" || *rewritten.Body.Current_step_id != "study_or_work" {
		t.Fatalf("rewrite = %+v", rewritten.Body)
	}
	for _, step := range []struct{ id, answer string }{
		{"study_or_work", "student"},
		{"identity", "candidate"},
		{"how_to_address", "friend"},
		{"exam_track", "language"},
		{"daily_time", "half-hour"},
	} {
		h.putOK(t, access, step.id, "", step.answer)
	}
	done := h.onboarding(t, access)
	if !done.Body.Completed || done.Body.Current_step_id != nil || done.Body.Next_step != qlutchapi.NextStepNEEDFIRSTPACK || !done.Body.Account.Onboarding_completed {
		t.Fatalf("done = %+v", done.Body)
	}
	if done.Body.Account.Is_new_user {
		t.Fatal("onboarding account must not reuse is_new_user")
	}

	profile := h.profile(t, access)
	if profile.Body.Locale != qlutchapi.LocaleZhHant || profile.Body.Region_code != "+86" {
		t.Fatalf("profile = %+v", profile.Body)
	}
	updated := h.putProfile(t, access, nil, regionPtr(qlutchapi.RegionCodeValue852))
	if updated.StatusCode != http.StatusOK || updated.Body.Region_code != "+852" || updated.Body.Locale != qlutchapi.LocaleZhHant {
		t.Fatalf("region pref = %+v raw %s", updated.Body, updated.Raw)
	}
	wrongRegion := h.send(t, "login-other-region", "+852", phone, "login", nil)
	if code := mustCode(t, wrongRegion.StatusCode, wrongRegion.Raw, wrongRegion.RequestID); code != "AUTH_ACCOUNT_NOT_FOUND" || wrongRegion.StatusCode != http.StatusNotFound {
		t.Fatalf("bound phone = %d %s", wrongRegion.StatusCode, code)
	}
	if len(h.sender.Codes()) != 1 {
		t.Fatal("wrong region must not send sms")
	}

	h.clock.Advance(identity.DefaultSendCooldown)
	loginSend := h.send(t, "login-1", "+86", phone, "login", nil)
	if loginSend.StatusCode != http.StatusOK {
		t.Fatalf("login send = %d %s", loginSend.StatusCode, loginSend.Raw)
	}
	logged := h.verify(t, "+86", phone, "login", testCode, nil)
	if logged.Body.Account.Is_new_user || logged.Body.Next_step != qlutchapi.NextStepNEEDFIRSTPACK || logged.Body.Profile.Region_code != "+86" {
		t.Fatalf("login = %+v", logged.Body)
	}
	if !h.svc.TestingSetFirstPack(context.Background(), logged.Body.Account.User_id, true) {
		t.Fatal("missing user")
	}
	home := h.refresh(t, logged.Body.Refresh_token, "")
	if home.Body.Next_step != qlutchapi.NextStepREADYHOME || home.Body.Account.Has_first_pack != true || home.Body.Account.Is_new_user {
		t.Fatalf("home = %+v", home.Body)
	}
}

func TestIdentityIdempotentSend(t *testing.T) {
	h := newIdentity(t)
	phone := "13800010002"
	body := map[string]string{"region_code": "+86", "phone_national": phone, "purpose": "register"}
	first := h.post(t, "/v1/identity/otp/send", body, map[string]string{"Idempotency-Key": "same"})
	if first.Code != http.StatusOK {
		t.Fatalf("first = %d %s", first.Code, first.Body.String())
	}
	replay := h.post(t, "/v1/identity/otp/send", body, map[string]string{"Idempotency-Key": "same"})
	if replay.Code != http.StatusOK || replay.Body.String() != first.Body.String() || len(h.sender.Codes()) != 1 {
		t.Fatalf("replay = %d %s codes %d", replay.Code, replay.Body.String(), len(h.sender.Codes()))
	}
	other := h.post(t, "/v1/identity/otp/send", map[string]string{"region_code": "+86", "phone_national": "13800010003", "purpose": "register"}, map[string]string{"Idempotency-Key": "same"})
	if other.Code != http.StatusConflict || mustCode(t, other.Code, other.Body.Bytes(), other.Header().Get("X-Request-Id")) != "IDEMPOTENCY_KEY_MISMATCH" {
		t.Fatalf("mismatch = %d %s", other.Code, other.Body.String())
	}
	if len(h.sender.Codes()) != 1 {
		t.Fatal("mismatch executed")
	}
	missing := h.post(t, "/v1/identity/otp/send", body, nil)
	if missing.Code != http.StatusBadRequest || mustCode(t, missing.Code, missing.Body.Bytes(), missing.Header().Get("X-Request-Id")) != "VALIDATION_MISSING_FIELD" {
		t.Fatalf("missing key = %d %s", missing.Code, missing.Body.String())
	}
}

func TestIdentityRevokeRefreshAndRealm(t *testing.T) {
	h := newIdentity(t)
	phone := "13800010004"
	if h.send(t, "r1", "+86", phone, "register", nil).StatusCode != http.StatusOK {
		t.Fatal("send")
	}
	session := h.verify(t, "+86", phone, "register", testCode, nil)
	access, refresh := session.Body.Access_token, session.Body.Refresh_token

	h.clock.Advance(identity.DefaultAccessTTL + time.Second)
	stale := h.refresh(t, refresh, access)
	if stale.StatusCode != http.StatusOK || stale.Body.Access_token == access || stale.Body.Refresh_token == refresh {
		t.Fatalf("refresh with expired access = %d %s", stale.StatusCode, stale.Raw)
	}
	if code := h.onboardingCode(t, access); code != "AUTH_EXPIRED" {
		t.Fatalf("old access = %s", code)
	}
	if code := h.refreshCode(t, refresh); code != "AUTH_EXPIRED" {
		t.Fatalf("old refresh = %s", code)
	}
	if h.onboarding(t, stale.Body.Access_token).StatusCode != http.StatusOK {
		t.Fatal("new access")
	}

	revoked := h.revoke(t, stale.Body.Access_token, stale.Body.Refresh_token)
	if revoked.StatusCode != http.StatusNoContent || len(revoked.Raw) != 0 || revoked.RequestID == "" {
		t.Fatalf("revoke = %d %q id %q", revoked.StatusCode, revoked.Raw, revoked.RequestID)
	}
	if code := h.onboardingCode(t, stale.Body.Access_token); code != "AUTH_EXPIRED" {
		t.Fatalf("revoked access = %s", code)
	}
	if code := h.refreshCode(t, stale.Body.Refresh_token); code != "AUTH_EXPIRED" {
		t.Fatalf("revoked refresh = %s", code)
	}
	again := h.revoke(t, "not-a-token", "")
	if again.StatusCode != http.StatusNoContent {
		t.Fatalf("repeat revoke = %d %s", again.StatusCode, again.Raw)
	}
	naked := h.revoke(t, "", "")
	if naked.StatusCode != http.StatusUnauthorized || mustCode(t, naked.StatusCode, naked.Raw, naked.RequestID) != "AUTH_REQUIRED" {
		t.Fatalf("naked = %d %s", naked.StatusCode, naked.Raw)
	}
	realm := h.onboardingCode(t, "qlu_i_internal")
	if realm != "AUTH_WRONG_REALM" {
		t.Fatalf("realm = %s", realm)
	}
	realmSend := h.clientSend(t, "realm", "+86", "13800010005", "register", "qlu_i_internal", nil)
	if realmSend.StatusCode != http.StatusForbidden || mustCode(t, realmSend.StatusCode, realmSend.Raw, realmSend.RequestID) != "AUTH_WRONG_REALM" {
		t.Fatalf("realm send = %d %s", realmSend.StatusCode, realmSend.Raw)
	}
	if len(h.sender.Codes()) != 1 {
		t.Fatal("wrong realm sent sms")
	}
}

func TestIdentityRateLimit(t *testing.T) {
	h := newIdentity(t)
	phone := "13800010006"
	if h.send(t, "c1", "+86", phone, "register", nil).StatusCode != http.StatusOK {
		t.Fatal("first")
	}
	limited := h.send(t, "c2", "+86", phone, "register", nil)
	if limited.StatusCode != http.StatusTooManyRequests || mustCode(t, limited.StatusCode, limited.Raw, limited.RequestID) != "RATE_LIMITED" {
		t.Fatalf("cooldown = %d %s", limited.StatusCode, limited.Raw)
	}
	wait := retryAfter(t, limited.Raw)
	if wait < 1 || wait > int(identity.DefaultSendCooldown/time.Second) {
		t.Fatalf("retry = %d", wait)
	}
	if len(h.sender.Codes()) != 1 {
		t.Fatal("limited send dispatched sms")
	}
	for i := 1; i < identity.DefaultSendBudget; i++ {
		h.clock.Advance(identity.DefaultSendCooldown)
		key := "budget-" + string(rune('a'+i))
		if res := h.send(t, key, "+86", phone, "register", nil); res.StatusCode != http.StatusOK {
			t.Fatalf("send %d = %d %s", i, res.StatusCode, res.Raw)
		}
	}
	h.clock.Advance(identity.DefaultSendCooldown)
	blocked := h.send(t, "budget-over", "+86", phone, "register", nil)
	if blocked.StatusCode != http.StatusTooManyRequests || retryAfter(t, blocked.Raw) < 1 {
		t.Fatalf("budget = %d %s", blocked.StatusCode, blocked.Raw)
	}

	verifyPhone := "13800010007"
	h.clock.Advance(identity.DefaultSendWindow)
	if h.send(t, "vsend", "+86", verifyPhone, "register", nil).StatusCode != http.StatusOK {
		t.Fatal("verify setup")
	}
	var last int
	for i := 0; i < identity.DefaultOTPMaxAttempts-1; i++ {
		res := h.verify(t, "+86", verifyPhone, "register", "000000", nil)
		last = res.StatusCode
		if code := mustCode(t, res.StatusCode, res.Raw, res.RequestID); code != "AUTH_OTP_INVALID" {
			t.Fatalf("attempt %d = %s", i, code)
		}
	}
	maxed := h.verify(t, "+86", verifyPhone, "register", "000000", nil)
	if maxed.StatusCode != http.StatusForbidden || mustCode(t, maxed.StatusCode, maxed.Raw, maxed.RequestID) != "AUTH_OTP_MAX_ATTEMPTS" {
		t.Fatalf("max = %d %s", maxed.StatusCode, maxed.Raw)
	}
	h.clock.Advance(identity.DefaultSendCooldown)
	if h.send(t, "vsend-2", "+86", verifyPhone, "register", nil).StatusCode != http.StatusOK {
		t.Fatal("resend")
	}
	for i := 0; i < identity.DefaultOTPMaxAttempts; i++ {
		_ = h.verify(t, "+86", verifyPhone, "register", "000000", nil)
	}
	over := h.verify(t, "+86", verifyPhone, "register", testCode, nil)
	if over.StatusCode != http.StatusTooManyRequests || mustCode(t, over.StatusCode, over.Raw, over.RequestID) != "RATE_LIMITED" {
		t.Fatalf("verify limit = %d %s (last invalid %d)", over.StatusCode, over.Raw, last)
	}
}

func TestIdentityOTPExpiryAndValidationCodes(t *testing.T) {
	h := newIdentity(t)
	phone := "13800010008"
	if h.send(t, "e1", "+86", phone, "register", nil).StatusCode != http.StatusOK {
		t.Fatal("send")
	}
	badFormat := h.verify(t, "+86", phone, "register", "12345", nil)
	if badFormat.StatusCode != http.StatusBadRequest || mustCode(t, badFormat.StatusCode, badFormat.Raw, badFormat.RequestID) != "VALIDATION_INVALID_OTP_FORMAT" {
		t.Fatalf("format = %d %s", badFormat.StatusCode, badFormat.Raw)
	}
	wrong := h.verify(t, "+86", phone, "register", "000000", nil)
	if wrong.StatusCode != http.StatusUnauthorized || mustCode(t, wrong.StatusCode, wrong.Raw, wrong.RequestID) != "AUTH_OTP_INVALID" {
		t.Fatalf("wrong = %d %s", wrong.StatusCode, wrong.Raw)
	}
	if strings.Contains(string(wrong.Raw), "000000") || strings.Contains(string(wrong.Raw), testCode) {
		t.Fatal("verify error echoed a code")
	}
	h.clock.Advance(identity.DefaultOTPTTL + time.Second)
	expired := h.verify(t, "+86", phone, "register", testCode, nil)
	if expired.StatusCode != http.StatusUnauthorized || mustCode(t, expired.StatusCode, expired.Raw, expired.RequestID) != "AUTH_OTP_EXPIRED" {
		t.Fatalf("expired = %d %s", expired.StatusCode, expired.Raw)
	}

	cases := []struct {
		body string
		code string
	}{
		{`{"phone_national":"13800010008","purpose":"register"}`, "VALIDATION_MISSING_FIELD"},
		{`{"region_code":"+1","phone_national":"13800010008","purpose":"register"}`, "VALIDATION_UNSUPPORTED_REGION"},
		{`{"region_code":"+86","phone_national":"138-000","purpose":"register"}`, "VALIDATION_INVALID_PHONE"},
		{`{"region_code":"+86","phone_national":"13800010008","purpose":"signup"}`, "VALIDATION_INVALID_PURPOSE"},
		{`{"region_code":"+86","phone_national":"13800010008","purpose":"register","locale":"en"}`, "VALIDATION_UNSUPPORTED_LOCALE"},
	}
	for _, tc := range cases {
		rec := h.postRaw(t, "/v1/identity/otp/send", tc.body, map[string]string{"Idempotency-Key": "val-" + tc.code})
		if rec.Code != http.StatusBadRequest || mustCode(t, rec.Code, rec.Body.Bytes(), rec.Header().Get("X-Request-Id")) != tc.code {
			t.Fatalf("%s: %d %s", tc.code, rec.Code, rec.Body.String())
		}
	}
	if h.onboardingCode(t, "") != "AUTH_REQUIRED" {
		t.Fatal("expected auth required")
	}
	hong := h.send(t, "hk", "+852", "51230001", "register", localePtr(qlutchapi.LocaleZhHans))
	if hong.StatusCode != http.StatusOK {
		t.Fatalf("hong kong = %d %s", hong.StatusCode, hong.Raw)
	}
}

func TestIdentityCodeIsNotLogged(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	h := newIdentity(t)
	res := h.send(t, "log-1", "+86", "13800010009", "register", nil)
	if res.StatusCode != http.StatusOK {
		t.Fatal(res.Raw)
	}
	bad := h.verify(t, "+86", "13800010009", "register", "000000", nil)
	if bad.StatusCode != http.StatusUnauthorized {
		t.Fatal(bad.Raw)
	}
	logged := buf.String()
	if strings.Contains(logged, testCode) || strings.Contains(logged, "000000") || strings.Contains(logged, "13800010009") {
		t.Fatalf("log leaked secret material: %s", logged)
	}
}

func TestIdentityNoSocialLogin(t *testing.T) {
	h := newIdentity(t)
	for _, path := range []string{"/v1/identity/oauth/wechat", "/v1/identity/oauth/alipay"} {
		rec := h.postRaw(t, path, `{}`, nil)
		if rec.Code != http.StatusNotFound || strings.Contains(rec.Body.String(), "access_token") {
			t.Fatalf("%s = %d %s", path, rec.Code, rec.Body.String())
		}
		if rec.Header().Get("X-Request-Id") == "" {
			t.Fatal("missing request id")
		}
	}
}

func TestIdentitySenderFailureDoesNotStick(t *testing.T) {
	sender := &flapSender{}
	h := newIdentityWith(t, identity.Options{Sender: sender, Code: fixedCode})
	sender.fail = true
	first := h.send(t, "retry", "+86", "13800010010", "register", nil)
	if first.StatusCode != http.StatusInternalServerError || mustCode(t, first.StatusCode, first.Raw, first.RequestID) != "INTERNAL" {
		t.Fatalf("fail = %d %s", first.StatusCode, first.Raw)
	}
	sender.fail = false
	second := h.send(t, "retry", "+86", "13800010010", "register", nil)
	if second.StatusCode != http.StatusOK || len(sender.codes) != 1 {
		t.Fatalf("retry = %d codes %d", second.StatusCode, len(sender.codes))
	}
}

type flapSender struct {
	fail  bool
	codes []string
}

func (f *flapSender) Send(_ context.Context, message identity.SMS) error {
	if f.fail {
		return io.EOF
	}
	f.codes = append(f.codes, message.Code)
	return nil
}

type identityHarness struct {
	svc    *identity.Service
	sender *identity.CaptureSender
	clock  *identity.ManualClock
	client *qlutchapi.Client
}

func newIdentity(t *testing.T) *identityHarness {
	t.Helper()
	sender := &identity.CaptureSender{}
	return newIdentityWith(t, identity.Options{Sender: sender, Code: fixedCode})
}

func newIdentityWith(t *testing.T, opts identity.Options) *identityHarness {
	t.Helper()
	clock := identity.NewManualClock(time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC))
	opts.Clock = clock
	if opts.Code == nil {
		opts.Code = fixedCode
	}
	svc, err := identity.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	httpSrv := httptest.NewServer(New(nil, svc).Handler(nil))
	t.Cleanup(httpSrv.Close)
	sender, _ := opts.Sender.(*identity.CaptureSender)
	return &identityHarness{
		svc:    svc,
		sender: sender,
		clock:  clock,
		client: &qlutchapi.Client{BaseURL: httpSrv.URL, HTTP: httpSrv.Client()},
	}
}

func fixedCode() (string, error) { return testCode, nil }

func (h *identityHarness) send(t *testing.T, key, region, phone, purpose string, locale *qlutchapi.Locale) qlutchapi.APIResponse[qlutchapi.OtpSendResponse] {
	t.Helper()
	return h.clientSend(t, key, region, phone, purpose, "", locale)
}

func (h *identityHarness) clientSend(t *testing.T, key, region, phone, purpose, bearer string, locale *qlutchapi.Locale) qlutchapi.APIResponse[qlutchapi.OtpSendResponse] {
	t.Helper()
	res, err := h.client.SendIdentityOtp(context.Background(), key, qlutchapi.OtpSendRequest{
		Region_code:    qlutchapi.RegionCode(region),
		Phone_national: phone,
		Purpose:        qlutchapi.OtpPurpose(purpose),
		Locale:         locale,
	}, qlutchapi.CallMeta{Bearer: bearer})
	if err != nil {
		t.Fatal(err)
	}
	if res.RequestID == "" {
		t.Fatal("missing request id")
	}
	return res
}

func (h *identityHarness) verify(t *testing.T, region, phone, purpose, code string, locale *qlutchapi.Locale) qlutchapi.APIResponse[qlutchapi.SessionResponse] {
	t.Helper()
	res, err := h.client.VerifyIdentityOtp(context.Background(), qlutchapi.OtpVerifyRequest{
		Region_code:    qlutchapi.RegionCode(region),
		Phone_national: phone,
		Purpose:        qlutchapi.OtpPurpose(purpose),
		Code:           code,
		Locale:         locale,
	}, qlutchapi.CallMeta{})
	if err != nil {
		t.Fatal(err)
	}
	if res.RequestID == "" {
		t.Fatal("missing request id")
	}
	return res
}

func (h *identityHarness) refresh(t *testing.T, token, bearer string) qlutchapi.APIResponse[qlutchapi.SessionResponse] {
	t.Helper()
	res, err := h.client.RefreshIdentitySession(context.Background(), qlutchapi.SessionRefreshRequest{Refresh_token: token}, qlutchapi.CallMeta{Bearer: bearer})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func (h *identityHarness) refreshCode(t *testing.T, token string) string {
	t.Helper()
	res := h.refresh(t, token, "")
	return mustCode(t, res.StatusCode, res.Raw, res.RequestID)
}

func (h *identityHarness) revoke(t *testing.T, access, refresh string) qlutchapi.APIResponse[qlutchapi.Empty] {
	t.Helper()
	var body *qlutchapi.SessionRevokeRequest
	if refresh != "" {
		body = &qlutchapi.SessionRevokeRequest{Refresh_token: &refresh}
	}
	res, err := h.client.RevokeIdentitySession(context.Background(), body, qlutchapi.CallMeta{Bearer: access})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func (h *identityHarness) onboarding(t *testing.T, access string) qlutchapi.APIResponse[qlutchapi.OnboardingState] {
	t.Helper()
	res, err := h.client.GetIdentityOnboarding(context.Background(), qlutchapi.CallMeta{Bearer: access})
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("onboarding %d %s", res.StatusCode, res.Raw)
	}
	return res
}

func (h *identityHarness) onboardingCode(t *testing.T, access string) string {
	t.Helper()
	res, err := h.client.GetIdentityOnboarding(context.Background(), qlutchapi.CallMeta{Bearer: access})
	if err != nil {
		t.Fatal(err)
	}
	return mustCode(t, res.StatusCode, res.Raw, res.RequestID)
}

func (h *identityHarness) putOK(t *testing.T, access, step, nickname, answer string) qlutchapi.APIResponse[qlutchapi.OnboardingState] {
	t.Helper()
	res := h.put(t, access, step, nickname, answer)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("put %s = %d %s", step, res.StatusCode, res.Raw)
	}
	return res
}

func (h *identityHarness) putStep(t *testing.T, access, step, nickname, answer string) string {
	t.Helper()
	res := h.put(t, access, step, nickname, answer)
	return mustCode(t, res.StatusCode, res.Raw, res.RequestID)
}

func (h *identityHarness) put(t *testing.T, access, step, nickname, answer string) qlutchapi.APIResponse[qlutchapi.OnboardingState] {
	t.Helper()
	body := qlutchapi.OnboardingWriteRequest{Step_id: qlutchapi.StepId(step)}
	if nickname != "" {
		body.Nickname = &nickname
	}
	if answer != "" {
		body.Answer = &answer
	}
	res, err := h.client.PutIdentityOnboarding(context.Background(), body, qlutchapi.CallMeta{Bearer: access})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func (h *identityHarness) profile(t *testing.T, access string) qlutchapi.APIResponse[qlutchapi.IdentityProfile] {
	t.Helper()
	res, err := h.client.GetIdentityProfile(context.Background(), qlutchapi.CallMeta{Bearer: access})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func (h *identityHarness) putProfile(t *testing.T, access string, locale *qlutchapi.Locale, region *qlutchapi.RegionCode) qlutchapi.APIResponse[qlutchapi.IdentityProfile] {
	t.Helper()
	res, err := h.client.PutIdentityProfile(context.Background(), qlutchapi.IdentityProfileUpdate{Locale: locale, Region_code: region}, qlutchapi.CallMeta{Bearer: access})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func (h *identityHarness) post(t *testing.T, path string, body any, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return h.postRaw(t, path, string(raw), headers)
}

func (h *identityHarness) postRaw(t *testing.T, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, h.client.BaseURL+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rec := httptest.NewRecorder()
	resp, err := h.client.HTTP.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	rec.Code = resp.StatusCode
	rec.Body = bytes.NewBuffer(payload)
	rec.Header().Set("X-Request-Id", resp.Header.Get("X-Request-Id"))
	return rec
}

func localePtr(value qlutchapi.Locale) *qlutchapi.Locale { return &value }

func regionPtr(value qlutchapi.RegionCode) *qlutchapi.RegionCode { return &value }

func mustCode(t *testing.T, status int, raw []byte, requestID string) string {
	t.Helper()
	var env qlutchapi.ErrorBody
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("status %d body %s: %v", status, raw, err)
	}
	if env.Error.Code == "" || env.Error.Message == "" || env.Request_id == "" || env.Request_id != requestID {
		t.Fatalf("envelope = %+v header %s", env, requestID)
	}
	if strings.Contains(env.Error.Message, testCode) {
		t.Fatalf("message leaked code: %s", env.Error.Message)
	}
	return env.Error.Code
}

func retryAfter(t *testing.T, raw []byte) int {
	t.Helper()
	var env qlutchapi.ErrorBody
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err)
	}
	var details struct {
		Retry int `json:"retry_after_seconds"`
	}
	if err := json.Unmarshal(env.Error.Details, &details); err != nil {
		t.Fatal(err)
	}
	return details.Retry
}
