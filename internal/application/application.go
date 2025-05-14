package application

import (
	"fmt"

	commonConfig "github.com/quadev-ltd/qd-common/pkg/config"
	"github.com/quadev-ltd/qd-common/pkg/grpcserver"
	"github.com/quadev-ltd/qd-common/pkg/log"

	"qd-image-analysis-api/internal/config"
	grpcFactory "qd-image-analysis-api/internal/grpcserver"
	"qd-image-analysis-api/internal/service"
)

type Applicationer interface {
	StartServer()
	Close()
	GetGRPCServerAddress() string
}

type Application struct {
	logger            log.Loggerer
	grpcServiceServer grpcserver.GRPCServicer
	grpcServerAddress string
	service           service.ImageAnalysisServicer
}

func NewApplication(config *config.Config, centralConfig *commonConfig.Config) Applicationer {
	logFactory := log.NewLogFactory(config.Environment)
	logger := logFactory.NewLogger()
	if centralConfig.TLSEnabled {
		logger.Info("TLS is enabled")
	} else {
		logger.Info("TLS is disabled")
	}

	imageAnalysisService, err := (&service.Factory{}).CreateService(config, centralConfig)
	if err != nil {
		logger.Error(err, "Failed to create image analysis service")
	}

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
		logger.Error(err, "Failed to create grpc server")
	}

	return New(grpcServiceServer, grpcServerAddress, imageAnalysisService, logger)
}

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

func (application *Application) StartServer() {
	application.logger.Info(fmt.Sprintf("Starting gRPC server on %s...", application.grpcServerAddress))
	err := application.grpcServiceServer.Serve()
	if err != nil {
		application.logger.Error(err, "Failed to serve grpc server")
		return
	}
}

func (application *Application) Close() {
	switch {
	case application.service == nil:
		application.logger.Error(nil, "Service is not created")
		return
	case application.grpcServiceServer == nil:
		application.logger.Error(nil, "gRPC server is not created")
		return
	}
	application.grpcServiceServer.Close()
	application.logger.Info("gRPC server closed")
}

func (application *Application) GetGRPCServerAddress() string {
	return application.grpcServerAddress
}
