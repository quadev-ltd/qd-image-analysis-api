package grpcserver

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	commonPB "github.com/quadev-ltd/qd-common/pb/gen/go/pb_image_analysis"
	commonLog "github.com/quadev-ltd/qd-common/pkg/log"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"qd-image-analysis-api/internal/service"
	"qd-image-analysis-api/internal/service/mock"
)

func TestProcessImageAndPrompt_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock.NewMockImageAnalysisServicer(ctrl)
	server := NewImageAnalysisServiceServer(mockService)

	logFactory := commonLog.NewLogFactory("test")
	logger := logFactory.NewLogger()
	ctx := context.WithValue(context.Background(), commonLog.LoggerKey, logger)

	testImageData := []byte("test-image-data")
	testPrompt := "test prompt"
	testMimeType := "image/png"
	testResponse := "test response"

	mockService.EXPECT().
		ProcessImageAndPrompt(gomock.Any(), testImageData, testMimeType, testPrompt).
		Return(testResponse, nil)

	request := &commonPB.ImagePromptRequest{
		ImageData: testImageData,
		Prompt:    testPrompt,
		MimeType:  testMimeType,
	}

	response, err := server.ProcessImageAndPrompt(ctx, request)

	assert.NoError(t, err)
	assert.Equal(t, testResponse, response.ResponseToPrompt)
}

func TestProcessImageAndPrompt_RateLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock.NewMockImageAnalysisServicer(ctrl)
	server := NewImageAnalysisServiceServer(mockService)

	logFactory := commonLog.NewLogFactory("test")
	logger := logFactory.NewLogger()
	ctx := context.WithValue(context.Background(), commonLog.LoggerKey, logger)

	testImageData := []byte("test-image-data")
	testPrompt := "test prompt"
	testMimeType := "image/png"

	// First few calls should succeed
	for i := 0; i < 1; i++ {
		mockService.EXPECT().
			ProcessImageAndPrompt(gomock.Any(), testImageData, testMimeType, testPrompt).
			Return("success", nil)
	}

	request := &commonPB.ImagePromptRequest{
		ImageData: testImageData,
		Prompt:    testPrompt,
		MimeType:  testMimeType,
	}

	// Make successful calls first
	for i := 0; i < 1; i++ {
		response, err := server.ProcessImageAndPrompt(ctx, request)
		assert.NoError(t, err)
		assert.Equal(t, "success", response.ResponseToPrompt)
	}

	// This call should trigger rate limit
	response, err := server.ProcessImageAndPrompt(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, response)

	status, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.ResourceExhausted.String(), status.Code().String())
	assert.Contains(t, status.Message(), "Too many requests")
}

func TestProcessImageAndPrompt_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock.NewMockImageAnalysisServicer(ctrl)
	server := NewImageAnalysisServiceServer(mockService)

	logFactory := commonLog.NewLogFactory("test")
	logger := logFactory.NewLogger()
	ctx := context.WithValue(context.Background(), commonLog.LoggerKey, logger)

	testImageData := []byte("test-image-data")
	testPrompt := "test prompt"
	testMimeType := "image/png"
	serviceError := &service.Error{
		Message: "Invalid argument",
	}

	mockService.EXPECT().
		ProcessImageAndPrompt(gomock.Any(), testImageData, testMimeType, testPrompt).
		Return("", serviceError)

	request := &commonPB.ImagePromptRequest{
		ImageData: testImageData,
		Prompt:    testPrompt,
		MimeType:  testMimeType,
	}

	response, err := server.ProcessImageAndPrompt(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, response)

	status, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument.String(), status.Code().String())
	assert.Contains(t, status.Message(), "Invalid argument")
}

func TestProcessImageAndPrompt_RegularError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock.NewMockImageAnalysisServicer(ctrl)
	server := NewImageAnalysisServiceServer(mockService)

	logFactory := commonLog.NewLogFactory("test")
	logger := logFactory.NewLogger()
	ctx := context.WithValue(context.Background(), commonLog.LoggerKey, logger)

	testImageData := []byte("test-image-data")
	testPrompt := "test prompt"
	testMimeType := "image/png"

	mockService.EXPECT().
		ProcessImageAndPrompt(gomock.Any(), testImageData, testMimeType, testPrompt).
		Return("", errors.New("unexpected error"))

	request := &commonPB.ImagePromptRequest{
		ImageData: testImageData,
		Prompt:    testPrompt,
		MimeType:  testMimeType,
	}

	response, err := server.ProcessImageAndPrompt(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, response)

	status, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal.String(), status.Code().String())
	assert.Contains(t, status.Message(), "Error processing image and prompt")
}
