package zhttpclient

import (
	"context"
	"crypto/tls"
	"github.com/zondax/golem/pkg/utils"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

type ZHTTPClient interface {
	SetRetryPolicy(retryPolicy *RetryPolicy) ZHTTPClient
	NewRequest() ZRequest
	Do(ctx context.Context, req *http.Request) (*Response, error)
}

type Config struct {
	Timeout    time.Duration
	TLSConfig  *tls.Config
	BaseClient *http.Client
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

	if config.BaseClient == nil {
		z.client = resty.New()
	} else {
		z.client = resty.NewWithClient(config.BaseClient)
	}

	if config.TLSConfig != nil {
		z.client.SetTLSClientConfig(config.TLSConfig)
	}
	z.client.SetTimeout(config.Timeout)
	return z
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
