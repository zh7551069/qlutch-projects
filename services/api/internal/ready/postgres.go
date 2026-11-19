package ready

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres pings a PostgreSQL pool. Connection errors are replaced so a DSN
// cannot leak into the response or the log line built from Ping.
type Postgres struct {
	Pool *pgxpool.Pool
}

// Name returns the probe name.
func (p Postgres) Name() string { return "database" }

// Ping checks that the pool can reach PostgreSQL.
func (p Postgres) Ping(ctx context.Context) error {
	if p.Pool == nil {
		return errors.New("database unreachable")
	}
	if err := p.Pool.Ping(ctx); err != nil {
		return errors.New("database unreachable")
	}
	return nil
}
