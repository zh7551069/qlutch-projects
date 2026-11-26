package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	qlutchapi "github.com/zh7551069/qlutch-projects/packages/contracts/gen/go"

	"github.com/zh7551069/qlutch-projects/services/api/internal/auth"
	"github.com/zh7551069/qlutch-projects/services/api/internal/identity"
	"github.com/zh7551069/qlutch-projects/services/api/internal/knowledge"
	"github.com/zh7551069/qlutch-projects/services/api/internal/middleware"
	"github.com/zh7551069/qlutch-projects/services/api/internal/ready"
)

// Server is the public HTTP surface. Learner identity and knowledge routes are mounted.
// The internal-domain middleware stays unmounted until that contract exists.
type Server struct {
	checks   []ready.Pinger
	ident    *identity.Service
	know     *knowledge.Service
	learner  http.Handler
	internal http.Handler
	reveal   *identity.StagingReveal
}

// New builds the public handler. A nil identity service gets an in-memory OTP service.
// Knowledge state for that service stays in memory. Production uses NewWithKnowledge.
func New(checks []ready.Pinger, sessions *identity.Service) *Server {
	return newServer(checks, sessions, nil)
}

// NewWithKnowledge builds the public handler with a caller-owned knowledge service.
func NewWithKnowledge(checks []ready.Pinger, sessions *identity.Service, know *knowledge.Service) *Server {
	return newServer(checks, sessions, know)
}

func newServer(checks []ready.Pinger, sessions *identity.Service, know *knowledge.Service) *Server {
	if sessions == nil {
		var err error
		sessions, err = identity.New(identity.Options{})
		if err != nil {
			panic(err)
		}
	}
	if know == nil {
		var err error
		know, err = knowledge.New(knowledge.Options{Identity: sessions})
		if err != nil {
			panic(err)
		}
	}
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	return &Server{
		checks:   checks,
		ident:    sessions,
		know:     know,
		learner:  auth.Require(auth.DomainLearner, auth.Unconfigured{}, next),
		internal: auth.Require(auth.DomainInternal, auth.Unconfigured{}, next),
	}
}

// SetStagingReveal attaches a staging-only OTP lookup. Nil leaves the route unmounted.
func (s *Server) SetStagingReveal(reveal *identity.StagingReveal) {
	if s == nil {
		return
	}
	s.reveal = reveal
}

// Handler is the public router.
func (s *Server) Handler(origins []string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /readyz", s.readyz)
	identity.Mount(mux, s.ident)
	knowledge.Mount(mux, s.know)
	identity.MountStagingReveal(mux, s.ident, s.reveal)
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
