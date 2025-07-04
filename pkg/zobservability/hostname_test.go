package zobservability

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildHostnameFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected string
	}{
		{
			name: "OTEL_RESOURCE_HOSTNAME override",
			envVars: map[string]string{
				envOtelResourceHostname: "custom-hostname",
				envKService:             "api",
				envKRevision:            "v1",
			},
			expected: "custom-hostname",
		},
		{
			name: "Cloud Run service with revision",
			envVars: map[string]string{
				envKService:  "kickstarter-api",
				envKRevision: "pr-71",
			},
			expected: "kickstarter-api-pr-71",
		},
		{
			name: "Cloud Run service with revision that already contains service name",
			envVars: map[string]string{
				envKService:  "api",
				envKRevision: "api-pr-71-abc123",
			},
			expected: "api-pr-71-abc123",
		},
		{
			name: "Cloud Run service without revision",
			envVars: map[string]string{
				envKService: "kickstarter-api",
			},
			expected: "kickstarter-api",
		},
		{
			name: "App Engine with version",
			envVars: map[string]string{
				envGAEService: "my-service",
				envGAEVersion: "v2",
			},
			expected: "my-service-v2",
		},
		{
			name: "App Engine without version",
			envVars: map[string]string{
				envGAEService: "my-service",
			},
			expected: "my-service",
		},
		{
			name: "Cloud Functions",
			envVars: map[string]string{
				envFunctionName: "process-webhook",
			},
			expected: "process-webhook",
		},
		{
			name: "GCP project with service name",
			envVars: map[string]string{
				envGoogleCloudProject: "my-project",
				envServiceName:        "api-service",
			},
			expected: "my-project-api-service",
		},
		{
			name: "GCP project without service name",
			envVars: map[string]string{
				envGoogleCloudProject: "my-project",
			},
			expected: "my-project",
		},
		{
			name: "HOSTNAME (not localhost)",
			envVars: map[string]string{
				envHostname: "pod-123-abc",
			},
			expected: "pod-123-abc",
		},
		{
			name: "HOSTNAME localhost (ignored)",
			envVars: map[string]string{
				envHostname: "localhost",
			},
			expected: "",
		},
		{
			name:     "No environment variables",
			envVars:  map[string]string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all relevant environment variables
			clearEnvVars := []string{
				envOtelResourceHostname,
				envKService,
				envKRevision,
				envGAEService,
				envGAEVersion,
				envFunctionName,
				envGoogleCloudProject,
				envServiceName,
				envHostname,
			}

			for _, env := range clearEnvVars {
				_ = os.Unsetenv(env)
			}

			// Set test environment variables
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}

			// Test the function
			result := buildHostnameFromEnv()
			assert.Equal(t, tt.expected, result, "buildHostnameFromEnv() = %q, want %q", result, tt.expected)

			// Clean up
			for key := range tt.envVars {
				_ = os.Unsetenv(key)
			}
		})
	}
}

