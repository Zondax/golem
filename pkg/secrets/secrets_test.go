package secrets

import (
	"context"
	"sync"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// mockProvider is a test provider for secrets
type mockProvider struct {
	prefix string
	value  string
	fail   bool
}

func (m mockProvider) IsSecretKey(ctx context.Context, key string) bool {
	return len(key) >= len(m.prefix) && key[:len(m.prefix)] == m.prefix
}

func (m mockProvider) GetSecret(ctx context.Context, secretPath string) (string, error) {
	if m.fail {
		return "", context.DeadlineExceeded
	}
	return m.value, nil
}

func TestRegisterAndResolveSecrets(t *testing.T) {
	viper.Reset()
	ResetProviders()

	viper.Set("mock_secret_key", "mock_path")
	RegisterProvider(mockProvider{prefix: "mock_", value: "resolved_secret", fail: false})

	ResolveSecrets(context.Background())
	assert.Equal(t, "resolved_secret", viper.GetString("mock_secret_key"))
}

func TestResolveSecrets_Error(t *testing.T) {
	viper.Reset()
	ResetProviders()

	viper.Set("mock_secret_key", "mock_path")
	RegisterProvider(mockProvider{prefix: "mock_", value: "", fail: true})

	ResolveSecrets(context.Background())
	// Should not replace the value if provider fails
	assert.Equal(t, "mock_path", viper.GetString("mock_secret_key"))
}

func TestRegisterProvider_Duplicate(t *testing.T) {
	ResetProviders()
	p := mockProvider{prefix: "mock_", value: "foo"}
	RegisterProvider(p)
	RegisterProvider(p)
	providers := getAllProviders()
	assert.Len(t, providers, 1)
}

func TestRegisterProvider_ThreadSafe(t *testing.T) {
	ResetProviders()
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			RegisterProvider(mockProvider{prefix: "mock_", value: "foo"})
		}()
	}
	wg.Wait()
	providers := getAllProviders()
	assert.Len(t, providers, 1)
}

func TestResetProviders(t *testing.T) {
	ResetProviders()
	RegisterProvider(mockProvider{prefix: "mock_", value: "foo"})
	ResetProviders()
	providers := getAllProviders()
	assert.Len(t, providers, 0)
}
