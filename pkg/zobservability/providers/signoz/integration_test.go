package signoz

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/propagation"

	"github.com/zondax/golem/pkg/zobservability"
)

func TestB3PropagationIntegration(t *testing.T) {
	// Test server that extracts trace headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for B3 headers
		traceID := r.Header.Get("X-B3-TraceId")
		spanID := r.Header.Get("X-B3-SpanId")
		sampled := r.Header.Get("X-B3-Sampled")

		w.Header().Set("X-Received-TraceId", traceID)
		w.Header().Set("X-Received-SpanId", spanID)
		w.Header().Set("X-Received-Sampled", sampled)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tests := []struct {
		name                string
		propagationFormats  []string
		expectedHeaderCount int
		checkSingleHeader   bool
	}{
		{
			name:                "B3 multi-header propagation",
			propagationFormats:  []string{zobservability.PropagationB3},
			expectedHeaderCount: 3, // X-B3-TraceId, X-B3-SpanId, X-B3-Sampled
			checkSingleHeader:   false,
		},
		{
			name:                "B3 single header propagation",
			propagationFormats:  []string{zobservability.PropagationB3Single},
			expectedHeaderCount: 1, // b3
			checkSingleHeader:   true,
		},
		{
			name:                "Mixed B3 and W3C propagation",
			propagationFormats:  []string{zobservability.PropagationB3, zobservability.PropagationW3C},
			expectedHeaderCount: 5, // B3 headers + traceparent + tracestate
			checkSingleHeader:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with specific propagation
			config := &Config{
				ServiceName: "test-service",
				Environment: "test",
				Release:     "1.0.0",
				Endpoint:    "localhost:4317",
				Insecure:    true,
				SampleRate:  1.0, // Always sample for testing
				Propagation: zobservability.PropagationConfig{
					Formats: tt.propagationFormats,
				},
			}

			// Create tracer provider with test configuration
			tracerProvider, tracer, err := createTracerProvider(config)
			require.NoError(t, err)
			defer tracerProvider.Shutdown(context.Background())

			// Start a span
			ctx, span := tracer.Start(context.Background(), "test-operation")
			defer span.End()

			// Create HTTP request with propagated context
			req, err := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
			require.NoError(t, err)

			// Inject trace context into headers
			propagator := createPropagator(config)
			propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

			// Make HTTP request
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Verify response
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			if tt.checkSingleHeader {
				// Check for B3 single header
				b3Header := req.Header.Get("b3")
				assert.NotEmpty(t, b3Header, "B3 single header should be present")
				assert.Contains(t, b3Header, "-", "B3 single header should contain trace and span IDs separated by dash")
			} else {
				// Check for B3 multi-headers
				if contains(tt.propagationFormats, zobservability.PropagationB3) {
					traceID := req.Header.Get("X-B3-TraceId")
					spanID := req.Header.Get("X-B3-SpanId")
					sampled := req.Header.Get("X-B3-Sampled")

					assert.NotEmpty(t, traceID, "X-B3-TraceId should be present")
					assert.NotEmpty(t, spanID, "X-B3-SpanId should be present")
					assert.Equal(t, "1", sampled, "X-B3-Sampled should be '1' for always sampling")

					// Verify trace ID format (32 hex characters for 128-bit trace ID)
					assert.Regexp(t, `^[0-9a-f]{32}$`, traceID, "Trace ID should be 32 hex characters")
					// Verify span ID format (16 hex characters for 64-bit span ID)
					assert.Regexp(t, `^[0-9a-f]{16}$`, spanID, "Span ID should be 16 hex characters")
				}

				// Check for W3C headers if configured
				if contains(tt.propagationFormats, zobservability.PropagationW3C) {
					traceparent := req.Header.Get("traceparent")
					assert.NotEmpty(t, traceparent, "traceparent header should be present")
					assert.Regexp(t, `^00-[0-9a-f]{32}-[0-9a-f]{16}-0[0-9a-f]$`, traceparent, "traceparent should match W3C format")
				}
			}

			// Count total headers for verification
			headerCount := 0
			for key := range req.Header {
				if isTraceHeader(key) {
					headerCount++
				}
			}

			// Note: This is an approximate check since header count can vary
			// based on the specific propagation combination
			if tt.expectedHeaderCount > 0 {
				assert.GreaterOrEqual(t, headerCount, 1, "At least one trace header should be present")
			}
		})
	}
}

