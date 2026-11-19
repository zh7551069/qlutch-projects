#!/usr/bin/env python3
"""Create the local S3 bucket with Signature Version 4. No cloud SDK."""

import datetime
import hashlib
import hmac
import os
import urllib.error
import urllib.request


def sign(method: str, host: str, path: str, region: str, access: str, secret: str) -> dict[str, str]:
    now = datetime.datetime.now(datetime.timezone.utc)
    amz_date = now.strftime("%Y%m%dT%H%M%SZ")
    date_stamp = now.strftime("%Y%m%d")
    payload_hash = hashlib.sha256(b"").hexdigest()
    canonical_headers = (
        f"host:{host}\n"
        f"x-amz-content-sha256:{payload_hash}\n"
        f"x-amz-date:{amz_date}\n"
    )
    signed_headers = "host;x-amz-content-sha256;x-amz-date"
    canonical = f"{method}\n{path}\n\n{canonical_headers}\n{signed_headers}\n{payload_hash}"
    scope = f"{date_stamp}/{region}/s3/aws4_request"
    string_to_sign = "AWS4-HMAC-SHA256\n" + amz_date + "\n" + scope + "\n" + hashlib.sha256(canonical.encode()).hexdigest()
    key = hmac.new(("AWS4" + secret).encode(), date_stamp.encode(), hashlib.sha256).digest()
    key = hmac.new(key, region.encode(), hashlib.sha256).digest()
    key = hmac.new(key, b"s3", hashlib.sha256).digest()
    key = hmac.new(key, b"aws4_request", hashlib.sha256).digest()
    signature = hmac.new(key, string_to_sign.encode(), hashlib.sha256).hexdigest()
    return {
        "Authorization": f"AWS4-HMAC-SHA256 Credential={access}/{scope}, SignedHeaders={signed_headers}, Signature={signature}",
        "x-amz-date": amz_date,
        "x-amz-content-sha256": payload_hash,
    }


def main() -> None:
    endpoint = os.environ["S3_ENDPOINT"].rstrip("/")
    region = os.environ["S3_REGION"]
    bucket = os.environ["S3_BUCKET"]
    access = os.environ["S3_ACCESS_KEY"]
    secret = os.environ["S3_SECRET_KEY"]
    host = endpoint.split("://", 1)[1]
    path = "/" + bucket
    headers = sign("PUT", host, path, region, access, secret)
    request = urllib.request.Request(endpoint + path, method="PUT", headers=headers)
    try:
        with urllib.request.urlopen(request, timeout=10) as response:
            print(f"bucket {bucket} status {response.status}")
    except urllib.error.HTTPError as exc:
        if exc.code in (200, 409):
            print(f"bucket {bucket} status {exc.code}")
            return
        raise SystemExit(f"bucket create failed: {exc.code}") from exc


if __name__ == "__main__":
    main()
