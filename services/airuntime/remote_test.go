package airuntime

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestProviderSourceStaysInsideItsBoundary(t *testing.T) {
	for _, name := range []string{"install.go", "remote.go", "parse.go", "audio.go"} {
		body, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		text := string(body)
		for _, banned := range []string{
			"https://", "http://", "sk-", "map_ready", "failed_retryable", "FailJob",
			"has_first_pack", "CommitMap", "temperature", "unit_price", "UnitPrice",
			"github.com/openai/", "github.com/aliyun/", "github.com/sashabaranov/go-openai",
		} {
			if strings.Contains(text, banned) {
				t.Fatalf("%s contains %s", name, banned)
			}
		}
	}
}

func TestConfiguredProviderAnswersS02AndText(t *testing.T) {
	fake := &fakeVendor{handler: func(path string, body []byte, model string) (int, any) {
		if strings.Contains(path, "multimodal-generation") {
			return http.StatusInternalServerError, map[string]any{"code": "wrong-protocol"}
		}
		var payload map[string]json.RawMessage
		if err := json.Unmarshal(body, &payload); err != nil {
			return http.StatusBadRequest, map[string]any{}
		}
		if _, ok := payload["temperature"]; ok {
			return http.StatusBadRequest, map[string]any{"error": map[string]string{"message": "temperature was sent"}}
		}
		if strings.Contains(string(body), "题量") {
			if model != defaultTextPrimary {
				return http.StatusBadRequest, map[string]any{}
			}
			return http.StatusOK, chatResult(`{"questions":[{"stem":"预习一题"},{"stem":"预习二题"}]}`, 0, 4, 3)
		}
		return http.StatusOK, chatResult(sampleMapJSON(), 6, 0, 0)
	}}
	srv := fake.server(t)
	rt, rec := installedRuntime(t, srv.URL, "test-key", "")
	note := []byte("ALPHA-PARAGRAPH\n\nBETA-PARAGRAPH")
	out, err := rt.RunS01(context.Background(), S01Input{
		ContentType: "text/plain", Body: note, PackID: "p", JobID: "job-text",
	})
	if err != nil || !mapEnterable(out.Map) {
		t.Fatalf("s01 %+v %v", out, err)
	}
	calls := fake.snapshot()
	if len(calls) != 2 {
		t.Fatalf("calls %d", len(calls))
	}
	for _, call := range calls {
		if !strings.Contains(call.path, "chat/completions") || strings.Contains(call.path, "multimodal-generation") {
			t.Fatalf("path %s", call.path)
		}
		if call.model != defaultTextPrimary || call.auth != "Bearer test-key" {
			t.Fatal("text call was not the configured model")
		}
		if strings.Contains(call.auth, "test-key") && strings.Contains(string(call.body), "test-key") {
			t.Fatal("key was copied into the body")
		}
		if strings.Contains(string(call.body), "ALPHA-PARAGRAPH") && strings.Contains(string(call.body), "BETA-PARAGRAPH") {
			t.Fatal("fragments were merged")
		}
		if !strings.Contains(string(call.body), templateS01) {
			t.Fatal("template version was not sent")
		}
	}
	s02, err := rt.RunS02(context.Background(), S02Input{
		Nodes: []NodeRef{{ID: "n1", Title: "考点"}}, PackID: "p", JobID: "job-text",
	})
	if err != nil || len(s02.Questions) != prefetchPerNode {
		t.Fatalf("s02 %+v %v", s02, err)
	}
	s02calls := fake.snapshot()[2:]
	if len(s02calls) != 1 || len(s02calls[0].body) == 0 {
		t.Fatal("s02 was not sent")
	}
	if strings.Contains(string(s02calls[0].body), "ALPHA-PARAGRAPH") || !strings.Contains(string(s02calls[0].body), "考点") || !strings.Contains(string(s02calls[0].body), "题量") {
		t.Fatal("s02 did not send only the node title and question count")
	}
	rows, err := rec.List(context.Background(), Query{JobID: "job-text"})
	if err != nil || len(rows) != 2 || !rows[0].Success || rows[0].Tokens != 12 || rows[0].Subclass != SubclassText || !rows[1].Success || rows[1].Tokens != 7 || rows[1].Tier != TierLight {
		t.Fatalf("rows %+v %v", rows, err)
	}
	beforeCache := len(fake.snapshot())
	if _, err := rt.RunS01(context.Background(), S01Input{
		ContentType: "text/plain", Body: []byte("同一份材料"), PackID: "p", JobID: "job-cache",
	}); err != nil {
		t.Fatal(err)
	}
	if len(fake.snapshot()) != beforeCache+1 {
		t.Fatal("first use of a note did not call the provider")
	}
	cached, err := rt.RunS01(context.Background(), S01Input{
		ContentType: "text/plain", Body: []byte("同一份材料"), PackID: "p", JobID: "job-cache-2",
	})
	if err != nil || !cached.CacheHit || len(fake.snapshot()) != beforeCache+1 {
		t.Fatalf("cache hit=%v calls=%d err=%v", cached.CacheHit, len(fake.snapshot()), err)
	}
	hit := true
	cachedRows, err := rec.List(context.Background(), Query{JobID: "job-cache-2", CacheHit: true, CacheHitSet: hit})
	if err != nil || len(cachedRows) != 1 || cachedRows[0].Tokens != 0 || !cachedRows[0].Success {
		t.Fatalf("cache row %+v %v", cachedRows, err)
	}
}

