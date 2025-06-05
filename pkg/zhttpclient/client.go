package zhttpclient

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/zondax/golem/pkg/utils"

	"github.com/go-resty/resty/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type ZHTTPClient interface {
	SetRetryPolicy(retryPolicy *RetryPolicy) ZHTTPClient
	NewRequest() ZRequest
	Do(ctx context.Context, req *http.Request) (*Response, error)
	GetHTTPClient() *http.Client
}

// OpenTelemetryConfig configures OpenTelemetry instrumentation for HTTP client
type OpenTelemetryConfig struct {
	// Enabled controls whether OpenTelemetry instrumentation is applied
	Enabled bool
	// OperationNameFunc is an optional function to customize operation names
	// If nil, default operation naming will be used
	// Signature: func(operation string, r *http.Request) string
	OperationNameFunc func(string, *http.Request) string
	// Filters is an optional function to filter which requests to instrument
	// If nil, all requests will be instrumented
	Filters func(*http.Request) bool
}

type Config struct {
	Timeout       time.Duration
	TLSConfig     *tls.Config
	BaseClient    *http.Client
	OpenTelemetry *OpenTelemetryConfig
}

// zHTTPClient abstracts over the resty.Client and provides per-request retry configurations.
type zHTTPClient struct {
	client      *resty.Client
	config      *Config
	retryPolicy RetryPolicy
}

func New(config Config) ZHTTPClient {
	z := &zHTTPClient{
		config: &config,
	}

	var baseClient *http.Client
	if config.BaseClient == nil {
		baseClient = &http.Client{}
	} else {
		baseClient = config.BaseClient
	}

	// Apply OpenTelemetry instrumentation if configured
	baseClient.Transport = z.configureTransport(baseClient.Transport, config.OpenTelemetry)

	z.client = resty.NewWithClient(baseClient)

	if config.TLSConfig != nil {
		z.client.SetTLSClientConfig(config.TLSConfig)
	}
	z.client.SetTimeout(config.Timeout)
	return z
}

// configureTransport applies OpenTelemetry instrumentation to the transport if enabled
func (z *zHTTPClient) configureTransport(transport http.RoundTripper, otelConfig *OpenTelemetryConfig) http.RoundTripper {
	// Return original transport if OpenTelemetry is not configured or disabled
	if otelConfig == nil || !otelConfig.Enabled {
		return transport
	}

	// Configure OpenTelemetry options
	opts := z.buildOpenTelemetryOptions(otelConfig)

	// Apply OpenTelemetry transport with configured options
	return otelhttp.NewTransport(transport, opts...)
}

// buildOpenTelemetryOptions constructs the OpenTelemetry options based on configuration
func (z *zHTTPClient) buildOpenTelemetryOptions(config *OpenTelemetryConfig) []otelhttp.Option {
	var opts []otelhttp.Option

	if config.OperationNameFunc != nil {
		opts = append(opts, otelhttp.WithSpanNameFormatter(config.OperationNameFunc))
	}

	if config.Filters != nil {
		opts = append(opts, otelhttp.WithFilter(config.Filters))
	}

	return opts
}

func (z *zHTTPClient) NewRequest() ZRequest {
	return newZRequest(z)
}

func (z *zHTTPClient) SetRetryPolicy(retryPolicy *RetryPolicy) ZHTTPClient {
	z.retryPolicy = *retryPolicy

	z.client.SetRetryCount(retryPolicy.MaxAttempts)
	z.client.SetRetryWaitTime(retryPolicy.WaitBeforeRetry)
	z.client.SetRetryMaxWaitTime(retryPolicy.MaxWaitBeforeRetry)

	z.client.AddRetryCondition(func(r *resty.Response, err error) bool {
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return true
			}
			return false
		}
		exists := false
		if retryPolicy.retryStatusCodes != nil {
			_, exists = retryPolicy.retryStatusCodes[r.StatusCode()]
		}
		return exists
	})

	// default backoff function is provided by resty
	if retryPolicy.backoffFn != nil {
		z.client.SetRetryAfter(func(c *resty.Client, r *resty.Response) (time.Duration, error) {
			attempt, _ := utils.IntToUInt(r.Request.Attempt)
			return retryPolicy.backoffFn(attempt, r.RawResponse, nil), nil
		})
	}
	return z
}

func (z *zHTTPClient) Do(ctx context.Context, req *http.Request) (*Response, error) {
	resp, err := z.client.GetClient().Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Response{
		Code: resp.StatusCode,
		Body: data,
	}, nil
}

func (z *zHTTPClient) GetHTTPClient() *http.Client {
	return z.client.GetClient()
}
