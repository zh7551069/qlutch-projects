package ready

import "context"

// Pinger is one dependency the readiness probe can reach.
type Pinger interface {
	Name() string
	Ping(ctx context.Context) error
}

// Static is a checker that always fails. It keeps the process servable when
// configuration cannot even build a client.
type Static struct {
	NameValue string
	Err       error
}

// Name returns the probe name.
func (s Static) Name() string { return s.NameValue }

// Ping returns the stored error.
func (s Static) Ping(context.Context) error { return s.Err }
