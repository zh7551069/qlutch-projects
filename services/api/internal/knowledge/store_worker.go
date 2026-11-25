package knowledge

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (p *postgresStore) ClaimJob(ctx context.Context, id string) (*job, bool, error) {
	row := p.pool.QueryRow(ctx, `UPDATE `+p.jobs+` SET status = 'running', updated_at = now() WHERE job_id = $1 AND status = 'accepted' RETURNING job_id, pack_id, material_id, user_id, source, status, progress_message, failure_code, failure_message`, id)
	item, err := scanJob(row)
	if err != nil {
		return nil, false, err
	}
	if item != nil {
		return item, false, nil
	}
	item, err = p.FindJob(ctx, id)
	if err != nil || item == nil {
		return nil, false, err
	}
	switch item.Status {
	case statusRunning:
		return item, false, nil
	case statusMapReady, statusFailedRetryable, statusFailedTerminal:
		return item, true, nil
	default:
		return nil, false, errStorage
	}
}

func (p *postgresStore) FailJob(ctx context.Context, id, status, message string) error {
	if status != statusFailedRetryable && status != statusFailedTerminal {
		return errConflict
	}
	tag, err := p.pool.Exec(ctx, `UPDATE `+p.jobs+` SET status = $2, failure_code = NULL, failure_message = $3, updated_at = now() WHERE job_id = $1 AND status IN ('accepted', 'running')`, id, status, message)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return errConflict
	}
	return nil
}

func (p *postgresStore) CommitMapReady(ctx context.Context, jobID, userID string, nodes []mapNode) error {
	if !enterableNodes(nodes) {
		return errConflict
	}
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var owner, status, packID string
	err = tx.QueryRow(ctx, `SELECT user_id, status, pack_id FROM `+p.jobs+` WHERE job_id = $1 FOR UPDATE`, jobID).Scan(&owner, &status, &packID)
	if errors.Is(err, pgx.ErrNoRows) {
		return errNotFound
	}
	if err != nil {
		return err
	}
	if owner != userID || status != statusRunning {
		return errConflict
	}
	if _, err = tx.Exec(ctx, `DELETE FROM `+p.nodes+` WHERE pack_id = $1`, packID); err != nil {
		return err
	}
	for _, node := range nodes {
		var parent any
		if node.ParentID != "" {
			parent = node.ParentID
		}
		_, err = tx.Exec(ctx, `INSERT INTO `+p.nodes+` (node_id, pack_id, job_id, parent_id, kind, title, position) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			node.ID, packID, jobID, parent, node.Kind, node.Title, node.Position)
		if err != nil {
			return err
		}
	}
	var ready int
	err = tx.QueryRow(ctx, `SELECT count(*) FROM `+p.nodes+` c
		JOIN `+p.nodes+` h ON h.parent_id = c.node_id AND h.kind = 'chapter' AND h.job_id = c.job_id
		JOIN `+p.nodes+` pt ON pt.parent_id = h.node_id AND pt.kind = 'point' AND pt.job_id = h.job_id
		WHERE c.job_id = $1 AND c.kind = 'catalog'`, jobID).Scan(&ready)
	if err != nil {
		return err
	}
	if ready < 1 {
		return errConflict
	}
	tag, err := tx.Exec(ctx, `UPDATE `+p.jobs+` SET status = 'map_ready', failure_code = NULL, failure_message = NULL, updated_at = now() WHERE job_id = $1 AND status = 'running'`, jobID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return errConflict
	}
	tag, err = tx.Exec(ctx, `UPDATE `+p.accounts+` SET has_first_pack = TRUE WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return errNotFound
	}
	return tx.Commit(ctx)
}

func (p *postgresStore) MapNodes(ctx context.Context, packID string) ([]mapNode, error) {
	rows, err := p.pool.Query(ctx, `SELECT node_id, pack_id, job_id, parent_id, kind, title, position FROM `+p.nodes+` WHERE pack_id = $1 ORDER BY position, node_id`, packID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []mapNode
	for rows.Next() {
		var node mapNode
		var parent *string
		if err := rows.Scan(&node.ID, &node.PackID, &node.JobID, &parent, &node.Kind, &node.Title, &node.Position); err != nil {
			return nil, err
		}
		if parent != nil {
			node.ParentID = *parent
		}
		out = append(out, node)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (p *postgresStore) InsertPrefetch(ctx context.Context, rowsIn []prefetchRow) error {
	for _, row := range rowsIn {
		_, err := p.pool.Exec(ctx, `INSERT INTO `+p.prefetch+` (prefetch_id, pack_id, job_id, node_id, position, stem) VALUES ($1, $2, $3, $4, $5, $6)`,
			row.ID, row.PackID, row.JobID, row.NodeID, row.Position, row.Stem)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *postgresStore) CountPrefetch(ctx context.Context, jobID string) (int, error) {
	var n int
	err := p.pool.QueryRow(ctx, `SELECT count(*) FROM `+p.prefetch+` WHERE job_id = $1`, jobID).Scan(&n)
	return n, err
}
