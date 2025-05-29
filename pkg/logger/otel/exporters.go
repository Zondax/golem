package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"

	"github.com/zondax/golem/pkg/logger"
)

// createExporter creates the appropriate OTLP exporter based on protocol
func (p *Provider) createExporter(config *logger.OpenTelemetryConfig) (log.Exporter, error) {
	// Validate endpoint is not empty
	if config.Endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}

	protocol := p.getProtocol(config)

	switch protocol {
	case ProtocolGRPC:
		return p.createGRPCExporter(config)
	case ProtocolHTTP:
		return p.createHTTPExporter(config)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// getProtocol returns the protocol to use, defaulting to HTTP if not specified
func (p *Provider) getProtocol(config *logger.OpenTelemetryConfig) string {
	if config.Protocol == "" {
		return ProtocolHTTP
	}
	return config.Protocol
}

// createHTTPExporter creates an HTTP OTLP exporter with the given configuration
func (p *Provider) createHTTPExporter(config *logger.OpenTelemetryConfig) (log.Exporter, error) {
	options := []otlploghttp.Option{
		otlploghttp.WithEndpoint(config.Endpoint),
	}

	if config.Insecure {
		options = append(options, otlploghttp.WithInsecure())
	}

	if len(config.Headers) > 0 {
		options = append(options, otlploghttp.WithHeaders(config.Headers))
	}

	return otlploghttp.New(context.Background(), options...)
}

// createGRPCExporter creates a gRPC OTLP exporter with the given configuration
func (p *Provider) createGRPCExporter(config *logger.OpenTelemetryConfig) (log.Exporter, error) {
	options := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(config.Endpoint),
	}

	if config.Insecure {
		options = append(options, otlploggrpc.WithInsecure())
	}

	if len(config.Headers) > 0 {
		options = append(options, otlploggrpc.WithHeaders(config.Headers))
	}

	return otlploggrpc.New(context.Background(), options...)
}
