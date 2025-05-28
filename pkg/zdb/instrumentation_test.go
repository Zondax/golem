package zdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zondax/golem/pkg/zdb/zdbconfig"
)

func TestGetConfigWithDefaults(t *testing.T) {
	t.Run("nil config returns disabled configuration", func(t *testing.T) {
		config := getConfigWithDefaults(nil)

		assert.False(t, config.Enabled)
	})

	t.Run("config with empty values gets defaults", func(t *testing.T) {
		userConfig := &zdbconfig.OpenTelemetryConfig{
			Enabled: true,
			// QueryFormatter is empty - should get default
			// DefaultAttributes is nil - should get empty map
		}

		config := getConfigWithDefaults(userConfig)

		assert.True(t, config.Enabled)
		assert.Equal(t, zdbconfig.QueryFormatterDefault, config.QueryFormatter)
		assert.NotNil(t, config.DefaultAttributes)
		assert.Empty(t, config.DefaultAttributes)
	})

	t.Run("config with set values preserves them", func(t *testing.T) {
		customAttrs := map[string]string{"service": "test"}
		userConfig := &zdbconfig.OpenTelemetryConfig{
			Enabled:                true,
			IncludeQueryParameters: true,
			QueryFormatter:         zdbconfig.QueryFormatterUpper,
			DefaultAttributes:      customAttrs,
			DisableMetrics:         true,
		}

		config := getConfigWithDefaults(userConfig)

		assert.True(t, config.Enabled)
		assert.True(t, config.IncludeQueryParameters)
		assert.Equal(t, zdbconfig.QueryFormatterUpper, config.QueryFormatter)
		assert.Equal(t, customAttrs, config.DefaultAttributes)
		assert.True(t, config.DisableMetrics)
	})
}

func TestInstrumentationManager_CreateQueryFormatter(t *testing.T) {
	tests := []struct {
		name           string
		queryFormatter string
		inputQuery     string
		expectedOutput string
		expectNil      bool
	}{
		{
			name:           "upper formatter",
			queryFormatter: zdbconfig.QueryFormatterUpper,
			inputQuery:     "select * from users",
			expectedOutput: "SELECT * FROM USERS",
			expectNil:      false,
		},
		{
			name:           "lower formatter",
			queryFormatter: zdbconfig.QueryFormatterLower,
			inputQuery:     "SELECT * FROM USERS",
			expectedOutput: "select * from users",
			expectNil:      false,
		},
		{
			name:           "none formatter hides query",
			queryFormatter: zdbconfig.QueryFormatterNone,
			inputQuery:     "SELECT * FROM users WHERE password = 'secret'",
			expectedOutput: "[QUERY HIDDEN]",
			expectNil:      false,
		},
		{
			name:           "default formatter returns nil",
			queryFormatter: zdbconfig.QueryFormatterDefault,
			inputQuery:     "SELECT * FROM users",
			expectNil:      true,
		},
		{
			name:           "empty formatter returns nil",
			queryFormatter: "",
			inputQuery:     "SELECT * FROM users",
			expectNil:      true,
		},
		{
			name:           "unknown formatter returns nil and logs warning",
			queryFormatter: "unknown",
			inputQuery:     "SELECT * FROM users",
			expectNil:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &instrumentationManager{
				config: zdbconfig.OpenTelemetryConfig{
					QueryFormatter: tt.queryFormatter,
				},
			}

			formatter := manager.createQueryFormatter()

			if tt.expectNil {
				assert.Nil(t, formatter)
			} else {
				assert.NotNil(t, formatter)
				result := formatter(tt.inputQuery)
				assert.Equal(t, tt.expectedOutput, result)
			}
		})
	}
}

