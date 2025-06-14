package signoz

import "errors"

var (
	// Configuration errors
	ErrMissingEndpoint    = errors.New("signoz endpoint is required")
	ErrMissingServiceName = errors.New("service name is required")
	ErrInvalidSampleRate  = errors.New("sample rate must be between 0 and 1")

	// Runtime errors
	ErrTracerProviderNil = errors.New("tracer provider is nil")
	ErrSpanNil           = errors.New("span is nil")
	ErrContextNil        = errors.New("context is nil")
)
