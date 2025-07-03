package zobservability

import (
	"os"
	"strings"
	"sync"
)

const (
	unknownHostFallback = "unknown-host"

	// Environment variable names for hostname detection
	envOtelResourceHostname = "OTEL_RESOURCE_HOSTNAME"
	envKService             = "K_SERVICE"            // Cloud Run service name
	envKRevision            = "K_REVISION"           // Cloud Run revision
	envGAEService           = "GAE_SERVICE"          // App Engine service
	envGAEVersion           = "GAE_VERSION"          // App Engine version
	envFunctionName         = "FUNCTION_NAME"        // Cloud Functions
	envGoogleCloudProject   = "GOOGLE_CLOUD_PROJECT" // GCP project ID
	envServiceName          = "SERVICE_NAME"         // Generic service name
	envHostname             = "HOSTNAME"             // Container/Pod hostname
)

var (
	// hostname cache - initialized once and reused
	cachedHostname string
	hostnameOnce   sync.Once
)

// initializeHostname initializes the hostname once using environment variables
// This is called only once per process lifecycle using sync.Once
func initializeHostname() {
	// Try to build a meaningful hostname from environment variables
	if hostname := buildHostnameFromEnv(); hostname != "" {
		cachedHostname = hostname
		return
	}

	// Fallback to os.Hostname()
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		cachedHostname = hostname
		return
	}

	// Final fallback
	cachedHostname = unknownHostFallback
}

// buildHostnameFromEnv creates a meaningful hostname using environment variables
// This avoids HTTP requests and provides better identification than container hostnames
func buildHostnameFromEnv() string {
	// Check for explicit hostname override first
	if hostname := os.Getenv(envOtelResourceHostname); hostname != "" {
		return hostname
	}

	// Try each platform detection strategy in priority order
	strategies := []func() string{
		buildCloudRunHostname,
		buildAppEngineHostname,
		buildCloudFunctionsHostname,
		buildGCPProjectHostname,
		buildContainerHostname,
	}

	for _, strategy := range strategies {
		if hostname := strategy(); hostname != "" {
			return hostname
		}
	}

	return ""
}

// buildCloudRunHostname builds hostname for Cloud Run environments
func buildCloudRunHostname() string {
	service := os.Getenv(envKService)
	if service == "" {
		return ""
	}

	revision := os.Getenv(envKRevision)
	if revision == "" {
		return service
	}

	// Avoid duplication if revision already contains the service name
	if strings.HasPrefix(revision, service) {
		return revision
	}

	return service + "-" + revision
}

// buildAppEngineHostname builds hostname for App Engine environments
func buildAppEngineHostname() string {
	service := os.Getenv(envGAEService)
	if service == "" {
		return ""
	}

	version := os.Getenv(envGAEVersion)
	if version == "" {
		return service
	}

	return service + "-" + version
}

// buildCloudFunctionsHostname builds hostname for Cloud Functions environments
func buildCloudFunctionsHostname() string {
	return os.Getenv(envFunctionName)
}

// buildGCPProjectHostname builds hostname using GCP project and service name
func buildGCPProjectHostname() string {
	project := os.Getenv(envGoogleCloudProject)
	if project == "" {
		return ""
	}

	service := os.Getenv(envServiceName)
	if service == "" {
		return project
	}

	return project + "-" + service
}

// buildContainerHostname builds hostname for container/Kubernetes environments
func buildContainerHostname() string {
	podName := os.Getenv(envHostname)
	if podName == "" || podName == "localhost" {
		return ""
	}

	return podName
}

// GetHostname returns the hostname - now ALWAYS included for service identification
// Hostname is crucial for:
// - Multi-server deployments (identifying which server handled the request)
// - Load balancer debugging (tracking requests across instances)
// - Performance analysis (comparing server performance)
// - Incident response (knowing exactly which server had issues)
//
// Uses sync.Once to ensure hostname detection only happens once per process lifecycle
func GetHostname() string {
	// Initialize hostname only once using sync.Once
	hostnameOnce.Do(initializeHostname)

	return cachedHostname
}