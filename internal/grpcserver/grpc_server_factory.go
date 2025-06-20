package grpcserver

import (
	"fmt"

	"github.com/quadev-ltd/qd-common/pb/gen/go/pb_image_analysis"
	"github.com/quadev-ltd/qd-common/pkg/grpcserver"
	"github.com/quadev-ltd/qd-common/pkg/log"
	commonTLS "github.com/quadev-ltd/qd-common/pkg/tls"
	"google.golang.org/grpc"

	"qd-image-analysis-api/internal/service"
)

// Factoryer defines the interface for creating gRPC server instances
type Factoryer interface {
	Create(
		grpcServerAddress string,
		imageAnalysisService service.ImageAnalysisServicer,
		logFactory log.Factoryer,
		tlsEnabled bool,
	) (grpcserver.GRPCServicer, error)
}

// Factory implements the Factoryer interface for creating gRPC server instances
type Factory struct{}

var _ Factoryer = &Factory{}

// Create builds and returns a new gRPC server instance with the specified configuration
func (grpcServerFactory *Factory) Create(
	grpcServerAddress string,
	imageAnalysisService service.ImageAnalysisServicer,
	logFactory log.Factoryer,
	tlsEnabled bool,
) (grpcserver.GRPCServicer, error) {
	const certFilePath = "certs/qd.image.analysis.api.crt"
	const keyFilePath = "certs/qd.image.analysis.api.key"

	grpcListener, err := commonTLS.CreateTLSListener(
		grpcServerAddress,
		certFilePath,
		keyFilePath,
		tlsEnabled,
	)
	if err != nil {
		return nil, err
	}

	imageAnalysisServiceGRPCServer := NewImageAnalysisServiceServer(imageAnalysisService)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(log.CreateLoggerInterceptor(logFactory)),
	)
	pb_image_analysis.RegisterImageAnalysisServiceServer(grpcServer, imageAnalysisServiceGRPCServer)

	if grpcListener == nil {
		fmt.Println("\n\n\n\n\n\n\n\n\ngrpcListener or grpcServer nil")
	}
	return grpcserver.NewGRPCService(grpcServer, grpcListener), nil
}
