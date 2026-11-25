// Package importworker is the production entry the worker process uses to run
// an accepted import. It is not an HTTP route. The provider is not installed here.
package importworker

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zh7551069/qlutch-projects/services/airuntime"
	"github.com/zh7551069/qlutch-projects/services/api/internal/identity"
	"github.com/zh7551069/qlutch-projects/services/api/internal/knowledge"
)

// Config is how the worker process reaches the stores. Secrets stay in the
// process environment. This struct is not serialized into logs.
type Config struct {
	DatabaseURL string
	Schema      string
	Runtime     *airuntime.Runtime
	S3Endpoint  string
	S3Region    string
	S3Bucket    string
	S3AccessKey string
	S3Secret    string

	pool  *pgxpool.Pool
	blobs knowledge.Blobs
}

// Processor runs queued import jobs through airuntime.
type Processor struct {
	svc   *knowledge.Service
	ident *identity.Service
	pool  *pgxpool.Pool
	owned bool
}

// Open prepares stores and a usage ledger. A nil runtime becomes an unconfigured
// airuntime. This function does not mark jobs map_ready by itself.
func Open(ctx context.Context, cfg Config) (*Processor, error) {
	if cfg.Runtime == nil {
		cfg.Runtime = airuntime.New()
	}
	pool := cfg.pool
	owned := false
	if pool == nil {
		if cfg.DatabaseURL == "" {
			return nil, errors.New("knowledge store unavailable")
		}
		opened, err := pgxpool.New(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, errors.New("knowledge store unavailable")
		}
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = opened.Ping(pingCtx)
		cancel()
		if err != nil {
			opened.Close()
			return nil, errors.New("knowledge store unavailable")
		}
		pool = opened
		owned = true
	}
	identStore, err := identity.OpenPostgres(ctx, pool, cfg.Schema)
	if err != nil {
		if owned {
			pool.Close()
		}
		return nil, errors.New("identity store unavailable")
	}
	knowStore, err := knowledge.OpenPostgres(ctx, pool, cfg.Schema)
	if err != nil {
		if owned {
			pool.Close()
		}
		return nil, errors.New("knowledge store unavailable")
	}
	learner, err := identity.New(identity.Options{Store: identStore})
	if err != nil {
		if owned {
			pool.Close()
		}
		return nil, errors.New("identity store unavailable")
	}
	if !cfg.Runtime.HasRecorder() {
		recorder, recErr := knowledge.NewUsageRecorder(pool, cfg.Schema)
		if recErr != nil {
			if owned {
				pool.Close()
			}
			return nil, errors.New("usage ledger unavailable")
		}
		cfg.Runtime.SetRecorder(recorder)
	}
	blobs := cfg.blobs
	if blobs == nil {
		blobs = knowledge.S3Store{
			Endpoint:  cfg.S3Endpoint,
			Region:    cfg.S3Region,
			Bucket:    cfg.S3Bucket,
			AccessKey: cfg.S3AccessKey,
			Secret:    cfg.S3Secret,
		}
	}
	svc, err := knowledge.New(knowledge.Options{
		Identity: learner,
		Store:    knowStore,
		Blobs:    blobs,
		Runtime:  cfg.Runtime,
	})
	if err != nil {
		if owned {
			pool.Close()
		}
		return nil, errors.New("knowledge store unavailable")
	}
	return &Processor{svc: svc, ident: learner, pool: pool, owned: owned}, nil
}

// Handle runs one queued payload. Business failure is settled on the job and
// returns nil so the queue shell does not invent a retry policy.
func (p *Processor) Handle(ctx context.Context, payload []byte) error {
	if p == nil || p.svc == nil {
		return errors.New("airuntime missing")
	}
	return p.svc.ProcessQueued(ctx, payload)
}

// Close releases a pool this processor opened.
func (p *Processor) Close() {
	if p != nil && p.owned && p.pool != nil {
		p.pool.Close()
	}
}
