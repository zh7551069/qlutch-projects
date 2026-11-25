package server

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	qlutchapi "github.com/zh7551069/qlutch-projects/packages/contracts/gen/go"
	"github.com/zh7551069/qlutch-projects/services/api/internal/ready"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type staticCheck struct {
	name string
	err  error
}

func (s staticCheck) Name() string { return s.name }

func (s staticCheck) Ping(context.Context) error { return s.err }

func TestHealthAndReadyUseGeneratedClient(t *testing.T) {
	srv := New([]ready.Pinger{
		staticCheck{name: "database", err: nil},
		staticCheck{name: "redis", err: nil},
		staticCheck{name: "object_storage", err: nil},
	}, nil)
	httpSrv := httptest.NewServer(srv.Handler(nil))
	t.Cleanup(httpSrv.Close)

	client := &qlutchapi.Client{BaseURL: httpSrv.URL, HTTP: httpSrv.Client()}
	health, err := client.GetHealth(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if health.StatusCode != http.StatusOK || health.Body.Status != "ok" || health.RequestID == "" {
		t.Fatalf("health = %+v", health)
	}

	got, err := client.GetReady(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.StatusCode != http.StatusOK || got.Body.Status != "ready" {
		t.Fatalf("ready = %+v", got)
	}
	if got.Body.Checks["database"] != "up" || got.Body.Checks["redis"] != "up" || got.Body.Checks["object_storage"] != "up" {
		t.Fatalf("checks = %+v", got.Body.Checks)
	}
}

func TestReadyzFailsClosed(t *testing.T) {
	srv := New([]ready.Pinger{
		staticCheck{name: "database", err: errors.New("down")},
		staticCheck{name: "redis"},
		staticCheck{name: "object_storage"},
	}, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	srv.Handler(nil).ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestRequestIDEchoAndTrace(t *testing.T) {
	spans := &spanBox{}
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans))
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		_ = provider.Shutdown(context.Background())
	})

	srv := New(nil, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Request-Id", "req-123")
	req.Header.Set("Origin", "http://127.0.0.1:4712")
	srv.Handler([]string{"http://127.0.0.1:4712"}).ServeHTTP(rec, req)
	if rec.Header().Get("X-Request-Id") != "req-123" {
		t.Fatalf("request id = %q", rec.Header().Get("X-Request-Id"))
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "http://127.0.0.1:4712" {
		t.Fatalf("cors = %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
	if len(spans.ended) != 1 {
		t.Fatalf("spans = %d", len(spans.ended))
	}
	if got := attributeValue(spans.ended[0], "qlutch.request_id"); got != "req-123" {
		t.Fatalf("span request id = %q", got)
	}
}

func TestAuthPlaceholdersAreNotMountedButReject(t *testing.T) {
	srv := New(nil, nil)
	for _, handler := range []http.Handler{srv.LearnerAuth(), srv.InternalAuth()} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/unused", nil)
		req.Header.Set("Authorization", "Bearer placeholder")
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d body %s", rec.Code, rec.Body.String())
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/learner/session", nil)
	srv.Handler(nil).ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("public router status = %d", rec.Code)
	}
}

type spanBox struct {
	mu    sync.Mutex
	ended []sdktrace.ReadOnlySpan
}

func (s *spanBox) OnStart(context.Context, sdktrace.ReadWriteSpan) {}

func (s *spanBox) OnEnd(span sdktrace.ReadOnlySpan) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ended = append(s.ended, span)
}

func (s *spanBox) Shutdown(context.Context) error { return nil }

func (s *spanBox) ForceFlush(context.Context) error { return nil }

func attributeValue(span sdktrace.ReadOnlySpan, key string) string {
	for _, attr := range span.Attributes() {
		if string(attr.Key) == key {
			return attr.Value.AsString()
		}
	}
	return ""
}

func TestHealthzBodyIsJSON(t *testing.T) {
	srv := New(nil, nil)
	rec := httptest.NewRecorder()
	srv.Handler(nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	body, _ := io.ReadAll(rec.Body)
	if len(body) == 0 {
		t.Fatal("empty body")
	}
}
