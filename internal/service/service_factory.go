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
		MockResponse: config.MockResponse,
		ModelConfig: ModelConfig{
			Provider:   ModelProvider(config.AIModel.Provider),
			ProjectID:  config.AIModel.ProjectID,
			Location:   config.AIModel.Location,
			ModelID:    config.AIModel.ModelID,
			EndpointID: config.AIModel.EndpointID,
			APIKey:     config.AIModel.APIKey,
			Parameters: make(map[string]interface{}),
		},
	}

	return NewImageAnalysisService(serviceConfig), nil
}
