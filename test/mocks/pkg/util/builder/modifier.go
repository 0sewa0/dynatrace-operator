// Code generated by mockery. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Modifier is an autogenerated mock type for the Modifier type
type Modifier[T interface{}] struct {
	mock.Mock
}

type Modifier_Expecter[T interface{}] struct {
	mock *mock.Mock
}

func (_m *Modifier[T]) EXPECT() *Modifier_Expecter[T] {
	return &Modifier_Expecter[T]{mock: &_m.Mock}
}

// Enabled provides a mock function with no fields
func (_m *Modifier[T]) Enabled() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Enabled")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Modifier_Enabled_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Enabled'
type Modifier_Enabled_Call[T interface{}] struct {
	*mock.Call
}

// Enabled is a helper method to define mock.On call
func (_e *Modifier_Expecter[T]) Enabled() *Modifier_Enabled_Call[T] {
	return &Modifier_Enabled_Call[T]{Call: _e.mock.On("Enabled")}
}

func (_c *Modifier_Enabled_Call[T]) Run(run func()) *Modifier_Enabled_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Modifier_Enabled_Call[T]) Return(_a0 bool) *Modifier_Enabled_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Modifier_Enabled_Call[T]) RunAndReturn(run func() bool) *Modifier_Enabled_Call[T] {
	_c.Call.Return(run)
	return _c
}

// Modify provides a mock function with given fields: _a0
func (_m *Modifier[T]) Modify(_a0 *T) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Modify")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*T) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Modifier_Modify_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Modify'
type Modifier_Modify_Call[T interface{}] struct {
	*mock.Call
}

// Modify is a helper method to define mock.On call
//   - _a0 *T
func (_e *Modifier_Expecter[T]) Modify(_a0 interface{}) *Modifier_Modify_Call[T] {
	return &Modifier_Modify_Call[T]{Call: _e.mock.On("Modify", _a0)}
}

func (_c *Modifier_Modify_Call[T]) Run(run func(_a0 *T)) *Modifier_Modify_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*T))
	})
	return _c
}

func (_c *Modifier_Modify_Call[T]) Return(_a0 error) *Modifier_Modify_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Modifier_Modify_Call[T]) RunAndReturn(run func(*T) error) *Modifier_Modify_Call[T] {
	_c.Call.Return(run)
	return _c
}

// NewModifier creates a new instance of Modifier. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewModifier[T interface{}](t interface {
	mock.TestingT
	Cleanup(func())
}) *Modifier[T] {
	mock := &Modifier[T]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
