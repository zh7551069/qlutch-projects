package airuntime

import (
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	envAPIKey      = "QLUTCH_LLM_API_KEY"
	envFallbackKey = "ALIYUN_MODEL_API_KEY"
	envBaseURL     = "QLUTCH_LLM_BASE_URL"
	envModelOCR    = "QLUTCH_LLM_MODEL_S01_OCR"
	envModelASR    = "QLUTCH_LLM_MODEL_S01_ASR"
	envModelText   = "QLUTCH_LLM_MODEL_S01_TEXT"
)

// Configuration defaults used only when the matching model variable is empty.
// A configured value replaces that primary. The alternate is used only when
// the primary is unavailable. These values do not select a tier.
var (
	defaultOCRPrimary   = "qwen3.5-ocr"
	defaultOCRFallback  = "qwen3.8-flash"
	defaultASRPrimary   = "qwen-audio-3.1-asr-flash"
	defaultASRFallback  = "qwen-audio-3.0-asr-flash"
	defaultTextPrimary  = "qwen3.8-flash"
	defaultTextFallback = "qwen3.7-plus"
)

// InstallFromEnv installs one provider when the process has both a key and a
// base URL. An empty key, including the fallback name, or an empty base URL
// leaves the runtime unconfigured. The base URL is not invented.
func InstallFromEnv(rt *Runtime) {
	installFrom(rt, os.Getenv)
}

func installFrom(rt *Runtime, getenv func(string) string) {
	if rt == nil || getenv == nil {
		return
	}
	key := strings.TrimSpace(getenv(envAPIKey))
	if key == "" {
		key = strings.TrimSpace(getenv(envFallbackKey))
	}
	root := strings.TrimSpace(getenv(envBaseURL))
	if key == "" || !usableRoot(root) {
		return
	}
	rt.SetProvider(&httpProvider{
		baseURL: root,
		key:     key,
		ocr: modelPair{
			primary:  modelOr(getenv, envModelOCR, defaultOCRPrimary),
			fallback: defaultOCRFallback,
		},
		asr: modelPair{
			primary:  modelOr(getenv, envModelASR, defaultASRPrimary),
			fallback: defaultASRFallback,
		},
		text: modelPair{
			primary:  modelOr(getenv, envModelText, defaultTextPrimary),
			fallback: defaultTextFallback,
		},
		http: &http.Client{CheckRedirect: refuseRedirect},
	})
}

func modelOr(getenv func(string) string, key, primary string) string {
	if value := strings.TrimSpace(getenv(key)); value != "" {
		return value
	}
	return primary
}

func usableRoot(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	return parsed.Host != ""
}

func refuseRedirect(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}
