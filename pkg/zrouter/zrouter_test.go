package zrouter

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/zrouter/domain"
	"net/http"
	"net/http/httptest"
	"testing"
)

type ZRouterSuite struct {
	suite.Suite
	router ZRouter
}

func (suite *ZRouterSuite) SetupTest() {
	suite.router = New(nil, &Config{AppVersion: "app_version", AppRevision: "app_revision"})
	logger.InitLogger(logger.Config{})
}

func (suite *ZRouterSuite) TestRegisterAndGetRoutes() {
	suite.router.GET("/get", func(ctx Context) (domain.ServiceResponse, error) {
		return domain.NewServiceResponse(http.StatusOK, []byte("GET OK")), nil
	})

	suite.router.POST("/post", func(ctx Context) (domain.ServiceResponse, error) {
		return domain.NewServiceResponse(http.StatusOK, []byte("POST OK")), nil
	})

	routes := suite.router.GetRegisteredRoutes()

	assert.Len(suite.T(), routes, 2)
	assert.Contains(suite.T(), routes, RegisteredRoute{Method: "GET", Path: "/get"})
	assert.Contains(suite.T(), routes, RegisteredRoute{Method: "POST", Path: "/post"})
}

func (suite *ZRouterSuite) TestRouteHandling() {
	suite.router.GET("/test", func(ctx Context) (domain.ServiceResponse, error) {
		return domain.NewServiceResponse(http.StatusOK, "test route"), nil
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	handler := suite.router.GetHandler()
	handler.ServeHTTP(recorder, req)

	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
	assert.Equal(suite.T(), "\"test route\"", recorder.Body.String())
}

func TestZRouterSuite(t *testing.T) {
	suite.Run(t, new(ZRouterSuite))
}

func TestValidateAppVersionAndRevision(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
			return
		}
		errorMessage, ok := r.(string)
		if !ok {
			t.Errorf("Expected panic with a string message but got %T", r)
			return
		}

		expectedMessage := "appVersion and appRevision are mandatory."
		if errorMessage != expectedMessage {
			t.Errorf("Expected panic with message %q but got %q", expectedMessage, errorMessage)
		}
	}()

	New(nil, nil)
}
