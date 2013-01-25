// Automatically generated by MockGen. DO NOT EDIT!
// Source: session.go

package seven5

import (
	http "net/http"
	auth "seven5/auth"
	gomock "code.google.com/p/gomock/gomock"
)

// Mock of SessionManager interface
type MockSessionManager struct {
	ctrl     *gomock.Controller
	recorder *_MockSessionManagerRecorder
}

// Recorder for MockSessionManager (not exported)
type _MockSessionManagerRecorder struct {
	mock *MockSessionManager
}

func NewMockSessionManager(ctrl *gomock.Controller) *MockSessionManager {
	mock := &MockSessionManager{ctrl: ctrl}
	mock.recorder = &_MockSessionManagerRecorder{mock}
	return mock
}

func (_m *MockSessionManager) EXPECT() *_MockSessionManagerRecorder {
	return _m.recorder
}

func (_m *MockSessionManager) Find(id string) (Session, error) {
	ret := _m.ctrl.Call(_m, "Find", id)
	ret0, _ := ret[0].(Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockSessionManagerRecorder) Find(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Find", arg0)
}

func (_m *MockSessionManager) Generate(c auth.OauthConnection, r *http.Request, state string, code string) (Session, error) {
	ret := _m.ctrl.Call(_m, "Generate", c, r, state, code)
	ret0, _ := ret[0].(Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockSessionManagerRecorder) Generate(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Generate", arg0, arg1, arg2, arg3)
}

func (_m *MockSessionManager) Destroy(id string) error {
	ret := _m.ctrl.Call(_m, "Destroy", id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockSessionManagerRecorder) Destroy(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Destroy", arg0)
}

// Mock of Session interface
type MockSession struct {
	ctrl     *gomock.Controller
	recorder *_MockSessionRecorder
}

// Recorder for MockSession (not exported)
type _MockSessionRecorder struct {
	mock *MockSession
}

func NewMockSession(ctrl *gomock.Controller) *MockSession {
	mock := &MockSession{ctrl: ctrl}
	mock.recorder = &_MockSessionRecorder{mock}
	return mock
}

func (_m *MockSession) EXPECT() *_MockSessionRecorder {
	return _m.recorder
}

func (_m *MockSession) SessionId() string {
	ret := _m.ctrl.Call(_m, "SessionId")
	ret0, _ := ret[0].(string)
	return ret0
}

func (_mr *_MockSessionRecorder) SessionId() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SessionId")
}
