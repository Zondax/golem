package domain

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

const (
	ContentTypeHeader          = "Content-Type"
	contentTypeApplicationJSON = "application/json; charset=utf-8"
	ContentTypeJSON            = "json"
)

var HeaderCacheControl = http.CanonicalHeaderKey("Cache-Control")

type ServiceResponse interface {
	Status() int
	Header() http.Header
	ResponseBytes() ([]byte, error)
	RespondMethod() string
	SetCache(maxAge int)
	Contents() interface{}
}

type baseServiceResponse struct {
	status int
	header http.Header
	cache  int
}

type jsonServiceResponse struct {
	baseServiceResponse
}

type defaultServiceResponse struct {
	jsonServiceResponse
	response      interface{}
	once          sync.Once
	responseBytes []byte
	marshalError  error
}

type customServiceResponse struct {
	jsonServiceResponse
	response []byte
}

func (b *baseServiceResponse) Status() int {
	return b.status
}

func (b *baseServiceResponse) SetCache(maxAge int) {
	b.cache = maxAge
}

func (b *baseServiceResponse) cloneHeader() http.Header {
	h := b.header
	if b.header == nil {
		h = http.Header{}
	}

	if b.cache > 0 {
		h[HeaderCacheControl] = []string{fmt.Sprintf("private, max-age=%d", b.cache)}
	}

	return h.Clone()
}

func (j *jsonServiceResponse) Header() http.Header {
	result := j.cloneHeader()
	if result.Get(ContentTypeHeader) == "" {
		result.Set(ContentTypeHeader, contentTypeApplicationJSON)
	}
	return result
}

func (j *jsonServiceResponse) RespondMethod() string {
	return ContentTypeJSON
}

func (d *defaultServiceResponse) ResponseBytes() ([]byte, error) {
	d.once.Do(func() {
		if d.response != nil {
			d.responseBytes, d.marshalError = json.Marshal(d.response)
		} else {
			d.responseBytes = []byte{}
		}
	})
	return d.responseBytes, d.marshalError
}

func (d *defaultServiceResponse) Contents() interface{} {
	return d.response
}

func (c *customServiceResponse) ResponseBytes() ([]byte, error) {
	return c.response, nil
}

func (c *customServiceResponse) Contents() interface{} {
	return c.response
}

func NewNoContentServiceResponse() ServiceResponse {
	return NewServiceResponse(http.StatusNoContent, nil)
}

func NewCreatedServiceResponse(response ...interface{}) ServiceResponse {
	var res interface{}
	if len(response) > 0 {
		res = response[0]
	}
	return NewServiceResponse(http.StatusCreated, res)
}

func NewServiceResponse(status int, response interface{}) ServiceResponse {
	return NewServiceResponseWith(status, response, nil)
}

func NewServiceResponseWith(status int, response interface{}, header http.Header) ServiceResponse {
	return &defaultServiceResponse{
		jsonServiceResponse: jsonServiceResponse{
			baseServiceResponse{
				header: header,
				status: status,
			},
		},
		response: response,
	}
}

func NewCustomServiceResponse(status int, response string) ServiceResponse {
	return NewCustomServiceResponseWith(status, response, nil)
}

func NewSuccessResponse(status int, data interface{}) ServiceResponse {
	return NewServiceResponse(status, data)
}

func NewErrorResponse(status int, errorCode, errMsg string) ServiceResponse {
	apiError := NewAPIErrorResponse(status, errorCode, errMsg)
	apiErrorBytes, err := json.Marshal(apiError)
	if err != nil {
		zap.S().Error(err.Error())
		return NewCustomServiceResponse(status, errMsg)
	}

	return NewCustomServiceResponseBytes(status, apiErrorBytes, nil)
}

func NewErrorNotFound(errMsg string) ServiceResponse {
	return NewErrorResponse(http.StatusNotFound, "ROUTE_NOT_FOUND", errMsg)
}

func NewCustomServiceResponseWith(status int, response string, header http.Header) ServiceResponse {
	return NewCustomServiceResponseBytes(status, []byte(response), header)
}

func NewCustomServiceResponseBytes(status int, bytes []byte, header http.Header) ServiceResponse {
	return &customServiceResponse{
		jsonServiceResponse: jsonServiceResponse{
			baseServiceResponse{
				header: header,
				status: status,
			},
		},
		response: bytes,
	}
}