func TestInstrumentationManager_BuildInstrumentationOptions(t *testing.T) {
	t.Run("disabled query parameters includes WithoutQueryVariables", func(t *testing.T) {
		manager := &instrumentationManager{
			config: zdbconfig.OpenTelemetryConfig{
				IncludeQueryParameters: false,
			},
		}

		opts := manager.buildInstrumentationOptions()

		// Should have at least one option (WithoutQueryVariables)
		assert.NotEmpty(t, opts)
	})

	t.Run("enabled query parameters does not include WithoutQueryVariables", func(t *testing.T) {
		manager := &instrumentationManager{
			config: zdbconfig.OpenTelemetryConfig{
				IncludeQueryParameters: true,
				QueryFormatter:         zdbconfig.QueryFormatterDefault, // This should not add options
				DefaultAttributes:      map[string]string{},             // Empty should not add options
				DisableMetrics:         false,                           // This should not add options
			},
		}

		opts := manager.buildInstrumentationOptions()

		// Should have no options since all are in their "default" state
		assert.Empty(t, opts)
	})

	t.Run("custom attributes are included", func(t *testing.T) {
		manager := &instrumentationManager{
			config: zdbconfig.OpenTelemetryConfig{
				DefaultAttributes: map[string]string{
					"service": "test-service",
					"version": "1.0.0",
				},
			},
		}

		opts := manager.buildInstrumentationOptions()

		// Should have at least one option (WithAttributes)
		assert.NotEmpty(t, opts)
	})

	t.Run("disabled metrics includes WithoutMetrics", func(t *testing.T) {
		manager := &instrumentationManager{
			config: zdbconfig.OpenTelemetryConfig{
				DisableMetrics: true,
			},
		}

		opts := manager.buildInstrumentationOptions()

		// Should have at least one option (WithoutMetrics)
		assert.NotEmpty(t, opts)
	})
}

func TestInstrumentationManager_GetQueryParameterOptions(t *testing.T) {
	t.Run("disabled includes WithoutQueryVariables option", func(t *testing.T) {
		manager := &instrumentationManager{
			config: zdbconfig.OpenTelemetryConfig{
				IncludeQueryParameters: false,
			},
		}

		opts := manager.getQueryParameterOptions()
		assert.Len(t, opts, 1)
	})

	t.Run("enabled returns empty options", func(t *testing.T) {
		manager := &instrumentationManager{
			config: zdbconfig.OpenTelemetryConfig{
				IncludeQueryParameters: true,
			},
		}

		opts := manager.getQueryParameterOptions()
		assert.Empty(t, opts)
	})
}

func TestInstrumentationManager_GetMetricsOptions(t *testing.T) {
	t.Run("disabled includes WithoutMetrics option", func(t *testing.T) {
		manager := &instrumentationManager{
			config: zdbconfig.OpenTelemetryConfig{
				DisableMetrics: true,
			},
		}

		opts := manager.getMetricsOptions()
		assert.Len(t, opts, 1)
	})

	t.Run("enabled returns empty options", func(t *testing.T) {
		manager := &instrumentationManager{
			config: zdbconfig.OpenTelemetryConfig{
				DisableMetrics: false,
			},
		}

		opts := manager.getMetricsOptions()
		assert.Empty(t, opts)
	})
}

func TestInstrumentationManager_GetDefaultAttributeOptions(t *testing.T) {
	t.Run("empty attributes return empty options", func(t *testing.T) {
		manager := &instrumentationManager{
			config: zdbconfig.OpenTelemetryConfig{
				DefaultAttributes: map[string]string{},
			},
		}

		opts := manager.getDefaultAttributeOptions()
		assert.Empty(t, opts)
	})

	t.Run("nil attributes return empty options", func(t *testing.T) {
		manager := &instrumentationManager{
			config: zdbconfig.OpenTelemetryConfig{
				DefaultAttributes: nil,
			},
		}

		opts := manager.getDefaultAttributeOptions()
		assert.Empty(t, opts)
	})

	t.Run("custom attributes return WithAttributes option", func(t *testing.T) {
		manager := &instrumentationManager{
			config: zdbconfig.OpenTelemetryConfig{
				DefaultAttributes: map[string]string{
					"service": "test",
					"version": "1.0",
				},
			},
		}

		opts := manager.getDefaultAttributeOptions()
		assert.Len(t, opts, 1)
	})
}

func TestInstrumentationManager_GetQueryFormatterOptions(t *testing.T) {
	t.Run("default formatter returns empty options", func(t *testing.T) {
		manager := &instrumentationManager{
			config: zdbconfig.OpenTelemetryConfig{
				QueryFormatter: zdbconfig.QueryFormatterDefault,
			},
		}

		opts := manager.getQueryFormatterOptions()
		assert.Empty(t, opts)
	})

	t.Run("custom formatter returns WithQueryFormatter option", func(t *testing.T) {
		manager := &instrumentationManager{
			config: zdbconfig.OpenTelemetryConfig{
				QueryFormatter: zdbconfig.QueryFormatterUpper,
			},
		}

		opts := manager.getQueryFormatterOptions()
		assert.Len(t, opts, 1)
	})
}
