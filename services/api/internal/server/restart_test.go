package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	qlutchapi "github.com/zh7551069/qlutch-projects/packages/contracts/gen/go"
	"github.com/zh7551069/qlutch-projects/services/api/internal/config"
	"github.com/zh7551069/qlutch-projects/services/api/internal/identity"
)

// TestIdentityStateSurvivesProcessBoundary writes onboarding and profile through
// one API service, then reads and continues them through new services that share
// only PostgreSQL. OTP challenges and idempotency keys stay in the first process
// and are absent from the next one.
func TestIdentityStateSurvivesProcessBoundary(t *testing.T) {
	pool, schema := identityPostgres(t)
	first := newBoundaryService(t, pool, schema)
	cn := "13800030001"
	hk := "51230001"
	fresh := "13800030099"

	if first.send(t, "cn-reg", "+86", cn, "register", localePtr(qlutchapi.LocaleZhHans)).StatusCode != http.StatusOK {
		t.Fatal("cn send")
	}
	cnSession := first.verify(t, "+86", cn, "register", testCode, localePtr(qlutchapi.LocaleZhHans))
	if cnSession.StatusCode != http.StatusOK || cnSession.Body.Next_step != qlutchapi.NextStepNEEDONBOARDING {
		t.Fatalf("cn verify = %d %s", cnSession.StatusCode, cnSession.Raw)
	}
	cnAccess := cnSession.Body.Access_token
	cnUser := cnSession.Body.Account.User_id
	first.putOK(t, cnAccess, "nickname", "小明", "")
	first.putOK(t, cnAccess, "age", "", "18-24")
	cnProfile := first.putProfile(t, cnAccess, localePtr(qlutchapi.LocaleZhHant), regionPtr(qlutchapi.RegionCodeValue852))
	if cnProfile.StatusCode != http.StatusOK || cnProfile.Body.Locale != qlutchapi.LocaleZhHant || cnProfile.Body.Region_code != "+852" {
		t.Fatalf("cn profile = %+v %s", cnProfile.Body, cnProfile.Raw)
	}

	if first.send(t, "hk-reg", "+852", hk, "register", localePtr(qlutchapi.LocaleZhHans)).StatusCode != http.StatusOK {
		t.Fatal("hk send")
	}
	hkSession := first.verify(t, "+852", hk, "register", testCode, localePtr(qlutchapi.LocaleZhHans))
	if hkSession.StatusCode != http.StatusOK {
		t.Fatalf("hk verify = %d %s", hkSession.StatusCode, hkSession.Raw)
	}
	hkAccess := hkSession.Body.Access_token
	hkRefresh := hkSession.Body.Refresh_token
	hkUser := hkSession.Body.Account.User_id
	first.putOK(t, hkAccess, "nickname", "阿明", "")

	if first.send(t, "boundary-same", "+86", fresh, "register", nil).StatusCode != http.StatusOK {
		t.Fatal("fresh send")
	}

	// Process boundary: a new service object, a new OTP pepper, and empty
	// in-memory OTP / idempotency / rate-limit maps. The first service is not
	// called again.
	second := newBoundaryService(t, pool, schema)
	exists := second.send(t, "cn-again", "+86", cn, "register", nil)
	if exists.StatusCode != http.StatusConflict || mustCode(t, exists.StatusCode, exists.Raw, exists.RequestID) != "AUTH_ACCOUNT_EXISTS" || len(second.sender.Codes()) != 0 {
		t.Fatalf("account did not survive: %d %s", exists.StatusCode, exists.Raw)
	}
	expired := second.verify(t, "+86", fresh, "register", testCode, nil)
	if expired.StatusCode != http.StatusUnauthorized || mustCode(t, expired.StatusCode, expired.Raw, expired.RequestID) != "AUTH_OTP_EXPIRED" {
		t.Fatalf("otp did not die with the first process: %d %s", expired.StatusCode, expired.Raw)
	}
	resent := second.send(t, "boundary-same", "+86", fresh, "register", nil)
	if resent.StatusCode != http.StatusOK || len(second.sender.Codes()) != 1 {
		t.Fatalf("idempotency survived the boundary: %d codes %d", resent.StatusCode, len(second.sender.Codes()))
	}

	resumed := second.onboarding(t, cnAccess)
	if resumed.Body.Account.User_id != cnUser || resumed.Body.Completed || resumed.Body.Next_step != qlutchapi.NextStepNEEDONBOARDING {
		t.Fatalf("resume account = %+v", resumed.Body)
	}
	if resumed.Body.Current_step_id == nil || *resumed.Body.Current_step_id != "study_or_work" || resumed.Body.Current_index != 3 {
		t.Fatalf("resume step = %+v index %d", resumed.Body.Current_step_id, resumed.Body.Current_index)
	}
	if resumed.Body.Nickname == nil || *resumed.Body.Nickname != "小明" {
		t.Fatalf("nickname = %+v", resumed.Body.Nickname)
	}
	assertOnboardingSteps(t, resumed.Body.Steps, map[string]string{"nickname": "小明", "age": "18-24"})
	if code := second.putStep(t, cnAccess, "daily_time", "", "half-hour"); code != "VALIDATION_STEP_OUT_OF_ORDER" {
		t.Fatalf("skip after restart = %s", code)
	}
	rewritten := second.putOK(t, cnAccess, "nickname", "小红", "")
	if rewritten.Body.Nickname == nil || *rewritten.Body.Nickname != "小红" || rewritten.Body.Current_step_id == nil || *rewritten.Body.Current_step_id != "study_or_work" {
		t.Fatalf("rewrite = %+v", rewritten.Body)
	}
	profile := second.profile(t, cnAccess)
	if profile.StatusCode != http.StatusOK || profile.Body.Locale != qlutchapi.LocaleZhHant || profile.Body.Region_code != "+852" {
		t.Fatalf("profile after restart = %+v %s", profile.Body, profile.Raw)
	}
	assertAccountRow(t, pool, schema, cnUser, "+86", cn, "zh-Hant", "+852")
	assertLoginPhone(t, second, "+86", "+852", cn)
	still := second.profile(t, cnAccess)
	if still.Body.Locale != qlutchapi.LocaleZhHant || still.Body.Region_code != "+852" {
		t.Fatalf("login send changed profile = %+v", still.Body)
	}

	rotated := second.refresh(t, hkRefresh, "")
	if rotated.StatusCode != http.StatusOK || rotated.Body.Account.User_id != hkUser || rotated.Body.Access_token == hkAccess || rotated.Body.Refresh_token == hkRefresh {
		t.Fatalf("refresh after restart = %d %s", rotated.StatusCode, rotated.Raw)
	}
	if code := second.onboardingCode(t, hkAccess); code != "AUTH_EXPIRED" {
		t.Fatalf("rotated access = %s", code)
	}
	hkState := second.onboarding(t, rotated.Body.Access_token)
	if hkState.Body.Current_step_id == nil || *hkState.Body.Current_step_id != "age" || hkState.Body.Current_index != 2 || hkState.Body.Next_step != qlutchapi.NextStepNEEDONBOARDING {
		t.Fatalf("hk resume = %+v", hkState.Body)
	}
	assertOnboardingSteps(t, hkState.Body.Steps, map[string]string{"nickname": "阿明"})
	hkProfile := second.profile(t, rotated.Body.Access_token)
	if hkProfile.Body.Locale != qlutchapi.LocaleZhHans || hkProfile.Body.Region_code != "+852" {
		t.Fatalf("hk profile = %+v", hkProfile.Body)
	}
	assertAccountRow(t, pool, schema, hkUser, "+852", hk, "zh-Hans", "+852")
	assertLoginPhone(t, second, "+852", "+86", hk)

	changed := second.putProfile(t, cnAccess, nil, regionPtr(qlutchapi.RegionCodeValue86))
	if changed.StatusCode != http.StatusOK || changed.Body.Region_code != "+86" || changed.Body.Locale != qlutchapi.LocaleZhHant {
		t.Fatalf("change pref = %+v %s", changed.Body, changed.Raw)
	}
	assertAccountRow(t, pool, schema, cnUser, "+86", cn, "zh-Hant", "+86")

	for _, step := range []struct{ id, answer string }{
		{"study_or_work", "student"},
		{"identity", "candidate"},
		{"how_to_address", "friend"},
		{"exam_track", "language"},
		{"daily_time", "half-hour"},
	} {
		second.putOK(t, cnAccess, step.id, "", step.answer)
	}
	done := second.onboarding(t, cnAccess)
	if !done.Body.Completed || done.Body.Current_step_id != nil || done.Body.Current_index != 7 || done.Body.Next_step != qlutchapi.NextStepNEEDFIRSTPACK {
		t.Fatalf("done = %+v", done.Body)
	}

	third := newBoundaryService(t, pool, schema)
	finalState := third.onboarding(t, cnAccess)
	if !finalState.Body.Completed || finalState.Body.Next_step != qlutchapi.NextStepNEEDFIRSTPACK || finalState.Body.Nickname == nil || *finalState.Body.Nickname != "小红" {
		t.Fatalf("third onboarding = %+v", finalState.Body)
	}
	assertOnboardingSteps(t, finalState.Body.Steps, map[string]string{
		"nickname": "小红", "age": "18-24", "study_or_work": "student", "identity": "candidate",
		"how_to_address": "friend", "exam_track": "language", "daily_time": "half-hour",
	})
	finalProfile := third.profile(t, cnAccess)
	if finalProfile.Body.Locale != qlutchapi.LocaleZhHant || finalProfile.Body.Region_code != "+86" {
		t.Fatalf("third profile = %+v", finalProfile.Body)
	}
	assertAccountRow(t, pool, schema, cnUser, "+86", cn, "zh-Hant", "+86")
	hkAgain := third.onboarding(t, rotated.Body.Access_token)
	if hkAgain.Body.Current_step_id == nil || *hkAgain.Body.Current_step_id != "age" {
		t.Fatalf("hk third = %+v", hkAgain.Body)
	}
}