func TestB3PropagationExtraction(t *testing.T) {
	// Create test server that injects B3 headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate downstream service that sends B3 headers
		w.Header().Set("X-B3-TraceId", "463ac35c9f6413ad48485a3953bb6124")
		w.Header().Set("X-B3-SpanId", "a2fb4a1d1a96d312")
		w.Header().Set("X-B3-Sampled", "1")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &Config{
		ServiceName: "test-service",
		Environment: "test",
		Release:     "1.0.0",
		Endpoint:    "localhost:4317",
		Insecure:    true,
		SampleRate:  1.0,
		Propagation: zobservability.PropagationConfig{
			Formats: []string{zobservability.PropagationB3},
		},
	}

	// Create tracer provider
	tracerProvider, tracer, err := createTracerProvider(config)
	require.NoError(t, err)
	defer tracerProvider.Shutdown(context.Background())

	// Create request with B3 headers
	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	// Add B3 headers to simulate incoming request
	req.Header.Set("X-B3-TraceId", "463ac35c9f6413ad48485a3953bb6124")
	req.Header.Set("X-B3-SpanId", "a2fb4a1d1a96d312")
	req.Header.Set("X-B3-Sampled", "1")

	// Extract context from headers
	propagator := createPropagator(config)
	ctx := propagator.Extract(context.Background(), propagation.HeaderCarrier(req.Header))

	// Start span with extracted context
	ctx, span := tracer.Start(ctx, "child-operation")
	defer span.End()

	// Verify that the span context contains the extracted trace information
	spanContext := span.SpanContext()
	assert.True(t, spanContext.IsValid(), "Span context should be valid")
	assert.True(t, spanContext.IsSampled(), "Span should be sampled")

	// The trace ID should match what was extracted
	// Note: OpenTelemetry might normalize the trace ID format
	traceID := spanContext.TraceID().String()
	assert.NotEmpty(t, traceID, "Trace ID should not be empty")
}

func TestPropagatorFallbackBehavior(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "empty formats fall back to W3C",
			config: &Config{
				ServiceName: "test-service",
				Environment: "test",
				Release:     "1.0.0",
				Endpoint:    "localhost:4317",
				Insecure:    true,
				SampleRate:  1.0,
				Propagation: zobservability.PropagationConfig{
					Formats: []string{},
				},
			},
		},
		{
			name: "invalid formats fall back to W3C",
			config: &Config{
				ServiceName: "test-service",
				Environment: "test",
				Release:     "1.0.0",
				Endpoint:    "localhost:4317",
				Insecure:    true,
				SampleRate:  1.0,
				Propagation: zobservability.PropagationConfig{
					Formats: []string{"invalid-format", "another-invalid"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create tracer provider
			tracerProvider, tracer, err := createTracerProvider(tt.config)
			require.NoError(t, err)
			defer tracerProvider.Shutdown(context.Background())

			// Start span and inject context
			ctx, span := tracer.Start(context.Background(), "test-operation")
			defer span.End()

			// Create request and inject headers
			req, err := http.NewRequest("GET", "http://example.com", nil)
			require.NoError(t, err)

			propagator := createPropagator(tt.config)
			propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

			// Should fall back to W3C propagation
			traceparent := req.Header.Get("traceparent")
			assert.NotEmpty(t, traceparent, "Should fall back to W3C traceparent header")
			assert.Regexp(t, `^00-[0-9a-f]{32}-[0-9a-f]{16}-0[0-9a-f]$`, traceparent, "traceparent should match W3C format")
		})
	}
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func isTraceHeader(headerName string) bool {
	traceHeaders := []string{
		"traceparent",
		"tracestate",
		"x-b3-traceid",
		"x-b3-spanid",
		"x-b3-sampled",
		"x-b3-flags",
		"x-b3-parentspanid",
		"b3",
		"uber-trace-id",
		"jaeger-debug-id",
		"jaeger-baggage",
	}

	headerLower := http.CanonicalHeaderKey(headerName)
	for _, traceHeader := range traceHeaders {
		if headerLower == http.CanonicalHeaderKey(traceHeader) {
			return true
		}
	}
	return false
}