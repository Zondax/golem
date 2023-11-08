package domain

import (
	"encoding/json"
	"net/http"
	"sync"
)

const (
	ContentTypeHeader          = "Content-Type"
	contentTypeApplicationJSON = "application/json; charset=utf-8"
	ContentTypeJSON            = "json"
)

type ServiceResponse interface {
	Status() int
	Header() http.Header
	ResponseBytes() ([]byte, error)
	ResponseFormat() string
	Contents() interface{}
}

type defaultServiceResponse struct {
	status        int
	header        http.Header
	response      interface{}
	once          sync.Once
	responseBytes []byte
	marshalError  error
}

func (d *defaultServiceResponse) Status() int {
	return d.status
}

func (d *defaultServiceResponse) Header() http.Header {
	h := d.header
	if h == nil {
		h = http.Header{}
	}
	if h.Get(ContentTypeHeader) == "" {
		h.Set(ContentTypeHeader, contentTypeApplicationJSON)
	}
	return h
}

func (d *defaultServiceResponse) ResponseFormat() string {
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

func NewServiceResponse(status int, response interface{}) ServiceResponse {
	return &defaultServiceResponse{
		status:   status,
		response: response,
		header:   nil,
	}
}

func NewServiceResponseWithHeader(status int, response interface{}, header http.Header) ServiceResponse {
	return &defaultServiceResponse{
		status:   status,
		response: response,
		header:   header,
	}
}

func NewErrorResponse(status int, errorCode, errMsg string) ServiceResponse {
	apiError := NewAPIErrorResponse(status, errorCode, errMsg)
	apiErrorBytes, err := json.Marshal(apiError)
	if err != nil {
		return NewServiceResponse(status, errMsg)
	}

	return &defaultServiceResponse{
		status:        status,
		response:      apiError,
		header:        nil,
		responseBytes: apiErrorBytes,
	}
}

func NewErrorNotFound(errMsg string) ServiceResponse {
	return NewErrorResponse(http.StatusNotFound, "ROUTE_NOT_FOUND", errMsg)
}
