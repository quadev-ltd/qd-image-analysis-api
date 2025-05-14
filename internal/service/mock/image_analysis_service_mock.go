package mock

import (
	"context"
	"reflect"

	"github.com/golang/mock/gomock"
)

// ImageAnalysisServicer is a mock of the ImageAnalysisServicer interface
type ImageAnalysisServicer struct {
	ctrl     *gomock.Controller
	recorder *ImageAnalysisServicerMockRecorder
}

// ImageAnalysisServicerMockRecorder is the mock recorder for ImageAnalysisServicer
type ImageAnalysisServicerMockRecorder struct {
	mock *ImageAnalysisServicer
}

// NewImageAnalysisServicer creates a new mock instance
func NewImageAnalysisServicer(ctrl *gomock.Controller) *ImageAnalysisServicer {
	mock := &ImageAnalysisServicer{ctrl: ctrl}
	mock.recorder = &ImageAnalysisServicerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *ImageAnalysisServicer) EXPECT() *ImageAnalysisServicerMockRecorder {
	return m.recorder
}

// ProcessImageAndPrompt mocks base method
func (m *ImageAnalysisServicer) ProcessImageAndPrompt(ctx context.Context, firebaseToken string, imageData []byte, prompt string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessImageAndPrompt", ctx, firebaseToken, imageData, prompt)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProcessImageAndPrompt indicates an expected call of ProcessImageAndPrompt
func (mr *ImageAnalysisServicerMockRecorder) ProcessImageAndPrompt(ctx, firebaseToken, imageData, prompt interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessImageAndPrompt", reflect.TypeOf((*ImageAnalysisServicer)(nil).ProcessImageAndPrompt), ctx, firebaseToken, imageData, prompt)
}
