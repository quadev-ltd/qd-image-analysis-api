package application

import (
	"fmt"

	commonConfig "github.com/quadev-ltd/qd-common/pkg/config"
	"github.com/quadev-ltd/qd-common/pkg/grpcserver"
	"github.com/quadev-ltd/qd-common/pkg/log"

	"qd-image-analysis-api/internal/ai"
	"qd-image-analysis-api/internal/config"
	grpcFactory "qd-image-analysis-api/internal/grpcserver"
	"qd-image-analysis-api/internal/service"
)

// Applicationer defines the interface for the application's core functionality
type Applicationer interface {
	StartServer()
	Close()
	GetGRPCServerAddress() string
}

// Application represents the main application structure that manages the gRPC server and services
type Application struct {
	logger            log.Loggerer
	grpcServiceServer grpcserver.GRPCServicer
	grpcServerAddress string
	service           service.ImageAnalysisServicer
}

// NewApplication creates a new instance of the application with the provided configuration
func NewApplication(config *config.Config, centralConfig *commonConfig.Config) (Applicationer, error) {
	logFactory := log.NewLogFactory(config.Environment)
	logger := logFactory.NewLogger()
	if centralConfig.TLSEnabled {
		logger.Info("TLS is enabled")
	} else {
		logger.Info("TLS is disabled")
	}

	aiAnalyser, err := ai.NewVertexAnalyzer(&config.VertexAI)
	if err != nil {
		logger.Error(err, "Failed to create AI analyzer")
		return nil, err
	}
	imageAnalysisService := service.NewImageAnalysisService(aiAnalyser)

	grpcServerAddress := fmt.Sprintf(
		"%s:%s",
		centralConfig.ImageAnalysisService.Host,
		centralConfig.ImageAnalysisService.Port,
	)

	grpcServiceServer, err := (&grpcFactory.Factory{}).Create(
		grpcServerAddress,
		imageAnalysisService,
		logFactory,
		centralConfig.TLSEnabled,
	)
	if err != nil {
		return nil, err
	}

	return New(grpcServiceServer, grpcServerAddress, imageAnalysisService, logger), nil
}

// New creates a new Application instance with the provided dependencies
func New(
	grpcServiceServer grpcserver.GRPCServicer,
	grpcServerAddress string,
	service service.ImageAnalysisServicer,
	logger log.Loggerer,
) Applicationer {
	return &Application{
		grpcServiceServer: grpcServiceServer,
		grpcServerAddress: grpcServerAddress,
		service:           service,
		logger:            logger,
	}
}

// StartServer starts the gRPC server and begins listening for requests
func (application *Application) StartServer() {
	application.logger.Info(fmt.Sprintf("Starting gRPC server on %s...", application.grpcServerAddress))
	err := application.grpcServiceServer.Serve()
	if err != nil {
		application.logger.Error(err, "Failed to serve grpc server")
		return
	}
}

// Close gracefully shuts down the application and its services
func (application *Application) Close() {
	switch {
	case application.service == nil:
		application.logger.Error(nil, "Service is not created")
		return
	case application.grpcServiceServer == nil:
		application.logger.Error(nil, "gRPC server is not created")
		return
	}
	err := application.grpcServiceServer.Close()
	if err != nil {
		application.logger.Error(err, "Failed to close gRPC server")
	}
	err = application.service.Close()
	if err != nil {
		application.logger.Error(err, "Failed to close service")
	}
	application.logger.Info("gRPC server closed")
}

// GetGRPCServerAddress returns the address where the gRPC server is listening
func (application *Application) GetGRPCServerAddress() string {
	return application.grpcServerAddress
}
