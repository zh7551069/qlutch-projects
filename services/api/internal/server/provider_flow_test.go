package server

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	qlutchapi "github.com/zh7551069/qlutch-projects/packages/contracts/gen/go"
	"github.com/zh7551069/qlutch-projects/services/airuntime"
)

func TestMissingProviderConfigFailsRetryable(t *testing.T) {
	clearProviderEnv(t)
	h := newKnowledge(t)
	access, _, _ := h.register(t, "13800043101")
	note := []byte("没有钥匙")
	pack, upload := uploadNote(t, h, access, note)
	imported := h.importJob(t, access, "prov-missing-key", pack, upload)
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		hits++
	}))
	t.Cleanup(srv.Close)
	rt := airuntime.New()
	rec := airuntime.NewMemoryRecorder()
	rt.SetRecorder(rec)
	t.Setenv("QLUTCH_LLM_BASE_URL", srv.URL)
	airuntime.InstallFromEnv(rt)
	if rt.Status() != "未配置提供方" {
		t.Fatalf("status %s", rt.Status())
	}
	h.know.SetRuntime(rt)
	if err := h.know.ProcessQueued(context.Background(), h.queue.Items()[0]); err != nil {
		t.Fatal(err)
	}
	got := h.job(t, access, imported.Body.Job_id)
	if got.Body.Status != qlutchapi.JobStatusFailedRetryable || !got.Body.Retryable || got.Body.Map_ready || got.Body.Account.Has_first_pack {
		t.Fatalf("job %+v", got.Body)
	}
	if ok, err := h.know.TestingMapEnterable(context.Background(), pack); err != nil || ok {
		t.Fatalf("map %v %v", ok, err)
	}
	rows, err := rec.List(context.Background(), airuntime.Query{SceneCode: airuntime.SceneS01, JobID: imported.Body.Job_id})
	if err != nil || len(rows) != 1 || rows[0].Success || rows[0].Subclass != airuntime.SubclassText || rows[0].Tokens != 0 {
		t.Fatalf("ledger %+v %v", rows, err)
	}
	if hits != 0 {
		t.Fatal("missing key called a provider")
	}
	if got.Body.Failure_message != nil && (strings.Contains(*got.Body.Failure_message, "test-key") || strings.Contains(*got.Body.Failure_message, srv.URL)) {
		t.Fatalf("message %q", *got.Body.Failure_message)
	}
}

func TestMissingBaseURLDoesNotInstall(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("QLUTCH_LLM_API_KEY", "test-key")
	t.Setenv("ALIYUN_MODEL_API_KEY", "fallback-key")
	rt := airuntime.New()
	airuntime.InstallFromEnv(rt)
	if rt.Status() != "未配置提供方" {
		t.Fatalf("status %s", rt.Status())
	}
	h := newKnowledge(t)
	access, _, _ := h.register(t, "13800043102")
	note := []byte("没有根地址")
	pack, upload := uploadNote(t, h, access, note)
	imported := h.importJob(t, access, "prov-missing-root", pack, upload)
	rec := airuntime.NewMemoryRecorder()
	rt.SetRecorder(rec)
	h.know.SetRuntime(rt)
	if err := h.know.ProcessQueued(context.Background(), h.queue.Items()[0]); err != nil {
		t.Fatal(err)
	}
	got := h.job(t, access, imported.Body.Job_id)
	if got.Body.Status != qlutchapi.JobStatusFailedRetryable || got.Body.Map_ready {
		t.Fatalf("job %+v", got.Body)
	}
	rows, err := rec.List(context.Background(), airuntime.Query{SceneCode: airuntime.SceneS01, JobID: imported.Body.Job_id})
	if err != nil || len(rows) != 1 || rows[0].Success {
		t.Fatalf("ledger %+v %v", rows, err)
	}
}

