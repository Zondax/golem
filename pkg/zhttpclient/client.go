package httpclient

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
)

type ZHTTPClient interface {
	SetRetryPolicy(retryPolicy *RetryPolicy) error
	Post(ctx context.Context, url string, body io.Reader, headers map[string]string) (int, []byte, error)
	Get(ctx context.Context, url string, headers map[string]string, params url.Values) (int, []byte, error)
	Do(ctx context.Context, req *http.Request) (int, []byte, error)
}

type Config struct {
	Timeout    time.Duration
	TLSConfig  *tls.Config
	BaseClient *http.Client
}

// zHTTPClient abstracts over the std http.Client and provides a retry mechanism.
type zHTTPClient struct {
	client *resty.Client
}

func NewZHTTPClient(config Config) ZHTTPClient {
	z := &zHTTPClient{}

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

func (z *zHTTPClient) SetRetryPolicy(retryPolicy *RetryPolicy) error {
	if retryPolicy == nil {
		return errors.New("retryPolicy cannot be nil")
	}
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
			return retryPolicy.backoffFn(uint(r.Request.Attempt), r.RawResponse, nil), nil
		})
	}
	return nil
}

func (z *zHTTPClient) Post(ctx context.Context, url string, body io.Reader, headers map[string]string) (int, []byte, error) {
	req := z.client.R().SetContext(ctx).SetHeaders(headers).SetBody(body)
	resp, err := req.Post(url)
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode(), resp.Body(), nil
}

func (z *zHTTPClient) Get(ctx context.Context, url string, headers map[string]string, params url.Values) (int, []byte, error) {
	req := z.client.R().SetContext(ctx).SetHeaders(headers).SetQueryParamsFromValues(params)

	resp, err := req.Get(url)
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode(), resp.Body(), nil
}

func (z *zHTTPClient) Do(ctx context.Context, req *http.Request) (int, []byte, error) {
	resp, err := z.client.GetClient().Do(req.WithContext(ctx))
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, data, nil
}
