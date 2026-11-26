package ready

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestObjectStorePingReadsStatus(t *testing.T) {
	var sawAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization") != ""
		if strings.HasPrefix(r.Host, "qlutch-local.") || strings.HasPrefix(r.Host, "missing.") {
			t.Errorf("local endpoint used virtual host %s", r.Host)
		}
		if r.URL.Path == "/missing" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.URL.Path != "/qlutch-local" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	store := ObjectStore{
		Endpoint:  srv.URL,
		Region:    "us-east-1",
		Bucket:    "qlutch-local",
		AccessKey: "qlutch",
		Secret:    "qlutch-local-dev",
	}
	if err := store.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !sawAuth {
		t.Fatal("expected a signed request")
	}

	store.Bucket = "missing"
	if err := store.Ping(context.Background()); err == nil {
		t.Fatal("expected a down check")
	}
}

func TestObjectStoreHeadBucketAddressStyle(t *testing.T) {
	const secret = "example-signing-secret"
	const bucket = "example-bucket"
	cases := []struct {
		endpoint string
		host     string
		path     string
	}{
		{"https://oss-cn-hangzhou.aliyuncs.com", "example-bucket.oss-cn-hangzhou.aliyuncs.com", "/"},
		{"https://oss-cn-hangzhou.aliyuncs.com/", "example-bucket.oss-cn-hangzhou.aliyuncs.com", "/"},
		{"http://127.0.0.1:4715", "127.0.0.1:4715", "/" + bucket},
		{"http://localhost:4715", "localhost:4715", "/" + bucket},
		{"http://[::1]:4715", "[::1]:4715", "/" + bucket},
		{"http://object:9000", "object:9000", "/" + bucket},
		{"http://minio:9000", "minio:9000", "/" + bucket},
	}
	for _, tc := range cases {
		t.Run(tc.endpoint, func(t *testing.T) {
			var saw string
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.Method != http.MethodHead {
					t.Errorf("method = %s", r.Method)
				}
				auth := r.Header.Get("Authorization")
				if auth == "" || strings.Contains(auth, secret) {
					t.Errorf("authorization missing or secret leaked")
				}
				saw = r.Host + " " + r.URL.EscapedPath()
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader("")),
					Header:     make(http.Header),
					Request:    r,
				}, nil
			})}
			store := ObjectStore{
				Endpoint:  tc.endpoint,
				Region:    "us-east-1",
				Bucket:    bucket,
				AccessKey: "EXAMPLEACCESSKEY",
				Secret:    secret,
				Client:    client,
			}
			if err := store.Ping(context.Background()); err != nil {
				t.Fatal(err)
			}
			if store.Endpoint != tc.endpoint || strings.Contains(tc.endpoint, bucket) {
				t.Fatalf("endpoint = %s", store.Endpoint)
			}
			if saw != tc.host+" "+tc.path {
				t.Fatalf("request = %s", saw)
			}
			if strings.HasPrefix(tc.host, bucket+".") && strings.Contains(tc.path, bucket) {
				t.Fatalf("bucket in both host and path: %s", saw)
			}
		})
	}
}

func TestObjectStorePingHidesCredentials(t *testing.T) {
	const secret = "example-signing-secret"
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusForbidden,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})}
	store := ObjectStore{
		Endpoint:  "https://oss-cn-hangzhou.aliyuncs.com",
		Region:    "us-east-1",
		Bucket:    "example-bucket",
		AccessKey: "EXAMPLEACCESSKEY",
		Secret:    secret,
		Client:    client,
	}
	err := store.Ping(context.Background())
	if err == nil {
		t.Fatal("expected status error")
	}
	if strings.Contains(err.Error(), secret) || strings.Contains(err.Error(), "EXAMPLEACCESSKEY") || strings.Contains(err.Error(), "example-bucket") {
		t.Fatal("status error included a credential or bucket")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
