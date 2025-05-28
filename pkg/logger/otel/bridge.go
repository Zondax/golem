package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zondax/golem/pkg/logger"
)

// Provider implements the OpenTelemetryProvider interface for OTLP logging
type Provider struct {
	loggerProvider *log.LoggerProvider
}

// NewProvider creates a new OpenTelemetry OTLP provider
func NewProvider() *Provider {
	return &Provider{}
}

// CreateLogger enhances an existing standard logger with OpenTelemetry integration
func (p *Provider) CreateLogger(config logger.Config, standardLogger *zap.Logger) (*zap.Logger, error) {
	if err := p.validateConfig(config); err != nil {
		return nil, err
	}

	provider, err := p.createLoggerProvider(config.OpenTelemetry)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenTelemetry provider: %w", err)
	}

	p.loggerProvider = provider

	// Parse the log level from config to get the minimum level
	logLevel, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		logLevel = zapcore.InfoLevel // Default to info if parsing fails
	}

	// Create OpenTelemetry core using the official bridge
	otelCore := otelzap.NewCore(config.OpenTelemetry.ServiceName,
		otelzap.WithLoggerProvider(provider),
	)

	// Create a level-filtering wrapper around the OpenTelemetry core
	filteredOtelCore := &levelFilterCore{
		Core:     otelCore,
		minLevel: logLevel,
	}

	// Combine the standard logger core with the level-filtered OpenTelemetry core
	combinedCore := zapcore.NewTee(standardLogger.Core(), filteredOtelCore)

	return p.createEnhancedLogger(combinedCore), nil
}

// Shutdown gracefully shuts down the OpenTelemetry logger provider
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.loggerProvider == nil {
		return nil
	}

	err := p.loggerProvider.Shutdown(ctx)
	p.loggerProvider = nil
	return err
}

// createEnhancedLogger creates the final enhanced logger with proper options
func (p *Provider) createEnhancedLogger(core zapcore.Core) *zap.Logger {
	return zap.New(core, zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))
}
