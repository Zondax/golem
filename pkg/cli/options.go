package cli

import "github.com/zondax/golem/pkg/secrets"

// LoadConfigOption is a type-safe option for LoadConfig.
// Is on purpose not put just a function, because we want to avoid side effects.
type LoadConfigOption interface {
	apply(*loadConfigOptions)
}

// loadConfigOptions is a struct that contains the options for LoadConfig.
// It is not exported and should not be used directly.
type loadConfigOptions struct {
	secretProviders []secrets.SecretProvider
}

func (o *loadConfigOptions) RegisterSecretProviders() {
	for _, p := range o.secretProviders {
		secrets.RegisterProvider(p)
	}
}

// withSecretProvidersOption is an option that registers secret providers.
type withSecretProvidersOption struct {
	providers []secrets.SecretProvider
}

// apply adds the given secret providers to the options struct.
func (o withSecretProvidersOption) apply(opts *loadConfigOptions) {
	opts.secretProviders = append(opts.secretProviders, o.providers...)
}

// WithSecretProviders registers the given secret providers for secret resolution.
// Pass this option to LoadConfig to specify which providers to use.
func WithSecretProviders(providers ...secrets.SecretProvider) LoadConfigOption {
	return withSecretProvidersOption{providers: providers}
}
