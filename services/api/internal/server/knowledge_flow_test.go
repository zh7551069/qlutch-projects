package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	qlutchapi "github.com/zh7551069/qlutch-projects/packages/contracts/gen/go"
	"github.com/zh7551069/qlutch-projects/services/api/internal/identity"
	"github.com/zh7551069/qlutch-projects/services/api/internal/knowledge"
)

func TestKnowledgePackAndImportDoNotCompleteFirstPack(t *testing.T) {
	h := newKnowledge(t)
	access, _, userID := h.register(t, "13800041001")
	h.onboard(t, access)

	created := h.createPack(t, access)
	if created.StatusCode != http.StatusCreated || created.Body.Pack_id == "" || created.Body.Account.Has_first_pack || created.Body.Next_step != qlutchapi.NextStepNEEDFIRSTPACK || created.Body.Account.Is_new_user {
		t.Fatalf("pack = %d %+v", created.StatusCode, created.Body)
	}
	state := h.onboarding(t, access)
	if state.Body.Account.Has_first_pack || state.Body.Next_step != qlutchapi.NextStepNEEDFIRSTPACK {
		t.Fatalf("after pack = %+v", state.Body)
	}

	upload := h.upload(t, access, created.Body.Pack_id, "paste_note", "text/plain", 11, "notes.txt")
	if upload.StatusCode != http.StatusOK || upload.Body.Upload_method != "PUT" || upload.Body.Upload_headers["Content-Type"] != "text/plain" || upload.Body.Expires_in != knowledge.DefaultPresignSeconds || upload.Body.Account.Has_first_pack {
		t.Fatalf("upload = %d %+v", upload.StatusCode, upload.Body)
	}
	if strings.Contains(string(upload.Raw), "object_key") || strings.Contains(upload.Body.Upload_url, "notes.txt") {
		t.Fatalf("upload leaked key or filename: %s", upload.Raw)
	}
	objectKey := strings.TrimPrefix(upload.Body.Upload_url, "http://127.0.0.1/upload/")
	if objectKey == "" || strings.Contains(objectKey, "notes.txt") {
		t.Fatalf("object key = %s", objectKey)
	}
	h.blobs.Complete(upload.Body.Upload_url, 11)

	imported := h.importJob(t, access, "import-1", created.Body.Pack_id, upload.Body.Material_id)
	if imported.StatusCode != http.StatusAccepted || imported.Body.Status != qlutchapi.JobStatusAccepted || imported.Body.Map_ready || imported.Body.Retryable || imported.Body.Account.Has_first_pack || imported.Body.Next_step != qlutchapi.NextStepNEEDFIRSTPACK {
		t.Fatalf("import = %d %+v raw %s", imported.StatusCode, imported.Body, imported.Raw)
	}
	if imported.Body.Source != qlutchapi.MaterialSourcePasteNote || imported.Body.Failure_code != nil || imported.Body.Failure_message != nil || imported.Body.Progress_message != nil {
		t.Fatalf("job shape = %+v", imported.Body)
	}
	if strings.Contains(string(imported.Raw), objectKey) || strings.Contains(string(imported.Raw), "notes.txt") {
		t.Fatalf("import leaked material: %s", imported.Raw)
	}
	if n, err := h.know.TestingJobCount(context.Background(), userID); err != nil || n != 1 {
		t.Fatalf("jobs = %d %v", n, err)
	}
	items := h.queue.Items()
	if len(items) != 1 || !strings.Contains(string(items[0]), imported.Body.Job_id) || strings.Contains(string(items[0]), "notes") || strings.Contains(string(items[0]), objectKey) {
		t.Fatalf("queue = %q", items)
	}
	var queued struct {
		JobID string `json:"job_id"`
	}
	if err := json.Unmarshal(items[0], &queued); err != nil || queued.JobID != imported.Body.Job_id {
		t.Fatalf("payload = %s %v", items[0], err)
	}

	got := h.job(t, access, imported.Body.Job_id)
	if got.StatusCode != http.StatusOK || got.Body.Status != qlutchapi.JobStatusAccepted || got.Body.Map_ready || got.Body.Job_id != imported.Body.Job_id {
		t.Fatalf("get = %+v", got.Body)
	}
	again := h.job(t, access, imported.Body.Job_id)
	if again.Body.Status != got.Body.Status || again.Body.Account.Has_first_pack {
		t.Fatalf("get changed the job: %+v", again.Body)
	}
	if h.onboarding(t, access).Body.Next_step != qlutchapi.NextStepNEEDFIRSTPACK {
		t.Fatal("accepted import opened the home gate")
	}
}

