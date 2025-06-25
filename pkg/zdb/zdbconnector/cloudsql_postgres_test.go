package zdbconnector

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zondax/golem/pkg/zdb/zdbconfig"
)

func TestCloudSQLPostgresConnector_ValidateConfig(t *testing.T) {
	connector := &CloudSQLPostgresConnector{}

	tests := []struct {
		name        string
		config      *zdbconfig.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "Cloud SQL disabled",
			config: &zdbconfig.Config{
				ConnectionParams: zdbconfig.ConnectionParams{
					CloudSQL: zdbconfig.CloudSQLConfig{
						Enabled: false,
					},
				},
			},
			expectError: true,
			errorMsg:    "cloud SQL is not enabled in configuration",
		},
		{
			name: "Missing instance name",
			config: &zdbconfig.Config{
				ConnectionParams: zdbconfig.ConnectionParams{
					CloudSQL: zdbconfig.CloudSQLConfig{
						Enabled:      true,
						InstanceName: "",
					},
				},
			},
			expectError: true,
			errorMsg:    "cloud SQL instance name is required",
		},
		{
			name: "Valid config with IAM auth",
			config: &zdbconfig.Config{
				ConnectionParams: zdbconfig.ConnectionParams{
					User: "test-user",
					Name: "test-db",
					CloudSQL: zdbconfig.CloudSQLConfig{
						Enabled:      true,
						InstanceName: "project:region:instance",
						UseIAMAuth:   true,
						UsePrivateIP: true,
					},
				},
				LogConfig: zdbconfig.LogConfig{
					LogLevel: "info",
				},
			},
			expectError: false,
		},
		{
			name: "Valid config with password auth",
			config: &zdbconfig.Config{
				ConnectionParams: zdbconfig.ConnectionParams{
					User:     "test-user",
					Password: "test-password",
					Name:     "test-db",
					CloudSQL: zdbconfig.CloudSQLConfig{
						Enabled:         true,
						InstanceName:    "project:region:instance",
						UseIAMAuth:      false,
						UsePrivateIP:    false,
						CredentialsFile: "/path/to/credentials.json",
						RefreshTimeout:  30,
					},
				},
				LogConfig: zdbconfig.LogConfig{
					LogLevel: "info",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := connector.Connect(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else if err != nil {
				// Note: We expect an error here because we're not actually connecting to a real Cloud SQL instance
				// but we should not get validation errors
				assert.NotContains(t, err.Error(), "cloud SQL is not enabled")
				assert.NotContains(t, err.Error(), "instance name is required")
			}
		})
	}
}

func TestBuildCloudSQLPostgresDSN(t *testing.T) {
	tests := []struct {
		name           string
		params         zdbconfig.ConnectionParams
		expectedParams []string // Using slice because map iteration order is not guaranteed
	}{
		{
			name: "Basic DSN with password",
			params: zdbconfig.ConnectionParams{
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
				CloudSQL: zdbconfig.CloudSQLConfig{
					UseIAMAuth: false,
				},
			},
			expectedParams: []string{"user=testuser", "database=testdb", "password=testpass", "sslmode=disable"},
		},
		{
			name: "DSN with IAM auth (no password)",
			params: zdbconfig.ConnectionParams{
				User: "test-sa@project.iam",
				Name: "testdb",
				CloudSQL: zdbconfig.CloudSQLConfig{
					UseIAMAuth: true,
				},
			},
			expectedParams: []string{"user=test-sa@project.iam", "database=testdb", "sslmode=disable"},
		},
		{
			name: "DSN with additional params",
			params: zdbconfig.ConnectionParams{
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
				Params:   "application_name=myapp connect_timeout=10",
				CloudSQL: zdbconfig.CloudSQLConfig{
					UseIAMAuth: false,
				},
			},
			expectedParams: []string{"user=testuser", "database=testdb", "password=testpass", "sslmode=disable", "application_name=myapp", "connect_timeout=10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCloudSQLPostgresDSN(tt.params)

			// Split the result and check that all expected parts are present
			resultParts := strings.Split(result, " ")
			assert.Len(t, resultParts, len(tt.expectedParams))

			// Check that all expected parts are in the result
			for _, expectedParam := range tt.expectedParams {
				assert.Contains(t, resultParts, expectedParam)
			}
		})
	}
}

func TestParseConnectionParams(t *testing.T) {
	tests := []struct {
		name     string
		params   string
		expected map[string]string
	}{
		{
			name:     "Empty params",
			params:   "",
			expected: map[string]string{},
		},
		{
			name:   "Single parameter",
			params: "application_name=myapp",
			expected: map[string]string{
				"application_name": "myapp",
			},
		},
		{
			name:   "Multiple parameters",
			params: "application_name=myapp connect_timeout=10 sslmode=require",
			expected: map[string]string{
				"application_name": "myapp",
				"connect_timeout":  "10",
				"sslmode":          "require",
			},
		},
		{
			name:   "Parameter with equals in value",
			params: "search_path=schema1,schema2 custom_param=key=value",
			expected: map[string]string{
				"search_path":  "schema1,schema2",
				"custom_param": "key=value",
			},
		},
		{
			name:   "Extra spaces",
			params: "  application_name=myapp   connect_timeout=10  ",
			expected: map[string]string{
				"application_name": "myapp",
				"connect_timeout":  "10",
			},
		},
		{
			name:   "Invalid parameter (no equals)",
			params: "application_name=myapp invalid_param connect_timeout=10",
			expected: map[string]string{
				"application_name": "myapp",
				"connect_timeout":  "10",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseConnectionParams(tt.params)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildDSNString(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]string
		expected []string // Using slice because map iteration order is not guaranteed
	}{
		{
			name:     "Empty params",
			params:   map[string]string{},
			expected: []string{},
		},
		{
			name: "Single parameter",
			params: map[string]string{
				"user": "testuser",
			},
			expected: []string{"user=testuser"},
		},
		{
			name: "Multiple parameters",
			params: map[string]string{
				"user":     "testuser",
				"database": "testdb",
				"sslmode":  "disable",
			},
			expected: []string{"user=testuser", "database=testdb", "sslmode=disable"},
		},
		{
			name: "Parameters with special characters",
			params: map[string]string{
				"user":        "test@example.com",
				"password":    "pass=word",
				"search_path": "schema1,schema2",
			},
			expected: []string{"user=test@example.com", "password=pass=word", "search_path=schema1,schema2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDSNString(tt.params)

			if len(tt.expected) == 0 {
				assert.Empty(t, result)
				return
			}

			// Split the result and check that all expected parts are present
			resultParts := strings.Split(result, " ")
			assert.Len(t, resultParts, len(tt.expected))

			// Check that all expected parts are in the result
			for _, expectedPart := range tt.expected {
				assert.Contains(t, resultParts, expectedPart)
			}
		})
	}
}
