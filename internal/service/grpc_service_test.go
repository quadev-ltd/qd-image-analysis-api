package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	commonPB "github.com/quadev-ltd/qd-common/pb/gen/go/pb_image_analysis"
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
	ctx := context.WithValue(context.Background(), commonLog.LoggerKey, logger)

	testImageData := []byte("test-image-data")
	testPrompt := "test prompt"
	testResponse := "test response"

	mockService.EXPECT().
		ProcessImageAndPrompt(gomock.Any(), testImageData, testPrompt).
		Return(testResponse, nil)

	request := &commonPB.ImagePromptRequest{
		ImageData: testImageData,
		Prompt:    testPrompt,
	}

	response, err := server.ProcessImageAndPrompt(ctx, request)

	assert.NoError(t, err)
	assert.Equal(t, testResponse, response.ResponseToPrompt)

	testError := errors.New("test error")
	mockService.EXPECT().
		ProcessImageAndPrompt(gomock.Any(), testImageData, testPrompt).
		Return("", testError)

	response, err = server.ProcessImageAndPrompt(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, response)
}
