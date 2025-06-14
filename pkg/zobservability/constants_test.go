package zobservability

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderConstants_WhenUsed_ShouldHaveCorrectValues(t *testing.T) {
	// Test provider constants
	assert.Equal(t, "sentry", ProviderSentry)
	assert.Equal(t, "signoz", ProviderSigNoz)
}

func TestEnvironmentConstants_WhenUsed_ShouldHaveCorrectValues(t *testing.T) {
	// Test environment constants
	assert.Equal(t, "production", EnvironmentProduction)
	assert.Equal(t, "development", EnvironmentDevelopment)
	assert.Equal(t, "staging", EnvironmentStaging)
	assert.Equal(t, "local", EnvironmentLocal)
}

func TestTagConstants_WhenUsed_ShouldHaveCorrectValues(t *testing.T) {
	// Test common tag keys
	assert.Equal(t, "operation", TagOperation)
	assert.Equal(t, "service", TagService)
	assert.Equal(t, "component", TagComponent)
	assert.Equal(t, "layer", TagLayer)
	assert.Equal(t, "method", TagMethod)

	// Test layer constants
	assert.Equal(t, "service", LayerService)
	assert.Equal(t, "repository", LayerRepository)
}

func TestResourceAttributeConstants_WhenUsed_ShouldHaveCorrectValues(t *testing.T) {
	// Test OpenTelemetry resource attribute keys
	assert.Equal(t, "service.name", ResourceServiceName)
	assert.Equal(t, "service.version", ResourceServiceVersion)
	assert.Equal(t, "service.type", ResourceServiceType)
	assert.Equal(t, "target.service", ResourceTargetService)
	assert.Equal(t, "deployment.environment", ResourceEnvironment)
	assert.Equal(t, "library.language", ResourceLanguage)
	assert.Equal(t, "host.name", ResourceHostName)
	assert.Equal(t, "process.pid", ResourceProcessPID)

	// Test resource attribute values
	assert.Equal(t, "go", ResourceLanguageGo)
}

func TestSpanAttributeConstants_WhenUsed_ShouldHaveCorrectValues(t *testing.T) {
	// Test span attribute keys
	assert.Equal(t, "level", SpanAttributeLevel)
	assert.Equal(t, "net.peer.name", SpanAttributeNetPeerName)
	assert.Equal(t, "http.url", SpanAttributeHTTPURL)
	assert.Equal(t, "http.target", SpanAttributeHTTPTarget)
	assert.Equal(t, "http.method", SpanAttributeHTTPMethod)
	assert.Equal(t, "http.scheme", SpanAttributeHTTPScheme)
	assert.Equal(t, "http.host", SpanAttributeHTTPHost)
	assert.Equal(t, "rpc.system", SpanAttributeRPCSystem)
	assert.Equal(t, "http.status_code", SpanAttributeHTTPStatusCode)
	assert.Equal(t, "rpc.grpc.status_code", SpanAttributeRPCGRPCStatusCode)
}

func TestRPCSystemConstants_WhenUsed_ShouldHaveCorrectValues(t *testing.T) {
	// Test RPC system values
	assert.Equal(t, "grpc", RPCSystemGRPC)
	assert.Equal(t, "http", RPCSystemHTTP)
}

func TestHTTPSchemeConstants_WhenUsed_ShouldHaveCorrectValues(t *testing.T) {
	// Test HTTP scheme values
	assert.Equal(t, "http", HTTPSchemeHTTP)
	assert.Equal(t, "https", HTTPSchemeHTTPS)
}

func TestUserAttributeConstants_WhenUsed_ShouldHaveCorrectValues(t *testing.T) {
	// Test user attribute keys
	assert.Equal(t, "user.id", UserAttributeID)
	assert.Equal(t, "user.email", UserAttributeEmail)
	assert.Equal(t, "user.username", UserAttributeUsername)
}

func TestFingerprintConstants_WhenUsed_ShouldHaveCorrectValues(t *testing.T) {
	// Test fingerprint constants
	assert.Equal(t, "fingerprint", FingerprintAttribute)
	assert.Equal(t, ",", FingerprintSeparator)
}

func TestTransactionStatusConstants_WhenUsed_ShouldHaveCorrectValues(t *testing.T) {
	// Test transaction status messages
	assert.Equal(t, "", TransactionSuccessMessage)
	assert.Equal(t, "transaction failed", TransactionFailureMessage)
	assert.Equal(t, "transaction cancelled", TransactionCancelledMessage)
}

func TestConstants_WhenGroupedByCategory_ShouldBeConsistent(t *testing.T) {
	// Test that related constants are consistent
	testCases := []struct {
		name        string
		constants   []string
		description string
	}{
		{
			name:        "providers",
			constants:   []string{ProviderSentry, ProviderSigNoz},
			description: "All provider constants should be non-empty",
		},
		{
			name:        "environments",
			constants:   []string{EnvironmentProduction, EnvironmentDevelopment, EnvironmentStaging, EnvironmentLocal},
			description: "All environment constants should be non-empty",
		},
		{
			name:        "tag_keys",
			constants:   []string{TagOperation, TagService, TagComponent, TagLayer, TagMethod},
			description: "All tag key constants should be non-empty",
		},
		{
			name:        "layers",
			constants:   []string{LayerService, LayerRepository},
			description: "All layer constants should be non-empty",
		},
		{
			name:        "resource_attributes",
			constants:   []string{ResourceServiceName, ResourceServiceVersion, ResourceServiceType, ResourceTargetService, ResourceEnvironment, ResourceLanguage, ResourceHostName, ResourceProcessPID},
			description: "All resource attribute constants should be non-empty",
		},
		{
			name:        "user_attributes",
			constants:   []string{UserAttributeID, UserAttributeEmail, UserAttributeUsername},
			description: "All user attribute constants should be non-empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, constant := range tc.constants {
				assert.NotEmpty(t, constant, tc.description)
			}
		})
	}
}

