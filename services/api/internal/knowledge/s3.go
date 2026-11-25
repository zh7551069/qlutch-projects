package knowledge

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/zh7551069/qlutch-projects/services/api/internal/ready"
)

// S3Store speaks the S3 API with query-signed PUT URLs and signed HEAD requests.
// It does not import a cloud vendor SDK.
type S3Store struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	Secret    string
	Expires   time.Duration
	Client    *http.Client
	Now       func() time.Time
}

// PresignPut implements Blobs.
func (s S3Store) PresignPut(ctx context.Context, objectKey, contentType string) (string, map[string]string, int64, error) {
	if err := ctx.Err(); err != nil {
		return "", nil, 0, errStorage
	}
	expires := s.Expires
	if expires <= 0 {
		expires = time.Duration(DefaultPresignSeconds) * time.Second
	}
	secs := int64(expires / time.Second)
	if secs < 1 {
		secs = 1
	}
	objectURL, err := s.objectURL(objectKey)
	if err != nil {
		return "", nil, 0, err
	}
	now := time.Now()
	if s.Now != nil {
		now = s.Now()
	}
	signed, headers, err := signQueryURL(http.MethodPut, objectURL.Scheme, objectURL.Host, objectURL.EscapedPath(), contentType, now, int(secs), s.Region, s.AccessKey, s.Secret)
	if err != nil {
		return "", nil, 0, errStorage
	}
	return signed, headers, secs, nil
}

// Stat implements Blobs. A missing object is found=false and a nil error.
func (s S3Store) Stat(ctx context.Context, objectKey string) (int64, bool, error) {
	objectURL, err := s.objectURL(objectKey)
	if err != nil {
		return 0, false, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, objectURL.String(), nil)
	if err != nil {
		return 0, false, errStorage
	}
	req.Host = objectURL.Host
	now := time.Now()
	if s.Now != nil {
		now = s.Now()
	}
	ready.SignS3Request(req, now, s.Region, s.AccessKey, s.Secret)
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, false, errStorage
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound {
		return 0, false, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || resp.ContentLength < 0 {
		return 0, false, errStorage
	}
	return resp.ContentLength, true, nil
}

func (s S3Store) objectURL(objectKey string) (*url.URL, error) {
	base, err := url.Parse(s.Endpoint)
	if err != nil || base.Scheme == "" || base.Host == "" || s.Bucket == "" || objectKey == "" {
		return nil, errStorage
	}
	base.RawQuery = ""
	base.Fragment = ""
	base.Path = strings.TrimRight(base.EscapedPath(), "/") + "/" + awsEncode(s.Bucket, true) + "/" + escapeKey(objectKey)
	base.RawPath = base.Path
	return base, nil
}

func signQueryURL(method, scheme, host, canonicalURI, contentType string, now time.Time, expires int, region, access, secret string) (string, map[string]string, error) {
	if method == "" || scheme == "" || host == "" || canonicalURI == "" || region == "" || access == "" || secret == "" || expires < 1 {
		return "", nil, errStorage
	}
	amzDate := now.UTC().Format("20060102T150405Z")
	dateStamp := now.UTC().Format("20060102")
	scope := dateStamp + "/" + region + "/s3/aws4_request"
	signedHeaderNames := "host"
	canonicalHeaders := "host:" + host + "\n"
	headers := map[string]string{}
	if contentType != "" {
		signedHeaderNames = "content-type;host"
		canonicalHeaders = "content-type:" + contentType + "\nhost:" + host + "\n"
		headers["Content-Type"] = contentType
	}
	pairs := []string{
		"X-Amz-Algorithm=" + awsEncode("AWS4-HMAC-SHA256", true),
		"X-Amz-Credential=" + awsEncode(access+"/"+scope, true),
		"X-Amz-Date=" + awsEncode(amzDate, true),
		"X-Amz-Expires=" + awsEncode(strconv.Itoa(expires), true),
		"X-Amz-SignedHeaders=" + awsEncode(signedHeaderNames, true),
	}
	sort.Strings(pairs)
	canonicalQuery := strings.Join(pairs, "&")
	canonical := method + "\n" +
		canonicalURI + "\n" +
		canonicalQuery + "\n" +
		canonicalHeaders + "\n" +
		signedHeaderNames + "\n" +
		"UNSIGNED-PAYLOAD"
	stringToSign := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + scope + "\n" + hex.EncodeToString(hashSHA256([]byte(canonical)))
	signature := hex.EncodeToString(hmacSHA256(signingKey(secret, dateStamp, region), []byte(stringToSign)))
	signed := scheme + "://" + host + canonicalURI + "?" + canonicalQuery + "&X-Amz-Signature=" + signature
	return signed, headers, nil
}

func escapeKey(objectKey string) string {
	parts := strings.Split(objectKey, "/")
	for i, part := range parts {
		parts[i] = awsEncode(part, true)
	}
	return strings.Join(parts, "/")
}

func awsEncode(value string, encodeSlash bool) string {
	var b strings.Builder
	for i := 0; i < len(value); i++ {
		c := value[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '-' || c == '~' || c == '.' || (c == '/' && !encodeSlash) {
			b.WriteByte(c)
			continue
		}
		const hexDigits = "0123456789ABCDEF"
		b.WriteByte('%')
		b.WriteByte(hexDigits[c>>4])
		b.WriteByte(hexDigits[c&15])
	}
	return b.String()
}

func signingKey(secret, date, region string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), []byte(date))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte("s3"))
	return hmacSHA256(kService, []byte("aws4_request"))
}

func hmacSHA256(key, value []byte) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(value)
	return mac.Sum(nil)
}

func hashSHA256(value []byte) []byte {
	sum := sha256.Sum256(value)
	return sum[:]
}
