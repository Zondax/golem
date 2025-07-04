package factory

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/zobservability"
	"github.com/zondax/golem/pkg/zobservability/providers/sentry"
	"github.com/zondax/golem/pkg/zobservability/providers/signoz"
)

type observerFactory func(config *zobservability.Config, serviceName string) (zobservability.Observer, error)

var providers = map[string]observerFactory{
	zobservability.ProviderSentry: newSentryObserver,
	zobservability.ProviderSigNoz: newSigNozObserver,
}

// NewObserver creates a new observer based on the provider and config
func NewObserver(config *zobservability.Config, serviceName string) (zobservability.Observer, error) {
	log := logger.NewLogger()

	log.Infof("Initializing observability - Provider: %s, Service: %s, Enabled: %t",
		config.Provider, serviceName, config.Enabled)

	if !config.Enabled {
		log.Infof("Observability disabled - using no-op observer")
		return zobservability.NewNoopObserver(), nil
	}

	factory, ok := providers[config.Provider]
	if !ok {
		log.Errorf("Unsupported observability provider: %s", config.Provider)
		return nil, fmt.Errorf("unsupported observability provider: %s", config.Provider)
	}

	observer, err := factory(config, serviceName)
	if err != nil {
		log.Errorf("Failed to initialize %s observer: %v", config.Provider, err)
		return nil, err
	}

	log.Infof("%s observability initialized successfully - Endpoint: %s",
		config.Provider, config.Address)

	return observer, nil
}

// Provider-specific factory functions

// newSentryObserver creates a Sentry observer with the provided configuration
func newSentryObserver(config *zobservability.Config, serviceName string) (zobservability.Observer, error) {
	sentryConfig := &sentry.Config{
		DSN:           config.Address,
		Environment:   config.Environment,
		Release:       config.Release,
		Debug:         config.Debug,
		ServiceName:   serviceName,
		SampleRate:    config.SampleRate,
		CaptureErrors: config.Middleware.CaptureErrors,
	}

	return sentry.NewObserver(sentryConfig)
}

// newSigNozObserver creates a SigNoz observer with the provided configuration
// Parses custom_config for SigNoz-specific settings using constants to avoid hardcoded strings
func newSigNozObserver(config *zobservability.Config, serviceName string) (zobservability.Observer, error) {
	// Parse headers from custom config using constants
	// Headers are identified by the "header_" prefix and converted to HTTP headers
	// Example: "header_signoz-access-token" becomes header "signoz-access-token"
	headers := make(map[string]string)
	for key, value := range config.CustomConfig {
		if strings.HasPrefix(key, signoz.ConfigKeyHeaderPrefix) {
			headerKey := strings.TrimPrefix(key, signoz.ConfigKeyHeaderPrefix)
			headers[headerKey] = value
		}
	}

	// Check if insecure mode is enabled using constants
	// Insecure mode disables TLS for development environments
	insecure := false
	if insecureStr, ok := config.CustomConfig[signoz.ConfigKeyInsecure]; ok {
		insecure = strings.ToLower(insecureStr) == signoz.ConfigKeyTrueValue
	}

	// Check if parent sampling should be ignored
	// DEFAULT: true - This fixes trace loss in Google Cloud Run and other cloud environments
	// where trace headers are automatically injected with sampling decisions
	ignoreParentSampling := true // Changed from false to true as default
	if ignoreParentStr, ok := config.CustomConfig[signoz.ConfigKeyIgnoreParentSampling]; ok {
		ignoreParentSampling = strings.ToLower(ignoreParentStr) == signoz.ConfigKeyTrueValue
	}

	// Parse SimpleSpan configuration
	useSimpleSpan := false
	if simpleSpanStr, ok := config.CustomConfig[signoz.ConfigKeyUseSimpleSpan]; ok {
		useSimpleSpan = strings.ToLower(simpleSpanStr) == signoz.ConfigKeyTrueValue
	}

	// Parse advanced batch configuration if present
	// BatchConfig controls performance and batching behavior
	batchConfig, err := parseBatchConfig(config.CustomConfig)
	if err != nil {
		return nil, err
	}

	// Parse advanced resource configuration if present
	// ResourceConfig controls what metadata is attached to traces
	var resourceConfig *signoz.ResourceConfig
	if resourceConfigRaw, ok := config.CustomConfig[signoz.ConfigKeyResourceConfig]; ok {
		resourceConfig = &signoz.ResourceConfig{}
		if err := mapstructure.Decode(resourceConfigRaw, resourceConfig); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", signoz.ConfigKeyResourceConfig, err)
		}
	}

	// Create SigNoz configuration with all parsed settings
	signozConfig := &signoz.Config{
		Endpoint:             config.Address,
		ServiceName:          serviceName,
		Environment:          config.Environment,
		Release:              config.Release,
		Debug:                config.Debug,
		Insecure:             insecure,
		Headers:              headers,
		SampleRate:           config.SampleRate,
		BatchConfig:          batchConfig,          // Optional: nil means use defaults
		ResourceConfig:       resourceConfig,       // Optional: nil means use defaults
		IgnoreParentSampling: ignoreParentSampling, // Critical for Google Cloud Run deployments
		UseSimpleSpan:        useSimpleSpan,        // Enable immediate span export
		Propagation:          config.Propagation,   // Copy propagation configuration
	}

	return signoz.NewObserver(signozConfig)
}

// Helper function to parse batch configuration
func parseBatchConfig(customConfig map[string]string) (*signoz.BatchConfig, error) {
	// Priority 1: Check for custom batch configuration first
	if batchConfigRaw, exists := customConfig[signoz.ConfigKeyBatchConfig]; exists {
		batchConfig := &signoz.BatchConfig{}
		if err := mapstructure.Decode(batchConfigRaw, batchConfig); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", signoz.ConfigKeyBatchConfig, err)
		}
		return batchConfig, nil
	}

	// Priority 2: If no custom config, check for predefined profile
	if batchProfile, exists := customConfig[signoz.ConfigKeyBatchProfile]; exists {
		// Predefined batch profile provided - use standardized configuration
		return signoz.GetBatchProfileConfig(batchProfile), nil
	}

	// Priority 3: If neither is provided, return nil (factory will use defaults)
	return nil, nil
}
