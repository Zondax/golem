package zrouter

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/zondax/golem/pkg/zrouter/domain"
	"net/http"
	"net/http/httptest"
	"testing"
)

type ChiHandlerAdapterSuite struct {
	suite.Suite
}

func (suite *ChiHandlerAdapterSuite) TestChiHandlerAdapter() {
	h := http.Header{}
	h.Add("Content-Type", "application/test")
	handlerFunc := func(ctx Context) (domain.ServiceResponse, error) {
		return domain.NewServiceResponseWithHeader(http.StatusOK, "Hello", h), nil
	}

	httpHandlerFunc := getChiHandler(handlerFunc)

	req, err := http.NewRequest("GET", "/test", bytes.NewBuffer(nil))
	assert.NoError(suite.T(), err)

	recorder := httptest.NewRecorder()

	httpHandlerFunc(recorder, req)

	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
	assert.Equal(suite.T(), "Hello", recorder.Body.String())
	assert.Equal(suite.T(), "application/test", recorder.Header().Get("Content-Type"))
}

func TestChiHandlerAdapterSuite(t *testing.T) {
	suite.Run(t, new(ChiHandlerAdapterSuite))
}
