package queue

import (
	"context"
	"log/slog"
	"time"
)

// PopFunc blocks until a payload is available, the timeout elapses, or ctx ends.
// ok is false when the wait ended without a payload.
type PopFunc func(ctx context.Context, timeout time.Duration) (payload []byte, ok bool, err error)

// Handler handles one opaque payload. The shell handler does not parse jobs.
type Handler func(ctx context.Context, payload []byte) error

// RetryHook observes a failed attempt. attempt starts at 1.
type RetryHook func(ctx context.Context, attempt int, err error)

// Options configures the empty consume loop.
type Options struct {
	Heartbeat   time.Duration
	MaxAttempts int
	Pop         PopFunc
	Handle      Handler
	OnRetry     RetryHook
	Sleep       func(time.Duration)
	Log         *slog.Logger
}

// Run consumes the queue until ctx is cancelled. Payloads are opaque.
// Failures go through OnRetry and are not parsed into business jobs.
func Run(ctx context.Context, opt Options) error {
	if opt.Heartbeat <= 0 {
		opt.Heartbeat = 30 * time.Second
	}
	if opt.MaxAttempts < 1 {
		opt.MaxAttempts = 3
	}
	if opt.Sleep == nil {
		opt.Sleep = time.Sleep
	}
	if opt.Log == nil {
		opt.Log = slog.Default()
	}
	if opt.Handle == nil {
		opt.Handle = func(context.Context, []byte) error { return nil }
	}
	opt.Log.Info("worker started", "heartbeat", opt.Heartbeat.String())
	next := time.Now().Add(opt.Heartbeat)
	for {
		if ctx.Err() != nil {
			opt.Log.Info("worker stopped")
			return nil
		}
		if !time.Now().Before(next) {
			opt.Log.Info("worker heartbeat")
			next = time.Now().Add(opt.Heartbeat)
		}
		wait := time.Until(next)
		if wait > time.Second {
			wait = time.Second
		}
		if wait < 0 {
			wait = 0
		}
		payload, ok, err := opt.Pop(ctx, wait)
		if err != nil {
			if ctx.Err() != nil {
				opt.Log.Info("worker stopped")
				return nil
			}
			opt.Log.Warn("queue read failed")
			if opt.OnRetry != nil {
				opt.OnRetry(ctx, 1, err)
			}
			opt.Sleep(time.Second)
			continue
		}
		if !ok {
			continue
		}
		err = withRetry(ctx, opt.MaxAttempts, opt.OnRetry, opt.Sleep, func() error {
			return opt.Handle(ctx, payload)
		})
		if err != nil {
			opt.Log.Warn("job dropped after retry", "bytes", len(payload))
		} else {
			opt.Log.Info("job shell acknowledged", "bytes", len(payload))
		}
	}
}

func withRetry(ctx context.Context, max int, hook RetryHook, sleep func(time.Duration), fn func() error) error {
	var err error
	for attempt := 1; attempt <= max; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err = fn()
		if err == nil {
			return nil
		}
		if hook != nil {
			hook(ctx, attempt, err)
		}
		if attempt < max {
			sleep(time.Duration(attempt) * 100 * time.Millisecond)
		}
	}
	return err
}
