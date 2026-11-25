package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	qlutchapi "github.com/zh7551069/qlutch-projects/packages/contracts/gen/go"
	"github.com/zh7551069/qlutch-projects/services/airuntime"
)

type flowProvider struct {
	s01     int
	s02     int
	limit   int
	s02err  error
	failS01 error
	onS01   func()
	onS02   func()
	echo    bool
	points  int
}

func (p *flowProvider) Complete(_ context.Context, req airuntime.ProviderRequest) (airuntime.ProviderResponse, error) {
	if req.SceneCode == airuntime.SceneS02 {
		p.s02++
		p.limit = req.QuestionLimit
		if len(req.Fragment) != 0 {
			return airuntime.ProviderResponse{}, errors.New("s02 received material")
		}
		if p.onS02 != nil {
			p.onS02()
		}
		if p.s02err != nil {
			return airuntime.ProviderResponse{}, p.s02err
		}
		return airuntime.ProviderResponse{Tokens: 1, Questions: []airuntime.Question{{Stem: "预习一题"}, {Stem: "预习二题"}, {Stem: "预习三题"}}}, nil
	}
	p.s01++
	if p.onS01 != nil {
		p.onS01()
	}
	if p.failS01 != nil {
		return airuntime.ProviderResponse{}, p.failS01
	}
	points := p.points
	if points < 1 {
		points = 1
	}
	var pts []airuntime.Point
	for i := 0; i < points; i++ {
		pts = append(pts, airuntime.Point{Title: "考点"})
	}
	title := "目录"
	if p.echo {
		title = string(req.Fragment)
	}
	return airuntime.ProviderResponse{Tokens: 2, Map: airuntime.MapTree{Catalogs: []airuntime.Catalog{{
		Title:    title,
		Chapters: []airuntime.Chapter{{Title: "章节", Points: pts}},
	}}}}, nil
}

func attachRuntime(h *knowledgeHarness, provider airuntime.Provider) (*airuntime.Runtime, *airuntime.MemoryRecorder) {
	rt := airuntime.New()
	rec := airuntime.NewMemoryRecorder()
	rt.SetRecorder(rec)
	if provider != nil {
		rt.SetProvider(provider)
	}
	h.know.SetRuntime(rt)
	return rt, rec
}

