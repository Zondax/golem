package zrouter

import (
	"github.com/stretchr/testify/mock"
	"net/http"
)

type MockContext struct {
	mock.Mock
}

func (m *MockContext) Request() *http.Request {
	args := m.Called()
	return args.Get(0).(*http.Request)
}

func (m *MockContext) BindJSON(obj interface{}) error {
	args := m.Called(obj)
	return args.Error(0)
}

func (m *MockContext) Header(key, value string) {
	m.Called(key, value)
}

func (m *MockContext) Param(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *MockContext) Query(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *MockContext) DefaultQuery(key, defaultValue string) string {
	args := m.Called(key, defaultValue)
	return args.String(0)
}
