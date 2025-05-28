package otel

import (
	"fmt"

	"go.opentelemetry.io/otel/sdk/log"

	"github.com/zondax/golem/pkg/logger"
)

// createLoggerProvider creates the OpenTelemetry log provider
func (p *Provider) createLoggerProvider(config *logger.OpenTelemetryConfig) (*log.LoggerProvider, error) {
	exporter, err := p.createExporter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
	}

	processor := log.NewBatchProcessor(exporter)
	provider := log.NewLoggerProvider(
		log.WithProcessor(processor),
	)

	return provider, nil
}