func TestA5MapIsVisibleBeforeMapReadyAndS02DoesNotGate(t *testing.T) {
	h := newKnowledge(t)
	access, _, _ := h.register(t, "13800043001")
	h.onboard(t, access)
	pack := h.createPack(t, access)
	if pack.Body.Account.Has_first_pack {
		t.Fatal("empty pack completed the first pack")
	}
	note := []byte("叶绿体把光能写成糖")
	upload := h.upload(t, access, pack.Body.Pack_id, "paste_note", "text/plain", int64(len(note)), "note.txt")
	key := strings.TrimPrefix(upload.Body.Upload_url, "http://127.0.0.1/upload/")
	h.blobs.Put(key, note)
	imported := h.importJob(t, access, "a5-import", pack.Body.Pack_id, upload.Body.Material_id)
	if imported.StatusCode != http.StatusAccepted || imported.Body.Status != qlutchapi.JobStatusAccepted || imported.Body.Map_ready || imported.Body.Account.Has_first_pack {
		t.Fatalf("import = %+v", imported.Body)
	}
	before := h.job(t, access, imported.Body.Job_id)
	if before.Body.Status != qlutchapi.JobStatusAccepted || before.Body.Map_ready {
		t.Fatalf("get before worker = %+v", before.Body)
	}
	if ok, err := h.know.TestingMapEnterable(context.Background(), pack.Body.Pack_id); err != nil || ok {
		t.Fatalf("map before worker = %v %v", ok, err)
	}

	var sawReady bool
	provider := &flowProvider{onS01: func() {
		mid := h.job(t, access, imported.Body.Job_id)
		ok, err := h.know.TestingMapEnterable(context.Background(), pack.Body.Pack_id)
		if err != nil || ok || mid.Body.Status == qlutchapi.JobStatusMapReady || mid.Body.Map_ready || mid.Body.Account.Has_first_pack {
			t.Fatalf("visible during s01 status=%s map=%v flag=%v enterable=%v", mid.Body.Status, mid.Body.Map_ready, mid.Body.Account.Has_first_pack, ok)
		}
	}, onS02: func() {
		ready := h.job(t, access, imported.Body.Job_id)
		ok, err := h.know.TestingMapEnterable(context.Background(), pack.Body.Pack_id)
		if err != nil || !ok || ready.Body.Status != qlutchapi.JobStatusMapReady || !ready.Body.Map_ready || ready.Body.Retryable || !ready.Body.Account.Has_first_pack || ready.Body.Next_step != qlutchapi.NextStepREADYHOME {
			t.Fatalf("s02 saw %+v enterable=%v err=%v", ready.Body, ok, err)
		}
		sawReady = true
	}}
	_, rec := attachRuntime(h, provider)
	if err := h.know.ProcessQueued(context.Background(), h.queue.Items()[0]); err != nil {
		t.Fatal(err)
	}
	if !sawReady || provider.s01 != 1 || provider.s02 != 1 || provider.limit >= 100 || provider.limit < 1 {
		t.Fatalf("provider s01=%d s02=%d limit=%d saw=%v", provider.s01, provider.s02, provider.limit, sawReady)
	}
	final := h.job(t, access, imported.Body.Job_id)
	if strings.Contains(string(final.Raw), "叶绿体") || strings.Contains(string(final.Raw), "sk-") {
		t.Fatalf("response leaked material: %s", final.Raw)
	}
	assertJobFields(t, final.Raw)
	rows, err := rec.List(context.Background(), QueryAll())
	if err != nil {
		t.Fatal(err)
	}
	again := h.job(t, access, imported.Body.Job_id)
	later, err := rec.List(context.Background(), QueryAll())
	if err != nil || len(later) != len(rows) || again.Body.Status != final.Body.Status || again.Body.Map_ready != final.Body.Map_ready {
		t.Fatalf("get changed state rows %d→%d status %s", len(rows), len(later), again.Body.Status)
	}
	if provider.s01 != 1 || provider.s02 != 1 {
		t.Fatal("get triggered a model call")
	}
	s01, err := rec.List(context.Background(), airuntime.Query{SceneCode: airuntime.SceneS01, Subclass: airuntime.SubclassText})
	if err != nil || len(s01) != 1 || !s01[0].Success || s01[0].CacheHit || s01[0].Caller != airuntime.CallerLearner || strings.Contains(s01[0].Fingerprint, "叶绿") {
		t.Fatalf("s01 = %+v %v", s01, err)
	}
	s02, err := rec.List(context.Background(), airuntime.Query{SceneCode: airuntime.SceneS02, Subclass: airuntime.SubclassPrefetch})
	if err != nil || len(s02) != 1 || !s02[0].Success || s02[0].JobID != imported.Body.Job_id {
		t.Fatalf("s02 = %+v %v", s02, err)
	}
	if n, err := h.know.TestingPrefetchCount(context.Background(), imported.Body.Job_id); err != nil || n < 1 || n >= 100 {
		t.Fatalf("prefetch = %d %v", n, err)
	}
	if h.onboarding(t, access).Body.Next_step != qlutchapi.NextStepREADYHOME {
		t.Fatal("session did not follow map_ready")
	}
	replay := h.importJob(t, access, "a5-import", pack.Body.Pack_id, upload.Body.Material_id)
	if replay.StatusCode != http.StatusAccepted || replay.Body.Job_id != imported.Body.Job_id || replay.Body.Status != qlutchapi.JobStatusMapReady {
		t.Fatalf("replay = %+v", replay.Body)
	}
}

