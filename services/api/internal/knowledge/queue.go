package knowledge

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"
)

// MemoryQueue records payloads in this process. Tests use it. It is not an HTTP route.
type MemoryQueue struct {
	mu    sync.Mutex
	items [][]byte
	Err   error
}

// NewMemoryQueue returns an empty queue.
func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{}
}

// Enqueue implements Enqueuer.
func (q *MemoryQueue) Enqueue(_ context.Context, payload []byte) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.Err != nil {
		return q.Err
	}
	q.items = append(q.items, append([]byte(nil), payload...))
	return nil
}

// Items returns a copy of the recorded payloads.
func (q *MemoryQueue) Items() [][]byte {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([][]byte, len(q.items))
	for i, item := range q.items {
		out[i] = append([]byte(nil), item...)
	}
	return out
}

// RedisQueue pushes opaque payloads onto a Redis list.
// The worker pops that list and runs ProcessQueued. The queue itself does not
// change import job status.
type RedisQueue struct {
	Client *redis.Client
	Key    string
}

// Enqueue implements Enqueuer.
func (q RedisQueue) Enqueue(ctx context.Context, payload []byte) error {
	if q.Client == nil || q.Key == "" {
		return errQueue
	}
	if err := q.Client.LPush(ctx, q.Key, string(payload)).Err(); err != nil {
		return errQueue
	}
	return nil
}
