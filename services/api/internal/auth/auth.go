package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// Domain identifies one of the two identity domains. Login channels are not wired.
type Domain string

const (
	// DomainLearner is the end-user identity domain shared by the three apps.
	DomainLearner Domain = "learner"
	// DomainInternal is the staff identity domain used by the admin console.
	DomainInternal Domain = "internal"
)

// ErrNotConfigured means this domain has no authenticator yet.
var ErrNotConfigured = errors.New("未配置")

// Authenticator is the empty seam for a future login channel.
type Authenticator interface {
	Authenticate(ctx context.Context, domain Domain, token string) error
}

// Unconfigured is the scaffold authenticator for both domains.
type Unconfigured struct{}

// Authenticate rejects every token. No login provider is called.
func (Unconfigured) Authenticate(context.Context, Domain, string) error {
	return ErrNotConfigured
}

// Require builds a placeholder middleware for one identity domain.
// Health probes do not use it. No protected route is mounted until a contract defines one.
func Require(domain Domain, authenticator Authenticator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearer(r.Header.Get("Authorization"))
		if token == "" || authenticator == nil {
			http.Error(w, "未认证", http.StatusUnauthorized)
			return
		}
		if err := authenticator.Authenticate(r.Context(), domain, token); err != nil {
			http.Error(w, "未配置", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func bearer(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}