func TestKnowledgeImportRules(t *testing.T) {
	h := newKnowledge(t)
	access, _, userID := h.register(t, "13800041002")
	h.onboard(t, access)
	packA := h.createPack(t, access).Body.Pack_id
	packB := h.createPack(t, access).Body.Pack_id
	if packA == packB {
		t.Fatal("packs collapsed")
	}
	upload := h.upload(t, access, packA, "paste_note", "text/plain", 4, "")
	early := h.importJob(t, access, "before-put", packA, upload.Body.Material_id)
	if early.StatusCode != http.StatusUnprocessableEntity || mustCode(t, early.StatusCode, early.Raw, early.RequestID) != "VALIDATION_MATERIAL_NOT_UPLOADED" {
		t.Fatalf("early = %d %s", early.StatusCode, early.Raw)
	}
	replayEarly := h.importJob(t, access, "before-put", packA, upload.Body.Material_id)
	if replayEarly.StatusCode != http.StatusUnprocessableEntity || len(h.queue.Items()) != 0 {
		t.Fatalf("replay early = %d queue %d", replayEarly.StatusCode, len(h.queue.Items()))
	}
	if n, _ := h.know.TestingJobCount(context.Background(), userID); n != 0 {
		t.Fatalf("jobs after early = %d", n)
	}
	h.blobs.Complete(upload.Body.Upload_url, 4)
	fixed := h.importJob(t, access, "after-put", packA, upload.Body.Material_id)
	if fixed.StatusCode != http.StatusAccepted || fixed.Body.Status != qlutchapi.JobStatusAccepted {
		t.Fatalf("fixed = %d %s", fixed.StatusCode, fixed.Raw)
	}
	replay := h.importJob(t, access, "after-put", packA, upload.Body.Material_id)
	if replay.StatusCode != http.StatusAccepted || replay.Body.Job_id != fixed.Body.Job_id || len(h.queue.Items()) != 1 {
		t.Fatalf("replay = %+v queue %d", replay.Body, len(h.queue.Items()))
	}
	other := h.importJob(t, access, "after-put", packB, upload.Body.Material_id)
	if other.StatusCode != http.StatusConflict || mustCode(t, other.StatusCode, other.Raw, other.RequestID) != "IDEMPOTENCY_KEY_MISMATCH" {
		t.Fatalf("mismatch = %d %s", other.StatusCode, other.Raw)
	}
	second := h.importJob(t, access, "second", packA, upload.Body.Material_id)
	if second.StatusCode != http.StatusConflict || mustCode(t, second.StatusCode, second.Raw, second.RequestID) != "CONFLICT" {
		t.Fatalf("conflict = %d %s", second.StatusCode, second.Raw)
	}
	if err := h.know.TestingSetJobStatus(context.Background(), fixed.Body.Job_id, string(qlutchapi.JobStatusRunning)); err != nil {
		t.Fatal(err)
	}
	whileRunning := h.importJob(t, access, "while-running", packA, upload.Body.Material_id)
	if whileRunning.StatusCode != http.StatusConflict {
		t.Fatalf("running = %d", whileRunning.StatusCode)
	}
	if h.onboarding(t, access).Body.Account.Has_first_pack {
		t.Fatal("running set has_first_pack")
	}
	if err := h.know.TestingSetJobStatus(context.Background(), fixed.Body.Job_id, string(qlutchapi.JobStatusFailedRetryable)); err != nil {
		t.Fatal(err)
	}
	retried := h.importJob(t, access, "retry", packA, upload.Body.Material_id)
	if retried.StatusCode != http.StatusAccepted || retried.Body.Job_id == fixed.Body.Job_id || retried.Body.Material_id != upload.Body.Material_id || retried.Body.Account.Has_first_pack {
		t.Fatalf("retry = %+v", retried.Body)
	}
	if err := h.know.TestingSetJobStatus(context.Background(), retried.Body.Job_id, string(qlutchapi.JobStatusMapReady)); err != nil {
		t.Fatal(err)
	}
	ready := h.job(t, access, retried.Body.Job_id)
	if !ready.Body.Map_ready || ready.Body.Status != qlutchapi.JobStatusMapReady || ready.Body.Retryable || !ready.Body.Account.Has_first_pack || ready.Body.Next_step != qlutchapi.NextStepREADYHOME {
		t.Fatalf("map ready = %+v", ready.Body)
	}
	stale := h.importJob(t, access, "retry", packA, upload.Body.Material_id)
	if stale.StatusCode != http.StatusAccepted || stale.Body.Job_id != retried.Body.Job_id || stale.Body.Status != qlutchapi.JobStatusMapReady || !stale.Body.Account.Has_first_pack {
		t.Fatalf("stale replay = %d %+v", stale.StatusCode, stale.Body)
	}
	if h.onboarding(t, access).Body.Next_step != qlutchapi.NextStepREADYHOME {
		t.Fatal("session did not follow map_ready")
	}
	kept := h.job(t, access, fixed.Body.Job_id)
	if kept.Body.Status != qlutchapi.JobStatusFailedRetryable {
		t.Fatalf("old job = %+v", kept.Body)
	}
}

