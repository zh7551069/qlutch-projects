package ready

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sort"
	"strings"
	"time"
)

// SignS3Request signs an S3-compatible request with AWS Signature Version 4.
// The application speaks the S3 API and does not import a cloud vendor SDK.
func SignS3Request(req *http.Request, now time.Time, region, accessKey, secret string) {
	payloadHash := emptyPayloadHash
	if req.Body == nil {
		req.Header.Set("x-amz-content-sha256", payloadHash)
	}
	amzDate := now.UTC().Format("20060102T150405Z")
	dateStamp := now.UTC().Format("20060102")
	req.Header.Set("x-amz-date", amzDate)
	req.Header.Set("x-amz-content-sha256", payloadHash)

	headers := map[string]string{
		"host": req.Host,
	}
	for key, values := range req.Header {
		if len(values) == 0 {
			continue
		}
		headers[strings.ToLower(key)] = strings.TrimSpace(values[0])
	}
	names := make([]string, 0, len(headers))
	for name := range headers {
		names = append(names, name)
	}
	sort.Strings(names)

	canonicalHeaders := ""
	for _, name := range names {
		canonicalHeaders += name + ":" + headers[name] + "\n"
	}
	signedHeaders := strings.Join(names, ";")
	canonicalURI := req.URL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}
	// CanonicalHeaders already ends in a newline. The signature spec adds one
	// more newline before the signed-header list.
	canonical := req.Method + "\n" +
		canonicalURI + "\n" +
		req.URL.RawQuery + "\n" +
		canonicalHeaders + "\n" +
		signedHeaders + "\n" +
		payloadHash
	scope := dateStamp + "/" + region + "/s3/aws4_request"
	stringToSign := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + scope + "\n" + hex.EncodeToString(hashSHA256([]byte(canonical)))
	signature := hex.EncodeToString(hmacSHA256(signingKey(secret, dateStamp, region), []byte(stringToSign)))
	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential="+accessKey+"/"+scope+", SignedHeaders="+signedHeaders+", Signature="+signature)
}

const emptyPayloadHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

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
