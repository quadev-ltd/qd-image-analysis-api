package service

import (
	"context"
	"testing"

	"github.com/quadev-ltd/qd-common/pkg/log"
	"github.com/stretchr/testify/assert"
)

type MockLogger struct{}

func (m *MockLogger) Info(msg string)  {}
func (m *MockLogger) Error(msg string) {}
func (m *MockLogger) Debug(msg string) {}
func (m *MockLogger) Warn(msg string)  {}

func createTestContext() context.Context {
	ctx := context.Background()
	logger := &MockLogger{}
	return log.NewContextWithLogger(ctx, logger)
}

func TestNewImageAnalysisService(t *testing.T) {
	config := ImageAnalysisServiceConfig{
		MockResponse: "test response",
		ModelConfig: ModelConfig{
			Provider: "",
		},
	}
	service := NewImageAnalysisService(config)
	assert.Equal(t, MockProvider, service.config.ModelConfig.Provider)
	
	config = ImageAnalysisServiceConfig{
		MockResponse: "test response",
		ModelConfig: ModelConfig{
			Provider: VertexAI,
		},
	}
	service = NewImageAnalysisService(config)
	assert.Equal(t, VertexAI, service.config.ModelConfig.Provider)
}

func TestProcessImageAndPrompt(t *testing.T) {
	ctx := createTestContext()
	
	mockConfig := ImageAnalysisServiceConfig{
		MockResponse: "mock test response",
		ModelConfig: ModelConfig{
			Provider: MockProvider,
		},
	}
	mockService := NewImageAnalysisService(mockConfig)
	
	firebaseToken := "valid-firebase-token"
	imageData := []byte("test-image-data")
	prompt := "What objects are in this image?"
	
	response, err := mockService.ProcessImageAndPrompt(ctx, firebaseToken, imageData, prompt)
	assert.NoError(t, err)
	assert.Equal(t, "mock test response", response)
	
	response, err = mockService.ProcessImageAndPrompt(ctx, "", imageData, prompt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "firebase token is required")
	
	response, err = mockService.ProcessImageAndPrompt(ctx, firebaseToken, []byte{}, prompt)
	assert.NoError(t, err)
	assert.Equal(t, "mock test response", response)
	
	response, err = mockService.ProcessImageAndPrompt(ctx, firebaseToken, imageData, "")
	assert.NoError(t, err)
	assert.Equal(t, "mock test response", response)
	
	unsupportedConfig := ImageAnalysisServiceConfig{
		ModelConfig: ModelConfig{
			Provider: "unsupported",
		},
	}
	unsupportedService := NewImageAnalysisService(unsupportedConfig)
	
	response, err = unsupportedService.ProcessImageAndPrompt(ctx, firebaseToken, imageData, prompt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported model provider")
}

func TestProcessWithMockProvider(t *testing.T) {
	ctx := createTestContext()
	
	customMockConfig := ImageAnalysisServiceConfig{
		MockResponse: "custom mock response",
		ModelConfig: ModelConfig{
			Provider: MockProvider,
		},
	}
	customMockService := NewImageAnalysisService(customMockConfig)
	
	imageData := []byte("test-image-data")
	prompt := "What objects are in this image?"
	
	response, err := customMockService.processWithMockProvider(ctx, imageData, prompt)
	assert.NoError(t, err)
	assert.Equal(t, "custom mock response", response)
	
	defaultMockConfig := ImageAnalysisServiceConfig{
		MockResponse: "",
		ModelConfig: ModelConfig{
			Provider: MockProvider,
		},
	}
	defaultMockService := NewImageAnalysisService(defaultMockConfig)
	
	response, err = defaultMockService.processWithMockProvider(ctx, imageData, prompt)
	assert.NoError(t, err)
	assert.Contains(t, response, "Mock analysis result for prompt")
	assert.Contains(t, response, "Detected objects")
}
