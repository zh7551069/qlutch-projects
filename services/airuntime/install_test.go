package airuntime

import (
	"os"
	"strings"
	"testing"
)

func TestInstallSourceUsesLockedNamesOnly(t *testing.T) {
	body, err := os.ReadFile("install.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, name := range []string{
		"QLUTCH_LLM_API_KEY",
		"QLUTCH_LLM_BASE_URL",
		"QLUTCH_LLM_MODEL_S01_OCR",
		"QLUTCH_LLM_MODEL_S01_ASR",
		"QLUTCH_LLM_MODEL_S01_TEXT",
		"ALIYUN_MODEL_API_KEY",
	} {
		if !strings.Contains(text, name) {
			t.Fatalf("missing %s", name)
		}
	}
	for _, banned := range []string{
		"DASHSCOPE_API_KEY",
		"QLUTCH_LLM_REGION",
		"QLUTCH_LLM_PROVIDER",
		`"QLUTCH_LLM_MODEL"`,
		"https://",
		"http://",
		"sk-",
	} {
		if strings.Contains(text, banned) {
			t.Fatalf("install.go contains %s", banned)
		}
	}
}

func TestMissingKeyDoesNotInstall(t *testing.T) {
	rt := New()
	installFrom(rt, func(key string) string {
		switch key {
		case envAPIKey, envFallbackKey:
			return ""
		case envBaseURL:
			return "http://127.0.0.1:9"
		default:
			return ""
		}
	})
	if rt.Status() != ErrProviderNotConfigured.Error() || rt.currentProvider() != nil {
		t.Fatalf("status %s provider installed", rt.Status())
	}
}

func TestMissingBaseURLDoesNotInstallOrGuess(t *testing.T) {
	for _, root := range []string{"", "   ", "localhost", "not-a-root"} {
		rt := New()
		installFrom(rt, func(key string) string {
			switch key {
			case envAPIKey:
				return "test-key"
			case envFallbackKey:
				t.Fatalf("fallback key read while primary key was present for %q", root)
			case envBaseURL:
				return root
			default:
				return ""
			}
			return ""
		})
		if rt.Status() != ErrProviderNotConfigured.Error() || rt.currentProvider() != nil {
			t.Fatalf("root %q installed a provider, status %s", root, rt.Status())
		}
	}
}

func TestPrimaryKeySkipsFallbackRead(t *testing.T) {
	rt := New()
	installFrom(rt, func(key string) string {
		switch key {
		case envAPIKey:
			return "test-key"
		case envFallbackKey:
			t.Fatal("fallback key was read")
		case envBaseURL:
			return "http://127.0.0.1:9"
		default:
			return ""
		}
		return ""
	})
	if rt.Status() != "ready" {
		t.Fatalf("status %s", rt.Status())
	}
}

func TestFallbackKeyIsReadOnly(t *testing.T) {
	t.Setenv(envAPIKey, "")
	t.Setenv(envFallbackKey, "fallback-key")
	t.Setenv(envBaseURL, "http://127.0.0.1:9")
	rt := New()
	InstallFromEnv(rt)
	if os.Getenv(envAPIKey) != "" {
		t.Fatal("fallback key was copied into the primary name")
	}
	if rt.Status() != "ready" {
		t.Fatalf("status %s", rt.Status())
	}
	provider, ok := rt.currentProvider().(*httpProvider)
	if !ok || provider.key != "fallback-key" {
		t.Fatal("fallback key was not used for the request header")
	}
}

func TestSingleModelNameDoesNotOverrideSubtypes(t *testing.T) {
	var asked []string
	rt := New()
	installFrom(rt, func(key string) string {
		asked = append(asked, key)
		switch key {
		case envAPIKey:
			return "test-key"
		case envBaseURL:
			return "http://127.0.0.1:9"
		case "QLUTCH_LLM_MODEL":
			return "should-not-be-read"
		default:
			return ""
		}
	})
	provider := rt.currentProvider().(*httpProvider)
	if provider.ocr.primary != defaultOCRPrimary || provider.asr.primary != defaultASRPrimary || provider.text.primary != defaultTextPrimary {
		t.Fatalf("models ocr=%s asr=%s text=%s", provider.ocr.primary, provider.asr.primary, provider.text.primary)
	}
	if provider.ocr.fallback != defaultOCRFallback || provider.asr.fallback != defaultASRFallback || provider.text.fallback != defaultTextFallback {
		t.Fatal("fallback defaults changed")
	}
	for _, key := range asked {
		if key == "QLUTCH_LLM_MODEL" || key == "QLUTCH_LLM_REGION" || key == "DASHSCOPE_API_KEY" {
			t.Fatalf("read %s", key)
		}
	}
}

func TestModelEnvReplacesOnlyItsPrimary(t *testing.T) {
	rt := New()
	installFrom(rt, func(key string) string {
		switch key {
		case envAPIKey:
			return "test-key"
		case envBaseURL:
			return "http://127.0.0.1:9"
		case envModelText:
			return "custom-text-model"
		default:
			return ""
		}
	})
	provider := rt.currentProvider().(*httpProvider)
	if provider.text.primary != "custom-text-model" || provider.text.fallback != defaultTextFallback {
		t.Fatalf("text primary %s fallback %s", provider.text.primary, provider.text.fallback)
	}
	if provider.ocr.primary != defaultOCRPrimary || provider.asr.primary != defaultASRPrimary {
		t.Fatal("one model variable changed another subtype")
	}
}
