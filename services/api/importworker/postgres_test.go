package importworker

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zh7551069/qlutch-projects/services/airuntime"
	"github.com/zh7551069/qlutch-projects/services/api/internal/config"
	"github.com/zh7551069/qlutch-projects/services/api/internal/knowledge"
)

type pgProvider struct {
	s01 int
	on  func(scene string)
}

func (p *pgProvider) Complete(_ context.Context, req airuntime.ProviderRequest) (airuntime.ProviderResponse, error) {
	if req.SceneCode == airuntime.SceneS01 {
		p.s01++
	}
	if p.on != nil {
		p.on(req.SceneCode)
	}
	if req.SceneCode == airuntime.SceneS02 {
		return airuntime.ProviderResponse{Tokens: 1, Questions: []airuntime.Question{{Stem: "预习"}}}, nil
	}
	return airuntime.ProviderResponse{Tokens: 2, Map: airuntime.MapTree{Catalogs: []airuntime.Catalog{{
		Title: "目录", Chapters: []airuntime.Chapter{{Title: "章节", Points: []airuntime.Point{{Title: "考点"}}}},
	}}}}, nil
}

func TestPostgresMapCommitsBeforeMapReady(t *testing.T) {
	pool, schema := testPostgres(t)
	note := []byte("第一段。\n\n第二段。")
	const (
		userID = "usr_a5pg"
		packID = "pack_a5pg"
		matID  = "mat_a5pg"
		jobID  = "job_a5pg"
		objKey = "m/usr_a5pg/obj_a5pg"
	)
	blobs := knowledge.NewMemoryBlobs()
	blobs.Put(objKey, note)
	provider := &pgProvider{}
	rt := airuntime.New()
	rt.SetProvider(provider)
	proc, err := Open(context.Background(), Config{Schema: schema, Runtime: rt, pool: pool, blobs: blobs})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(proc.Close)
	insertImport(t, pool, schema, userID, packID, matID, jobID, objKey, len(note))
	provider.on = func(scene string) {
		status := pgText(t, pool, schema, "knowledge_jobs", "status", "job_id", jobID)
		nodes := pgCount(t, pool, schema, "knowledge_map_nodes", "pack_id", packID)
		flag := pgBool(t, pool, schema, "identity_accounts", "has_first_pack", "user_id", userID)
		if scene == airuntime.SceneS01 {
			if status != "running" || nodes != 0 || flag {
				t.Fatalf("during s01 status=%s nodes=%d flag=%v", status, nodes, flag)
			}
			return
		}
		if status != "map_ready" || nodes < 3 || !flag {
			t.Fatalf("during s02 status=%s nodes=%d flag=%v", status, nodes, flag)
		}
	}
	payload, err := json.Marshal(map[string]string{"job_id": jobID})
	if err != nil {
		t.Fatal(err)
	}
	if err := proc.Handle(context.Background(), payload); err != nil {
		t.Fatal(err)
	}
	if provider.s01 != 2 {
		t.Fatalf("s01 calls = %d", provider.s01)
	}
	if pgText(t, pool, schema, "knowledge_jobs", "status", "job_id", jobID) != "map_ready" {
		t.Fatal("status was not committed")
	}
	var success bool
	err = pool.QueryRow(context.Background(), `SELECT success FROM `+pgx.Identifier{schema}.Sanitize()+`.knowledge_usage_ledger WHERE job_id = $1 AND scene_code = 'S01'`, jobID).Scan(&success)
	if err != nil || !success {
		t.Fatalf("s01 ledger success=%v err=%v", success, err)
	}
}

