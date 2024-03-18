package zhttpclient

import (
	"context"
	"io"
	"net/url"

	"github.com/go-resty/resty/v2"
)

type ZRequest interface {
	SetURL(url string) ZRequest
	SetHeaders(headers map[string]string) ZRequest
	SetBody(body io.Reader) ZRequest
	SetQueryParams(params url.Values) ZRequest
	SetRetryPolicy(retryPolicy *RetryPolicy) ZRequest
	SetResult(result interface{}) ZRequest
	SetError(err interface{}) ZRequest
	Post(ctx context.Context) (*Response, error)
	Get(ctx context.Context) (*Response, error)
}

type Response struct {
	Code   int
	Result interface{}
	Error  interface{}
	Body   []byte
}

type zRequest struct {
	c       *zHTTPClient
	request *resty.Request
	url     string
}

func newZRequest(client *zHTTPClient) ZRequest {
	// only used to enforce retry policies at the request level
	c := New(*client.config).(*zHTTPClient)
	tmp := client.retryPolicy
	c.SetRetryPolicy(&tmp)

	return &zRequest{
		c:       c,
		request: c.client.R(),
	}
}

func (r *zRequest) SetURL(url string) ZRequest {
	r.url = url
	return r
}

func (r *zRequest) SetHeaders(headers map[string]string) ZRequest {
	for k, v := range headers {
		r.request.Header.Add(k, v)
	}
	return r
}

func (r *zRequest) SetBody(body io.Reader) ZRequest {
	r.request.SetBody(body)
	// r.request.SetResult(res interface{})
	return r
}

func (r *zRequest) SetResult(result interface{}) ZRequest {
	r.request.SetResult(result)
	return r
}

func (r *zRequest) SetError(err interface{}) ZRequest {
	r.request.SetError(err)
	return r
}

func (r *zRequest) SetQueryParams(params url.Values) ZRequest {
	r.request.SetQueryParamsFromValues(params)
	return r
}

func (r *zRequest) SetRetryPolicy(retryPolicy *RetryPolicy) ZRequest {
	// override client retry policy
	r.c.client.RetryConditions = nil
	r.c.client.RetryAfter = nil

	r.c.SetRetryPolicy(retryPolicy)
	return r
}

func (r *zRequest) Post(ctx context.Context) (*Response, error) {
	resp, err := r.request.SetContext(ctx).Post(r.url)
	if err != nil {
		return nil, err
	}
	return &Response{
		Code:   resp.StatusCode(),
		Result: resp.Result(),
		Error:  resp.Error(),
		Body:   resp.Body(),
	}, nil
}

func (r *zRequest) Get(ctx context.Context) (*Response, error) {
	resp, err := r.request.SetContext(ctx).Get(r.url)
	if err != nil {
		return nil, err
	}

	return &Response{
		Code:   resp.StatusCode(),
		Result: resp.Result(),
		Error:  resp.Error(),
		Body:   resp.Body(),
	}, nil
}
