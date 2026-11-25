package knowledge

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var schemaName = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

//go:embed schema.sql
var knowledgeSchemaSQL string

type postgresStore struct {
	pool      *pgxpool.Pool
	packs     string
	materials string
	jobs      string
	skips     string
	idem      string
	nodes     string
	prefetch  string
	accounts  string
}

// OpenPostgres migrates knowledge tables into schema and returns a store.
// An empty schema uses public. Returned errors do not include a DSN or password.
func OpenPostgres(ctx context.Context, pool *pgxpool.Pool, schema string) (Store, error) {
	if pool == nil {
		return nil, errors.New("knowledge store unavailable")
	}
	if schema == "" {
		schema = "public"
	}
	if !schemaName.MatchString(schema) {
		return nil, errors.New("knowledge store unavailable")
	}
	ident := pgx.Identifier{schema}.Sanitize()
	store := &postgresStore{
		pool:      pool,
		packs:     ident + ".knowledge_packs",
		materials: ident + ".knowledge_materials",
		jobs:      ident + ".knowledge_jobs",
		skips:     ident + ".knowledge_first_pack_skips",
		idem:      ident + ".knowledge_idempotency",
		nodes:     ident + ".knowledge_map_nodes",
		prefetch:  ident + ".knowledge_prefetch",
		accounts:  ident + ".identity_accounts",
	}
	if err := store.migrate(ctx, ident); err != nil {
		return nil, redactStoreErr(err)
	}
	return store, nil
}

func (p *postgresStore) migrate(ctx context.Context, ident string) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS "+ident); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "SET LOCAL search_path TO "+ident); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, knowledgeSchemaSQL); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p *postgresStore) InsertPack(ctx context.Context, item pack) error {
	_, err := p.pool.Exec(ctx, `INSERT INTO `+p.packs+` (pack_id, user_id) VALUES ($1, $2)`, item.ID, item.UserID)
	return mapWriteErr(err)
}

