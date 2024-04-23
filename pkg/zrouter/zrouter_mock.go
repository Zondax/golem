package zrouter

import (
	"github.com/stretchr/testify/mock"
	"github.com/zondax/golem/pkg/zrouter/zmiddlewares"
	"net/http"
)

type MockZRouter struct {
	mock.Mock
}

func (m *MockZRouter) Run(addr ...string) error {
	args := m.Called(addr)
	return args.Error(0)
}

func (m *MockZRouter) GET(path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes {
	args := m.Called(path, handler, middlewares)
	return args.Get(0).(Routes)
}

func (m *MockZRouter) POST(path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes {
	args := m.Called(path, handler, middlewares)
	return args.Get(0).(Routes)
}

func (m *MockZRouter) PUT(path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes {
	args := m.Called(path, handler, middlewares)
	return args.Get(0).(Routes)
}

func (m *MockZRouter) PATCH(path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes {
	args := m.Called(path, handler, middlewares)
	return args.Get(0).(Routes)
}

func (m *MockZRouter) DELETE(path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes {
	args := m.Called(path, handler, middlewares)
	return args.Get(0).(Routes)
}

func (m *MockZRouter) Route(method, path string, handler HandlerFunc, middlewares ...zmiddlewares.Middleware) Routes {
	args := m.Called(method, path, handler, middlewares)
	return args.Get(0).(Routes)
}

func (m *MockZRouter) Use(middlewares ...zmiddlewares.Middleware) Routes {
	args := m.Called(middlewares)
	return args.Get(0).(Routes)
}

func (m *MockZRouter) Group(prefix string) Routes {
	args := m.Called(prefix)
	return args.Get(0).(Routes)
}

func (m *MockZRouter) Handle(pattern string, handler HandlerFunc) {
	m.Called(pattern, handler)
}

func (m *MockZRouter) Mount(pattern string, handler HandlerFunc) {
	m.Called(pattern, handler)
}

func (m *MockZRouter) ServeFiles(routePattern string, httpHandler http.Handler) {
	m.Called(routePattern, httpHandler)
}

func (m *MockZRouter) NoRoute(handler HandlerFunc) {
	m.Called(handler)
}

func (m *MockZRouter) GetRegisteredRoutes() []RegisteredRoute {
	args := m.Called()
	return args.Get(0).([]RegisteredRoute)
}

func (m *MockZRouter) SetDefaultMiddlewares(loggingOptions zmiddlewares.LoggingMiddlewareOptions) {
	m.Called(loggingOptions)
}

func (m *MockZRouter) GetHandler() http.Handler {
	args := m.Called()
	return args.Get(0).(http.Handler)
}

func (m *MockZRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m.Called(w, req)
}
