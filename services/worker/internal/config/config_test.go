package config

import "testing"

func TestStagingImportFailIsExactEnv(t *testing.T) {
	t.Setenv("QLUTCH_ENV", "staging")
	if !Load().StagingImportFail() {
		t.Fatal("staging did not enable the import failure trigger")
	}
	t.Setenv("QLUTCH_ENV", " staging ")
	if !Load().StagingImportFail() {
		t.Fatal("trimmed staging did not enable the import failure trigger")
	}
	for _, name := range []string{"", "local", "production", "prod", "STAGING"} {
		t.Setenv("QLUTCH_ENV", name)
		if Load().StagingImportFail() {
			t.Fatalf("environment %q enabled the import failure trigger", name)
		}
	}
}
