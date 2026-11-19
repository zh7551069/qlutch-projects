package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireBothDomainsStayUnconfigured(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next should not run")
	})
	for _, domain := range []Domain{DomainLearner, DomainInternal} {
		handler := Require(domain, Unconfigured{}, next)
		missing := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, missing)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s missing token status = %d", domain, rec.Code)
		}

		present := httptest.NewRequest(http.MethodGet, "/", nil)
		present.Header.Set("Authorization", "Bearer placeholder")
		rec = httptest.NewRecorder()
		handler.ServeHTTP(rec, present)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s configured token status = %d", domain, rec.Code)
		}
		if !stringsContains(rec.Body.String(), "未配置") {
			t.Fatalf("%s body = %q", domain, rec.Body.String())
		}
	}
}

func stringsContains(value, part string) bool {
	return len(value) >= len(part) && (value == part || len(part) == 0 || contains(value, part))
}

func contains(value, part string) bool {
	for i := 0; i+len(part) <= len(value); i++ {
		if value[i:i+len(part)] == part {
			return true
		}
	}
	return false
}
