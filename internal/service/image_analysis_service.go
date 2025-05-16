package service

import (
	"context"
	"fmt"
	"time"

	"github.com/quadev-ltd/qd-common/pkg/log"
)

// ImageAnalysisServiceConfig holds the configuration for the image analysis service
type ImageAnalysisServiceConfig struct {
	MockResponse string
}

// ImageAnalysisServicer defines the interface for image analysis operations
type ImageAnalysisServicer interface {
	ProcessImageAndPrompt(ctx context.Context, imageData []byte, prompt string) (string, error)
}

// ImageAnalysisService implements the ImageAnalysisServicer interface
type ImageAnalysisService struct {
	config ImageAnalysisServiceConfig
}

var _ ImageAnalysisServicer = &ImageAnalysisService{}

// NewImageAnalysisService creates a new instance of the image analysis service
func NewImageAnalysisService(config ImageAnalysisServiceConfig) *ImageAnalysisService {
	return &ImageAnalysisService{
		config: config,
	}
}

// ProcessImageAndPrompt processes an image with the given prompt and returns the analysis result
func (service *ImageAnalysisService) ProcessImageAndPrompt(ctx context.Context, imageData []byte, prompt string) (string, error) {
	logger, err := log.GetLoggerFromContext(ctx)
	if err != nil {
		return "", err
	}

	logger.Info(fmt.Sprintf("Processing image of size %d bytes with prompt: %s", len(imageData), prompt))

	time.Sleep(500 * time.Millisecond)

	response := service.config.MockResponse
	if response == "" {
		response = fmt.Sprintf("Mock analysis result for prompt: %s. Image size: %d bytes.", prompt, len(imageData))
	}

	return response, nil
}