func newBoundaryService(t *testing.T, pool *pgxpool.Pool, schema string) *identityHarness {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	store, err := identity.OpenPostgres(ctx, pool, schema)
	if err != nil {
		t.Fatal(err)
	}
	return newIdentityWith(t, identity.Options{Sender: &identity.CaptureSender{}, Store: store, Code: fixedCode})
}

func identityPostgres(t *testing.T) (*pgxpool.Pool, string) {
	t.Helper()
	dsn := os.Getenv("QLUTCH_TEST_DATABASE_URL")
	required := dsn != ""
	if dsn == "" {
		dsn = config.Load().DatabaseURL
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		if required {
			t.Fatal("database url is invalid")
		}
		t.Skip("postgres is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		if required {
			t.Fatal("postgres is not running")
		}
		t.Skip("postgres is not running")
	}
	t.Cleanup(pool.Close)
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		t.Fatal(err)
	}
	schema := "a3_" + hex.EncodeToString(buf)
	t.Cleanup(func() {
		dropCtx, dropCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer dropCancel()
		_, _ = pool.Exec(dropCtx, "DROP SCHEMA IF EXISTS "+pgx.Identifier{schema}.Sanitize()+" CASCADE")
	})
	return pool, schema
}

func assertAccountRow(t *testing.T, pool *pgxpool.Pool, schema, userID, region, national, locale, pref string) {
	t.Helper()
	var gotRegion, gotNational, gotLocale, gotPref string
	table := pgx.Identifier{schema}.Sanitize() + ".identity_accounts"
	err := pool.QueryRow(context.Background(), `SELECT phone_region, phone_national, locale, region_pref FROM `+table+` WHERE user_id = $1`, userID).
		Scan(&gotRegion, &gotNational, &gotLocale, &gotPref)
	if err != nil {
		t.Fatal("account row missing")
	}
	if gotRegion != region || gotNational != national || gotLocale != locale || gotPref != pref {
		t.Fatalf("stored identity region=%s locale=%s pref=%s national_ok=%v", gotRegion, gotLocale, gotPref, gotNational == national)
	}
}

