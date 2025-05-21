// Package secrets provides a pluggable system for resolving secrets from multiple providers (GCP, AWS, etc.)
// into your Viper configuration. Register one or more providers, then call ResolveSecrets after loading your config.
//
// Example usage:
//
//	import (
//	    "github.com/zondax/golem/pkg/secrets"
//	    "github.com/zondax/golem/pkg/secrets/providers"
//	)
//
//	func main() {
//	    secrets.RegisterProvider(providers.GcpProvider{})
//	    // ... load your Viper config ...
//	    secrets.ResolveSecrets()
//	    // ... use your config as usual ...
//	}
package secrets

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/spf13/viper"
)

var providersMap sync.Map // map[string]SecretProvider

// SecretProvider defines the interface for secret providers (GCP, AWS, etc.)
type SecretProvider interface {
	IsSecretKey(ctx context.Context, key string) bool
	GetSecret(ctx context.Context, secretPath string) (string, error)
}

func RegisterProvider(p SecretProvider) {
	key := providerKey(p)
	providersMap.LoadOrStore(key, p)
}

func ResetProviders() {
	providersMap = sync.Map{}
}

// ResolveSecrets scans all Viper keys, and for each key that matches a provider,
// it fetches the secret and replaces the value in Viper.
func ResolveSecrets() {
	ctx := context.Background()
	for _, key := range viper.AllKeys() {
		resolveSecretForKey(ctx, key)
	}
}

// resolveSecretForKey checks all registered providers for a given key and replaces its value if a provider matches.
func resolveSecretForKey(ctx context.Context, key string) {
	for _, provider := range getAllProviders() {
		if provider.IsSecretKey(ctx, key) {
			secretPath := viper.GetString(key)
			secretValue, err := provider.GetSecret(ctx, secretPath)
			if err != nil {
				log.Printf("[secrets] Error resolving secret for key %s: %v", key, err)
				continue
			}
			viper.Set(key, secretValue)
		}
	}
}

// getAllProviders returns all registered providers as a slice.
func getAllProviders() []SecretProvider {
	var result []SecretProvider
	providersMap.Range(func(_, value interface{}) bool {
		result = append(result, value.(SecretProvider))
		return true
	})
	return result
}

// providerKey returns a unique key for a provider (by type).
func providerKey(p SecretProvider) string {
	return fmt.Sprintf("%T", p)
}
