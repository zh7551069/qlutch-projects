package airuntime

import (
	"context"
	"errors"
	"testing"
)

func TestCallRequiresSceneCodeAndProvider(t *testing.T) {
	rt := New()
	if rt.Status() != "未配置提供方" {
		t.Fatalf("status = %q", rt.Status())
	}

	_, err := rt.Call(context.Background(), LedgerEntry{})
	if !errors.Is(err, ErrSceneCodeRequired) {
		t.Fatalf("empty scene error = %v", err)
	}

	_, err = rt.Call(context.Background(), LedgerEntry{SceneCode: "scaffold"})
	if !errors.Is(err, ErrProviderNotConfigured) {
		t.Fatalf("provider error = %v", err)
	}
}
