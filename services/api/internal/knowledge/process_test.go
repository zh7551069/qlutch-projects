package knowledge

import (
	"os"
	"strings"
	"testing"
)

func TestWorkerPathIsNotAnHTTPRoute(t *testing.T) {
	httpSrc := mustRead(t, "http.go")
	for _, banned := range []string{"ProcessImport", "ProcessQueued", "TestingSetJobStatus", "RunS01", "RunS02"} {
		if strings.Contains(httpSrc, banned) {
			t.Fatalf("http.go contains %s", banned)
		}
	}
	processSrc := mustRead(t, "process.go")
	if strings.Contains(processSrc, "TestingSetJobStatus") {
		t.Fatal("production import path uses the test fixture")
	}
	routes := strings.Count(httpSrc, "mux.HandleFunc(")
	if routes != 5 {
		t.Fatalf("routes = %d", routes)
	}
}

func TestImportAcceptanceSourceDoesNotRunS01(t *testing.T) {
	body := mustRead(t, "service.go")
	start := strings.Index(body, "func (s *Service) Import(")
	end := strings.Index(body, "func (s *Service) Job(")
	if start < 0 || end < start {
		t.Fatal("import function was not found")
	}
	importFn := body[start:end]
	for _, banned := range []string{"RunS01", "RunS02", "statusMapReady", "CompleteFirstPack", "map_ready"} {
		if strings.Contains(importFn, banned) {
			t.Fatalf("import contains %s", banned)
		}
	}
}

func TestKnowledgeDoesNotCallProviderHTTP(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		body := mustRead(t, entry.Name())
		for _, banned := range []string{"chat/completions", "multimodal-generation", "InstallFromEnv", "QLUTCH_LLM_API_KEY"} {
			if strings.Contains(body, banned) {
				t.Fatalf("%s contains %s", entry.Name(), banned)
			}
		}
	}
}

func TestNoProviderSDKInWorkerPath(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		body := mustRead(t, entry.Name())
		for _, banned := range []string{"github.com/sashabaranov/go-openai", "github.com/openai/", "github.com/anthropics/"} {
			if strings.Contains(body, banned) {
				t.Fatalf("%s imports %s", entry.Name(), banned)
			}
		}
	}
}
