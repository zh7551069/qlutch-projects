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
	newAt := strings.Index(text, "airuntime.New()")
	installAt := strings.Index(text, "airuntime.InstallFromEnv(")
	openAt := strings.Index(text, "importworker.Open")
	if newAt < 0 || installAt < newAt || openAt < installAt {
		t.Fatal("worker main does not install through the airuntime entry")
	}
	if !strings.Contains(text, "StagingImportFail: cfg.StagingImportFail()") {
		t.Fatal("worker main does not pass the staging import failure switch")
	}
	for _, banned := range []string{"SetProvider", "TestingSetJobStatus", "map_ready", "QLUTCH_FAKE", "openai", "sk-"} {
		if strings.Contains(text, banned) {
			t.Fatalf("worker main contains %s", banned)
		}
	}
}

func TestImportWorkerDoesNotInstallProvider(t *testing.T) {
	body, err := os.ReadFile("../api/importworker/importworker.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if strings.Contains(text, "SetProvider") || strings.Contains(text, "InstallFromEnv") {
		t.Fatal("importworker installs a provider")
	}
	if !strings.Contains(text, "SetRecorder") {
		t.Fatal("importworker no longer installs the usage ledger")
	}
	if !strings.Contains(text, "SetStagingImportFail") {
		t.Fatal("importworker does not apply the staging import failure switch")
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
		"QLUTCH_ENV",
	} {
		if !strings.Contains(text, name) {
			t.Fatalf("missing %s", name)
		}
	}
}