func TestASRUsesMultimodalAndNotChatCompletions(t *testing.T) {
	audio := wavBytes(8, 1)
	fake := &fakeVendor{handler: func(path string, body []byte, model string) (int, any) {
		if strings.Contains(path, "chat/completions") {
			return http.StatusBadRequest, map[string]any{"error": map[string]string{"message": "audio reached chat completions"}}
		}
		if model != defaultASRPrimary || !strings.Contains(string(body), "data:audio/") || !strings.Contains(path, "multimodal-generation/generation") {
			return http.StatusBadRequest, map[string]any{}
		}
		return http.StatusOK, mmResult(sampleMapJSON(), 9)
	}}
	srv := fake.server(t)
	rt, rec := installedRuntime(t, srv.URL, "test-key", "")
	out, err := rt.RunS01(context.Background(), S01Input{
		ContentType: "audio/wav", Body: audio, PackID: "p", JobID: "job-asr",
	})
	if err != nil || !mapEnterable(out.Map) {
		t.Fatalf("asr %+v %v", out, err)
	}
	for _, call := range fake.snapshot() {
		if strings.Contains(call.path, "chat/completions") || call.model != defaultASRPrimary {
			t.Fatalf("asr path %s model %s", call.path, call.model)
		}
	}
	rows, err := rec.List(context.Background(), Query{JobID: "job-asr"})
	if err != nil || len(rows) != 1 || !rows[0].Success || rows[0].Subclass != SubclassASR || rows[0].Tier != TierRecognize || rows[0].Tokens != 9 {
		t.Fatalf("rows %+v %v", rows, err)
	}
}

func TestASRTranscriptIsStructuredWithoutSendingAudioToChat(t *testing.T) {
	audio := wavBytes(8, 1)
	fake := &fakeVendor{handler: func(path string, body []byte, model string) (int, any) {
		if strings.Contains(path, "multimodal-generation") {
			if model != defaultASRPrimary || strings.Contains(path, "chat/completions") {
				return http.StatusBadRequest, map[string]any{}
			}
			return http.StatusOK, mmResult("转写一句", 5)
		}
		if model != defaultTextPrimary || strings.Contains(string(body), "data:audio/") || strings.Contains(string(body), defaultASRPrimary) {
			return http.StatusBadRequest, map[string]any{}
		}
		if !strings.Contains(string(body), "转写一句") {
			return http.StatusBadRequest, map[string]any{}
		}
		return http.StatusOK, chatResult(sampleMapJSON(), 7, 0, 0)
	}}
	srv := fake.server(t)
	rt, rec := installedRuntime(t, srv.URL, "test-key", "")
	out, err := rt.RunS01(context.Background(), S01Input{
		ContentType: "audio/wav", Body: audio, PackID: "p", JobID: "job-asr-text",
	})
	if err != nil || !mapEnterable(out.Map) {
		t.Fatalf("asr %+v %v", out, err)
	}
	calls := fake.snapshot()
	if len(calls) != 2 || strings.Contains(calls[0].path, "chat/completions") || !strings.Contains(calls[1].path, "chat/completions") {
		t.Fatalf("calls %#v", callPaths(calls))
	}
	rows, err := rec.List(context.Background(), Query{JobID: "job-asr-text"})
	if err != nil || len(rows) != 1 || rows[0].Tokens != 12 || rows[0].Subclass != SubclassASR || rows[0].Tier != TierRecognize {
		t.Fatalf("rows %+v %v", rows, err)
	}
}

func TestLongAudioFailsWithoutACall(t *testing.T) {
	fake := &fakeVendor{handler: func(string, []byte, string) (int, any) {
		return http.StatusOK, mmResult(sampleMapJSON(), 1)
	}}
	srv := fake.server(t)
	rt, rec := installedRuntime(t, srv.URL, "test-key", "")
	_, err := rt.RunS01(context.Background(), S01Input{
		ContentType: "audio/wav", Body: wavBytes(8, 301), PackID: "p", JobID: "job-long",
	})
	if !errors.Is(err, errShortAudio) {
		t.Fatalf("err %v", err)
	}
	if len(fake.snapshot()) != 0 {
		t.Fatal("long audio called the provider")
	}
	rows, err := rec.List(context.Background(), Query{JobID: "job-long"})
	if err != nil || len(rows) != 1 || rows[0].Success || rows[0].Tokens != 0 || rows[0].Subclass != SubclassASR {
		t.Fatalf("rows %+v %v", rows, err)
	}
}

