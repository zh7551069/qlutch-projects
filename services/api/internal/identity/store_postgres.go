package identity

import (
	"context"
	_ "embed"
	"errors"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var schemaName = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

//go:embed schema.sql
var identitySchemaSQL string

// postgresStore keeps accounts, onboarding, preferences, and session hashes in PostgreSQL.
// It does not cache rows. A new process with the same database sees the same records.
type postgresStore struct {
	pool     *pgxpool.Pool
	accounts string
	answers  string
	sessions string
}

// OpenPostgres migrates identity tables into schema and returns a store.
// An empty schema uses public. Returned errors do not include a DSN or password.
func OpenPostgres(ctx context.Context, pool *pgxpool.Pool, schema string) (Store, error) {
	if pool == nil {
		return nil, errors.New("identity store unavailable")
	}
	if schema == "" {
		schema = "public"
	}
	if !schemaName.MatchString(schema) {
		return nil, errors.New("identity store unavailable")
	}
	ident := pgx.Identifier{schema}.Sanitize()
	store := &postgresStore{
		pool:     pool,
		accounts: ident + ".identity_accounts",
		answers:  ident + ".identity_onboarding_answers",
		sessions: ident + ".identity_sessions",
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
	if _, err := tx.Exec(ctx, identitySchemaSQL); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p *postgresStore) FindByPhone(ctx context.Context, region, national string) (*user, error) {
	row := p.pool.QueryRow(ctx, `SELECT user_id, phone_region, phone_national, locale, region_pref, has_first_pack FROM `+p.accounts+` WHERE phone_region = $1 AND phone_national = $2`, region, national)
	return p.scanAccount(ctx, row)
}

func (p *postgresStore) FindByID(ctx context.Context, id string) (*user, error) {
	row := p.pool.QueryRow(ctx, `SELECT user_id, phone_region, phone_national, locale, region_pref, has_first_pack FROM `+p.accounts+` WHERE user_id = $1`, id)
	return p.scanAccount(ctx, row)
}

func (p *postgresStore) scanAccount(ctx context.Context, row pgx.Row) (*user, error) {
	var account user
	err := row.Scan(&account.ID, &account.PhoneRegion, &account.PhoneNational, &account.Locale, &account.RegionPref, &account.HasFirstPack)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	account.Answers = map[string]string{}
	rows, err := p.pool.Query(ctx, `SELECT step_id, answer FROM `+p.answers+` WHERE user_id = $1`, account.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var step, answer string
		if err := rows.Scan(&step, &answer); err != nil {
			return nil, err
		}
		account.Answers[step] = answer
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &account, nil
}

func (p *postgresStore) Insert(ctx context.Context, account *user) error {
	_, err := p.pool.Exec(ctx, `INSERT INTO `+p.accounts+` (user_id, phone_region, phone_national, locale, region_pref, has_first_pack) VALUES ($1, $2, $3, $4, $5, $6)`,
		account.ID, account.PhoneRegion, account.PhoneNational, account.Locale, account.RegionPref, account.HasFirstPack)
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return errAccountExists
	}
	return err
}

func (p *postgresStore) DeleteAccount(ctx context.Context, id string) error {
	if _, err := p.pool.Exec(ctx, `DELETE FROM `+p.sessions+` WHERE user_id = $1`, id); err != nil {
		return err
	}
	if _, err := p.pool.Exec(ctx, `DELETE FROM `+p.answers+` WHERE user_id = $1`, id); err != nil {
		return err
	}
	_, err := p.pool.Exec(ctx, `DELETE FROM `+p.accounts+` WHERE user_id = $1`, id)
	return err
}

func (p *postgresStore) SaveProfile(ctx context.Context, id, locale, regionPref string) error {
	tag, err := p.pool.Exec(ctx, `UPDATE `+p.accounts+` SET locale = $2, region_pref = $3 WHERE user_id = $1`, id, locale, regionPref)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return errors.New("account missing")
	}
	return nil
}

func (p *postgresStore) SaveAnswer(ctx context.Context, id, step, answer string) error {
	tag, err := p.pool.Exec(ctx, `INSERT INTO `+p.answers+` (user_id, step_id, answer) VALUES ($1, $2, $3) ON CONFLICT (user_id, step_id) DO UPDATE SET answer = EXCLUDED.answer, updated_at = now()`, id, step, answer)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return errors.New("answer was not stored")
	}
	return nil
}

func (p *postgresStore) SaveFirstPack(ctx context.Context, id string, has bool) error {
	tag, err := p.pool.Exec(ctx, `UPDATE `+p.accounts+` SET has_first_pack = $2 WHERE user_id = $1`, id, has)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return errors.New("account missing")
	}
	return nil
}

func (p *postgresStore) InsertSession(ctx context.Context, sess *session) error {
	_, err := p.pool.Exec(ctx, `INSERT INTO `+p.sessions+` (session_id, user_id, access_hash, refresh_hash, access_expiry, refresh_expiry) VALUES ($1, $2, $3, $4, $5, $6)`,
		sess.ID, sess.UserID, sess.AccessHash[:], sess.RefreshHash[:], sess.AccessExpiry, sess.RefreshExpiry)
	return err
}

func (p *postgresStore) DeleteSession(ctx context.Context, id string) error {
	_, err := p.pool.Exec(ctx, `DELETE FROM `+p.sessions+` WHERE session_id = $1`, id)
	return err
}

func (p *postgresStore) FindSessionByAccess(ctx context.Context, hash [32]byte) (*session, error) {
	return p.findSession(ctx, `SELECT session_id, user_id, access_hash, refresh_hash, access_expiry, refresh_expiry FROM `+p.sessions+` WHERE access_hash = $1`, hash)
}

func (p *postgresStore) FindSessionByRefresh(ctx context.Context, hash [32]byte) (*session, error) {
	return p.findSession(ctx, `SELECT session_id, user_id, access_hash, refresh_hash, access_expiry, refresh_expiry FROM `+p.sessions+` WHERE refresh_hash = $1`, hash)
}

func (p *postgresStore) findSession(ctx context.Context, query string, hash [32]byte) (*session, error) {
	var sess session
	var access, refresh []byte
	err := p.pool.QueryRow(ctx, query, hash[:]).Scan(&sess.ID, &sess.UserID, &access, &refresh, &sess.AccessExpiry, &sess.RefreshExpiry)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(access) != 32 || len(refresh) != 32 {
		return nil, errors.New("session hash")
	}
	copy(sess.AccessHash[:], access)
	copy(sess.RefreshHash[:], refresh)
	return &sess, nil
}

func redactStoreErr(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	lower := strings.ToLower(msg)
	if strings.Contains(msg, "://") || strings.Contains(lower, "password") || strings.Contains(lower, "qlutch-local-dev") {
		return errors.New("identity store unavailable")
	}
	return err
}