func TestConstants_WhenUsedAsMapKeys_ShouldBeUnique(t *testing.T) {
	// Test that tag constants can be used as unique map keys
	tagMap := map[string]bool{
		TagOperation: true,
		TagService:   true,
		TagComponent: true,
		TagLayer:     true,
		TagMethod:    true,
	}

	assert.Len(t, tagMap, 5, "All tag constants should be unique")

	// Test that resource attribute constants can be used as unique map keys
	resourceMap := map[string]bool{
		ResourceServiceName:    true,
		ResourceServiceVersion: true,
		ResourceServiceType:    true,
		ResourceTargetService:  true,
		ResourceEnvironment:    true,
		ResourceLanguage:       true,
		ResourceHostName:       true,
		ResourceProcessPID:     true,
	}

	assert.Len(t, resourceMap, 8, "All resource attribute constants should be unique")
}

func TestConstants_WhenUsedInRealScenarios_ShouldWorkCorrectly(t *testing.T) {
	// Test that constants work in real-world scenarios
	t.Run("provider_selection", func(t *testing.T) {
		providers := []string{ProviderSentry, ProviderSigNoz}
		for _, provider := range providers {
			assert.Contains(t, []string{"sentry", "signoz"}, provider)
		}
	})

	t.Run("environment_validation", func(t *testing.T) {
		environments := []string{EnvironmentProduction, EnvironmentDevelopment, EnvironmentStaging, EnvironmentLocal}
		for _, env := range environments {
			assert.NotEmpty(t, env)
			assert.NotContains(t, env, " ") // Should not contain spaces
		}
	})

	t.Run("tag_usage", func(t *testing.T) {
		tags := map[string]string{
			TagOperation: "create-user",
			TagService:   "user-service",
			TagComponent: "user-repository",
			TagLayer:     LayerService,
			TagMethod:    "CreateUser",
		}

		for key, value := range tags {
			assert.NotEmpty(t, key)
			assert.NotEmpty(t, value)
		}
	})
}

func TestConstants_WhenComparedToStandards_ShouldFollowConventions(t *testing.T) {
	// Test that OpenTelemetry constants follow semantic conventions
	t.Run("opentelemetry_semantic_conventions", func(t *testing.T) {
		// Resource attributes should follow service.* pattern
		assert.Contains(t, ResourceServiceName, "service.")
		assert.Contains(t, ResourceServiceVersion, "service.")
		assert.Contains(t, ResourceServiceType, "service.")

		// HTTP attributes should follow http.* pattern
		assert.Contains(t, SpanAttributeHTTPURL, "http.")
		assert.Contains(t, SpanAttributeHTTPTarget, "http.")
		assert.Contains(t, SpanAttributeHTTPMethod, "http.")
		assert.Contains(t, SpanAttributeHTTPScheme, "http.")
		assert.Contains(t, SpanAttributeHTTPHost, "http.")
		assert.Contains(t, SpanAttributeHTTPStatusCode, "http.")

		// Network attributes should follow net.* pattern
		assert.Contains(t, SpanAttributeNetPeerName, "net.")

		// RPC attributes should follow rpc.* pattern
		assert.Contains(t, SpanAttributeRPCSystem, "rpc.")
		assert.Contains(t, SpanAttributeRPCGRPCStatusCode, "rpc.")

		// User attributes should follow user.* pattern
		assert.Contains(t, UserAttributeID, "user.")
		assert.Contains(t, UserAttributeEmail, "user.")
		assert.Contains(t, UserAttributeUsername, "user.")
	})
}

func TestConstants_WhenUsedForValidation_ShouldProvideCorrectValues(t *testing.T) {
	// Test constants that are used for validation
	t.Run("supported_providers", func(t *testing.T) {
		supportedProviders := []string{ProviderSentry, ProviderSigNoz}

		// Should contain expected providers
		assert.Contains(t, supportedProviders, "sentry")
		assert.Contains(t, supportedProviders, "signoz")

		// Should not be empty
		for _, provider := range supportedProviders {
			assert.NotEmpty(t, provider)
		}
	})

	t.Run("supported_environments", func(t *testing.T) {
		supportedEnvironments := []string{EnvironmentProduction, EnvironmentDevelopment, EnvironmentStaging, EnvironmentLocal}

		// Should contain expected environments
		assert.Contains(t, supportedEnvironments, "production")
		assert.Contains(t, supportedEnvironments, "development")
		assert.Contains(t, supportedEnvironments, "staging")
		assert.Contains(t, supportedEnvironments, "local")

		// Should not be empty
		for _, env := range supportedEnvironments {
			assert.NotEmpty(t, env)
		}
	})
}
