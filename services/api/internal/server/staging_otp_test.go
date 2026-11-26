package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/zh7551069/qlutch-projects/services/api/internal/identity"
)

const stagingFixtureToken = "fixture-staging-token"

func TestStagingRevealGateRejectsOtherEnvironments(t *testing.T) {
	if identity.NewStagingReveal("staging", stagingFixtureToken) == nil {
		t.Fatal("staging with a token did not construct a reveal gate")
	}
	for _, envName := range []string{"production", "local", "", "prod", "STAGING", "staging "} {
		if envName == "staging " {
			// A trailing space is trimmed and therefore still staging.
			if identity.NewStagingReveal(envName, stagingFixtureToken) == nil {
				t.Fatal("trimmed staging name did not construct a reveal gate")
			}
			continue
		}
		if identity.NewStagingReveal(envName, stagingFixtureToken) != nil {
			t.Fatalf("environment %q constructed a reveal gate", envName)
		}
	}
	if identity.NewStagingReveal("staging", "") != nil || identity.NewStagingReveal("staging", "   ") != nil {
		t.Fatal("blank token constructed a reveal gate")
	}
}

func TestStagingRevealReturnsFixtureAndOtherEnvsDoNot(t *testing.T) {
	stagingSrv, stagingLog := revealServer(t, "staging", stagingFixtureToken)
	sendStatus, sendBody := postJSON(t, stagingSrv.URL+"/v1/identity/otp/send", map[string]string{
		"Idempotency-Key": "stage-send",
	}, `{"region_code":"+86","phone_national":"13800018037","purpose":"register"}`)
	if sendStatus != http.StatusOK {
		t.Fatalf("send status = %d", sendStatus)
	}
	assertNoFixture(t, "send response", sendBody)
	revealStatus, revealBody := postJSON(t, stagingSrv.URL+"/internal/staging/otp/reveal", map[string]string{
		"Authorization": "Bearer " + stagingFixtureToken,
	}, `{"region_code":"+86","phone_national":"13800018037","purpose":"register"}`)
	if revealStatus != http.StatusOK {
		t.Fatal("staging reveal did not return the active fixture")
	}
	var revealed struct {
		Code string `json:"code"`
	}
	if json.Unmarshal(revealBody, &revealed) != nil || revealed.Code != testCode {
		t.Fatal("staging reveal did not return the active fixture")
	}
	assertNoFixture(t, "staging log", stagingLog.Bytes())
	if strings.Contains(stagingLog.String(), stagingFixtureToken) || strings.Contains(stagingLog.String(), "13800018037") {
		t.Fatal("staging log contained fixture material")
	}
	wrongStatus, wrongBody := postJSON(t, stagingSrv.URL+"/internal/staging/otp/reveal", map[string]string{
		"Authorization": "Bearer wrong-fixture-token",
	}, `{"region_code":"+86","phone_national":"13800018037","purpose":"register"}`)
	if wrongStatus != http.StatusNotFound {
		t.Fatal("wrong token was not rejected")
	}
	assertNoFixture(t, "wrong token response", wrongBody)
	verifyStatus, verifyBody := postJSON(t, stagingSrv.URL+"/v1/identity/otp/verify", nil,
		`{"region_code":"+86","phone_national":"13800018037","purpose":"register","code":"`+revealed.Code+`"}`)
	if verifyStatus != http.StatusOK {
		t.Fatal("verify with the revealed fixture failed")
	}
	assertNoFixture(t, "verify response", verifyBody)

	for _, envName := range []string{"production", "local", "prod", "STAGING"} {
		srv, logBuf := revealServer(t, envName, stagingFixtureToken)
		status, body := postJSON(t, srv.URL+"/v1/identity/otp/send", map[string]string{
			"Idempotency-Key": "other-send",
		}, `{"region_code":"+86","phone_national":"13800018038","purpose":"register"}`)
		if status != http.StatusOK {
			t.Fatalf("%s send status = %d", envName, status)
		}
		assertNoFixture(t, envName+" send response", body)
		revealCode, revealRaw := postJSON(t, srv.URL+"/internal/staging/otp/reveal", map[string]string{
			"Authorization": "Bearer " + stagingFixtureToken,
		}, `{"region_code":"+86","phone_national":"13800018038","purpose":"register"}`)
		if revealCode != http.StatusNotFound {
			t.Fatalf("%s reveal status = %d", envName, revealCode)
		}
		assertNoFixture(t, envName+" reveal response", revealRaw)
		assertNoFixture(t, envName+" log", logBuf.Bytes())
		if strings.Contains(logBuf.String(), stagingFixtureToken) {
			t.Fatalf("%s log contained the fixture token", envName)
		}
	}

	closed, closedLog := revealServer(t, "staging", "")
	closedStatus, closedBody := postJSON(t, closed.URL+"/internal/staging/otp/reveal", map[string]string{
		"Authorization": "Bearer " + stagingFixtureToken,
	}, `{"region_code":"+86","phone_national":"13800018039","purpose":"register"}`)
	if closedStatus != http.StatusNotFound {
		t.Fatal("staging without a token mounted reveal")
	}
	assertNoFixture(t, "closed reveal response", closedBody)
	assertNoFixture(t, "closed log", closedLog.Bytes())
}

func TestLearnerContractOmitsStagingReveal(t *testing.T) {
	for _, rel := range []string{
		"../../../../packages/contracts/openapi.yaml",
		"../../../../docs/contracts/新用户流程 A1 身份契约（2026-11-25）.md",
	} {
		body, err := os.ReadFile(rel)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(body), "/internal/staging/otp/reveal") {
			t.Fatalf("%s includes the staging reveal route", rel)
		}
	}
}

func revealServer(t *testing.T, envName, token string) (*httptest.Server, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })
	reveal := identity.NewStagingReveal(envName, token)
	svc, err := identity.New(identity.Options{
		Sender:     &identity.CaptureSender{},
		Code:       fixedCode,
		Clock:      identity.NewManualClock(time.Date(2026, 11, 26, 8, 0, 0, 0, time.UTC)),
		RetainCode: reveal != nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	public := New(nil, svc)
	public.SetStagingReveal(reveal)
	srv := httptest.NewServer(public.Handler(nil))
	t.Cleanup(srv.Close)
	return srv, &buf
}

func postJSON(t *testing.T, url string, headers map[string]string, body string) (int, []byte) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return res.StatusCode, raw
}

func assertNoFixture(t *testing.T, label string, body []byte) {
	t.Helper()
	if strings.Contains(string(body), testCode) {
		t.Fatalf("%s contained the fixture code", label)
	}
}
