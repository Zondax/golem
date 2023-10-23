package zrouter

import (
	"github.com/stretchr/testify/mock"
	"github.com/zondax/golem/pkg/zrouter/zmiddlewares"
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