func TestKnowledgeValidationAndAuth(t *testing.T) {
	h := newKnowledge(t)
	access, _, _ := h.register(t, "13800041003")
	pack := h.createPack(t, access).Body.Pack_id
	otherAccess, _, _ := h.register(t, "13800041004")

	cases := []struct {
		source, content string
		size            int64
		filename        string
		want            string
		status          int
	}{
		{"douyin", "video/mp4", 10, "", "VALIDATION_INVALID_SOURCE", http.StatusBadRequest},
		{"isbn", "text/plain", 10, "", "VALIDATION_INVALID_SOURCE", http.StatusBadRequest},
		{"bilibili", "video/mp4", 10, "", "VALIDATION_INVALID_SOURCE", http.StatusBadRequest},
		{"camera", "text/plain", 10, "", "VALIDATION_UNSUPPORTED_MEDIA", http.StatusBadRequest},
		{"paste_note", "image/jpeg", 10, "", "VALIDATION_UNSUPPORTED_MEDIA", http.StatusBadRequest},
		{"file", "application/pdf", knowledge.DefaultMaxBytes + 1, "book.pdf", "VALIDATION_MEDIA_TOO_LARGE", http.StatusBadRequest},
		{"file", "application/pdf", 10, "a/book.pdf", "VALIDATION_MISSING_FIELD", http.StatusBadRequest},
	}
	for _, tc := range cases {
		res := h.upload(t, access, pack, tc.source, tc.content, tc.size, tc.filename)
		if res.StatusCode != tc.status || mustCode(t, res.StatusCode, res.Raw, res.RequestID) != tc.want {
			t.Fatalf("%s %s = %d %s", tc.source, tc.content, res.StatusCode, res.Raw)
		}
	}
	tooBig := h.upload(t, access, pack, "file", "application/pdf", knowledge.DefaultMaxBytes+1, "book.pdf")
	var env qlutchapi.ErrorBody
	if err := json.Unmarshal(tooBig.Raw, &env); err != nil {
		t.Fatal(err)
	}
	var details struct {
		Max int64 `json:"max_bytes"`
	}
	if err := json.Unmarshal(env.Error.Details, &details); err != nil || details.Max != knowledge.DefaultMaxBytes {
		t.Fatalf("details = %s", env.Error.Details)
	}
	missingPack := h.upload(t, access, "pack_missing", "camera", "image/jpeg", 10, "")
	if missingPack.StatusCode != http.StatusNotFound {
		t.Fatalf("missing pack = %d", missingPack.StatusCode)
	}
	foreign := h.upload(t, otherAccess, pack, "camera", "image/jpeg", 8, "")
	if foreign.StatusCode != http.StatusNotFound {
		t.Fatalf("foreign pack = %d %s", foreign.StatusCode, foreign.Raw)
	}
	okUpload := h.upload(t, access, pack, "camera", "image/jpeg", 8, "shot.jpg")
	h.blobs.Complete(okUpload.Body.Upload_url, 8)
	foreignImport := h.importJob(t, otherAccess, "foreign", pack, okUpload.Body.Material_id)
	if foreignImport.StatusCode != http.StatusNotFound {
		t.Fatalf("foreign import = %d", foreignImport.StatusCode)
	}
	wrongSize := h.upload(t, access, pack, "file", "audio/mpeg", 6, "a.mp3")
	h.blobs.Complete(wrongSize.Body.Upload_url, 5)
	badSize := h.importJob(t, access, "size", pack, wrongSize.Body.Material_id)
	if badSize.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("size = %d %s", badSize.StatusCode, badSize.Raw)
	}
	noKey := h.importJob(t, access, "", pack, okUpload.Body.Material_id)
	if noKey.StatusCode != http.StatusBadRequest || mustCode(t, noKey.StatusCode, noKey.Raw, noKey.RequestID) != "VALIDATION_MISSING_FIELD" {
		t.Fatalf("missing key = %d %s", noKey.StatusCode, noKey.Raw)
	}
	anon := h.upload(t, "", pack, "camera", "image/jpeg", 8, "")
	if anon.StatusCode != http.StatusUnauthorized || mustCode(t, anon.StatusCode, anon.Raw, anon.RequestID) != "AUTH_REQUIRED" {
		t.Fatalf("anon = %d %s", anon.StatusCode, anon.Raw)
	}
	realm := h.raw(t, http.MethodPost, "/v1/knowledge/packs", `{}`, "", "qlu_i_internal")
	if realm.StatusCode != http.StatusForbidden || mustCode(t, realm.StatusCode, realm.Body.Bytes(), realm.Header.Get("X-Request-Id")) != "AUTH_WRONG_REALM" {
		t.Fatalf("realm = %d %s", realm.StatusCode, realm.Body.String())
	}
	for _, path := range []string{"/v1/knowledge/douyin", "/v1/knowledge/bilibili", "/v1/knowledge/isbn"} {
		res := h.raw(t, http.MethodPost, path, `{}`, "", access)
		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("%s = %d", path, res.StatusCode)
		}
	}
	unknown := h.job(t, access, "job_missing")
	if unknown.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown job = %d", unknown.StatusCode)
	}
	imported := h.importJob(t, access, "owned", pack, okUpload.Body.Material_id)
	if imported.StatusCode != http.StatusAccepted {
		t.Fatalf("owned = %d %s", imported.StatusCode, imported.Raw)
	}
	hidden := h.job(t, otherAccess, imported.Body.Job_id)
	if hidden.StatusCode != http.StatusNotFound {
		t.Fatalf("hidden job = %d", hidden.StatusCode)
	}
}