func TestA5RetryReusesTheObjectKeyAndCache(t *testing.T) {
	h := newKnowledge(t)
	access, _, _ := h.register(t, "13800043002")
	h.onboard(t, access)
	pack := h.createPack(t, access).Body.Pack_id
	note := []byte("叶绿体复试材料")
	upload := h.upload(t, access, pack, "paste_note", "text/plain", int64(len(note)), "")
	key := strings.TrimPrefix(upload.Body.Upload_url, "http://127.0.0.1/upload/")
	h.blobs.Put(key, note)
	first := h.importJob(t, access, "a5-fail", pack, upload.Body.Material_id)
	provider := &flowProvider{failS01: errors.New("sk-secret 叶绿体复试材料")}
	_, rec := attachRuntime(h, provider)
	if err := h.know.ProcessQueued(context.Background(), h.queue.Items()[0]); err != nil {
		t.Fatal(err)
	}
	failed := h.job(t, access, first.Body.Job_id)
	if failed.Body.Status != qlutchapi.JobStatusFailedRetryable || !failed.Body.Retryable || failed.Body.Map_ready || failed.Body.Account.Has_first_pack || failed.Body.Failure_code != nil {
		t.Fatalf("failed = %+v", failed.Body)
	}
	if failed.Body.Failure_message == nil || strings.Contains(*failed.Body.Failure_message, "sk-secret") || strings.Contains(*failed.Body.Failure_message, "叶绿体") {
		t.Fatalf("message = %q", stringOrEmpty(failed.Body.Failure_message))
	}
	objectKey, err := h.know.TestingMaterialKey(context.Background(), upload.Body.Material_id)
	if err != nil || objectKey == "" || objectKey != key {
		t.Fatalf("key = %s %v", objectKey, err)
	}
	second := h.importJob(t, access, "a5-retry", pack, upload.Body.Material_id)
	if second.StatusCode != http.StatusAccepted || second.Body.Job_id == first.Body.Job_id || second.Body.Material_id != upload.Body.Material_id {
		t.Fatalf("retry import = %+v", second.Body)
	}
	kept, err := h.know.TestingMaterialKey(context.Background(), upload.Body.Material_id)
	if err != nil || kept != objectKey {
		t.Fatalf("reused key = %s %v", kept, err)
	}
	if h.job(t, access, first.Body.Job_id).Body.Status != qlutchapi.JobStatusFailedRetryable {
		t.Fatal("old job was reopened")
	}
	provider.failS01 = nil
	items := h.queue.Items()
	if err := h.know.ProcessQueued(context.Background(), items[len(items)-1]); err != nil {
		t.Fatal(err)
	}
	ready := h.job(t, access, second.Body.Job_id)
	if ready.Body.Status != qlutchapi.JobStatusMapReady || !ready.Body.Account.Has_first_pack || h.job(t, access, first.Body.Job_id).Body.Status != qlutchapi.JobStatusFailedRetryable {
		t.Fatalf("ready = %+v", ready.Body)
	}
	third := h.importJob(t, access, "a5-cache", pack, upload.Body.Material_id)
	calls := provider.s01
	if err := h.know.ProcessQueued(context.Background(), h.queue.Items()[len(h.queue.Items())-1]); err != nil {
		t.Fatal(err)
	}
	if provider.s01 != calls {
		t.Fatalf("cache still called the provider: %d → %d", calls, provider.s01)
	}
	hit := true
	rows, err := rec.List(context.Background(), airuntime.Query{SceneCode: airuntime.SceneS01, JobID: third.Body.Job_id, CacheHit: true, CacheHitSet: hit})
	if err != nil || len(rows) != 1 || !rows[0].Success || rows[0].Tokens != 0 {
		t.Fatalf("cache row = %+v %v", rows, err)
	}
	failedRows, err := rec.List(context.Background(), airuntime.Query{SceneCode: airuntime.SceneS01, JobID: first.Body.Job_id})
	if err != nil || len(failedRows) != 1 || failedRows[0].Success || failedRows[0].Fingerprint == "" || failedRows[0].Fingerprint != rows[0].Fingerprint {
		t.Fatalf("failed ledger = %+v hit=%+v", failedRows, rows)
	}
}

func TestA5UnconfiguredProviderDoesNotMarkReady(t *testing.T) {
	h := newKnowledge(t)
	access, _, _ := h.register(t, "13800043003")
	note := []byte("没有提供方")
	pack, upload := uploadNote(t, h, access, note)
	imported := h.importJob(t, access, "a5-none", pack, upload)
	_, rec := attachRuntime(h, nil)
	if err := h.know.ProcessQueued(context.Background(), h.queue.Items()[0]); err != nil {
		t.Fatal(err)
	}
	got := h.job(t, access, imported.Body.Job_id)
	if got.StatusCode != http.StatusOK || got.Body.Status != qlutchapi.JobStatusFailedRetryable || !got.Body.Retryable || got.Body.Map_ready || got.Body.Account.Has_first_pack {
		t.Fatalf("job = %+v", got.Body)
	}
	if ok, err := h.know.TestingMapEnterable(context.Background(), pack); err != nil || ok {
		t.Fatalf("map = %v %v", ok, err)
	}
	rows, err := rec.List(context.Background(), airuntime.Query{SceneCode: airuntime.SceneS01, JobID: imported.Body.Job_id})
	if err != nil || len(rows) != 1 || rows[0].Success || rows[0].Subclass != airuntime.SubclassText {
		t.Fatalf("ledger = %+v %v", rows, err)
	}
	s02, err := rec.List(context.Background(), airuntime.Query{SceneCode: airuntime.SceneS02})
	if err != nil || len(s02) != 0 {
		t.Fatalf("s02 = %+v", s02)
	}
	if h.onboarding(t, access).Body.Next_step != qlutchapi.NextStepNEEDONBOARDING {
		t.Fatal("failure changed next_step")
	}
}

