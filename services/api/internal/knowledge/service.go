package knowledge

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/zh7551069/qlutch-projects/services/api/internal/identity"
)

// Options configures the knowledge service.
// Nil Store, Blobs, and Queue use in-process stand-ins. Production passes PostgreSQL,
// an S3-compatible Blobs, and the worker queue.
type Options struct {
	Identity       *identity.Service
	Store          Store
	Blobs          Blobs
	Queue          Enqueuer
	Now            func() time.Time
	MaxBytes       int64
	IdempotencyTTL time.Duration
}

// Service is the learner knowledge API.
type Service struct {
	mu       sync.Mutex
	identity *identity.Service
	store    Store
	blobs    Blobs
	queue    Enqueuer
	now      func() time.Time
	maxBytes int64
	idemTTL  time.Duration
}

// New builds a knowledge service. Identity is required.
func New(opts Options) (*Service, error) {
	if opts.Identity == nil {
		return nil, errors.New("knowledge identity missing")
	}
	store := opts.Store
	if store == nil {
		store = newMemoryStore()
	}
	blobs := opts.Blobs
	if blobs == nil {
		blobs = NewMemoryBlobs()
	}
	queue := opts.Queue
	if queue == nil {
		queue = NewMemoryQueue()
	}
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	maxBytes := opts.MaxBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}
	idemTTL := opts.IdempotencyTTL
	if idemTTL <= 0 {
		idemTTL = DefaultIdempotencyTTL
	}
	return &Service{
		identity: opts.Identity,
		store:    store,
		blobs:    blobs,
		queue:    queue,
		now:      now,
		maxBytes: maxBytes,
		idemTTL:  idemTTL,
	}, nil
}

