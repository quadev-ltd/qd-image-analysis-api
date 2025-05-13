package grpcserver

import (
	commonPB "github.com/quadev-ltd/qd-common/pb/gen/go/pb_image_analysis"
	"github.com/quadev-ltd/qd-common/pkg/grpcserver"
	"github.com/quadev-ltd/qd-common/pkg/log"
	commonTLS "github.com/quadev-ltd/qd-common/pkg/tls"
	"google.golang.org/grpc"

	"qd-image-analysis-api/internal/service"
)

type Factoryer interface {
	Create(
		grpcServerAddress string,
		imageAnalysisService service.ImageAnalysisServicer,
		logFactory log.Factoryer,
		tlsEnabled bool,
	) (grpcserver.GRPCServicer, error)
}

type Factory struct{}

var _ Factoryer = &Factory{}

func (grpcServerFactory *Factory) Create(
	grpcServerAddress string,
	imageAnalysisService service.ImageAnalysisServicer,
	logFactory log.Factoryer,
	tlsEnabled bool,
) (grpcserver.GRPCServicer, error) {
	const certFilePath = "certs/qd.image-analysis.api.crt"
	const keyFilePath = "certs/qd.image-analysis.api.key"
	
	grpcListener, err := commonTLS.CreateTLSListener(
		grpcServerAddress,
		certFilePath,
		keyFilePath,
		tlsEnabled,
	)
	if err != nil {
		return nil, err
	}

	imageAnalysisServiceGRPCServer := service.NewImageAnalysisServiceServer(imageAnalysisService)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(log.CreateLoggerInterceptor(logFactory)),
	)
	commonPB.RegisterImageAnalysisServiceServer(grpcServer, imageAnalysisServiceGRPCServer)

	return grpcserver.NewGRPCService(grpcServer, grpcListener), nil
}
