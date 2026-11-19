package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zh7551069/qlutch-projects/services/airuntime"
	"github.com/zh7551069/qlutch-projects/services/worker/internal/config"
	"github.com/zh7551069/qlutch-projects/services/worker/internal/queue"
)

type quietRedisLog struct{}

func (quietRedisLog) Printf(context.Context, string, ...interface{}) {}

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	runtime := airuntime.New()
	log.Info("ai runtime", "status", runtime.Status())

	redis.SetLogger(quietRedisLog{})
	client := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer func() { _ = client.Close() }()

	err := queue.Run(ctx, queue.Options{
		Heartbeat: cfg.Heartbeat,
		Log:       log,
		Pop: func(ctx context.Context, timeout time.Duration) ([]byte, bool, error) {
			values, popErr := client.BRPop(ctx, timeout, cfg.QueueKey).Result()
			if errors.Is(popErr, redis.Nil) {
				return nil, false, nil
			}
			if popErr != nil {
				return nil, false, popErr
			}
			if len(values) != 2 {
				return nil, false, nil
			}
			return []byte(values[1]), true, nil
		},
		Handle: func(context.Context, []byte) error { return nil },
		OnRetry: func(_ context.Context, attempt int, err error) {
			log.Warn("retry hook", "attempt", attempt, "err", err.Error())
		},
	})
	if err != nil {
		log.Error("worker stopped", "err", err.Error())
		os.Exit(1)
	}
}
