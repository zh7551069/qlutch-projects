package ready

import (
	"net"
	"strings"
)

// UsePathStyle reports whether endpointHost should keep the bucket in the
// request path. IP addresses, localhost, and single-label names stay
// path-style so local MinIO keeps working: the default endpoint is an IP,
// and the compose service name "object" is a single label. A dotted DNS
// name uses virtual-hosted style, which regional endpoints require.
func UsePathStyle(endpointHost string) bool {
	name := endpointHost
	if h, _, err := net.SplitHostPort(endpointHost); err == nil {
		name = h
	}
	name = strings.TrimPrefix(name, "[")
	name = strings.TrimSuffix(name, "]")
	if strings.EqualFold(name, "localhost") {
		return true
	}
	if net.ParseIP(name) != nil {
		return true
	}
	return !strings.Contains(name, ".")
}

// BucketHost is the Host header for bucket on endpointHost.
// Virtual-hosted style is "{bucket}.{endpointHost}". Path-style returns
// endpointHost unchanged. The configured endpoint string is not rewritten,
// and the bucket is not also placed in the path when it is on the host.
func BucketHost(endpointHost, bucket string) (host string, pathStyle bool) {
	if UsePathStyle(endpointHost) {
		return endpointHost, true
	}
	return bucket + "." + endpointHost, false
}
