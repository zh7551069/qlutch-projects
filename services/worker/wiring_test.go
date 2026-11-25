package sdkguard

import (
	"os"
	"strings"
	"testing"
)

func TestProductionWorkerDoesNotShortcutMapReady(t *testing.T) {
	body, err := os.ReadFile("cmd/worker/main.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "airuntime.New()") || !strings.Contains(text, "importworker.Open") {
		t.Fatal("worker main does not use the airuntime entry")
	}
	for _, banned := range []string{"SetProvider", "TestingSetJobStatus", "map_ready", "QLUTCH_FAKE", "openai", "sk-"} {
		if strings.Contains(text, banned) {
			t.Fatalf("worker main contains %s", banned)
		}
	}
}

func TestWorkerConfigKeepsExistingNames(t *testing.T) {
	body, err := os.ReadFile("internal/config/config.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, name := range []string{
		"QLUTCH_WORKER_QUEUE",
		"QLUTCH_WORKER_HEARTBEAT",
		"QLUTCH_S3_ENDPOINT",
		"QLUTCH_S3_REGION",
		"QLUTCH_S3_BUCKET",
		"QLUTCH_S3_ACCESS_KEY",
		"QLUTCH_S3_SECRET_KEY",
		"QLUTCH_DATABASE_URL",
	} {
		if !strings.Contains(text, name) {
			t.Fatalf("missing %s", name)
		}
	}
}
