package otel

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	"github.com/zondax/golem/pkg/logger"
)

// createLoggerProvider creates the OpenTelemetry log provider with proper resource configuration
func (p *Provider) createLoggerProvider(config *logger.OpenTelemetryConfig) (*log.LoggerProvider, error) {
	exporter, err := p.createExporter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
	}

	res, err := p.createResource(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	processor := log.NewBatchProcessor(exporter)
	provider := log.NewLoggerProvider(
		log.WithProcessor(processor),
		log.WithResource(res),
	)

	return provider, nil
}

// createResource creates the OpenTelemetry resource with service identification attributes
// All attributes come from configuration, no environment variables are used
func (p *Provider) createResource(config *logger.OpenTelemetryConfig) (*resource.Resource, error) {
	attrs, err := p.buildResourceAttributes(config)
	if err != nil {
		return nil, err
	}

	return resource.NewWithAttributes(
		semconv.SchemaURL,
		attrs...,
	), nil
}

// buildResourceAttributes constructs all resource attributes from configuration
func (p *Provider) buildResourceAttributes(config *logger.OpenTelemetryConfig) ([]attribute.KeyValue, error) {
	var attrs []attribute.KeyValue

	// Service name is required
	if config.ServiceName == "" {
		return nil, fmt.Errorf("service name is required")
	}
	attrs = append(attrs, semconv.ServiceName(config.ServiceName))

	// Service version - use configured version or default
	serviceVersion := config.ServiceVersion
	if serviceVersion == "" {
		serviceVersion = "unknown"
	}
	attrs = append(attrs, semconv.ServiceVersion(serviceVersion))

	// Environment - only if specified in config
	if config.Environment != "" {
		attrs = append(attrs, semconv.DeploymentEnvironment(config.Environment))
	}

	// Hostname - only if specified in config
	if config.Hostname != "" {
		attrs = append(attrs, semconv.HostName(config.Hostname))
	}

	return attrs, nil
}
