package server

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	qlutchapi "github.com/zh7551069/qlutch-projects/packages/contracts/gen/go"
	"github.com/zh7551069/qlutch-projects/services/api/internal/identity"
	"github.com/zh7551069/qlutch-projects/services/api/internal/knowledge"
)

func TestKnowledgeStateSurvivesProcessBoundary(t *testing.T) {
	pool, schema := identityPostgres(t)
	blobs := knowledge.NewMemoryBlobs()
	first := newBoundaryKnowledge(t, pool, schema, blobs)

	access, _, userID := first.register(t, "13800042001")
	first.onboard(t, access)
	pack := first.createPack(t, access)
	if pack.Body.Account.Has_first_pack {
		t.Fatal("empty pack completed the first pack")
	}
	upload := first.upload(t, access, pack.Body.Pack_id, "paste_note", "text/plain", 9, "note.txt")
	blobs.Complete(upload.Body.Upload_url, 9)
	imported := first.importJob(t, access, "durable-import", pack.Body.Pack_id, upload.Body.Material_id)
	if imported.StatusCode != http.StatusAccepted || imported.Body.Status != qlutchapi.JobStatusAccepted || imported.Body.Account.Has_first_pack {
		t.Fatalf("import = %d %+v", imported.StatusCode, imported.Body)
	}
	if err := first.know.TestingSetJobStatus(context.Background(), imported.Body.Job_id, string(qlutchapi.JobStatusMapReady)); err != nil {
		t.Fatal(err)
	}

	skipAccess, _, skipUser := first.register(t, "13800042002")
	skipped := first.skip(t, skipAccess, "durable-skip", nil)
	if skipped.StatusCode != http.StatusOK || !skipped.Body.Account.Has_first_pack || skipped.Body.Next_step != qlutchapi.NextStepNEEDONBOARDING {
		t.Fatalf("skip = %+v", skipped.Body)
	}
	if n, err := first.know.TestingJobCount(context.Background(), skipUser); err != nil || n != 0 {
		t.Fatalf("skip jobs = %d %v", n, err)
	}

	second := newBoundaryKnowledge(t, pool, schema, blobs)
	resumed := second.job(t, access, imported.Body.Job_id)
	if resumed.StatusCode != http.StatusOK || resumed.Body.Status != qlutchapi.JobStatusMapReady || !resumed.Body.Map_ready || !resumed.Body.Account.Has_first_pack || resumed.Body.Next_step != qlutchapi.NextStepREADYHOME {
		t.Fatalf("resumed job = %d %+v %s", resumed.StatusCode, resumed.Body, resumed.Raw)
	}
	replay := second.importJob(t, access, "durable-import", pack.Body.Pack_id, upload.Body.Material_id)
	if replay.StatusCode != http.StatusAccepted || replay.Body.Job_id != imported.Body.Job_id || replay.Body.Status != qlutchapi.JobStatusMapReady {
		t.Fatalf("replay after restart = %d %+v", replay.StatusCode, replay.Body)
	}
	if second.onboarding(t, access).Body.Next_step != qlutchapi.NextStepREADYHOME {
		t.Fatal("refresh path bounced to the first-pack gate")
	}
	skipState := second.onboarding(t, skipAccess)
	if !skipState.Body.Account.Has_first_pack || skipState.Body.Next_step != qlutchapi.NextStepNEEDONBOARDING {
		t.Fatalf("skip after restart = %+v", skipState.Body)
	}
	if n, err := second.know.TestingJobCount(context.Background(), skipUser); err != nil || n != 0 {
		t.Fatalf("skip jobs after restart = %d %v", n, err)
	}
	assertFlag(t, pool, schema, userID, true)
	assertFlag(t, pool, schema, skipUser, true)
	assertSkip(t, pool, schema, skipUser)
	assertJobStatus(t, pool, schema, imported.Body.Job_id, "map_ready")
	assertPackCount(t, pool, schema, userID, 1)
}

func newBoundaryKnowledge(t *testing.T, pool *pgxpool.Pool, schema string, blobs *knowledge.MemoryBlobs) *knowledgeHarness {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	identStore, err := identity.OpenPostgres(ctx, pool, schema)
	if err != nil {
		t.Fatal(err)
	}
	knowStore, err := knowledge.OpenPostgres(ctx, pool, schema)
	if err != nil {
		t.Fatal(err)
	}
	return newKnowledgeWith(t, identStore, knowStore, blobs)
}

func assertFlag(t *testing.T, pool *pgxpool.Pool, schema, userID string, want bool) {
	t.Helper()
	var got bool
	table := pgx.Identifier{schema}.Sanitize() + ".identity_accounts"
	err := pool.QueryRow(context.Background(), `SELECT has_first_pack FROM `+table+` WHERE user_id = $1`, userID).Scan(&got)
	if err != nil || got != want {
		t.Fatalf("has_first_pack stored=%v err=%v", got, err)
	}
}

func assertSkip(t *testing.T, pool *pgxpool.Pool, schema, userID string) {
	t.Helper()
	var got string
	table := pgx.Identifier{schema}.Sanitize() + ".knowledge_first_pack_skips"
	err := pool.QueryRow(context.Background(), `SELECT user_id FROM `+table+` WHERE user_id = $1`, userID).Scan(&got)
	if err != nil || got != userID {
		t.Fatal("skip row missing")
	}
}

func assertJobStatus(t *testing.T, pool *pgxpool.Pool, schema, jobID, want string) {
	t.Helper()
	var got string
	table := pgx.Identifier{schema}.Sanitize() + ".knowledge_jobs"
	err := pool.QueryRow(context.Background(), `SELECT status FROM `+table+` WHERE job_id = $1`, jobID).Scan(&got)
	if err != nil || got != want {
		t.Fatalf("job status = %s err=%v", got, err)
	}
}

func assertPackCount(t *testing.T, pool *pgxpool.Pool, schema, userID string, want int) {
	t.Helper()
	var got int
	table := pgx.Identifier{schema}.Sanitize() + ".knowledge_packs"
	err := pool.QueryRow(context.Background(), `SELECT count(*) FROM `+table+` WHERE user_id = $1`, userID).Scan(&got)
	if err != nil || got != want {
		t.Fatalf("packs = %d err=%v", got, err)
	}
}
