package ready

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

// Redis pings a Redis client.
type Redis struct {
	Client *redis.Client
}

// Name returns the probe name.
func (r Redis) Name() string { return "redis" }

// Ping checks that Redis answers PING.
func (r Redis) Ping(ctx context.Context) error {
	if r.Client == nil {
		return errors.New("redis unreachable")
	}
	if err := r.Client.Ping(ctx).Err(); err != nil {
		return errors.New("redis unreachable")
	}
	return nil
}