func TestInstalledProviderRunsS01AndS02(t *testing.T) {
	clearProviderEnv(t)
	var mu sync.Mutex
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		mu.Lock()
		calls = append(calls, r.URL.Path+" "+string(raw))
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(string(raw), "题量") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"choices": []any{map[string]any{"message": map[string]any{"content": `{"questions":[{"stem":"预习一题"}]}`}}},
				"usage":   map[string]any{"total_tokens": 3},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []any{map[string]any{"message": map[string]any{"content": `{"catalogs":[{"title":"目录","chapters":[{"title":"章节","points":[{"title":"考点"}]}]}]}`}}},
			"usage":   map[string]any{"total_tokens": 4},
		})
	}))
	t.Cleanup(srv.Close)
	t.Setenv("QLUTCH_LLM_API_KEY", "test-key")
	t.Setenv("QLUTCH_LLM_BASE_URL", srv.URL)
	h := newKnowledge(t)
	access, _, _ := h.register(t, "13800043103")
	h.onboard(t, access)
	note := []byte("叶绿体笔记")
	pack, upload := uploadNote(t, h, access, note)
	imported := h.importJob(t, access, "prov-installed", pack, upload)
	rt := airuntime.New()
	rec := airuntime.NewMemoryRecorder()
	rt.SetRecorder(rec)
	airuntime.InstallFromEnv(rt)
	if rt.Status() != "ready" {
		t.Fatalf("status %s", rt.Status())
	}
	h.know.SetRuntime(rt)
	if _, err := rt.Call(context.Background(), airuntime.LedgerEntry{
		SceneCode: airuntime.SceneS01, Subclass: airuntime.SubclassText, Tier: airuntime.TierLight,
		TemplateVersion: "s01-map.v1", Fragments: [][]byte{note},
	}); err != nil {
		t.Fatal(err)
	}
	mid := h.job(t, access, imported.Body.Job_id)
	if mid.Body.Status != qlutchapi.JobStatusAccepted || mid.Body.Map_ready {
		t.Fatalf("adapter changed job %+v", mid.Body)
	}
	if ok, err := h.know.TestingMapEnterable(context.Background(), pack); err != nil || ok {
		t.Fatalf("adapter wrote a map %v %v", ok, err)
	}
	if err := h.know.ProcessQueued(context.Background(), h.queue.Items()[0]); err != nil {
		t.Fatal(err)
	}
	got := h.job(t, access, imported.Body.Job_id)
	if got.Body.Status != qlutchapi.JobStatusMapReady || !got.Body.Map_ready || !got.Body.Account.Has_first_pack || got.Body.Next_step != qlutchapi.NextStepREADYHOME {
		t.Fatalf("job %+v", got.Body)
	}
	if n, err := h.know.TestingPrefetchCount(context.Background(), imported.Body.Job_id); err != nil || n != 1 {
		t.Fatalf("prefetch %d %v", n, err)
	}
	s01, err := rec.List(context.Background(), airuntime.Query{SceneCode: airuntime.SceneS01, JobID: imported.Body.Job_id})
	if err != nil || len(s01) != 1 || !s01[0].Success || s01[0].Tokens != 4 || s01[0].Subclass != airuntime.SubclassText {
		t.Fatalf("s01 %+v %v", s01, err)
	}
	s02, err := rec.List(context.Background(), airuntime.Query{SceneCode: airuntime.SceneS02, JobID: imported.Body.Job_id})
	if err != nil || len(s02) != 1 || !s02[0].Success || s02[0].Tokens != 3 {
		t.Fatalf("s02 %+v %v", s02, err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(calls) < 3 {
		t.Fatalf("calls %d", len(calls))
	}
	for _, call := range calls {
		if !strings.Contains(call, "chat/completions") {
			t.Fatalf("path %s", call)
		}
	}
}

func TestLongAudioJobStaysRetryable(t *testing.T) {
	clearProviderEnv(t)
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		hits++
	}))
	t.Cleanup(srv.Close)
	t.Setenv("QLUTCH_LLM_API_KEY", "test-key")
	t.Setenv("QLUTCH_LLM_BASE_URL", srv.URL)
	h := newKnowledge(t)
	access, _, _ := h.register(t, "13800043104")
	body := flowWav(8, 301)
	pack := h.createPack(t, access).Body.Pack_id
	upload := h.upload(t, access, pack, "file", "audio/wav", int64(len(body)), "clip.wav")
	h.blobs.Put(strings.TrimPrefix(upload.Body.Upload_url, "http://127.0.0.1/upload/"), body)
	imported := h.importJob(t, access, "prov-long-audio", pack, upload.Body.Material_id)
	rt := airuntime.New()
	rec := airuntime.NewMemoryRecorder()
	rt.SetRecorder(rec)
	airuntime.InstallFromEnv(rt)
	h.know.SetRuntime(rt)
	if err := h.know.ProcessQueued(context.Background(), h.queue.Items()[0]); err != nil {
		t.Fatal(err)
	}
	got := h.job(t, access, imported.Body.Job_id)
	if got.Body.Status != qlutchapi.JobStatusFailedRetryable || got.Body.Map_ready || got.Body.Account.Has_first_pack {
		t.Fatalf("job %+v", got.Body)
	}
	if hits != 0 {
		t.Fatal("long audio was sent")
	}
	rows, err := rec.List(context.Background(), airuntime.Query{JobID: imported.Body.Job_id, SceneCode: airuntime.SceneS01})
	if err != nil || len(rows) != 1 || rows[0].Success || rows[0].Subclass != airuntime.SubclassASR || rows[0].Tokens != 0 {
		t.Fatalf("ledger %+v %v", rows, err)
	}
	if ok, err := h.know.TestingMapEnterable(context.Background(), pack); err != nil || ok {
		t.Fatalf("map %v %v", ok, err)
	}
}

