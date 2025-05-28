package service

import (
	"context"
	"fmt"

	"github.com/quadev-ltd/qd-common/pkg/log"

	"qd-image-analysis-api/internal/ai"
)

// ImageAnalysisServicer defines the interface for image analysis operations
type ImageAnalysisServicer interface {
	ProcessImageAndPrompt(ctx context.Context, imageData []byte, mimeType string, prompt string) (string, error)
	Close() error
}

// ImageAnalysisService implements the ImageAnalysisServicer interface
type ImageAnalysisService struct {
	analyzer ai.Analyzer
}

var _ ImageAnalysisServicer = &ImageAnalysisService{}

// NewImageAnalysisService creates a new instance of the image analysis service
func NewImageAnalysisService(analyzer ai.Analyzer) *ImageAnalysisService {
	return &ImageAnalysisService{analyzer: analyzer}
}

// ProcessImageAndPrompt processes an image with a given prompt using the configured analyzer.
// It validates the input parameters and returns the analysis result or an error if the processing fails.
func (imageAnalysisService *ImageAnalysisService) ProcessImageAndPrompt(ctx context.Context, imageData []byte, mimeType string, prompt string) (string, error) {
	logger, err := log.GetLoggerFromContext(ctx)
	if err != nil {
		return "", err
	}

	switch {
	case len(imageData) == 0:
		return "", &Error{
			Message: "no image provided",
		}
	case prompt == "":
		return "", &Error{
			Message: "no prompt provided",
		}
	case mimeType != "image/jpeg" && mimeType != "image/png":
		return "", &Error{
			Message: fmt.Sprintf("unsupported mime type %q", mimeType),
		}
	}

	logger.Info(fmt.Sprintf("Processing image of size %d bytes with prompt: %s", len(imageData), prompt))
	return imageAnalysisService.analyzer.Analyze(ctx, imageData, mimeType, prompt)
}

// Close closes the image analysis service and its underlying analyzer.
// It should be called when the service is no longer needed.
func (imageAnalysisService *ImageAnalysisService) Close() error {
	return imageAnalysisService.analyzer.Close()
}
