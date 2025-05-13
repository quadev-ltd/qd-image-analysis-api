package mock

import (
	"context"
	"reflect"

	"github.com/golang/mock/gomock"
)

type MockImageAnalysisServicer struct {
	ctrl     *gomock.Controller
	recorder *MockImageAnalysisServicerMockRecorder
}

type MockImageAnalysisServicerMockRecorder struct {
	mock *MockImageAnalysisServicer
}

func NewMockImageAnalysisServicer(ctrl *gomock.Controller) *MockImageAnalysisServicer {
	mock := &MockImageAnalysisServicer{ctrl: ctrl}
	mock.recorder = &MockImageAnalysisServicerMockRecorder{mock}
	return mock
}

func (m *MockImageAnalysisServicer) EXPECT() *MockImageAnalysisServicerMockRecorder {
	return m.recorder
}

func (m *MockImageAnalysisServicer) ProcessImageAndPrompt(ctx context.Context, firebaseToken string, imageData []byte, prompt string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessImageAndPrompt", ctx, firebaseToken, imageData, prompt)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockImageAnalysisServicerMockRecorder) ProcessImageAndPrompt(ctx, firebaseToken, imageData, prompt interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessImageAndPrompt", reflect.TypeOf((*MockImageAnalysisServicer)(nil).ProcessImageAndPrompt), ctx, firebaseToken, imageData, prompt)
}