func TestA5MissingMaterialIsTerminal(t *testing.T) {
	h := newKnowledge(t)
	access, _, _ := h.register(t, "13800043004")
	h.onboard(t, access)
	note := []byte("稍后被删掉")
	pack, upload := uploadNote(t, h, access, note)
	imported := h.importJob(t, access, "a5-gone", pack, upload)
	key, err := h.know.TestingMaterialKey(context.Background(), upload)
	if err != nil {
		t.Fatal(err)
	}
	h.blobs.Remove(key)
	provider := &flowProvider{}
	_, rec := attachRuntime(h, provider)
	if err := h.know.ProcessQueued(context.Background(), h.queue.Items()[0]); err != nil {
		t.Fatal(err)
	}
	if provider.s01 != 0 {
		t.Fatal("missing material called the provider")
	}
	got := h.job(t, access, imported.Body.Job_id)
	if got.Body.Status != qlutchapi.JobStatusFailedTerminal || got.Body.Retryable || got.Body.Map_ready || got.Body.Account.Has_first_pack {
		t.Fatalf("terminal = %+v", got.Body)
	}
	if got.Body.Failure_message == nil || strings.Contains(*got.Body.Failure_message, "稍后") {
		t.Fatalf("message = %+v", got.Body.Failure_message)
	}
	rows, err := rec.List(context.Background(), airuntime.Query{JobID: imported.Body.Job_id, SceneCode: airuntime.SceneS01})
	if err != nil || len(rows) != 1 || rows[0].Success {
		t.Fatalf("ledger = %+v %v", rows, err)
	}
}

func TestA5S02FailureDoesNotCloseTheGate(t *testing.T) {
	h := newKnowledge(t)
	access, _, _ := h.register(t, "13800043005")
	h.onboard(t, access)
	note := []byte("地图可以进")
	pack, upload := uploadNote(t, h, access, note)
	imported := h.importJob(t, access, "a5-s02", pack, upload)
	provider := &flowProvider{s02err: errors.New("prefetch down")}
	_, rec := attachRuntime(h, provider)
	if err := h.know.ProcessQueued(context.Background(), h.queue.Items()[0]); err != nil {
		t.Fatal(err)
	}
	got := h.job(t, access, imported.Body.Job_id)
	if got.Body.Status != qlutchapi.JobStatusMapReady || !got.Body.Map_ready || !got.Body.Account.Has_first_pack || got.Body.Next_step != qlutchapi.NextStepREADYHOME {
		t.Fatalf("gate = %+v", got.Body)
	}
	if n, err := h.know.TestingPrefetchCount(context.Background(), imported.Body.Job_id); err != nil || n != 0 {
		t.Fatalf("prefetch = %d %v", n, err)
	}
	rows, err := rec.List(context.Background(), airuntime.Query{SceneCode: airuntime.SceneS02, JobID: imported.Body.Job_id})
	if err != nil || len(rows) != 1 || rows[0].Success {
		t.Fatalf("s02 = %+v %v", rows, err)
	}
}

func TestA5MapReadyBeforeOnboardingStaysOnboarding(t *testing.T) {
	h := newKnowledge(t)
	access, _, _ := h.register(t, "13800043006")
	note := []byte("还没建档")
	pack, upload := uploadNote(t, h, access, note)
	imported := h.importJob(t, access, "a5-onboard", pack, upload)
	attachRuntime(h, &flowProvider{})
	if err := h.know.ProcessQueued(context.Background(), h.queue.Items()[0]); err != nil {
		t.Fatal(err)
	}
	got := h.job(t, access, imported.Body.Job_id)
	if !got.Body.Map_ready || !got.Body.Account.Has_first_pack || got.Body.Next_step != qlutchapi.NextStepNEEDONBOARDING {
		t.Fatalf("job = %+v", got.Body)
	}
	if h.onboarding(t, access).Body.Next_step != qlutchapi.NextStepNEEDONBOARDING {
		t.Fatal("refresh left onboarding")
	}
}

func uploadNote(t *testing.T, h *knowledgeHarness, access string, note []byte) (string, string) {
	t.Helper()
	pack := h.createPack(t, access).Body.Pack_id
	upload := h.upload(t, access, pack, "paste_note", "text/plain", int64(len(note)), "")
	h.blobs.Put(strings.TrimPrefix(upload.Body.Upload_url, "http://127.0.0.1/upload/"), note)
	return pack, upload.Body.Material_id
}

func assertJobFields(t *testing.T, raw []byte) {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{
		"job_id": true, "pack_id": true, "material_id": true, "source": true,
		"status": true, "map_ready": true, "retryable": true, "progress_message": true,
		"failure_code": true, "failure_message": true, "account": true, "next_step": true,
	}
	for key := range fields {
		if !allowed[key] {
			t.Fatalf("unexpected job field %s", key)
		}
	}
}

func QueryAll() airuntime.Query { return airuntime.Query{} }

func stringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
