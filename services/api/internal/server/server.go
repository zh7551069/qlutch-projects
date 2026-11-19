package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	qlutchapi "github.com/zh7551069/qlutch-projects/packages/contracts/gen/go"

	"github.com/zh7551069/qlutch-projects/services/api/internal/auth"
	"github.com/zh7551069/qlutch-projects/services/api/internal/middleware"
	"github.com/zh7551069/qlutch-projects/services/api/internal/ready"
)

// Server is the probe HTTP surface. Auth middlewares for both identity domains
// are constructed here and are not mounted: no login route exists yet.
type Server struct {
	checks   []ready.Pinger
	learner  http.Handler
	internal http.Handler
}

// New builds the public handler and the two unmounted auth placeholders.
func New(checks []ready.Pinger) *Server {
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	return &Server{
		checks:   checks,
		learner:  auth.Require(auth.DomainLearner, auth.Unconfigured{}, next),
		internal: auth.Require(auth.DomainInternal, auth.Unconfigured{}, next),
	}
}

// Handler is the public probe router.
func (s *Server) Handler(origins []string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /readyz", s.readyz)
	return middleware.RequestID(middleware.Tracing(middleware.CORS(origins, mux)))
}

// LearnerAuth is the unmounted learner-domain placeholder.
func (s *Server) LearnerAuth() http.Handler { return s.learner }

// InternalAuth is the unmounted internal-domain placeholder.
func (s *Server) InternalAuth() http.Handler { return s.internal }

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, qlutchapi.HealthStatus{Status: "ok"})
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	checks := make(map[string]string, len(s.checks))
	readyNow := true
	for _, check := range s.checks {
		if err := check.Ping(ctx); err != nil {
			readyNow = false
			checks[check.Name()] = "down"
			continue
		}
		checks[check.Name()] = "up"
	}
	status := "ready"
	code := http.StatusOK
	if !readyNow {
		status = "not_ready"
		code = http.StatusServiceUnavailable
	}
	writeJSON(w, code, qlutchapi.ReadyStatus{Status: status, Checks: checks})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
