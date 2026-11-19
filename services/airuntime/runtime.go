package airuntime

import (
	"context"
	"errors"
)

// ErrProviderNotConfigured is returned until a provider adapter is configured.
var ErrProviderNotConfigured = errors.New("未配置提供方")

// ErrSceneCodeRequired means a call arrived without the accounting scene code.
var ErrSceneCodeRequired = errors.New("缺少场景码")

// LedgerEntry is the accounting key for one model call.
// SceneCode is required. The remaining ledger fields belong to a later contract.
type LedgerEntry struct {
	SceneCode string
}

// CallResult is the empty result of a call that did not reach a provider.
type CallResult struct{}

// Runtime is the unified model entry used by the API and worker processes.
type Runtime struct{}

// New returns the scaffold runtime. It has no provider.
func New() *Runtime {
	return &Runtime{}
}

// Status reports the current provider state.
func (r *Runtime) Status() string {
	return ErrProviderNotConfigured.Error()
}

// Call is the only model entry. Accounting is keyed by scene code.
// No cache or tier routing is performed.
func (r *Runtime) Call(ctx context.Context, entry LedgerEntry) (CallResult, error) {
	if err := ctx.Err(); err != nil {
		return CallResult{}, err
	}
	if entry.SceneCode == "" {
		return CallResult{}, ErrSceneCodeRequired
	}
	return CallResult{}, ErrProviderNotConfigured
}
