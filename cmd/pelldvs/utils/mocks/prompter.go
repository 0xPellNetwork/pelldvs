// Code generated by mockery. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Prompter is an autogenerated mock type for the Prompter type
type Prompter struct {
	mock.Mock
}

// Confirm provides a mock function with given fields: prompt
func (_m *Prompter) Confirm(prompt string) (bool, error) {
	ret := _m.Called(prompt)

	if len(ret) == 0 {
		panic("no return value specified for Confirm")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (bool, error)); ok {
		return rf(prompt)
	}
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(prompt)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(prompt)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// InputHiddenString provides a mock function with given fields: prompt, help, validator
func (_m *Prompter) InputHiddenString(prompt string, help string, validator func(string) error) (string, error) {
	ret := _m.Called(prompt, help, validator)

	if len(ret) == 0 {
		panic("no return value specified for InputHiddenString")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, func(string) error) (string, error)); ok {
		return rf(prompt, help, validator)
	}
	if rf, ok := ret.Get(0).(func(string, string, func(string) error) string); ok {
		r0 = rf(prompt, help, validator)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string, string, func(string) error) error); ok {
		r1 = rf(prompt, help, validator)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// InputInteger provides a mock function with given fields: prompt, defValue, help, validator
func (_m *Prompter) InputInteger(prompt string, defValue string, help string, validator func(int64) error) (int64, error) {
	ret := _m.Called(prompt, defValue, help, validator)

	if len(ret) == 0 {
		panic("no return value specified for InputInteger")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, string, func(int64) error) (int64, error)); ok {
		return rf(prompt, defValue, help, validator)
	}
	if rf, ok := ret.Get(0).(func(string, string, string, func(int64) error) int64); ok {
		r0 = rf(prompt, defValue, help, validator)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(string, string, string, func(int64) error) error); ok {
		r1 = rf(prompt, defValue, help, validator)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// InputString provides a mock function with given fields: prompt, defValue, help, validator
func (_m *Prompter) InputString(prompt string, defValue string, help string, validator func(string) error) (string, error) {
	ret := _m.Called(prompt, defValue, help, validator)

	if len(ret) == 0 {
		panic("no return value specified for InputString")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, string, func(string) error) (string, error)); ok {
		return rf(prompt, defValue, help, validator)
	}
	if rf, ok := ret.Get(0).(func(string, string, string, func(string) error) string); ok {
		r0 = rf(prompt, defValue, help, validator)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string, string, string, func(string) error) error); ok {
		r1 = rf(prompt, defValue, help, validator)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Select provides a mock function with given fields: prompt, options
func (_m *Prompter) Select(prompt string, options []string) (string, error) {
	ret := _m.Called(prompt, options)

	if len(ret) == 0 {
		panic("no return value specified for Select")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string, []string) (string, error)); ok {
		return rf(prompt, options)
	}
	if rf, ok := ret.Get(0).(func(string, []string) string); ok {
		r0 = rf(prompt, options)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string, []string) error); ok {
		r1 = rf(prompt, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewPrompter creates a new instance of Prompter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPrompter(t interface {
	mock.TestingT
	Cleanup(func())
}) *Prompter {
	mock := &Prompter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
