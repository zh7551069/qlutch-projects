package knowledge

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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
		t.Fatal("presign query missing content-type or expiry")
	}
	parsed, err := url.Parse(signed)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Host != "127.0.0.1:4715" || parsed.Path != "/qlutch-local/m/usr_abc/obj_def" {
		t.Fatalf("host=%s path=%s", parsed.Host, parsed.Path)
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
		if strings.HasPrefix(r.Host, "qlutch-local.") {
			t.Errorf("local endpoint used virtual host %s", r.Host)
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

func TestObjectURLAddressStyle(t *testing.T) {
	const bucket = "example-bucket"
	const key = "m/usr_abc/obj_def"
	cases := []struct {
		endpoint string
		host     string
		path     string
	}{
		{"https://oss-cn-hangzhou.aliyuncs.com", "example-bucket.oss-cn-hangzhou.aliyuncs.com", "/" + key},
		{"https://oss-cn-hangzhou.aliyuncs.com/", "example-bucket.oss-cn-hangzhou.aliyuncs.com", "/" + key},
		{"https://storage.example.com:8443", "example-bucket.storage.example.com:8443", "/" + key},
		{"http://127.0.0.1:4715", "127.0.0.1:4715", "/" + bucket + "/" + key},
		{"http://127.0.0.1:4715/", "127.0.0.1:4715", "/" + bucket + "/" + key},
		{"http://localhost:4715", "localhost:4715", "/" + bucket + "/" + key},
		{"http://[::1]:4715", "[::1]:4715", "/" + bucket + "/" + key},
		{"http://10.1.2.3:9000", "10.1.2.3:9000", "/" + bucket + "/" + key},
		{"http://object:9000", "object:9000", "/" + bucket + "/" + key},
		{"http://minio:9000", "minio:9000", "/" + bucket + "/" + key},
	}
	for _, tc := range cases {
		t.Run(tc.endpoint, func(t *testing.T) {
			store := S3Store{Endpoint: tc.endpoint, Bucket: bucket}
			got, err := store.objectURL(key)
			if err != nil {
				t.Fatal(err)
			}
			if store.Endpoint != tc.endpoint {
				t.Fatalf("endpoint rewritten to %s", store.Endpoint)
			}
			if strings.Contains(tc.endpoint, bucket) {
				t.Fatal("fixture wrote the bucket into the endpoint")
			}
			if got.Host != tc.host || got.EscapedPath() != tc.path {
				t.Fatalf("host=%s path=%s", got.Host, got.EscapedPath())
			}
			inPath := strings.Contains(got.EscapedPath(), "/"+bucket+"/")
			onHost := strings.HasPrefix(got.Host, bucket+".")
			if inPath == onHost {
				t.Fatalf("bucket placement host=%s path=%s", got.Host, got.EscapedPath())
			}
		})
	}
}

func TestPresignPutAddressStyle(t *testing.T) {
	const secret = "example-signing-secret"
	const bucket = "example-bucket"
	const key = "m/usr_abc/obj_def"
	cases := []struct {
		endpoint string
		host     string
		path     string
	}{
		{"https://oss-cn-hangzhou.aliyuncs.com", "example-bucket.oss-cn-hangzhou.aliyuncs.com", "/" + key},
		{"http://127.0.0.1:4715", "127.0.0.1:4715", "/" + bucket + "/" + key},
		{"http://localhost:9000", "localhost:9000", "/" + bucket + "/" + key},
		{"http://object:9000", "object:9000", "/" + bucket + "/" + key},
	}
	for _, tc := range cases {
		t.Run(tc.endpoint, func(t *testing.T) {
			store := S3Store{
				Endpoint:  tc.endpoint,
				Region:    "us-east-1",
				Bucket:    bucket,
				AccessKey: "EXAMPLEACCESSKEY",
				Secret:    secret,
				Now:       func() time.Time { return time.Date(2026, 11, 26, 0, 0, 0, 0, time.UTC) },
			}
			signed, headers, _, err := store.PresignPut(context.Background(), key, "text/plain")
			if err != nil {
				t.Fatal(err)
			}
			if store.Endpoint != tc.endpoint {
				t.Fatalf("endpoint rewritten to %s", store.Endpoint)
			}
			parsed, err := url.Parse(signed)
			if err != nil {
				t.Fatal(err)
			}
			if parsed.Host != tc.host || parsed.Path != tc.path {
				t.Fatalf("host=%s path=%s", parsed.Host, parsed.Path)
			}
			if strings.Contains(parsed.Path, "/"+bucket+"/") && strings.HasPrefix(parsed.Host, bucket+".") {
				t.Fatalf("bucket in both host and path: host=%s path=%s", parsed.Host, parsed.Path)
			}
			if parsed.Host == "oss-cn-hangzhou.aliyuncs.com" || strings.Contains(signed, "oss-cn-hangzhou.aliyuncs.com/"+bucket+"/") {
				t.Fatal("presign stayed path-style on the regional host")
			}
			if headers["Content-Type"] != "text/plain" {
				t.Fatalf("headers missing content type")
			}
			sig := parsed.Query().Get("X-Amz-Signature")
			if parsed.Query().Get("X-Amz-Algorithm") != "AWS4-HMAC-SHA256" || len(sig) != 64 {
				t.Fatal("presign query missing algorithm or signature")
			}
			if strings.Contains(signed, secret) {
				t.Fatal("secret leaked into the url")
			}
		})
	}
}

func TestStatAndReadFollowAddressStyle(t *testing.T) {
	const secret = "example-signing-secret"
	const bucket = "example-bucket"
	const key = "m/usr_abc/obj_def"
	cases := []struct {
		endpoint string
		host     string
		path     string
	}{
		{"https://oss-cn-hangzhou.aliyuncs.com", "example-bucket.oss-cn-hangzhou.aliyuncs.com", "/" + key},
		{"http://127.0.0.1:4715", "127.0.0.1:4715", "/" + bucket + "/" + key},
		{"http://localhost:4715", "localhost:4715", "/" + bucket + "/" + key},
		{"http://[::1]:4715", "[::1]:4715", "/" + bucket + "/" + key},
		{"http://object:9000", "object:9000", "/" + bucket + "/" + key},
	}
	for _, tc := range cases {
		t.Run(tc.endpoint, func(t *testing.T) {
			var saw []string
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.Header.Get("Authorization") == "" || strings.Contains(r.Header.Get("Authorization"), secret) {
					t.Errorf("authorization missing or secret leaked")
				}
				saw = append(saw, r.Method+" "+r.Host+" "+r.URL.EscapedPath())
				return &http.Response{
					StatusCode:    http.StatusOK,
					ContentLength: 3,
					Body:          io.NopCloser(strings.NewReader("abc")),
					Header:        make(http.Header),
					Request:       r,
				}, nil
			})}
			store := S3Store{
				Endpoint:  tc.endpoint,
				Region:    "us-east-1",
				Bucket:    bucket,
				AccessKey: "EXAMPLEACCESSKEY",
				Secret:    secret,
				Client:    client,
				Now:       func() time.Time { return time.Date(2026, 11, 26, 0, 0, 0, 0, time.UTC) },
			}
			if _, _, err := store.Stat(context.Background(), key); err != nil {
				t.Fatal(err)
			}
			body, found, err := store.Read(context.Background(), key)
			if err != nil || !found || string(body) != "abc" {
				t.Fatalf("read found=%v err=%v len=%d", found, err, len(body))
			}
			want := tc.host + " " + tc.path
			if len(saw) != 2 || saw[0] != "HEAD "+want || saw[1] != "GET "+want {
				t.Fatalf("requests = %#v", saw)
			}
			if store.Endpoint != tc.endpoint {
				t.Fatalf("endpoint rewritten to %s", store.Endpoint)
			}
		})
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
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