func TestGetHostname_Caching(t *testing.T) {
	// Helper to reset hostname cache for testing
	resetHostnameCache := func() {
		cachedHostname = ""
		hostnameOnce = sync.Once{}
	}

	t.Run("caches hostname after first call", func(t *testing.T) {
		resetHostnameCache()

		// Set up environment
		_ = os.Setenv(envKService, "test-service")
		_ = os.Setenv(envKRevision, "test-revision")
		defer func() {
			_ = os.Unsetenv(envKService)
			_ = os.Unsetenv(envKRevision)
		}()

		// First call should initialize
		hostname1 := GetHostname()
		expected := "test-service-test-revision"
		assert.Equal(t, expected, hostname1, "First call: GetHostname() = %q, want %q", hostname1, expected)

		// Change environment (should not affect cached result)
		_ = os.Setenv(envKService, "different-service")

		// Second call should return cached value
		hostname2 := GetHostname()
		assert.Equal(t, expected, hostname2, "Second call: GetHostname() = %q, want %q (cached)", hostname2, expected)

		// Verify it's the same instance (cached)
		assert.Equal(t, hostname1, hostname2, "Hostname should be cached and return same value")
	})

	t.Run("fallback behavior when no env vars", func(t *testing.T) {
		resetHostnameCache()

		// Clear all environment variables
		clearEnvVars := []string{
			envOtelResourceHostname,
			envKService,
			envKRevision,
			envGAEService,
			envGAEVersion,
			envFunctionName,
			envGoogleCloudProject,
			envServiceName,
			envHostname,
		}

		for _, env := range clearEnvVars {
			_ = os.Unsetenv(env)
		}

		hostname := GetHostname()

		// Should either be os.Hostname() result or unknownHostFallback
		assert.NotEmpty(t, hostname, "GetHostname() should never return empty string")

		// Should be either a real hostname or the fallback
		if hostname == unknownHostFallback {
			// If it's the fallback, that's fine
			assert.Equal(t, unknownHostFallback, hostname)
		} else {
			// If it's not the fallback, it should be a non-empty string from os.Hostname()
			assert.NotEmpty(t, hostname, "Hostname from os.Hostname() should not be empty")
		}
	})

	t.Run("sync.Once ensures single execution", func(t *testing.T) {
		resetHostnameCache()

		_ = os.Setenv(envKService, "once-test")
		defer func() { _ = os.Unsetenv(envKService) }()

		// Call multiple times concurrently
		const numGoroutines = 10
		results := make([]string, numGoroutines)
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				results[index] = GetHostname()
			}(i)
		}

		wg.Wait()

		// All results should be identical
		expected := results[0]
		for i, result := range results {
			assert.Equal(t, expected, result, "Goroutine %d: GetHostname() = %q, want %q", i, result, expected)
		}
	})
}

func TestInitializeHostname_Priority(t *testing.T) {
	// Reset the cache
	cachedHostname = ""
	hostnameOnce = sync.Once{}

	t.Run("prioritizes environment variables correctly", func(t *testing.T) {
		// Set multiple env vars to test priority
		_ = os.Setenv(envOtelResourceHostname, "override-hostname")
		_ = os.Setenv(envKService, "service")
		_ = os.Setenv(envKRevision, "revision")
		defer func() {
			_ = os.Unsetenv(envOtelResourceHostname)
			_ = os.Unsetenv(envKService)
			_ = os.Unsetenv(envKRevision)
		}()

		initializeHostname()

		// Should use the override hostname (highest priority)
		assert.Equal(t, "override-hostname", cachedHostname, "initializeHostname() cached %q, want %q", cachedHostname, "override-hostname")
	})
}

func TestGetHostname_Integration(t *testing.T) {
	t.Run("returns non-empty hostname", func(t *testing.T) {
		hostname := GetHostname()

		// Should never return empty string
		assert.NotEmpty(t, hostname, "GetHostname() should never return empty string")
	})

	t.Run("consistent results across calls", func(t *testing.T) {
		// Call GetHostname multiple times to ensure it's consistent
		hostname1 := GetHostname()
		hostname2 := GetHostname()

		assert.Equal(t, hostname1, hostname2, "GetHostname() should return consistent results")
		assert.NotEmpty(t, hostname1, "Hostname should not be empty")
	})
}

func TestBuildCloudRunHostname(t *testing.T) {
	tests := []struct {
		name     string
		service  string
		revision string
		expected string
	}{
		{
			name:     "service and revision both set",
			service:  "my-service",
			revision: "v1-abc123",
			expected: "my-service-v1-abc123",
		},
		{
			name:     "revision contains service name",
			service:  "api",
			revision: "api-v1-abc123",
			expected: "api-v1-abc123",
		},
		{
			name:     "only service set",
			service:  "my-service",
			revision: "",
			expected: "my-service",
		},
		{
			name:     "no service set",
			service:  "",
			revision: "v1-abc123",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			_ = os.Unsetenv(envKService)
			_ = os.Unsetenv(envKRevision)

			// Set test values
			if tt.service != "" {
				_ = os.Setenv(envKService, tt.service)
			}
			if tt.revision != "" {
				_ = os.Setenv(envKRevision, tt.revision)
			}

			result := buildCloudRunHostname()
			assert.Equal(t, tt.expected, result)

			// Clean up
			_ = os.Unsetenv(envKService)
			_ = os.Unsetenv(envKRevision)
		})
	}
}

