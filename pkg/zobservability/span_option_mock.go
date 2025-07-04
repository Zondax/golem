// Code generated by mockery. DO NOT EDIT.

package zobservability

import mock "github.com/stretchr/testify/mock"

// MockSpanOption is an autogenerated mock type for the SpanOption type
type MockSpanOption struct {
	mock.Mock
}

type MockSpanOption_Expecter struct {
	mock *mock.Mock
}

func (_m *MockSpanOption) EXPECT() *MockSpanOption_Expecter {
	return &MockSpanOption_Expecter{mock: &_m.Mock}
}

// ApplySpan provides a mock function with given fields: _a0
func (_m *MockSpanOption) ApplySpan(_a0 Span) {
	_m.Called(_a0)
}

// MockSpanOption_ApplySpan_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ApplySpan'
type MockSpanOption_ApplySpan_Call struct {
	*mock.Call
}

// ApplySpan is a helper method to define mock.On call
//   - _a0 Span
func (_e *MockSpanOption_Expecter) ApplySpan(_a0 interface{}) *MockSpanOption_ApplySpan_Call {
	return &MockSpanOption_ApplySpan_Call{Call: _e.mock.On("ApplySpan", _a0)}
}

func (_c *MockSpanOption_ApplySpan_Call) Run(run func(_a0 Span)) *MockSpanOption_ApplySpan_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(Span))
	})
	return _c
}

func (_c *MockSpanOption_ApplySpan_Call) Return() *MockSpanOption_ApplySpan_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockSpanOption_ApplySpan_Call) RunAndReturn(run func(Span)) *MockSpanOption_ApplySpan_Call {
	_c.Run(run)
	return _c
}

// NewMockSpanOption creates a new instance of MockSpanOption. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockSpanOption(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSpanOption {
	mock := &MockSpanOption{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