func TestOCRAndDocumentStayOnTheirProtocols(t *testing.T) {
	fake := &fakeVendor{handler: func(path string, body []byte, model string) (int, any) {
		if strings.Contains(path, "multimodal-generation") {
			return http.StatusBadRequest, map[string]any{}
		}
		if strings.Contains(string(body), "data:image/jpeg;base64,") {
			if model != defaultOCRPrimary {
				return http.StatusBadRequest, map[string]any{}
			}
			return http.StatusOK, chatResult(sampleMapJSON(), 2, 0, 0)
		}
		if model != defaultTextPrimary || !strings.Contains(string(body), "%PDF-1.7") {
			return http.StatusBadRequest, map[string]any{}
		}
		return http.StatusOK, chatResult(sampleMapJSON(), 1, 0, 0)
	}}
	srv := fake.server(t)
	rt, rec := installedRuntime(t, srv.URL, "test-key", "")
	image := []byte{0xFF, 0xD8, 0xFF, 0x00, 0x11}
	if _, err := rt.RunS01(context.Background(), S01Input{ContentType: "image/jpeg", Body: image, PackID: "p", JobID: "job-img"}); err != nil {
		t.Fatal(err)
	}
	pdf := append([]byte("%PDF-1.7\n"), bytesRepeat(20)...)
	if _, err := rt.RunS01(context.Background(), S01Input{ContentType: "application/pdf", Body: pdf, PackID: "p", JobID: "job-pdf"}); err != nil {
		t.Fatal(err)
	}
	imgRows, err := rec.List(context.Background(), Query{JobID: "job-img"})
	if err != nil || len(imgRows) != 1 || imgRows[0].Subclass != SubclassOCR || imgRows[0].Tier != TierRecognize {
		t.Fatalf("image %+v", imgRows)
	}
	pdfRows, err := rec.List(context.Background(), Query{JobID: "job-pdf"})
	if err != nil || len(pdfRows) != 1 || pdfRows[0].Subclass != SubclassText || pdfRows[0].Tier != TierStandard {
		t.Fatalf("pdf %+v", pdfRows)
	}
	noteRows, err := rec.List(context.Background(), Query{Subclass: SubclassOCR, JobID: "job-pdf"})
	if err != nil || len(noteRows) != 0 {
		t.Fatal("document was recorded as recognition")
	}
}

func TestFallbackDoesNotChangeTierOrRetryTransientFailure(t *testing.T) {
	fake := &fakeVendor{handler: func(path string, body []byte, model string) (int, any) {
		image := strings.Contains(string(body), "data:image/")
		if image && model == defaultOCRPrimary {
			return http.StatusNotFound, map[string]any{"error": map[string]string{"code": "model_not_found", "message": "model not found"}}
		}
		if image && model == defaultOCRFallback && strings.Contains(path, "chat/completions") {
			return http.StatusOK, chatResult(sampleMapJSON(), 4, 0, 0)
		}
		if !image && model == defaultTextPrimary {
			return http.StatusBadGateway, map[string]any{"error": map[string]string{"message": "test-key vendor-secret"}}
		}
		return http.StatusBadRequest, map[string]any{}
	}}
	srv := fake.server(t)
	rt, rec := installedRuntime(t, srv.URL, "test-key", "")
	if _, err := rt.RunS01(context.Background(), S01Input{ContentType: "image/jpeg", Body: []byte{0xFF, 0xD8, 0xFF, 0x01}, PackID: "p", JobID: "job-fall"}); err != nil {
		t.Fatal(err)
	}
	rows, err := rec.List(context.Background(), Query{JobID: "job-fall"})
	if err != nil || len(rows) != 1 || !rows[0].Success || rows[0].Tier != TierRecognize || rows[0].Subclass != SubclassOCR || rows[0].Tokens != 4 {
		t.Fatalf("fallback row %+v", rows)
	}
	_, err = rt.RunS01(context.Background(), S01Input{ContentType: "text/plain", Body: []byte("笔记"), PackID: "p", JobID: "job-down"})
	if err == nil || strings.Contains(err.Error(), "test-key") || strings.Contains(err.Error(), "vendor-secret") {
		t.Fatalf("err %v", err)
	}
	textCalls := 0
	for _, call := range fake.snapshot() {
		if strings.Contains(string(call.body), "data:image/") {
			continue
		}
		textCalls++
		if call.model == defaultTextFallback {
			t.Fatal("transient failure changed model")
		}
	}
	if textCalls != 1 {
		t.Fatalf("text calls %d", textCalls)
	}
	failed, err := rec.List(context.Background(), Query{JobID: "job-down"})
	if err != nil || len(failed) != 1 || failed[0].Success || failed[0].Tokens != 0 || failed[0].Tier != TierLight {
		t.Fatalf("failed %+v", failed)
	}
}