func TestPostgresUnconfiguredProviderStaysRetryable(t *testing.T) {
	pool, schema := testPostgres(t)
	note := []byte("没有提供方")
	blobs := knowledge.NewMemoryBlobs()
	blobs.Put("m/obj", note)
	rt := airuntime.New()
	proc, err := Open(context.Background(), Config{Schema: schema, Runtime: rt, pool: pool, blobs: blobs})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(proc.Close)
	insertImport(t, pool, schema, "usr_a5none", "pack_a5none", "mat_a5none", "job_a5none", "m/obj", len(note))
	payload, _ := json.Marshal(map[string]string{"job_id": "job_a5none"})
	if err := proc.Handle(context.Background(), payload); err != nil {
		t.Fatal(err)
	}
	if pgText(t, pool, schema, "knowledge_jobs", "status", "job_id", "job_a5none") != "failed_retryable" {
		t.Fatal("status")
	}
	if pgBool(t, pool, schema, "identity_accounts", "has_first_pack", "user_id", "usr_a5none") {
		t.Fatal("flag")
	}
	if pgCount(t, pool, schema, "knowledge_map_nodes", "pack_id", "pack_a5none") != 0 {
		t.Fatal("map")
	}
	var success bool
	err = pool.QueryRow(context.Background(), `SELECT success FROM `+pgx.Identifier{schema}.Sanitize()+`.knowledge_usage_ledger WHERE job_id = $1 AND scene_code = 'S01'`, "job_a5none").Scan(&success)
	if err != nil || success {
		t.Fatalf("ledger success=%v err=%v", success, err)
	}
}

func insertImport(t *testing.T, pool *pgxpool.Pool, schema, userID, packID, matID, jobID, objectKey string, size int) {
	t.Helper()
	prefix := pgx.Identifier{schema}.Sanitize()
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `INSERT INTO `+prefix+`.identity_accounts (user_id, phone_region, phone_national, locale, region_pref, has_first_pack) VALUES ($1, '+86', $2, 'zh-Hans', '+86', false)`, userID, "138"+userID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO `+prefix+`.knowledge_packs (pack_id, user_id) VALUES ($1, $2)`, packID, userID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO `+prefix+`.knowledge_materials (material_id, pack_id, user_id, source, content_type, byte_size, object_key) VALUES ($1, $2, $3, 'paste_note', 'text/plain', $4, $5)`, matID, packID, userID, size, objectKey); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO `+prefix+`.knowledge_jobs (job_id, pack_id, material_id, user_id, source, status) VALUES ($1, $2, $3, $4, 'paste_note', 'accepted')`, jobID, packID, matID, userID); err != nil {
		t.Fatal(err)
	}
}

func testPostgres(t *testing.T) (*pgxpool.Pool, string) {
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
	if _, err := hex.Decode(buf, []byte("0011223344556677")); err != nil {
		t.Fatal(err)
	}
	schema := "a5_" + hex.EncodeToString(buf)[:12]
	t.Cleanup(func() {
		dropCtx, dropCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer dropCancel()
		_, _ = pool.Exec(dropCtx, "DROP SCHEMA IF EXISTS "+pgx.Identifier{schema}.Sanitize()+" CASCADE")
	})
	return pool, schema
}

func pgText(t *testing.T, pool *pgxpool.Pool, schema, table, column, keyColumn, key string) string {
	t.Helper()
	var got string
	err := pool.QueryRow(context.Background(), `SELECT `+column+` FROM `+pgx.Identifier{schema}.Sanitize()+`.`+table+` WHERE `+keyColumn+` = $1`, key).Scan(&got)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func pgCount(t *testing.T, pool *pgxpool.Pool, schema, table, keyColumn, key string) int {
	t.Helper()
	var n int
	err := pool.QueryRow(context.Background(), `SELECT count(*) FROM `+pgx.Identifier{schema}.Sanitize()+`.`+table+` WHERE `+keyColumn+` = $1`, key).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func pgBool(t *testing.T, pool *pgxpool.Pool, schema, table, column, keyColumn, key string) bool {
	t.Helper()
	var got bool
	err := pool.QueryRow(context.Background(), `SELECT `+column+` FROM `+pgx.Identifier{schema}.Sanitize()+`.`+table+` WHERE `+keyColumn+` = $1`, key).Scan(&got)
	if err != nil {
		t.Fatal(err)
	}
	return got
}