func TestBuildAppEngineHostname(t *testing.T) {
	tests := []struct {
		name     string
		service  string
		version  string
		expected string
	}{
		{
			name:     "service and version both set",
			service:  "default",
			version:  "20230101t123456",
			expected: "default-20230101t123456",
		},
		{
			name:     "only service set",
			service:  "my-service",
			version:  "",
			expected: "my-service",
		},
		{
			name:     "no service set",
			service:  "",
			version:  "20230101t123456",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			_ = os.Unsetenv(envGAEService)
			_ = os.Unsetenv(envGAEVersion)

			// Set test values
			if tt.service != "" {
				_ = os.Setenv(envGAEService, tt.service)
			}
			if tt.version != "" {
				_ = os.Setenv(envGAEVersion, tt.version)
			}

			result := buildAppEngineHostname()
			assert.Equal(t, tt.expected, result)

			// Clean up
			_ = os.Unsetenv(envGAEService)
			_ = os.Unsetenv(envGAEVersion)
		})
	}
}

func TestBuildGCPProjectHostname(t *testing.T) {
	tests := []struct {
		name     string
		project  string
		service  string
		expected string
	}{
		{
			name:     "project and service both set",
			project:  "my-project",
			service:  "api-service",
			expected: "my-project-api-service",
		},
		{
			name:     "only project set",
			project:  "my-project",
			service:  "",
			expected: "my-project",
		},
		{
			name:     "no project set",
			project:  "",
			service:  "api-service",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			_ = os.Unsetenv(envGoogleCloudProject)
			_ = os.Unsetenv(envServiceName)

			// Set test values
			if tt.project != "" {
				_ = os.Setenv(envGoogleCloudProject, tt.project)
			}
			if tt.service != "" {
				_ = os.Setenv(envServiceName, tt.service)
			}

			result := buildGCPProjectHostname()
			assert.Equal(t, tt.expected, result)

			// Clean up
			_ = os.Unsetenv(envGoogleCloudProject)
			_ = os.Unsetenv(envServiceName)
		})
	}
}

func TestBuildCloudFunctionsHostname(t *testing.T) {
	tests := []struct {
		name         string
		functionName string
		expected     string
	}{
		{
			name:         "function name set",
			functionName: "process-webhook",
			expected:     "process-webhook",
		},
		{
			name:         "function name empty",
			functionName: "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			_ = os.Unsetenv(envFunctionName)

			// Set test value
			if tt.functionName != "" {
				_ = os.Setenv(envFunctionName, tt.functionName)
			}

			result := buildCloudFunctionsHostname()
			assert.Equal(t, tt.expected, result)

			// Clean up
			_ = os.Unsetenv(envFunctionName)
		})
	}
}

func TestBuildContainerHostname(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		expected string
	}{
		{
			name:     "valid container hostname",
			hostname: "pod-123-abc",
			expected: "pod-123-abc",
		},
		{
			name:     "localhost ignored",
			hostname: "localhost",
			expected: "",
		},
		{
			name:     "empty hostname",
			hostname: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			_ = os.Unsetenv(envHostname)

			// Set test value
			if tt.hostname != "" {
				_ = os.Setenv(envHostname, tt.hostname)
			}

			result := buildContainerHostname()
			assert.Equal(t, tt.expected, result)

			// Clean up
			_ = os.Unsetenv(envHostname)
		})
	}
}
