package mocks

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockPrompter is a mock of Prompter interface.
type MockPrompter struct {
	ctrl     *gomock.Controller
	recorder *MockPrompterMockRecorder
}

// MockPrompterMockRecorder is the mock recorder for MockPrompter.
type MockPrompterMockRecorder struct {
	mock *MockPrompter
}

// NewMockPrompter creates a new mock instance.
func NewMockPrompter(ctrl *gomock.Controller) *MockPrompter {
	mock := &MockPrompter{ctrl: ctrl}
	mock.recorder = &MockPrompterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPrompter) EXPECT() *MockPrompterMockRecorder {
	return m.recorder
}

// Confirm mocks base method.
func (m *MockPrompter) Confirm(arg0 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Confirm", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Confirm indicates an expected call of Confirm.
func (mr *MockPrompterMockRecorder) Confirm(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Confirm", reflect.TypeOf((*MockPrompter)(nil).Confirm), arg0)
}

// InputHiddenString mocks base method.
func (m *MockPrompter) InputHiddenString(arg0, arg1 string, arg2 func(string) error) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InputHiddenString", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InputHiddenString indicates an expected call of InputHiddenString.
func (mr *MockPrompterMockRecorder) InputHiddenString(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InputHiddenString", reflect.TypeOf((*MockPrompter)(nil).InputHiddenString), arg0, arg1, arg2)
}

// InputInteger mocks base method.
func (m *MockPrompter) InputInteger(arg0, arg1, arg2 string, arg3 func(int64) error) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InputInteger", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InputInteger indicates an expected call of InputInteger.
func (mr *MockPrompterMockRecorder) InputInteger(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InputInteger", reflect.TypeOf((*MockPrompter)(nil).InputInteger), arg0, arg1, arg2, arg3)
}

// InputString mocks base method.
func (m *MockPrompter) InputString(arg0, arg1, arg2 string, arg3 func(string) error) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InputString", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InputString indicates an expected call of InputString.
func (mr *MockPrompterMockRecorder) InputString(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InputString", reflect.TypeOf((*MockPrompter)(nil).InputString), arg0, arg1, arg2, arg3)
}

// Select mocks base method.
func (m *MockPrompter) Select(arg0 string, arg1 []string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Select", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Select indicates an expected call of Select.
func (mr *MockPrompterMockRecorder) Select(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Select", reflect.TypeOf((*MockPrompter)(nil).Select), arg0, arg1)
}
