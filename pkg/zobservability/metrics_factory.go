package zobservability

import (
	"fmt"
	"strings"
)

// MetricsProviderType represents the type of metrics provider
type MetricsProviderType string

const (
	MetricsProviderOpenTelemetry MetricsProviderType = "opentelemetry"
	MetricsProviderNoop          MetricsProviderType = "noop"
)

// NewMetricsProvider creates a new metrics provider based on the configuration
func NewMetricsProvider(name string, config MetricsConfig) (MetricsProvider, error) {
	if !config.Enabled {
		return NewNoopMetricsProvider(name), nil
	}

	providerType := MetricsProviderType(strings.ToLower(config.Provider))

	switch providerType {
	case MetricsProviderOpenTelemetry:
		return NewOpenTelemetryMetricsProvider(name, config)
	case MetricsProviderNoop:
		return NewNoopMetricsProvider(name), nil
	default:
		return nil, fmt.Errorf("unsupported metrics provider: %s", config.Provider)
	}
}