func TestKnowledgeEnqueueFailureDoesNotStick(t *testing.T) {
	h := newKnowledge(t)
	access, _, userID := h.register(t, "13800041005")
	pack := h.createPack(t, access).Body.Pack_id
	upload := h.upload(t, access, pack, "paste_note", "text/plain", 3, "")
	h.blobs.Complete(upload.Body.Upload_url, 3)
	h.queue.Err = io.EOF
	failed := h.importJob(t, access, "queue-down", pack, upload.Body.Material_id)
	if failed.StatusCode != http.StatusInternalServerError || mustCode(t, failed.StatusCode, failed.Raw, failed.RequestID) != "INTERNAL" {
		t.Fatalf("enqueue = %d %s", failed.StatusCode, failed.Raw)
	}
	if n, _ := h.know.TestingJobCount(context.Background(), userID); n != 0 || h.onboarding(t, access).Body.Account.Has_first_pack {
		t.Fatal("failed enqueue left a job or completed the first pack")
	}
	h.queue.Err = nil
	ok := h.importJob(t, access, "queue-down", pack, upload.Body.Material_id)
	if ok.StatusCode != http.StatusAccepted || ok.Body.Status != qlutchapi.JobStatusAccepted || ok.Body.Account.Has_first_pack {
		t.Fatalf("retry enqueue = %d %+v", ok.StatusCode, ok.Body)
	}
}

