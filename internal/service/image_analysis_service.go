package service

import (
	"context"
	"fmt"
	"time"

	"github.com/quadev-ltd/qd-common/pkg/log"
)

type ImageAnalysisServiceConfig struct {
	MockResponse string
}

type ImageAnalysisServicer interface {
	ProcessImageAndPrompt(ctx context.Context, firebaseToken string, imageData []byte, prompt string) (string, error)
}

type ImageAnalysisService struct {
	config ImageAnalysisServiceConfig
}

var _ ImageAnalysisServicer = &ImageAnalysisService{}

func NewImageAnalysisService(config ImageAnalysisServiceConfig) *ImageAnalysisService {
	return &ImageAnalysisService{
		config: config,
	}
}

func (service *ImageAnalysisService) ProcessImageAndPrompt(ctx context.Context, firebaseToken string, imageData []byte, prompt string) (string, error) {
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
