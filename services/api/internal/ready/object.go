package ready

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// ObjectStore checks an S3-compatible bucket with a signed HeadBucket request.
// DNS endpoints use virtual-hosted style ({bucket}.{host}/). IP, localhost,
// and single-label hosts stay path-style so local MinIO keeps working.
type ObjectStore struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	Secret    string
	Client    *http.Client
	Now       func() time.Time
}

// Name returns the probe name.
func (o ObjectStore) Name() string { return "object_storage" }

// Ping reports whether the bucket answers a signed HEAD.
func (o ObjectStore) Ping(ctx context.Context) error {
	endpoint, err := url.Parse(o.Endpoint)
	if err != nil || endpoint.Host == "" || o.Bucket == "" {
		return errors.New("object storage unreachable")
	}
	endpoint.RawQuery = ""
	endpoint.Fragment = ""
	host, pathStyle := BucketHost(endpoint.Host, o.Bucket)
	endpoint.Host = host
	if pathStyle {
		endpoint.Path = "/" + url.PathEscape(o.Bucket)
	} else {
		endpoint.Path = "/"
	}
	endpoint.RawPath = ""
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, endpoint.String(), nil)
	if err != nil {
		return errors.New("object storage unreachable")
	}
	req.Host = endpoint.Host
	now := time.Now()
	if o.Now != nil {
		now = o.Now()
	}
	SignS3Request(req, now, o.Region, o.AccessKey, o.Secret)
	client := o.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("object storage unreachable")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("object storage status %d", resp.StatusCode)
	}
	return nil
}
