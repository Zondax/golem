package cli

import (
	"context"
	"testing"

	"github.com/zondax/golem/pkg/secrets"
)

type mockProvider struct {
	id string
}

func (m mockProvider) IsSecretKey(_ context.Context, _ string) bool          { return false }
func (m mockProvider) GetSecret(_ context.Context, _ string) (string, error) { return "", nil }

func TestWithSecretProviders_AddsProviders(t *testing.T) {
	opt := WithSecretProviders(mockProvider{"a"}, mockProvider{"b"})
	var opts loadConfigOptions
	opt.apply(&opts)
	if len(opts.secretProviders) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(opts.secretProviders))
	}
	if mp, ok := opts.secretProviders[0].(mockProvider); !ok || mp.id != "a" {
		t.Errorf("first provider not set correctly")
	}
	if mp, ok := opts.secretProviders[1].(mockProvider); !ok || mp.id != "b" {
		t.Errorf("second provider not set correctly")
	}
}

// This test does not check the global effect of RegisterSecretProviders, only that it does not panic and accepts providers.
func TestRegisterSecretProviders_DoesNotPanic(t *testing.T) {
	opts := loadConfigOptions{
		secretProviders: []secrets.SecretProvider{
			mockProvider{"x"}, mockProvider{"y"},
		},
	}
	// Should execute without panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RegisterSecretProviders panicked: %v", r)
		}
	}()
	opts.RegisterSecretProviders()
}

func TestWithSecretProviders_NoSideEffects(t *testing.T) {
	_ = WithSecretProviders(mockProvider{"z"})
	// There is no way to check global effects here, only that it does not panic
}
