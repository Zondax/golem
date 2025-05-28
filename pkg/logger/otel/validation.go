package otel

import (
	"fmt"

	"github.com/zondax/golem/pkg/logger"
)

// validateConfig validates the OpenTelemetry configuration
func (p *Provider) validateConfig(config logger.Config) error {
	if config.OpenTelemetry == nil {
		return fmt.Errorf("OpenTelemetry configuration is nil")
	}

	otelConfig := config.OpenTelemetry
	if otelConfig.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}

	if otelConfig.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}

	if otelConfig.Protocol != "" && !p.isSupportedProtocol(otelConfig.Protocol) {
		return fmt.Errorf("unsupported protocol: %s", otelConfig.Protocol)
	}

	return nil
}

// isSupportedProtocol checks if the protocol is supported
func (p *Provider) isSupportedProtocol(protocol string) bool {
	return protocol == ProtocolHTTP || protocol == ProtocolGRPC
}
