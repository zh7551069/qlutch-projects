package identity

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/zh7551069/qlutch-projects/services/api/internal/middleware"
)

// Handler serves the learner identity routes.
type Handler struct {
	svc *Service
	log *slog.Logger
}

// Mount attaches identity routes to mux.
func Mount(mux *http.ServeMux, svc *Service) {
	h := &Handler{svc: svc, log: slog.Default()}
	h.mount(mux)
}

// MountWithLog attaches identity routes and logs result codes without OTP or phone numbers.
func MountWithLog(mux *http.ServeMux, svc *Service, log *slog.Logger) {
	if log == nil {
		Mount(mux, svc)
		return
	}
	h := &Handler{svc: svc, log: log}
	h.mount(mux)
}

func (h *Handler) mount(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/identity/otp/send", h.guard(h.send))
	mux.HandleFunc("POST /v1/identity/otp/verify", h.guard(h.verify))
	mux.HandleFunc("POST /v1/identity/session/refresh", h.guard(h.refresh))
	mux.HandleFunc("POST /v1/identity/session/revoke", h.guard(h.revoke))
	mux.HandleFunc("GET /v1/identity/onboarding", h.guard(h.onboarding))
	mux.HandleFunc("PUT /v1/identity/onboarding", h.guard(h.writeOnboarding))
	mux.HandleFunc("GET /v1/identity/profile", h.guard(h.profile))
	mux.HandleFunc("PUT /v1/identity/profile", h.guard(h.writeProfile))
}

func (h *Handler) guard(next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(bearerToken(r), prefixInternal) {
			h.finish(w, r, Result{Err: codeErr("AUTH_WRONG_REALM")})
			return
		}
		next(w, r)
	}
}

func (h *Handler) send(w http.ResponseWriter, r *http.Request) {
	raw, err := readBody(r)
	if err != nil {
		h.finish(w, r, Result{Err: codeErr("VALIDATION_MISSING_FIELD")})
		return
	}
	h.finish(w, r, h.svc.Send(r.Context(), r.Header.Get("Idempotency-Key"), raw, r.RemoteAddr))
}

func (h *Handler) verify(w http.ResponseWriter, r *http.Request) {
	raw, err := readBody(r)
	if err != nil {
		h.finish(w, r, Result{Err: codeErr("VALIDATION_MISSING_FIELD")})
		return
	}
	// Authorization is ignored. A stale access token must not affect verify.
	h.finish(w, r, h.svc.Verify(raw, r.RemoteAddr))
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	raw, err := readBody(r)
	if err != nil {
		h.finish(w, r, Result{Err: codeErr("VALIDATION_MISSING_FIELD")})
		return
	}
	// Refresh uses the refresh token only. An expired access token is not consulted.
	h.finish(w, r, h.svc.Refresh(raw))
}

func (h *Handler) revoke(w http.ResponseWriter, r *http.Request) {
	raw, err := readBody(r)
	if err != nil {
		h.finish(w, r, Result{Err: codeErr("VALIDATION_MISSING_FIELD")})
		return
	}
	h.finish(w, r, h.svc.Revoke(bearerToken(r), raw))
}

func (h *Handler) onboarding(w http.ResponseWriter, r *http.Request) {
	h.finish(w, r, h.svc.Onboarding(bearerToken(r)))
}

func (h *Handler) writeOnboarding(w http.ResponseWriter, r *http.Request) {
	raw, err := readBody(r)
	if err != nil {
		h.finish(w, r, Result{Err: codeErr("VALIDATION_MISSING_FIELD")})
		return
	}
	h.finish(w, r, h.svc.WriteOnboarding(bearerToken(r), raw))
}

func (h *Handler) profile(w http.ResponseWriter, r *http.Request) {
	h.finish(w, r, h.svc.Profile(bearerToken(r)))
}

func (h *Handler) writeProfile(w http.ResponseWriter, r *http.Request) {
	raw, err := readBody(r)
	if err != nil {
		h.finish(w, r, Result{Err: codeErr("VALIDATION_MISSING_FIELD")})
		return
	}
	h.finish(w, r, h.svc.UpdateProfile(bearerToken(r), raw))
}

func (h *Handler) finish(w http.ResponseWriter, r *http.Request, result Result) {
	id := currentID(w, r)
	resultCode := "ok"
	if result.Err != nil {
		resultCode = result.Err.Code
		writeError(w, id, result.Err)
	} else if result.Status == http.StatusNoContent {
		resultCode = "no_content"
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusNoContent)
	} else {
		writeJSON(w, result.Status, result.Body)
	}
	if h.log != nil {
		h.log.Info("identity", "path", r.URL.Path, "result", resultCode)
	}
}

func writeError(w http.ResponseWriter, requestID string, err *Error) {
	payload := struct {
		Error struct {
			Code    string         `json:"code"`
			Message string         `json:"message"`
			Details map[string]any `json:"details,omitempty"`
		} `json:"error"`
		RequestID string `json:"request_id"`
	}{}
	payload.Error.Code = err.Code
	payload.Error.Message = err.Message
	payload.Error.Details = err.Details
	payload.RequestID = requestID
	raw, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		raw = []byte(`{"error":{"code":"INTERNAL","message":"服务暂时不可用"},"request_id":` + jsonString(requestID) + `}`)
		writeJSON(w, http.StatusInternalServerError, raw)
		return
	}
	writeJSON(w, statusOf(err.Code), raw)
}

func writeJSON(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func readBody(r *http.Request) ([]byte, error) {
	defer func() { _ = r.Body.Close() }()
	return io.ReadAll(io.LimitReader(r.Body, maxBodyBytes+1))
}

func bearerToken(r *http.Request) string {
	const scheme = "Bearer "
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, scheme) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, scheme))
}

func currentID(w http.ResponseWriter, r *http.Request) string {
	if id := w.Header().Get("X-Request-Id"); id != "" {
		return id
	}
	if id, ok := middleware.RequestIDFrom(r.Context()); ok {
		w.Header().Set("X-Request-Id", id)
		return id
	}
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		w.Header().Set("X-Request-Id", "unknown")
		return "unknown"
	}
	id := hex.EncodeToString(buf)
	w.Header().Set("X-Request-Id", id)
	return id
}

func jsonString(value string) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return `""`
	}
	return string(raw)
}