func TestKnowledgeSkip(t *testing.T) {
	h := newKnowledge(t)
	earlyAccess, earlyRefresh, earlyUser := h.register(t, "13800041006")
	skipped := h.skip(t, earlyAccess, "skip-early", nil)
	if skipped.StatusCode != http.StatusOK || !skipped.Body.Account.Has_first_pack || skipped.Body.Next_step != qlutchapi.NextStepNEEDONBOARDING {
		t.Fatalf("early skip = %d %+v", skipped.StatusCode, skipped.Body)
	}
	if n, _ := h.know.TestingJobCount(context.Background(), earlyUser); n != 0 {
		t.Fatalf("skip created jobs = %d", n)
	}
	recorded, err := h.know.TestingSkipRecorded(context.Background(), earlyUser)
	if err != nil || !recorded {
		t.Fatal("skip was not stored")
	}
	replay := h.skip(t, earlyAccess, "skip-early", &qlutchapi.FirstPackSkipRequest{})
	if replay.StatusCode != http.StatusOK || !replay.Body.Account.Has_first_pack || replay.Body.Next_step != qlutchapi.NextStepNEEDONBOARDING {
		t.Fatalf("replay skip = %+v", replay.Body)
	}
	if h.onboarding(t, earlyAccess).Body.Next_step != qlutchapi.NextStepNEEDONBOARDING {
		t.Fatal("incomplete onboarding left P03")
	}
	mismatch := h.raw(t, http.MethodPost, "/v1/knowledge/first-pack/skip", `{"extra":true}`, "skip-early", earlyAccess)
	if mismatch.StatusCode != http.StatusConflict || mustCode(t, mismatch.StatusCode, mismatch.Body.Bytes(), mismatch.Header.Get("X-Request-Id")) != "IDEMPOTENCY_KEY_MISMATCH" {
		t.Fatalf("skip mismatch = %d %s", mismatch.StatusCode, mismatch.Body.String())
	}
	h.onboard(t, earlyAccess)
	if h.onboarding(t, earlyAccess).Body.Next_step != qlutchapi.NextStepREADYHOME {
		t.Fatal("finished onboarding after skip did not reach home")
	}
	refreshed := h.refresh(t, earlyRefresh)
	if refreshed.Body.Next_step != qlutchapi.NextStepREADYHOME || !refreshed.Body.Account.Has_first_pack {
		t.Fatalf("refresh = %+v", refreshed.Body)
	}

	homeAccess, homeRefresh, homeUser := h.register(t, "13800041007")
	h.onboard(t, homeAccess)
	if h.onboarding(t, homeAccess).Body.Next_step != qlutchapi.NextStepNEEDFIRSTPACK {
		t.Fatal("expected first pack gate")
	}
	homeSkip := h.skip(t, homeAccess, "skip-home", nil)
	if homeSkip.StatusCode != http.StatusOK || homeSkip.Body.Next_step != qlutchapi.NextStepREADYHOME || !homeSkip.Body.Account.Has_first_pack {
		t.Fatalf("home skip = %+v", homeSkip.Body)
	}
	again := h.skip(t, homeAccess, "skip-home-again", nil)
	if again.StatusCode != http.StatusOK || !again.Body.Account.Has_first_pack || again.Body.Next_step != qlutchapi.NextStepREADYHOME {
		t.Fatalf("second skip = %d %+v", again.StatusCode, again.Body)
	}
	missingKey := h.skip(t, homeAccess, "", nil)
	if missingKey.StatusCode != http.StatusBadRequest {
		t.Fatalf("missing skip key = %d %s", missingKey.StatusCode, missingKey.Raw)
	}
	if h.refresh(t, homeRefresh).Body.Next_step != qlutchapi.NextStepREADYHOME {
		t.Fatal("refresh bounced to P04")
	}

	readyAccess, _, readyUser := h.register(t, "13800041008")
	h.onboard(t, readyAccess)
	pack := h.createPack(t, readyAccess).Body.Pack_id
	upload := h.upload(t, readyAccess, pack, "paste_note", "text/plain", 2, "")
	h.blobs.Complete(upload.Body.Upload_url, 2)
	imported := h.importJob(t, readyAccess, "then-skip", pack, upload.Body.Material_id)
	if err := h.know.TestingSetJobStatus(context.Background(), imported.Body.Job_id, string(qlutchapi.JobStatusMapReady)); err != nil {
		t.Fatal(err)
	}
	after := h.skip(t, readyAccess, "skip-after-ready", nil)
	if after.StatusCode != http.StatusOK || !after.Body.Account.Has_first_pack || after.Body.Next_step != qlutchapi.NextStepREADYHOME {
		t.Fatalf("skip after ready = %+v", after.Body)
	}
	kept := h.job(t, readyAccess, imported.Body.Job_id)
	if kept.Body.Status != qlutchapi.JobStatusMapReady || kept.Body.Pack_id != pack {
		t.Fatalf("skip cleared the job: %+v", kept.Body)
	}
	if n, _ := h.know.TestingJobCount(context.Background(), readyUser); n != 1 {
		t.Fatalf("jobs after skip = %d", n)
	}
	if n, _ := h.know.TestingJobCount(context.Background(), homeUser); n != 0 {
		t.Fatal("home skip created a job")
	}
}

