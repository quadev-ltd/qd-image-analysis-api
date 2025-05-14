package service

import (
	commonConfig "github.com/quadev-ltd/qd-common/pkg/config"

	"qd-image-analysis-api/internal/config"
)

// Factoryer defines the interface for creating image analysis service instances
type Factoryer interface {
	CreateService(config *config.Config, centralConfig *commonConfig.Config) (ImageAnalysisServicer, error)
}

// Factory implements the Factoryer interface for creating image analysis service instances
type Factory struct{}

var _ Factoryer = &Factory{}

// CreateService creates a new instance of the image analysis service with the provided configuration
func (factory *Factory) CreateService(config *config.Config, centralConfig *commonConfig.Config) (ImageAnalysisServicer, error) {
	serviceConfig := ImageAnalysisServiceConfig{
		MockResponse: "This is a mock response from the image analysis service.",
	}

	return NewImageAnalysisService(serviceConfig), nil
}
