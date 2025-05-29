package zdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/zdb/zdbconfig"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

// instrumentationManager handles OpenTelemetry instrumentation for GORM
type instrumentationManager struct {
	config zdbconfig.OpenTelemetryConfig
}

// setupOpenTelemetryInstrumentation configures and adds OpenTelemetry instrumentation to GORM
func setupOpenTelemetryInstrumentation(db *gorm.DB, config *zdbconfig.OpenTelemetryConfig) error {
	manager := &instrumentationManager{
		config: getConfigWithDefaults(config),
	}

	return manager.instrumentDatabase(db)
}

// instrumentDatabase applies OpenTelemetry instrumentation to the database
func (im *instrumentationManager) instrumentDatabase(db *gorm.DB) error {
	if !im.config.Enabled {
		logger.GetLoggerFromContext(context.Background()).Debug("OpenTelemetry database instrumentation is disabled")
		return nil
	}

	// Build instrumentation options
	opts := im.buildInstrumentationOptions()

	// Create and register the plugin
	plugin := otelgorm.NewPlugin(opts...)
	if err := db.Use(plugin); err != nil {
		return fmt.Errorf("failed to add OpenTelemetry plugin to database: %w", err)
	}

	logger.GetLoggerFromContext(context.Background()).Info("OpenTelemetry database instrumentation enabled with custom configuration")
	return nil
}

// buildInstrumentationOptions constructs OpenTelemetry plugin options based on configuration
func (im *instrumentationManager) buildInstrumentationOptions() []otelgorm.Option {
	var opts []otelgorm.Option

	// Configure query parameter inclusion
	opts = append(opts, im.getQueryParameterOptions()...)

	// Configure query formatter
	opts = append(opts, im.getQueryFormatterOptions()...)

	// Configure default attributes
	opts = append(opts, im.getDefaultAttributeOptions()...)

	// Configure metrics
	opts = append(opts, im.getMetricsOptions()...)

	return opts
}

// getQueryParameterOptions returns options for query parameter handling
func (im *instrumentationManager) getQueryParameterOptions() []otelgorm.Option {
	if !im.config.IncludeQueryParameters {
		return []otelgorm.Option{otelgorm.WithoutQueryVariables()}
	}
	return []otelgorm.Option{}
}

// getQueryFormatterOptions returns options for query formatting
func (im *instrumentationManager) getQueryFormatterOptions() []otelgorm.Option {
	formatter := im.createQueryFormatter()
	if formatter != nil {
		return []otelgorm.Option{otelgorm.WithQueryFormatter(formatter)}
	}
	return []otelgorm.Option{}
}

// createQueryFormatter creates a query formatter function based on configuration
func (im *instrumentationManager) createQueryFormatter() func(string) string {
	switch strings.ToLower(im.config.QueryFormatter) {
	case zdbconfig.QueryFormatterUpper:
		return strings.ToUpper
	case zdbconfig.QueryFormatterLower:
		return strings.ToLower
	case zdbconfig.QueryFormatterNone:
		return func(string) string { return "[QUERY HIDDEN]" }
	case zdbconfig.QueryFormatterDefault, "":
		return nil // Use default formatter
	default:
		logger.GetLoggerFromContext(context.Background()).Warnf("Unknown query formatter: %s, using default", im.config.QueryFormatter)
		return nil
	}
}

// getDefaultAttributeOptions returns options for default attributes
func (im *instrumentationManager) getDefaultAttributeOptions() []otelgorm.Option {
	if len(im.config.DefaultAttributes) == 0 {
		return []otelgorm.Option{}
	}

	attrs := make([]attribute.KeyValue, 0, len(im.config.DefaultAttributes))
	for key, value := range im.config.DefaultAttributes {
		attrs = append(attrs, attribute.String(key, value))
	}

	return []otelgorm.Option{otelgorm.WithAttributes(attrs...)}
}

// getMetricsOptions returns options for metrics configuration
func (im *instrumentationManager) getMetricsOptions() []otelgorm.Option {
	if im.config.DisableMetrics {
		return []otelgorm.Option{otelgorm.WithoutMetrics()}
	}
	return []otelgorm.Option{}
}

// getConfigWithDefaults returns configuration with sensible defaults applied
func getConfigWithDefaults(userConfig *zdbconfig.OpenTelemetryConfig) zdbconfig.OpenTelemetryConfig {
	// If no config provided, return disabled configuration
	if userConfig == nil {
		return zdbconfig.OpenTelemetryConfig{
			Enabled: false,
		}
	}

	// Start with user configuration
	config := *userConfig

	// Apply defaults only for empty/zero values
	if config.QueryFormatter == "" {
		config.QueryFormatter = zdbconfig.QueryFormatterDefault
	}

	if config.DefaultAttributes == nil {
		config.DefaultAttributes = make(map[string]string)
	}

	return config
}