type knowledgeHarness struct {
	know   *knowledge.Service
	blobs  *knowledge.MemoryBlobs
	queue  *knowledge.MemoryQueue
	client *qlutchapi.Client
	base   string
	http   *http.Client
}

func newKnowledge(t *testing.T) *knowledgeHarness {
	t.Helper()
	return newKnowledgeWith(t, nil, nil, nil)
}

func newKnowledgeWith(t *testing.T, identStore identity.Store, knowStore knowledge.Store, blobs *knowledge.MemoryBlobs) *knowledgeHarness {
	t.Helper()
	clock := identity.NewManualClock(time.Date(2026, 11, 26, 1, 0, 0, 0, time.UTC))
	ident, err := identity.New(identity.Options{Sender: &identity.CaptureSender{}, Code: fixedCode, Clock: clock, Store: identStore})
	if err != nil {
		t.Fatal(err)
	}
	if blobs == nil {
		blobs = knowledge.NewMemoryBlobs()
	}
	queue := knowledge.NewMemoryQueue()
	know, err := knowledge.New(knowledge.Options{
		Identity: ident,
		Store:    knowStore,
		Blobs:    blobs,
		Queue:    queue,
		Now:      clock.Now,
	})
	if err != nil {
		t.Fatal(err)
	}
	httpSrv := httptest.NewServer(NewWithKnowledge(nil, ident, know).Handler(nil))
	t.Cleanup(httpSrv.Close)
	return &knowledgeHarness{
		know:   know,
		blobs:  blobs,
		queue:  queue,
		client: &qlutchapi.Client{BaseURL: httpSrv.URL, HTTP: httpSrv.Client()},
		base:   httpSrv.URL,
		http:   httpSrv.Client(),
	}
}

func (h *knowledgeHarness) register(t *testing.T, phone string) (string, string, string) {
	t.Helper()
	sent, err := h.client.SendIdentityOtp(context.Background(), "reg-"+phone, qlutchapi.OtpSendRequest{
		Region_code:    "+86",
		Phone_national: phone,
		Purpose:        qlutchapi.OtpPurposeRegister,
	}, qlutchapi.CallMeta{})
	if err != nil || sent.StatusCode != http.StatusOK {
		t.Fatalf("send %s = %v %d %s", phone, err, sent.StatusCode, sent.Raw)
	}
	verified, err := h.client.VerifyIdentityOtp(context.Background(), qlutchapi.OtpVerifyRequest{
		Region_code:    "+86",
		Phone_national: phone,
		Purpose:        qlutchapi.OtpPurposeRegister,
		Code:           testCode,
	}, qlutchapi.CallMeta{})
	if err != nil || verified.StatusCode != http.StatusOK {
		t.Fatalf("verify %s = %v %d %s", phone, err, verified.StatusCode, verified.Raw)
	}
	return verified.Body.Access_token, verified.Body.Refresh_token, verified.Body.Account.User_id
}

func (h *knowledgeHarness) onboard(t *testing.T, access string) {
	t.Helper()
	h.putOK(t, access, "nickname", "小明", "")
	for _, step := range []struct{ id, answer string }{
		{"age", "18-24"},
		{"study_or_work", "student"},
		{"identity", "candidate"},
		{"how_to_address", "friend"},
		{"exam_track", "language"},
		{"daily_time", "half-hour"},
	} {
		h.putOK(t, access, step.id, "", step.answer)
	}
}

