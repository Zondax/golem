package zobservability

// Provider constants
const (
	ProviderSentry = "sentry"
	ProviderSigNoz = "signoz"
)

// Environment constants for observability configuration
const (
	EnvironmentProduction  = "production"
	EnvironmentDevelopment = "development"
	EnvironmentStaging     = "staging"
	EnvironmentLocal       = "local"
)

// Common tag keys for spans and transactions
const (
	TagOperation = "operation"
	TagService   = "service"
	TagComponent = "component"
	TagLayer     = "layer"
	TagMethod    = "method"

	LayerService    = "service"
	LayerRepository = "repository"
)

// OpenTelemetry Resource attribute keys - these are standard across all providers
const (
	ResourceServiceName    = "service.name"
	ResourceServiceVersion = "service.version"
	ResourceServiceType    = "service.type"
	ResourceTargetService  = "target.service"
	ResourceEnvironment    = "deployment.environment"
	ResourceLanguage       = "library.language"
	ResourceHostName       = "host.name"
	ResourceProcessPID     = "process.pid"
)

// OpenTelemetry Resource attribute values
const (
	ResourceLanguageGo = "go"
)

// OpenTelemetry Span attribute keys
const (
	SpanAttributeLevel = "level"
)

// External API Monitoring attributes (SigNoz semantic conventions)
// These attributes are used by SigNoz to automatically detect and categorize external API calls
const (
	// Network attributes for external API detection
	SpanAttributeNetPeerName = "net.peer.name" // Domain or host of the external service (e.g., "api.stripe.com")
	SpanAttributeHTTPURL     = "http.url"      // Complete URL of the request (e.g., "https://api.stripe.com/v1/charges")
	SpanAttributeHTTPTarget  = "http.target"   // Path portion of the URL (e.g., "/v1/charges")
	SpanAttributeHTTPMethod  = "http.method"   // HTTP method (GET, POST, etc.)
	SpanAttributeHTTPScheme  = "http.scheme"   // HTTP scheme (http, https)
	SpanAttributeHTTPHost    = "http.host"     // HTTP host header value

	// gRPC attributes for external service calls
	SpanAttributeRPCSystem = "rpc.system" // RPC system identifier (e.g., "grpc")

	// Response attributes
	SpanAttributeHTTPStatusCode    = "http.status_code"     // HTTP response status code
	SpanAttributeRPCGRPCStatusCode = "rpc.grpc.status_code" // gRPC status code
)

// External API categorization values
const (
	// RPC system values
	RPCSystemGRPC = "grpc"
	RPCSystemHTTP = "http"

	// HTTP schemes
	HTTPSchemeHTTP  = "http"
	HTTPSchemeHTTPS = "https"
)

// User attribute keys
const (
	UserAttributeID       = "user.id"
	UserAttributeEmail    = "user.email"
	UserAttributeUsername = "user.username"
)

// Fingerprint attribute key
const (
	FingerprintAttribute = "fingerprint"
	FingerprintSeparator = ","
)

// Transaction status messages
const (
	TransactionSuccessMessage   = ""
	TransactionFailureMessage   = "transaction failed"
	TransactionCancelledMessage = "transaction cancelled"
)
