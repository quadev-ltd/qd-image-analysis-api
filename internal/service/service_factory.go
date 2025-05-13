package service

import (
	commonConfig "github.com/quadev-ltd/qd-common/pkg/config"

	"qd-image-analysis-api/internal/config"
)

type Factoryer interface {
	CreateService(config *config.Config, centralConfig *commonConfig.Config) (ImageAnalysisServicer, error)
}

type Factory struct{}

var _ Factoryer = &Factory{}

func (factory *Factory) CreateService(config *config.Config, centralConfig *commonConfig.Config) (ImageAnalysisServicer, error) {
	serviceConfig := ImageAnalysisServiceConfig{
		MockResponse: "This is a mock response from the image analysis service.",
	}

	return NewImageAnalysisService(serviceConfig), nil
}
