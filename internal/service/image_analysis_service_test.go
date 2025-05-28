package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/quadev-ltd/qd-common/pkg/log"
	loggerMock "github.com/quadev-ltd/qd-common/pkg/log/mock"
	"github.com/stretchr/testify/assert"

	"qd-image-analysis-api/internal/ai/mock"
)

func TestNewImageAnalysisService(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockAnalyzer := mock.NewMockAnalyzer(controller)
	service := NewImageAnalysisService(mockAnalyzer)

	assert.NotNil(t, service)
	assert.Equal(t, mockAnalyzer, service.analyzer)
}

func TestProcessImageAndPrompt_Success(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockAnalyzer := mock.NewMockAnalyzer(controller)
	mockLogger := loggerMock.NewMockLoggerer(controller)
	service := NewImageAnalysisService(mockAnalyzer)

	ctx := context.WithValue(context.Background(), log.LoggerKey, mockLogger)
	ctx = log.AddCorrelationIDToOutgoingContext(ctx, "test-correlation-id")

	imageData := []byte("test-image-data")
	mimeType := "image/png"
	prompt := "What is in this image?"
	expectedResponse := "# Image Analysis\n\nThis is a test image."

	mockLogger.EXPECT().
		Info(gomock.Any()).
		Times(1)

	mockAnalyzer.EXPECT().
		Analyze(ctx, imageData, mimeType, prompt).
		Return(expectedResponse, nil)

	response, err := service.ProcessImageAndPrompt(ctx, imageData, mimeType, prompt)

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
}

func TestProcessImageAndPrompt_NoImage(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockAnalyzer := mock.NewMockAnalyzer(controller)
	mockLogger := loggerMock.NewMockLoggerer(controller)
	service := NewImageAnalysisService(mockAnalyzer)

	ctx := context.WithValue(context.Background(), log.LoggerKey, mockLogger)
	ctx = log.AddCorrelationIDToOutgoingContext(ctx, "test-correlation-id")

	response, err := service.ProcessImageAndPrompt(ctx, nil, "image/png", "test prompt")

	assert.Error(t, err)
	assert.Empty(t, response)
	assert.Equal(t, "no image provided", err.Error())
}

func TestProcessImageAndPrompt_NoPrompt(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockAnalyzer := mock.NewMockAnalyzer(controller)
	mockLogger := loggerMock.NewMockLoggerer(controller)
	service := NewImageAnalysisService(mockAnalyzer)

	ctx := context.WithValue(context.Background(), log.LoggerKey, mockLogger)
	ctx = log.AddCorrelationIDToOutgoingContext(ctx, "test-correlation-id")

	response, err := service.ProcessImageAndPrompt(ctx, []byte("test"), "image/png", "")

	assert.Error(t, err)
	assert.Empty(t, response)
	assert.Equal(t, "no prompt provided", err.Error())
}

func TestProcessImageAndPrompt_InvalidMimeType(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockAnalyzer := mock.NewMockAnalyzer(controller)
	mockLogger := loggerMock.NewMockLoggerer(controller)
	service := NewImageAnalysisService(mockAnalyzer)

	ctx := context.WithValue(context.Background(), log.LoggerKey, mockLogger)
	ctx = log.AddCorrelationIDToOutgoingContext(ctx, "test-correlation-id")

	response, err := service.ProcessImageAndPrompt(ctx, []byte("test"), "image/gif", "test prompt")

	assert.Error(t, err)
	assert.Empty(t, response)
	assert.Equal(t, "unsupported mime type \"image/gif\"", err.Error())
}

func TestProcessImageAndPrompt_AnalyzerError(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockAnalyzer := mock.NewMockAnalyzer(controller)
	mockLogger := loggerMock.NewMockLoggerer(controller)
	service := NewImageAnalysisService(mockAnalyzer)

	ctx := context.WithValue(context.Background(), log.LoggerKey, mockLogger)
	ctx = log.AddCorrelationIDToOutgoingContext(ctx, "test-correlation-id")

	imageData := []byte("test-image-data")
	mimeType := "image/png"
	prompt := "What is in this image?"
	expectedError := errors.New("analyzer error")

	mockLogger.EXPECT().
		Info(gomock.Any()).
		Times(1)

	mockAnalyzer.EXPECT().
		Analyze(ctx, imageData, mimeType, prompt).
		Return("", expectedError)

	response, err := service.ProcessImageAndPrompt(ctx, imageData, mimeType, prompt)

	assert.Error(t, err)
	assert.Empty(t, response)
	assert.Equal(t, expectedError, err)
}

func TestProcessImageAndPrompt_NoLoggerInContext(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockAnalyzer := mock.NewMockAnalyzer(controller)
	service := NewImageAnalysisService(mockAnalyzer)

	ctx := context.Background()
	response, err := service.ProcessImageAndPrompt(ctx, []byte("test"), "image/png", "test prompt")

	assert.Error(t, err)
	assert.Empty(t, response)
	assert.Contains(t, err.Error(), "Logger not found in context")
}
