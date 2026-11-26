package identity

import (
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

const stagingRevealPath = "POST /internal/staging/otp/reveal"

// StagingReveal is the in-process gate for the staging OTP lookup.
// It is not part of the learner contract. A nil value mounts nothing.
type StagingReveal struct {
	token string
}

// NewStagingReveal returns a gate only when the process environment name is
// exactly "staging" and token is non-empty. Every other name, including
// production and local, returns nil. The token is not logged.
func NewStagingReveal(envName, token string) *StagingReveal {
	if strings.TrimSpace(envName) != "staging" {
		return nil
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}
	return &StagingReveal{token: token}
}

func (g *StagingReveal) match(got string) bool {
	if g == nil || g.token == "" {
		return false
	}
	want := []byte(g.token)
	have := []byte(strings.TrimSpace(got))
	sameLen := len(have) == len(want)
	if !sameLen {
		have = want
	}
	return subtle.ConstantTimeCompare(have, want) == 1 && sameLen
}

// MountStagingReveal registers the lookup route only when reveal is non-nil
// and the service is retaining codes. Other processes keep the learner routes only.
func MountStagingReveal(mux *http.ServeMux, svc *Service, reveal *StagingReveal) {
	if mux == nil || svc == nil || reveal == nil || !svc.retainCode {
		return
	}
	mux.HandleFunc(stagingRevealPath, func(w http.ResponseWriter, r *http.Request) {
		raw, err := readBody(r)
		if err != nil || !reveal.match(bearerToken(r)) {
			writeStagingMiss(w)
			return
		}
		code, ok := svc.revealCode(raw)
		if !ok {
			writeStagingMiss(w)
			return
		}
		payload, mErr := json.Marshal(struct {
			Code string `json:"code"`
		}{Code: code})
		if mErr != nil {
			writeStagingMiss(w)
			return
		}
		slog.Default().Info("staging otp reveal", "result", "ok")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
	})
}

func writeStagingMiss(w http.ResponseWriter) {
	slog.Default().Info("staging otp reveal", "result", "miss")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(`{"error":{"code":"NOT_FOUND"}}`))
}
