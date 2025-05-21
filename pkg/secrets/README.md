# secrets package

A pluggable system for resolving secrets from multiple providers (GCP, AWS, etc.) into your Viper configuration.

## Features
- Register one or more secret providers (GCP, AWS, Vault, ...)
- Automatically resolve and inject secrets into your Viper config
- Easy to extend with new providers

## Usage (Recommended)

Use the CLI config loader to automatically register providers and resolve secrets:

```go
import (
    "github.com/zondax/golem/pkg/cli"
    "github.com/zondax/golem/pkg/secrets/providers"
)

type MyConfig struct {
    // ... your config fields ...
}

func (c *MyConfig) SetDefaults() { /* ... */ }
func (c *MyConfig) Validate() error { return nil }

func main() {
    cfg, err := cli.LoadConfig[MyConfig](
        cli.WithSecretProviders(providers.GcpProvider{}),
    )
    if err != nil {
        // handle error
    }
    // Use cfg as usual
}
```

## Manual Usage (Advanced)

If you are not using the CLI config loader, you can register providers and resolve secrets manually:

```go
import (
    "github.com/zondax/golem/pkg/secrets"
    "github.com/zondax/golem/pkg/secrets/providers"
)

func main() {
    // Register the GCP provider (or others)
    secrets.RegisterProvider(providers.GcpProvider{})

    // Load your Viper config as usual
    // ...

    // Resolve secrets (replaces secret keys with their real values)
    secrets.ResolveSecrets()

    // Use your config as usual
    // ...
}
```

## Adding a new provider

1. Implement the `SecretProvider` interface:
   ```go
   type SecretProvider interface {
       IsSecretKey(ctx context.Context, key string) bool
       GetSecret(ctx context.Context, secretPath string) (string, error)
   }
   ```
2. Register your provider with `cli.WithSecretProviders(...)` (recommended) or `secrets.RegisterProvider(...)` before calling `ResolveSecrets()`.

## GCP Provider
See [`providers/gcp.go`](./providers/gcp.go) for the GCP Secret Manager implementation.