func (h *knowledgeHarness) putOK(t *testing.T, access, step, nickname, answer string) {
	t.Helper()
	body := qlutchapi.OnboardingWriteRequest{Step_id: qlutchapi.StepId(step)}
	if nickname != "" {
		body.Nickname = &nickname
	}
	if answer != "" {
		body.Answer = &answer
	}
	res, err := h.client.PutIdentityOnboarding(context.Background(), body, qlutchapi.CallMeta{Bearer: access})
	if err != nil || res.StatusCode != http.StatusOK {
		t.Fatalf("put %s = %v %d %s", step, err, res.StatusCode, res.Raw)
	}
}

func (h *knowledgeHarness) onboarding(t *testing.T, access string) qlutchapi.APIResponse[qlutchapi.OnboardingState] {
	t.Helper()
	res, err := h.client.GetIdentityOnboarding(context.Background(), qlutchapi.CallMeta{Bearer: access})
	if err != nil || res.StatusCode != http.StatusOK {
		t.Fatalf("onboarding = %v %d %s", err, res.StatusCode, res.Raw)
	}
	return res
}

func (h *knowledgeHarness) refresh(t *testing.T, token string) qlutchapi.APIResponse[qlutchapi.SessionResponse] {
	t.Helper()
	res, err := h.client.RefreshIdentitySession(context.Background(), qlutchapi.SessionRefreshRequest{Refresh_token: token}, qlutchapi.CallMeta{})
	if err != nil || res.StatusCode != http.StatusOK {
		t.Fatalf("refresh = %v %d %s", err, res.StatusCode, res.Raw)
	}
	return res
}

func (h *knowledgeHarness) createPack(t *testing.T, access string) qlutchapi.APIResponse[qlutchapi.KnowledgePack] {
	t.Helper()
	res, err := h.client.CreateKnowledgePack(context.Background(), &qlutchapi.KnowledgePackCreateRequest{}, qlutchapi.CallMeta{Bearer: access})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func (h *knowledgeHarness) upload(t *testing.T, access, pack, source, content string, size int64, filename string) qlutchapi.APIResponse[qlutchapi.MaterialUploadUrlResponse] {
	t.Helper()
	body := qlutchapi.MaterialUploadUrlRequest{
		Pack_id:      pack,
		Source:       qlutchapi.MaterialSource(source),
		Content_type: content,
		Byte_size:    size,
	}
	if filename != "" {
		body.Filename = &filename
	}
	res, err := h.client.CreateKnowledgeMaterialUploadUrl(context.Background(), body, qlutchapi.CallMeta{Bearer: access})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func (h *knowledgeHarness) importJob(t *testing.T, access, key, pack, material string) qlutchapi.APIResponse[qlutchapi.KnowledgeJob] {
	t.Helper()
	res, err := h.client.CreateKnowledgeImport(context.Background(), key, qlutchapi.ImportRequest{Pack_id: pack, Material_id: material}, qlutchapi.CallMeta{Bearer: access})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func (h *knowledgeHarness) job(t *testing.T, access, id string) qlutchapi.APIResponse[qlutchapi.KnowledgeJob] {
	t.Helper()
	res, err := h.client.GetKnowledgeJob(context.Background(), id, qlutchapi.CallMeta{Bearer: access})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func (h *knowledgeHarness) skip(t *testing.T, access, key string, body *qlutchapi.FirstPackSkipRequest) qlutchapi.APIResponse[qlutchapi.FirstPackSkipResponse] {
	t.Helper()
	res, err := h.client.SkipKnowledgeFirstPack(context.Background(), key, body, qlutchapi.CallMeta{Bearer: access})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

type rawResponse struct {
	StatusCode int
	Header     http.Header
	Body       *bytes.Buffer
}

func (h *knowledgeHarness) raw(t *testing.T, method, path, body, key, bearer string) rawResponse {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, h.base+path, reader)
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	res, err := h.http.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	payload, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return rawResponse{StatusCode: res.StatusCode, Header: res.Header, Body: bytes.NewBuffer(payload)}
}
