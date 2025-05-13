package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	commonLog "github.com/quadev-ltd/qd-common/pkg/log"
	"github.com/stretchr/testify/assert"

	"qd-image-analysis-api/internal/service/mock"
)

func TestProcessImageAndPrompt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock.NewMockImageAnalysisServicer(ctrl)

	server := NewImageAnalysisServiceServer(mockService)

	logFactory := commonLog.NewLogFactory("test")
	logger := logFactory.NewLogger()
	ctx := commonLog.ContextWithLogger(context.Background(), logger)

	testToken := "test-firebase-token"
	testImageData := []byte("test-image-data")
	testPrompt := "test prompt"
	testResponse := "test response"

	mockService.EXPECT().
		ProcessImageAndPrompt(ctx, testToken, testImageData, testPrompt).
		Return(testResponse, nil)

	response, err := server.ProcessImageAndPrompt(ctx, testToken, testImageData, testPrompt)

	assert.NoError(t, err)
	assert.Equal(t, testResponse, response)

	testError := errors.New("test error")
	mockService.EXPECT().
		ProcessImageAndPrompt(ctx, testToken, testImageData, testPrompt).
		Return("", testError)

	response, err = server.ProcessImageAndPrompt(ctx, testToken, testImageData, testPrompt)

	assert.Error(t, err)
	assert.Equal(t, "", response)
}