func TestS02FailureDoesNotRewriteMapReady(t *testing.T) {
	clearProviderEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(string(raw), "题量") {
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{"message": "test-key down"}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []any{map[string]any{"message": map[string]any{"content": `{"catalogs":[{"title":"目录","chapters":[{"title":"章节","points":[{"title":"考点"}]}]}]}`}}},
			"usage":   map[string]any{"total_tokens": 2},
		})
	}))
	t.Cleanup(srv.Close)
	t.Setenv("QLUTCH_LLM_API_KEY", "test-key")
	t.Setenv("QLUTCH_LLM_BASE_URL", srv.URL)
	h := newKnowledge(t)
	access, _, _ := h.register(t, "13800043105")
	h.onboard(t, access)
	note := []byte("地图已经可以进入")
	pack, upload := uploadNote(t, h, access, note)
	imported := h.importJob(t, access, "prov-s02-fail", pack, upload)
	rt := airuntime.New()
	rec := airuntime.NewMemoryRecorder()
	rt.SetRecorder(rec)
	airuntime.InstallFromEnv(rt)
	h.know.SetRuntime(rt)
	if err := h.know.ProcessQueued(context.Background(), h.queue.Items()[0]); err != nil {
		t.Fatal(err)
	}
	got := h.job(t, access, imported.Body.Job_id)
	if got.Body.Status != qlutchapi.JobStatusMapReady || !got.Body.Map_ready || !got.Body.Account.Has_first_pack || got.Body.Next_step != qlutchapi.NextStepREADYHOME {
		t.Fatalf("job %+v", got.Body)
	}
	if got.Body.Failure_message != nil && strings.Contains(*got.Body.Failure_message, "test-key") {
		t.Fatalf("message %q", *got.Body.Failure_message)
	}
	if n, err := h.know.TestingPrefetchCount(context.Background(), imported.Body.Job_id); err != nil || n != 0 {
		t.Fatalf("prefetch %d %v", n, err)
	}
	s02, err := rec.List(context.Background(), airuntime.Query{SceneCode: airuntime.SceneS02, JobID: imported.Body.Job_id})
	if err != nil || len(s02) != 1 || s02[0].Success || s02[0].Tokens != 0 {
		t.Fatalf("s02 %+v %v", s02, err)
	}
	s01, err := rec.List(context.Background(), airuntime.Query{SceneCode: airuntime.SceneS01, JobID: imported.Body.Job_id})
	if err != nil || len(s01) != 1 || !s01[0].Success {
		t.Fatalf("s01 %+v %v", s01, err)
	}
}

func clearProviderEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"QLUTCH_LLM_API_KEY",
		"ALIYUN_MODEL_API_KEY",
		"QLUTCH_LLM_BASE_URL",
		"QLUTCH_LLM_MODEL_S01_OCR",
		"QLUTCH_LLM_MODEL_S01_ASR",
		"QLUTCH_LLM_MODEL_S01_TEXT",
		"QLUTCH_LLM_MODEL",
		"QLUTCH_LLM_REGION",
		"QLUTCH_LLM_PROVIDER",
		"DASHSCOPE_API_KEY",
	} {
		t.Setenv(name, "")
	}
}

func flowWav(sampleRate, seconds int) []byte {
	dataLen := sampleRate * seconds
	buf := make([]byte, 44+dataLen)
	copy(buf[0:], "RIFF")
	binary.LittleEndian.PutUint32(buf[4:], uint32(36+dataLen))
	copy(buf[8:], "WAVE")
	copy(buf[12:], "fmt ")
	binary.LittleEndian.PutUint32(buf[16:], 16)
	binary.LittleEndian.PutUint16(buf[20:], 1)
	binary.LittleEndian.PutUint16(buf[22:], 1)
	binary.LittleEndian.PutUint32(buf[24:], uint32(sampleRate))
	binary.LittleEndian.PutUint32(buf[28:], uint32(sampleRate))
	binary.LittleEndian.PutUint16(buf[32:], 1)
	binary.LittleEndian.PutUint16(buf[34:], 8)
	copy(buf[36:], "data")
	binary.LittleEndian.PutUint32(buf[40:], uint32(dataLen))
	return buf
}
