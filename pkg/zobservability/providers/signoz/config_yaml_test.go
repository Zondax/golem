package signoz

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYAMLParsing(t *testing.T) {
	yamlConfig := `
provider: signoz
enabled: true
span_counting:
  enabled: true
  log_span_counts: true
environment: local
address: "ingest.eu.signoz.cloud:443"
sample_rate: 1.0
`

	var cfg Config
	err := yaml.Unmarshal([]byte(yamlConfig), &cfg)
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	t.Logf("Parsed config: %+v", cfg)
	t.Logf("SpanCountingConfig: %+v", cfg.SpanCountingConfig)

	spanCountingConfig := cfg.GetSpanCountingConfig()
	if !spanCountingConfig.Enabled {
		t.Errorf("Expected span counting to be enabled, got %v", spanCountingConfig.Enabled)
	}

	if !spanCountingConfig.LogSpanCounts {
		t.Errorf("Expected log span counts to be enabled, got %v", spanCountingConfig.LogSpanCounts)
	}
}