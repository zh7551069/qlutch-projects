package knowledge

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

// Handler serves the learner knowledge routes.
type Handler struct {
	svc *Service
	log *slog.Logger
}

// Mount attaches knowledge routes to mux.
func Mount(mux *http.ServeMux, svc *Service) {
	h := &Handler{svc: svc, log: slog.Default()}
	mux.HandleFunc("POST /v1/knowledge/packs", h.createPack)
	mux.HandleFunc("POST /v1/knowledge/materials/upload-url", h.uploadURL)
	mux.HandleFunc("POST /v1/knowledge/imports", h.importJob)
	mux.HandleFunc("GET /v1/knowledge/jobs/{id}", h.job)
	mux.HandleFunc("POST /v1/knowledge/first-pack/skip", h.skip)
}

func (h *Handler) createPack(w http.ResponseWriter, r *http.Request) {
	raw, err := readBody(r)
	if err != nil {
		h.finish(w, r, Result{Err: codeErr("VALIDATION_MISSING_FIELD")})
		return
	}
	h.finish(w, r, h.svc.CreatePack(r.Context(), bearerToken(r), raw))
}

func (h *Handler) uploadURL(w http.ResponseWriter, r *http.Request) {
	raw, err := readBody(r)
	if err != nil {
		h.finish(w, r, Result{Err: codeErr("VALIDATION_MISSING_FIELD")})
		return
	}
	h.finish(w, r, h.svc.UploadURL(r.Context(), bearerToken(r), raw))
}

func (h *Handler) importJob(w http.ResponseWriter, r *http.Request) {
	raw, err := readBody(r)
	if err != nil {
		h.finish(w, r, Result{Err: codeErr("VALIDATION_MISSING_FIELD")})
		return
	}
	h.finish(w, r, h.svc.Import(r.Context(), bearerToken(r), r.Header.Get("Idempotency-Key"), raw))
}

func (h *Handler) job(w http.ResponseWriter, r *http.Request) {
	h.finish(w, r, h.svc.Job(r.Context(), bearerToken(r), r.PathValue("id")))
}

func (h *Handler) skip(w http.ResponseWriter, r *http.Request) {
	raw, err := readBody(r)
	if err != nil {
		h.finish(w, r, Result{Err: codeErr("VALIDATION_MISSING_FIELD")})
		return
	}
	h.finish(w, r, h.svc.Skip(r.Context(), bearerToken(r), r.Header.Get("Idempotency-Key"), raw))
}

func (h *Handler) finish(w http.ResponseWriter, r *http.Request, result Result) {
	id := currentID(w, r)
	resultCode := "ok"
	if result.Err != nil {
		resultCode = result.Err.Code
		writeError(w, id, result.Err)
	} else {
		writeJSON(w, result.Status, result.Body)
	}
	if h.log != nil {
		h.log.Info("knowledge", "path", r.URL.Path, "result", resultCode)
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
