package signoz

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYAMLParsing(t *testing.T) {
	yamlConfig := `
provider: signoz
enabled: true
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
}