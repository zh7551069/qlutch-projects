package ready

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestObjectStorePingReadsStatus(t *testing.T) {
	var sawAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization") != ""
		if r.URL.Path == "/missing" {
			w.WriteHeader(http.StatusNotFound)
			return
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