func (p *postgresStore) FindPack(ctx context.Context, id string) (*pack, error) {
	var item pack
	err := p.pool.QueryRow(ctx, `SELECT pack_id, user_id FROM `+p.packs+` WHERE pack_id = $1`, id).Scan(&item.ID, &item.UserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (p *postgresStore) InsertMaterial(ctx context.Context, item material) error {
	_, err := p.pool.Exec(ctx, `INSERT INTO `+p.materials+` (material_id, pack_id, user_id, source, content_type, byte_size, filename, object_key) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		item.ID, item.PackID, item.UserID, item.Source, item.ContentType, item.ByteSize, nullString(item.Filename), item.ObjectKey)
	return mapWriteErr(err)
}

func (p *postgresStore) FindMaterial(ctx context.Context, id string) (*material, error) {
	var item material
	var filename *string
	err := p.pool.QueryRow(ctx, `SELECT material_id, pack_id, user_id, source, content_type, byte_size, filename, object_key FROM `+p.materials+` WHERE material_id = $1`, id).
		Scan(&item.ID, &item.PackID, &item.UserID, &item.Source, &item.ContentType, &item.ByteSize, &filename, &item.ObjectKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if filename != nil {
		item.Filename = *filename
	}
	return &item, nil
}

func (p *postgresStore) AcceptJob(ctx context.Context, item job) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var owner string
	err = tx.QueryRow(ctx, `SELECT user_id FROM `+p.packs+` WHERE pack_id = $1 FOR UPDATE`, item.PackID).Scan(&owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return errNotFound
	}
	if err != nil {
		return err
	}
	if owner != item.UserID {
		return errNotFound
	}
	var materialOwner, materialPack string
	err = tx.QueryRow(ctx, `SELECT user_id, pack_id FROM `+p.materials+` WHERE material_id = $1`, item.MaterialID).Scan(&materialOwner, &materialPack)
	if errors.Is(err, pgx.ErrNoRows) {
		return errNotFound
	}
	if err != nil {
		return err
	}
	if materialOwner != item.UserID || materialPack != item.PackID {
		return errNotFound
	}
	var active int
	err = tx.QueryRow(ctx, `SELECT count(*) FROM `+p.jobs+` WHERE pack_id = $1 AND status IN ('accepted', 'running')`, item.PackID).Scan(&active)
	if err != nil {
		return err
	}
	if active > 0 {
		return errConflict
	}
	_, err = tx.Exec(ctx, `INSERT INTO `+p.jobs+` (job_id, pack_id, material_id, user_id, source, status) VALUES ($1, $2, $3, $4, $5, $6)`,
		item.ID, item.PackID, item.MaterialID, item.UserID, item.Source, item.Status)
	if err != nil {
		return mapWriteErr(err)
	}
	return tx.Commit(ctx)
}

func (p *postgresStore) DeleteJob(ctx context.Context, id string) error {
	_, err := p.pool.Exec(ctx, `DELETE FROM `+p.jobs+` WHERE job_id = $1 AND status = 'accepted'`, id)
	return err
}

func (p *postgresStore) FindJob(ctx context.Context, id string) (*job, error) {
	row := p.pool.QueryRow(ctx, `SELECT job_id, pack_id, material_id, user_id, source, status, progress_message, failure_code, failure_message FROM `+p.jobs+` WHERE job_id = $1`, id)
	return scanJob(row)
}

func (p *postgresStore) UpdateJobStatus(ctx context.Context, id, status string) (*job, error) {
	row := p.pool.QueryRow(ctx, `UPDATE `+p.jobs+` SET status = $2, updated_at = now() WHERE job_id = $1 RETURNING job_id, pack_id, material_id, user_id, source, status, progress_message, failure_code, failure_message`, id, status)
	item, err := scanJob(row)
	if err != nil {
		return nil, mapWriteErr(err)
	}
	if item == nil {
		return nil, errNotFound
	}
	return item, nil
}

func scanJob(row pgx.Row) (*job, error) {
	var item job
	var progress, failureCode, failureMessage *string
	err := row.Scan(&item.ID, &item.PackID, &item.MaterialID, &item.UserID, &item.Source, &item.Status, &progress, &failureCode, &failureMessage)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if progress != nil {
		item.ProgressMessage = *progress
	}
	if failureCode != nil {
		item.FailureCode = *failureCode
	}
	if failureMessage != nil {
		item.FailureMessage = *failureMessage
	}
	return &item, nil
}

func (p *postgresStore) CountJobs(ctx context.Context, userID string) (int, error) {
	var n int
	err := p.pool.QueryRow(ctx, `SELECT count(*) FROM `+p.jobs+` WHERE user_id = $1`, userID).Scan(&n)
	return n, err
}

func (p *postgresStore) RecordSkip(ctx context.Context, userID string, at time.Time) error {
	_, err := p.pool.Exec(ctx, `INSERT INTO `+p.skips+` (user_id, skipped_at) VALUES ($1, $2) ON CONFLICT (user_id) DO NOTHING`, userID, at)
	return err
}

func (p *postgresStore) SkipRecorded(ctx context.Context, userID string) (bool, error) {
	var found int
	err := p.pool.QueryRow(ctx, `SELECT 1 FROM `+p.skips+` WHERE user_id = $1`, userID).Scan(&found)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (p *postgresStore) SaveIdempotency(ctx context.Context, userID, op, key string, row idemRow) error {
	var details []byte
	var err error
	if row.ErrDetails != nil {
		details, err = json.Marshal(row.ErrDetails)
		if err != nil {
			return err
		}
	}
	var jobID any
	if row.JobID != "" {
		jobID = row.JobID
	}
	var errCode any
	if row.ErrCode != "" {
		errCode = row.ErrCode
	}
	_, err = p.pool.Exec(ctx, `INSERT INTO `+p.idem+` (user_id, op, idem_key, body_hash, job_id, err_code, err_details, expires_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		userID, op, key, row.Hash[:], jobID, errCode, details, row.Expires)
	return err
}

func (p *postgresStore) FindIdempotency(ctx context.Context, userID, op, key string, now time.Time) (*idemRow, error) {
	var hash []byte
	var jobID, errCode *string
	var details []byte
	var expires time.Time
	err := p.pool.QueryRow(ctx, `SELECT body_hash, job_id, err_code, err_details, expires_at FROM `+p.idem+` WHERE user_id = $1 AND op = $2 AND idem_key = $3`, userID, op, key).
		Scan(&hash, &jobID, &errCode, &details, &expires)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !expires.After(now) {
		_, _ = p.pool.Exec(ctx, `DELETE FROM `+p.idem+` WHERE user_id = $1 AND op = $2 AND idem_key = $3`, userID, op, key)
		return nil, nil
	}
	if len(hash) != 32 {
		return nil, errors.New("idempotency hash")
	}
	row := &idemRow{Expires: expires}
	copy(row.Hash[:], hash)
	if jobID != nil {
		row.JobID = *jobID
	}
	if errCode != nil {
		row.ErrCode = *errCode
	}
	if len(details) > 0 {
		if err := json.Unmarshal(details, &row.ErrDetails); err != nil {
			return nil, err
		}
	}
	return row, nil
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func mapWriteErr(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && (pgErr.Code == "23505" || pgErr.Code == "23503") {
		if pgErr.Code == "23503" {
			return errNotFound
		}
		return errConflict
	}
	return err
}

func redactStoreErr(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	lower := strings.ToLower(msg)
	if strings.Contains(msg, "://") || strings.Contains(lower, "password") || strings.Contains(lower, "qlutch-local-dev") {
		return errors.New("knowledge store unavailable")
	}
	return err
}
