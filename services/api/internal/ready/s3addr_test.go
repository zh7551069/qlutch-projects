package ready

import "testing"

func TestUsePathStyle(t *testing.T) {
	pathStyle := []string{
		"127.0.0.1:4715",
		"127.0.0.1",
		"10.0.0.8:9000",
		"localhost:4715",
		"localhost",
		"LOCALHOST:9000",
		"[::1]:4715",
		"[::1]",
		"::1",
		"object:9000",
		"minio:9000",
		"object",
	}
	for _, host := range pathStyle {
		if !UsePathStyle(host) {
			t.Fatalf("expected path-style for %s", host)
		}
		got, path := BucketHost(host, "example-bucket")
		if !path || got != host {
			t.Fatalf("BucketHost(%s) = %s path=%v", host, got, path)
		}
	}

	virtual := []string{
		"oss-cn-hangzhou.aliyuncs.com",
		"oss-cn-hangzhou.aliyuncs.com:443",
		"storage.example.com",
		"storage.example.com:8443",
	}
	for _, host := range virtual {
		if UsePathStyle(host) {
			t.Fatalf("expected virtual-hosted style for %s", host)
		}
		got, path := BucketHost(host, "example-bucket")
		if path || got != "example-bucket."+host {
			t.Fatalf("BucketHost(%s) = %s path=%v", host, got, path)
		}
	}
}
