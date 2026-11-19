package queue

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestWithRetryInvokesHookThenSucceeds(t *testing.T) {
	attempts := 0
	hooks := 0
	err := withRetry(context.Background(), 3, func(context.Context, int, error) {
		hooks++
	}, func(time.Duration) {}, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("not yet")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if attempts != 3 || hooks != 2 {
		t.Fatalf("attempts=%d hooks=%d", attempts, hooks)
	}
}

func TestWithRetryStopsAtMax(t *testing.T) {
	hooks := 0
	err := withRetry(context.Background(), 2, func(context.Context, int, error) {
		hooks++
	}, func(time.Duration) {}, func() error {
		return errors.New("failed")
	})
	if err == nil || hooks != 2 {
		t.Fatalf("err=%v hooks=%d", err, hooks)
	}
}

func TestRunAcknowledgesOpaquePayloadAndHeartbeats(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	pops := 0
	var got int
	logs := slog.New(slog.NewTextHandler(io.Discard, nil))
	err := Run(ctx, Options{
		Heartbeat:   20 * time.Millisecond,
		MaxAttempts: 2,
		Sleep:       func(time.Duration) {},
		Log:         logs,
		Pop: func(context.Context, time.Duration) ([]byte, bool, error) {
			pops++
			if pops == 1 {
				return []byte("opaque"), true, nil
			}
			time.Sleep(30 * time.Millisecond)
			if pops > 3 {
				cancel()
			}
			return nil, false, nil
		},
		Handle: func(_ context.Context, payload []byte) error {
			got = len(payload)
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != 6 {
		t.Fatalf("payload length was not observed: %d", got)
	}
}
