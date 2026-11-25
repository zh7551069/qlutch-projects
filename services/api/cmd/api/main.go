package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/zh7551069/qlutch-projects/services/airuntime"
	"github.com/zh7551069/qlutch-projects/services/api/internal/auth"
	"github.com/zh7551069/qlutch-projects/services/api/internal/config"
	"github.com/zh7551069/qlutch-projects/services/api/internal/identity"
	"github.com/zh7551069/qlutch-projects/services/api/internal/ready"
	"github.com/zh7551069/qlutch-projects/services/api/internal/server"
	"github.com/zh7551069/qlutch-projects/services/api/internal/telemetry"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("api stopped", "err", err.Error())
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdown, err := telemetry.Install(context.Background(), cfg.OTLPEndpoint, "qlutch-api")
	if err != nil {
		log.Warn("telemetry export disabled", "reason", "otlp exporter was not created")
		shutdown = func(context.Context) error { return nil }
	}
	defer func() {
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdown(shutCtx)
	}()

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Error("database", "reason", "invalid database url")
		return errors.New("database url is invalid")
	}
	defer pool.Close()
	migCtx, migCancel := context.WithTimeout(context.Background(), 10*time.Second)
	store, err := identity.OpenPostgres(migCtx, pool, "")
	migCancel()
	if err != nil {
		log.Error("identity", "reason", "schema was not created")
		return errors.New("identity store unavailable")
	}
	checks, closeDeps := openChecks(cfg, pool)
	defer closeDeps()

	rt := airuntime.New()
	log.Info("ai runtime", "status", rt.Status())
	learner, err := identity.New(identity.Options{Store: store})
	if err != nil {
		return err
	}
	log.Info("identity", "learner", "otp-session", "sms", "mock", "store", "postgres", "internal", string(auth.DomainInternal))

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           server.New(checks, learner).Handler(cfg.CORSOrigins),
		ReadHeaderTimeout: 5 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		log.Info("api listening", "addr", cfg.HTTPAddr, "env", cfg.EnvName)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
		return <-errCh
	case err := <-errCh:
		return err
	}
}

type quietRedisLog struct{}

func (quietRedisLog) Printf(context.Context, string, ...interface{}) {}

func openChecks(cfg config.Config, pool *pgxpool.Pool) ([]ready.Pinger, func()) {
	redis.SetLogger(quietRedisLog{})
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	checks := []ready.Pinger{
		ready.Postgres{Pool: pool},
		ready.Redis{Client: redisClient},
		ready.ObjectStore{
			Endpoint:  cfg.S3Endpoint,
			Region:    cfg.S3Region,
			Bucket:    cfg.S3Bucket,
			AccessKey: cfg.S3AccessKey,
			Secret:    cfg.S3SecretKey,
		},
	}
	return checks, func() { _ = redisClient.Close() }
}
