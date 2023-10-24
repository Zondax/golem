package zrouter

import (
	"bytes"
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
	suite.Require().NoError(err)

	recorder := httptest.NewRecorder()

	httpHandlerFunc(recorder, req)

	suite.Equal(http.StatusOK, recorder.Code)
	suite.Equal("\"Hello\"", recorder.Body.String())
	suite.Equal("application/test", recorder.Header().Get("Content-Type"))
}

func TestChiHandlerAdapterSuite(t *testing.T) {
	suite.Run(t, new(ChiHandlerAdapterSuite))
}
