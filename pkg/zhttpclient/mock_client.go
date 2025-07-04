// Code generated by mockery v2.20.0. DO NOT EDIT.

package zhttpclient

import (
	context "context"
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// MockZHTTPClient is an autogenerated mock type for the ZHTTPClient type
type MockZHTTPClient struct {
	mock.Mock
}

// Do provides a mock function with given fields: ctx, req
func (_m *MockZHTTPClient) Do(ctx context.Context, req *http.Request) (*Response, error) {
	ret := _m.Called(ctx, req)

	var r0 *Response
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *http.Request) (*Response, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *http.Request) *Response); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Response)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *http.Request) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewRequest provides a mock function with given fields:
func (_m *MockZHTTPClient) NewRequest() ZRequest {
	ret := _m.Called()

	var r0 ZRequest
	if rf, ok := ret.Get(0).(func() ZRequest); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(ZRequest)
		}
	}

	return r0
}

// SetRetryPolicy provides a mock function with given fields: retryPolicy
func (_m *MockZHTTPClient) SetRetryPolicy(retryPolicy *RetryPolicy) ZHTTPClient {
	ret := _m.Called(retryPolicy)

	var r0 ZHTTPClient
	if rf, ok := ret.Get(0).(func(*RetryPolicy) ZHTTPClient); ok {
		r0 = rf(retryPolicy)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(ZHTTPClient)
		}
	}

	return r0
}


func (m *MockZHTTPClient) GetHTTPClient() *http.Client {
	ret := m.Called()

	var r0 *http.Client
	if rf, ok := ret.Get(0).(func() *http.Client); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Client)
		}
	}

	return r0
}
type mockConstructorTestingTNewMockZHTTPClient interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockZHTTPClient creates a new instance of MockZHTTPClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockZHTTPClient(t mockConstructorTestingTNewMockZHTTPClient) *MockZHTTPClient {
	mock := &MockZHTTPClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