func TestAdapterDoesNotAcceptMaterialOnS02(t *testing.T) {
	fake := &fakeVendor{handler: func(string, []byte, string) (int, any) {
		return http.StatusOK, chatResult(`{"questions":[{"stem":"不该出现"}]}`, 1, 0, 0)
	}}
	srv := fake.server(t)
	rt, _ := installedRuntime(t, srv.URL, "test-key", "")
	_, err := rt.currentProvider().Complete(context.Background(), ProviderRequest{
		SceneCode: SceneS02, Subclass: SubclassPrefetch, Fragment: []byte("教材字节"),
		NodeTitle: "考点", QuestionLimit: 2, TemplateVersion: templateS02,
	})
	if !errors.Is(err, errMaterial) || len(fake.snapshot()) != 0 {
		t.Fatalf("err %v calls %d", err, len(fake.snapshot()))
	}
}

func TestParseUsageDoesNotInventTokens(t *testing.T) {
	_, tokens, err := parseChat([]byte(`{"choices":[{"message":{"content":"ok"}}],"usage":{"prompt_tokens":2,"completion_tokens":3}}`))
	if err != nil || tokens != 5 {
		t.Fatalf("tokens %d %v", tokens, err)
	}
	_, tokens, err = parseChat([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	if err != nil || tokens != 0 {
		t.Fatalf("missing usage %d %v", tokens, err)
	}
	_, _, err = parseChat([]byte(`{"error":{"code":"bad","message":"test-key vendor-secret"}}`))
	if err == nil || strings.Contains(err.Error(), "test-key") || strings.Contains(err.Error(), "vendor-secret") {
		t.Fatalf("err %v", err)
	}
}

func installedRuntime(t *testing.T, root, key, textModel string) (*Runtime, *MemoryRecorder) {
	t.Helper()
	rt, rec := readyRuntime(nil)
	installFrom(rt, func(name string) string {
		switch name {
		case envAPIKey:
			return key
		case envBaseURL:
			return root
		case envModelText:
			return textModel
		default:
			return ""
		}
	})
	if rt.Status() != "ready" {
		t.Fatalf("status %s", rt.Status())
	}
	return rt, rec
}

func sampleMapJSON() string {
	return `{"catalogs":[{"title":"目录","chapters":[{"title":"章节","points":[{"title":"考点"}]}]}]}`
}

func chatResult(content string, total, prompt, completion int) map[string]any {
	usage := map[string]any{}
	if total > 0 {
		usage["total_tokens"] = total
	}
	if prompt > 0 || completion > 0 {
		usage["prompt_tokens"] = prompt
		usage["completion_tokens"] = completion
	}
	return map[string]any{
		"choices": []any{map[string]any{"message": map[string]any{"role": "assistant", "content": content}}},
		"usage":   usage,
	}
}

func mmResult(content string, total int) map[string]any {
	return map[string]any{
		"output": map[string]any{
			"choices": []any{map[string]any{"message": map[string]any{
				"content": []any{map[string]any{"text": content}},
			}}},
		},
		"usage": map[string]any{"total_tokens": total},
	}
}

type capturedCall struct {
	path  string
	body  []byte
	auth  string
	model string
}

type fakeVendor struct {
	mu      sync.Mutex
	calls   []capturedCall
	handler func(path string, body []byte, model string) (int, any)
}

func (f *fakeVendor) server(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(io.LimitReader(r.Body, maxResponseBytes))
		if err != nil {
			http.Error(w, "read", http.StatusBadRequest)
			return
		}
		var probe struct {
			Model string `json:"model"`
		}
		_ = json.Unmarshal(raw, &probe)
		f.mu.Lock()
		f.calls = append(f.calls, capturedCall{path: r.URL.Path, body: append([]byte(nil), raw...), auth: r.Header.Get("Authorization"), model: probe.Model})
		f.mu.Unlock()
		status, payload := f.handler(r.URL.Path, raw, probe.Model)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(payload)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func (f *fakeVendor) snapshot() []capturedCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]capturedCall, len(f.calls))
	copy(out, f.calls)
	return out
}

func callPaths(calls []capturedCall) []string {
	out := make([]string, len(calls))
	for i, call := range calls {
		out[i] = call.path
	}
	return out
}

func bytesRepeat(n int) []byte {
	return []byte(strings.Repeat("a", n))
}