func assertLoginPhone(t *testing.T, h *identityHarness, bound, other, national string) {
	t.Helper()
	before := len(h.sender.Codes())
	missing := h.send(t, "miss-"+other+"-"+national, other, national, "login", nil)
	if code := mustCode(t, missing.StatusCode, missing.Raw, missing.RequestID); code != "AUTH_ACCOUNT_NOT_FOUND" || missing.StatusCode != http.StatusNotFound {
		t.Fatalf("other region login = %d %s", missing.StatusCode, code)
	}
	if len(h.sender.Codes()) != before {
		t.Fatal("wrong region sent sms")
	}
	found := h.send(t, "hit-"+bound+"-"+national, bound, national, "login", nil)
	if found.StatusCode != http.StatusOK {
		t.Fatalf("bound login = %d %s", found.StatusCode, found.Raw)
	}
}

func assertOnboardingSteps(t *testing.T, steps []qlutchapi.OnboardingStepState, answers map[string]string) {
	t.Helper()
	if len(steps) != 7 {
		t.Fatalf("steps = %d", len(steps))
	}
	order := []string{"nickname", "age", "study_or_work", "identity", "how_to_address", "exam_track", "daily_time"}
	for i, id := range order {
		step := steps[i]
		if string(step.Step_id) != id || step.Index != int64(i+1) {
			t.Fatalf("step %d = %+v", i, step)
		}
		want, ok := answers[id]
		if !ok {
			if step.Status != qlutchapi.OnboardingStepStateStatusPending || step.Answer != nil {
				t.Fatalf("pending %s = %+v", id, step)
			}
			continue
		}
		if step.Status != qlutchapi.OnboardingStepStateStatusCompleted || step.Answer == nil || *step.Answer != want {
			t.Fatalf("answer %s = %+v", id, step)
		}
	}
}
