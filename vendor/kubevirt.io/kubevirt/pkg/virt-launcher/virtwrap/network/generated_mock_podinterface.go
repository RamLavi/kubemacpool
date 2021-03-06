// Automatically generated by MockGen. DO NOT EDIT!
// Source: podinterface.go

package network

import (
	gomock "github.com/golang/mock/gomock"
)

// Mock of BindMechanism interface
type MockBindMechanism struct {
	ctrl     *gomock.Controller
	recorder *_MockBindMechanismRecorder
}

// Recorder for MockBindMechanism (not exported)
type _MockBindMechanismRecorder struct {
	mock *MockBindMechanism
}

func NewMockBindMechanism(ctrl *gomock.Controller) *MockBindMechanism {
	mock := &MockBindMechanism{ctrl: ctrl}
	mock.recorder = &_MockBindMechanismRecorder{mock}
	return mock
}

func (_m *MockBindMechanism) EXPECT() *_MockBindMechanismRecorder {
	return _m.recorder
}

func (_m *MockBindMechanism) discoverPodNetworkInterface() error {
	ret := _m.ctrl.Call(_m, "discoverPodNetworkInterface")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockBindMechanismRecorder) discoverPodNetworkInterface() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "discoverPodNetworkInterface")
}

func (_m *MockBindMechanism) preparePodNetworkInterfaces() error {
	ret := _m.ctrl.Call(_m, "preparePodNetworkInterfaces")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockBindMechanismRecorder) preparePodNetworkInterfaces() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "preparePodNetworkInterfaces")
}

func (_m *MockBindMechanism) decorateConfig() error {
	ret := _m.ctrl.Call(_m, "decorateConfig")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockBindMechanismRecorder) decorateConfig() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "decorateConfig")
}

func (_m *MockBindMechanism) loadCachedInterface(name string) (bool, error) {
	ret := _m.ctrl.Call(_m, "loadCachedInterface", name)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockBindMechanismRecorder) loadCachedInterface(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "loadCachedInterface", arg0)
}

func (_m *MockBindMechanism) setCachedInterface(name string) error {
	ret := _m.ctrl.Call(_m, "setCachedInterface", name)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockBindMechanismRecorder) setCachedInterface(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "setCachedInterface", arg0)
}
