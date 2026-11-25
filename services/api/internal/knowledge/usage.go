package knowledge

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zh7551069/qlutch-projects/services/airuntime"
)

// UsageRecorder appends S01 and S02 rows in PostgreSQL.
// OpenPostgres must already have created the table. Rows are not updated.
type UsageRecorder struct {
	pool  *pgxpool.Pool
	table string
}

// NewUsageRecorder returns a ledger for the schema OpenPostgres migrated.
func NewUsageRecorder(pool *pgxpool.Pool, schema string) (*UsageRecorder, error) {
	if pool == nil {
		return nil, errors.New("usage ledger unavailable")
	}
	if schema == "" {
		schema = "public"
	}
	if !schemaName.MatchString(schema) {
		return nil, errors.New("usage ledger unavailable")
	}
	return &UsageRecorder{pool: pool, table: pgx.Identifier{schema}.Sanitize() + ".knowledge_usage_ledger"}, nil
}

// Append implements airuntime.Recorder.
func (u *UsageRecorder) Append(ctx context.Context, entry airuntime.Entry) error {
	if u == nil || u.pool == nil {
		return errors.New("usage ledger unavailable")
	}
	_, err := u.pool.Exec(ctx, `INSERT INTO `+u.table+` (
		entry_id, scene_code, subclass, tier, cache_hit, success, retry_index, tokens,
		metric_kind, metric_value, pack_id, job_id, template_version, caller, fingerprint, created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`,
		entry.ID, entry.SceneCode, entry.Subclass, entry.Tier, entry.CacheHit, entry.Success, entry.RetryIndex, entry.Tokens,
		entry.MetricKind, entry.MetricValue, entry.PackID, entry.JobID, entry.TemplateVersion, entry.Caller, entry.Fingerprint, entry.At)
	return err
}

// List implements airuntime.Recorder.
func (u *UsageRecorder) List(ctx context.Context, query airuntime.Query) ([]airuntime.Entry, error) {
	if u == nil || u.pool == nil {
		return nil, errors.New("usage ledger unavailable")
	}
	sqlText := `SELECT entry_id, scene_code, subclass, tier, cache_hit, success, retry_index, tokens, metric_kind, metric_value, pack_id, job_id, template_version, caller, fingerprint, created_at FROM ` + u.table + ` WHERE 1=1`
	var args []any
	if query.SceneCode != "" {
		args = append(args, query.SceneCode)
		sqlText += fmt.Sprintf(" AND scene_code = $%d", len(args))
	}
	if query.Subclass != "" {
		args = append(args, query.Subclass)
		sqlText += fmt.Sprintf(" AND subclass = $%d", len(args))
	}
	if query.JobID != "" {
		args = append(args, query.JobID)
		sqlText += fmt.Sprintf(" AND job_id = $%d", len(args))
	}
	if query.CacheHitSet {
		args = append(args, query.CacheHit)
		sqlText += fmt.Sprintf(" AND cache_hit = $%d", len(args))
	}
	sqlText += ` ORDER BY created_at, entry_id`
	rows, err := u.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []airuntime.Entry
	for rows.Next() {
		var entry airuntime.Entry
		if err := rows.Scan(&entry.ID, &entry.SceneCode, &entry.Subclass, &entry.Tier, &entry.CacheHit, &entry.Success, &entry.RetryIndex, &entry.Tokens, &entry.MetricKind, &entry.MetricValue, &entry.PackID, &entry.JobID, &entry.TemplateVersion, &entry.Caller, &entry.Fingerprint, &entry.At); err != nil {
			return nil, err
		}
		out = append(out, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
