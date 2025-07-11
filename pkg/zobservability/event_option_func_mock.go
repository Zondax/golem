// Code generated by mockery. DO NOT EDIT.

package zobservability

import mock "github.com/stretchr/testify/mock"

// MockeventOptionFunc is an autogenerated mock type for the eventOptionFunc type
type MockeventOptionFunc struct {
	mock.Mock
}

type MockeventOptionFunc_Expecter struct {
	mock *mock.Mock
}

func (_m *MockeventOptionFunc) EXPECT() *MockeventOptionFunc_Expecter {
	return &MockeventOptionFunc_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: _a0
func (_m *MockeventOptionFunc) Execute(_a0 Event) {
	_m.Called(_a0)
}

// MockeventOptionFunc_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type MockeventOptionFunc_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - _a0 Event
func (_e *MockeventOptionFunc_Expecter) Execute(_a0 interface{}) *MockeventOptionFunc_Execute_Call {
	return &MockeventOptionFunc_Execute_Call{Call: _e.mock.On("Execute", _a0)}
}

func (_c *MockeventOptionFunc_Execute_Call) Run(run func(_a0 Event)) *MockeventOptionFunc_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(Event))
	})
	return _c
}

func (_c *MockeventOptionFunc_Execute_Call) Return() *MockeventOptionFunc_Execute_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockeventOptionFunc_Execute_Call) RunAndReturn(run func(Event)) *MockeventOptionFunc_Execute_Call {
	_c.Run(run)
	return _c
}

// NewMockeventOptionFunc creates a new instance of MockeventOptionFunc. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockeventOptionFunc(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockeventOptionFunc {
	mock := &MockeventOptionFunc{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