// CreatePack creates an empty pack. It does not set has_first_pack.
func (s *Service) CreatePack(ctx context.Context, accessToken string, raw []byte) Result {
	account, authErr := s.identity.AccountFromAccess(ctx, accessToken)
	if authErr != nil {
		return Result{Err: fromIdentity(authErr)}
	}
	raw = optionalObject(raw)
	if _, err := canonicalJSON(raw); err != nil {
		return Result{Err: codeErr("VALIDATION_MISSING_FIELD")}
	}
	if decErr := decodeObject(raw, &packBody{}); decErr != nil {
		return Result{Err: decErr}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	item := pack{ID: newID("pack_"), UserID: account.UserID}
	if err := s.store.InsertPack(ctx, item); err != nil {
		return Result{Err: codeErr("INTERNAL")}
	}
	fresh, authErr := s.identity.AccountFromAccess(ctx, accessToken)
	if authErr != nil {
		return Result{Err: fromIdentity(authErr)}
	}
	return s.marshal(http.StatusCreated, packView{PackID: item.ID, Account: accountOf(fresh), NextStep: fresh.NextStep})
}

// UploadURL issues a presigned PUT. It does not set has_first_pack and does not return the object key as its own field.
func (s *Service) UploadURL(ctx context.Context, accessToken string, raw []byte) Result {
	account, authErr := s.identity.AccountFromAccess(ctx, accessToken)
	if authErr != nil {
		return Result{Err: fromIdentity(authErr)}
	}
	var body uploadBody
	if decErr := decodeObject(raw, &body); decErr != nil {
		return Result{Err: decErr}
	}
	if body.PackID == "" {
		return Result{Err: codeErr("VALIDATION_MISSING_FIELD")}
	}
	if mediaErr := mediaError(body.Source, body.ContentType); mediaErr != nil {
		return Result{Err: mediaErr}
	}
	if body.ByteSize < 1 {
		return Result{Err: codeErr("VALIDATION_MISSING_FIELD")}
	}
	if body.ByteSize > s.maxBytes {
		return Result{Err: tooLarge(s.maxBytes)}
	}
	if body.Filename != nil && !validFilename(*body.Filename) {
		return Result{Err: codeErr("VALIDATION_MISSING_FIELD")}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	owner, err := s.store.FindPack(ctx, body.PackID)
	if err != nil {
		return Result{Err: codeErr("INTERNAL")}
	}
	if owner == nil || owner.UserID != account.UserID {
		return Result{Err: codeErr("NOT_FOUND")}
	}
	item := material{
		ID:          newID("mat_"),
		PackID:      body.PackID,
		UserID:      account.UserID,
		Source:      body.Source,
		ContentType: body.ContentType,
		ByteSize:    body.ByteSize,
		ObjectKey:   "m/" + account.UserID + "/" + newID("obj_"),
	}
	if body.Filename != nil {
		item.Filename = *body.Filename
	}
	uploadURL, headers, expiresIn, err := s.blobs.PresignPut(ctx, item.ObjectKey, item.ContentType)
	if err != nil || uploadURL == "" || expiresIn < 1 {
		return Result{Err: codeErr("INTERNAL")}
	}
	if err := s.store.InsertMaterial(ctx, item); err != nil {
		return Result{Err: codeErr("INTERNAL")}
	}
	fresh, authErr := s.identity.AccountFromAccess(ctx, accessToken)
	if authErr != nil {
		return Result{Err: fromIdentity(authErr)}
	}
	return s.marshal(http.StatusOK, uploadView{
		MaterialID: item.ID,
		PackID:     item.PackID,
		Source:     item.Source,
		UploadURL:  uploadURL,
		Method:     http.MethodPut,
		Headers:    headers,
		ExpiresIn:  expiresIn,
		Account:    accountOf(fresh),
		NextStep:   fresh.NextStep,
	})
}

// Import accepts an uploaded material and enqueues it. Status stays accepted.
// It does not set has_first_pack and does not run map generation.
func (s *Service) Import(ctx context.Context, accessToken, idemKey string, raw []byte) Result {
	account, authErr := s.identity.AccountFromAccess(ctx, accessToken)
	if authErr != nil {
		return Result{Err: fromIdentity(authErr)}
	}
	if !validIdempotencyKey(idemKey) {
		return Result{Err: codeErr("VALIDATION_MISSING_FIELD")}
	}
	canon, err := canonicalJSON(raw)
	if err != nil {
		return Result{Err: codeErr("VALIDATION_MISSING_FIELD")}
	}
	hash := bodyHash(canon)
	s.mu.Lock()
	defer s.mu.Unlock()
	if result, done := s.replay(ctx, accessToken, account.UserID, opImport, idemKey, hash); done {
		return result
	}
	var body importBody
	if decErr := decodeObject(raw, &body); decErr != nil {
		return s.remember(ctx, account.UserID, opImport, idemKey, hash, Result{Err: decErr}, "")
	}
	if body.PackID == "" || body.MaterialID == "" {
		return s.remember(ctx, account.UserID, opImport, idemKey, hash, Result{Err: codeErr("VALIDATION_MISSING_FIELD")}, "")
	}
	mat, err := s.store.FindMaterial(ctx, body.MaterialID)
	if err != nil {
		return Result{Err: codeErr("INTERNAL")}
	}
	if mat == nil || mat.UserID != account.UserID || mat.PackID != body.PackID {
		return s.remember(ctx, account.UserID, opImport, idemKey, hash, Result{Err: codeErr("NOT_FOUND")}, "")
	}
	size, found, err := s.blobs.Stat(ctx, mat.ObjectKey)
	if err != nil {
		return Result{Err: codeErr("INTERNAL")}
	}
	if !found || size != mat.ByteSize {
		return s.remember(ctx, account.UserID, opImport, idemKey, hash, Result{Err: codeErr("VALIDATION_MATERIAL_NOT_UPLOADED")}, "")
	}
	item := job{
		ID:         newID("job_"),
		PackID:     mat.PackID,
		MaterialID: mat.ID,
		UserID:     account.UserID,
		Source:     mat.Source,
		Status:     statusAccepted,
	}
	if err := s.store.AcceptJob(ctx, item); err != nil {
		if errors.Is(err, errConflict) {
			return s.remember(ctx, account.UserID, opImport, idemKey, hash, Result{Err: codeErr("CONFLICT")}, "")
		}
		if errors.Is(err, errNotFound) {
			return s.remember(ctx, account.UserID, opImport, idemKey, hash, Result{Err: codeErr("NOT_FOUND")}, "")
		}
		return Result{Err: codeErr("INTERNAL")}
	}
	payload, err := json.Marshal(queuedJob{JobID: item.ID})
	if err != nil {
		_ = s.store.DeleteJob(ctx, item.ID)
		return Result{Err: codeErr("INTERNAL")}
	}
	if err := s.queue.Enqueue(ctx, payload); err != nil {
		_ = s.store.DeleteJob(ctx, item.ID)
		return Result{Err: codeErr("INTERNAL")}
	}
	result := s.jobResult(ctx, accessToken, &item, http.StatusAccepted)
	if result.Err != nil {
		_ = s.store.DeleteJob(ctx, item.ID)
		return result
	}
	return s.remember(ctx, account.UserID, opImport, idemKey, hash, result, item.ID)
}

// Job returns the current job. It does not change status.
func (s *Service) Job(ctx context.Context, accessToken, jobID string) Result {
	account, authErr := s.identity.AccountFromAccess(ctx, accessToken)
	if authErr != nil {
		return Result{Err: fromIdentity(authErr)}
	}
	if jobID == "" {
		return Result{Err: codeErr("NOT_FOUND")}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	item, err := s.store.FindJob(ctx, jobID)
	if err != nil {
		return Result{Err: codeErr("INTERNAL")}
	}
	if item == nil || item.UserID != account.UserID {
		return Result{Err: codeErr("NOT_FOUND")}
	}
	return s.jobResult(ctx, accessToken, item, http.StatusOK)
}

// Skip records that the learner has no material. It does not create a pack or a job.
func (s *Service) Skip(ctx context.Context, accessToken, idemKey string, raw []byte) Result {
	account, authErr := s.identity.AccountFromAccess(ctx, accessToken)
	if authErr != nil {
		return Result{Err: fromIdentity(authErr)}
	}
	if !validIdempotencyKey(idemKey) {
		return Result{Err: codeErr("VALIDATION_MISSING_FIELD")}
	}
	raw = optionalObject(raw)
	canon, err := canonicalJSON(raw)
	if err != nil {
		return Result{Err: codeErr("VALIDATION_MISSING_FIELD")}
	}
	hash := bodyHash(canon)
	s.mu.Lock()
	defer s.mu.Unlock()
	if result, done := s.replay(ctx, accessToken, account.UserID, opSkip, idemKey, hash); done {
		return result
	}
	if decErr := decodeObject(raw, &skipBody{}); decErr != nil {
		return s.remember(ctx, account.UserID, opSkip, idemKey, hash, Result{Err: decErr}, "")
	}
	if err := s.store.RecordSkip(ctx, account.UserID, s.now()); err != nil {
		return Result{Err: codeErr("INTERNAL")}
	}
	if err := s.identity.CompleteFirstPack(ctx, account.UserID); err != nil {
		return Result{Err: codeErr("INTERNAL")}
	}
	result := s.skipResult(ctx, accessToken)
	if result.Err != nil {
		return result
	}
	return s.remember(ctx, account.UserID, opSkip, idemKey, hash, result, "")
}

// TestingSetJobStatus moves a job for tests. It is not an HTTP route.
// map_ready is the import write point for has_first_pack. Other statuses do not clear it.
// Import acceptance does not call this.
func (s *Service) TestingSetJobStatus(ctx context.Context, jobID, status string) error {
	if !knownStatus(status) || jobID == "" {
		return errors.New("status")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	item, err := s.store.UpdateJobStatus(ctx, jobID, status)
	if err != nil {
		return err
	}
	if status != statusMapReady {
		return nil
	}
	return s.identity.CompleteFirstPack(ctx, item.UserID)
}

// TestingJobCount returns how many import jobs the learner has.
func (s *Service) TestingJobCount(ctx context.Context, userID string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.store.CountJobs(ctx, userID)
}

// TestingSkipRecorded reports whether the skip row is stored.
func (s *Service) TestingSkipRecorded(ctx context.Context, userID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.store.SkipRecorded(ctx, userID)
}

func (s *Service) replay(ctx context.Context, accessToken, userID, op, key string, hash [32]byte) (Result, bool) {
	row, err := s.store.FindIdempotency(ctx, userID, op, key, s.now())
	if err != nil {
		return Result{Err: codeErr("INTERNAL")}, true
	}
	if row == nil {
		return Result{}, false
	}
	if !hmac.Equal(row.Hash[:], hash[:]) {
		return Result{Err: codeErr("IDEMPOTENCY_KEY_MISMATCH")}, true
	}
	if row.ErrCode != "" {
		return Result{Err: &Error{Code: row.ErrCode, Message: messages[row.ErrCode], Details: row.ErrDetails}}, true
	}
	if op == opSkip {
		return s.skipResult(ctx, accessToken), true
	}
	item, err := s.store.FindJob(ctx, row.JobID)
	if err != nil || item == nil || item.UserID != userID {
		return Result{Err: codeErr("INTERNAL")}, true
	}
	return s.jobResult(ctx, accessToken, item, http.StatusAccepted), true
}

func (s *Service) remember(ctx context.Context, userID, op, key string, hash [32]byte, result Result, jobID string) Result {
	if result.Err != nil && (result.Err.Code == "INTERNAL" || result.Err.Code == "IDEMPOTENCY_KEY_MISMATCH") {
		return result
	}
	row := idemRow{Hash: hash, JobID: jobID, Expires: s.now().Add(s.idemTTL)}
	if result.Err != nil {
		row.ErrCode = result.Err.Code
		row.ErrDetails = result.Err.Details
	}
	if err := s.store.SaveIdempotency(ctx, userID, op, key, row); err != nil {
		if result.Err != nil {
			return result
		}
		if jobID != "" {
			_ = s.store.DeleteJob(ctx, jobID)
		}
		return Result{Err: codeErr("INTERNAL")}
	}
	return result
}

func (s *Service) jobResult(ctx context.Context, accessToken string, item *job, status int) Result {
	fresh, authErr := s.identity.AccountFromAccess(ctx, accessToken)
	if authErr != nil {
		return Result{Err: fromIdentity(authErr)}
	}
	return s.marshal(status, jobView{
		JobID:       item.ID,
		PackID:      item.PackID,
		MaterialID:  item.MaterialID,
		Source:      item.Source,
		Status:      item.Status,
		MapReady:    item.Status == statusMapReady,
		Retryable:   item.Status == statusFailedRetryable,
		Progress:    progressPtr(item.ProgressMessage),
		FailureCode: textPtr(item.Status, item.FailureCode),
		FailureText: textPtr(item.Status, item.FailureMessage),
		Account:     accountOf(fresh),
		NextStep:    fresh.NextStep,
	})
}

func (s *Service) skipResult(ctx context.Context, accessToken string) Result {
	fresh, authErr := s.identity.AccountFromAccess(ctx, accessToken)
	if authErr != nil {
		return Result{Err: fromIdentity(authErr)}
	}
	return s.marshal(http.StatusOK, skipView{Account: accountOf(fresh), NextStep: fresh.NextStep})
}

func (s *Service) marshal(status int, body any) Result {
	payload, err := json.Marshal(body)
	if err != nil {
		return Result{Err: codeErr("INTERNAL")}
	}
	return Result{Status: status, Body: payload}
}

func fromIdentity(err *identity.Error) *Error {
	if err == nil {
		return nil
	}
	return &Error{Code: err.Code, Message: err.Message, Details: err.Details}
}

func newID(prefix string) string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return prefix + "unknown"
	}
	return prefix + hex.EncodeToString(buf)
}

// Result is an HTTP result.
type Result struct {
	Status int
	Body   []byte
	Err    *Error
}
