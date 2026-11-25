package identity

import (
	"sync"
	"time"
)

// Clock supplies the time used for expiry and rate limits.
type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

// ManualClock is a test clock.
type ManualClock struct {
	mu sync.Mutex
	at time.Time
}

// NewManualClock starts a clock at at.
func NewManualClock(at time.Time) *ManualClock {
	return &ManualClock{at: at}
}

// Now implements Clock.
func (c *ManualClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.at
}

// Advance moves the clock forward.
func (c *ManualClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.at = c.at.Add(d)
}
