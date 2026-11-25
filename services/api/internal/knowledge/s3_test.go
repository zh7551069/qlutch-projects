package knowledge

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestQueryPresignMatchesAWSExample(t *testing.T) {
	// Documented Signature Version 4 sample. Not a live credential.
	const secret = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	signed, headers, err := signQueryURL(
		http.MethodGet,
		"https",
		"examplebucket.s3.amazonaws.com",
		"/test.txt",
		"",
		time.Date(2013, 5, 24, 0, 0, 0, 0, time.UTC),
		86400,
		"us-east-1",
		"AKIAIOSFODNN7EXAMPLE",
		secret,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(headers) != 0 {
		t.Fatalf("headers = %#v", headers)
	}
	if strings.Contains(signed, secret) {
		t.Fatal("secret leaked into the url")
	}
	// Canonical request is the documented virtual-hosted GET of /test.txt:
	// UNSIGNED-PAYLOAD, signed header host, and the query string below.
	const want = "https://examplebucket.s3.amazonaws.com/test.txt?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAIOSFODNN7EXAMPLE%2F20130524%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20130524T000000Z&X-Amz-Expires=86400&X-Amz-SignedHeaders=host&X-Amz-Signature=aeeed9bbccd4d02ee5c0109b86d86835f995330da4c265957d157751f604d404"
	if signed != want {
		t.Fatalf("signed =\n%s\nwant\n%s", signed, want)
	}
}

func TestS3PresignPutSignsContentType(t *testing.T) {
	const secret = "local-signing-secret"
	store := S3Store{
		Endpoint:  "http://127.0.0.1:4715",
		Region:    "us-east-1",
		Bucket:    "qlutch-local",
		AccessKey: "qlutch",
		Secret:    secret,
		Expires:   time.Duration(DefaultPresignSeconds) * time.Second,
		Now:       func() time.Time { return time.Date(2026, 11, 26, 0, 0, 0, 0, time.UTC) },
	}
	signed, headers, expires, err := store.PresignPut(context.Background(), "m/usr_abc/obj_def", "text/plain")
	if err != nil {
		t.Fatal(err)
	}
	if expires != DefaultPresignSeconds || headers["Content-Type"] != "text/plain" {
		t.Fatalf("expires %d headers %#v", expires, headers)
	}
	if !strings.Contains(signed, "X-Amz-SignedHeaders=content-type%3Bhost") || !strings.Contains(signed, "X-Amz-Expires=900") {
		t.Fatalf("url = %s", signed)
	}
	if strings.Contains(signed, secret) || strings.Contains(signed, "notes.txt") {
		t.Fatal("secret or filename leaked")
	}
}

func TestS3StatUsesSignedHead(t *testing.T) {
	const secret = "local-signing-secret"
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Errorf("method = %s", r.Method)
		}
		sawAuth = r.Header.Get("Authorization")
		if !strings.Contains(r.URL.Path, "/qlutch-local/m/abc") {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Length", "11")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	store := S3Store{
		Endpoint:  srv.URL,
		Region:    "us-east-1",
		Bucket:    "qlutch-local",
		AccessKey: "qlutch",
		Secret:    secret,
		Client:    srv.Client(),
		Now:       func() time.Time { return time.Date(2026, 11, 26, 0, 0, 0, 0, time.UTC) },
	}
	size, found, err := store.Stat(context.Background(), "m/abc")
	if err != nil || !found || size != 11 {
		t.Fatalf("stat = %d %v %v", size, found, err)
	}
	if sawAuth == "" || strings.Contains(sawAuth, secret) {
		t.Fatalf("authorization = %s", sawAuth)
	}

	missing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(missing.Close)
	store.Endpoint = missing.URL
	store.Client = missing.Client()
	size, found, err = store.Stat(context.Background(), "m/abc")
	if err != nil || found || size != 0 {
		t.Fatalf("missing = %d %v %v", size, found, err)
	}
}

func TestFirstPackWriteSites(t *testing.T) {
	body := mustRead(t, "service.go")
	if strings.Count(body, "CompleteFirstPack(") != 2 {
		t.Fatalf("CompleteFirstPack calls = %d", strings.Count(body, "CompleteFirstPack("))
	}
	importFn := body[strings.Index(body, "func (s *Service) Import("):strings.Index(body, "func (s *Service) Job(")]
	if strings.Contains(importFn, "statusMapReady") || strings.Contains(importFn, "CompleteFirstPack") {
		t.Fatal("import acceptance writes map_ready or has_first_pack")
	}
}

func mustRead(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
