package service

import (
	"context"

	commonPB "github.com/quadev-ltd/qd-common/pb/gen/go/pb_image_analysis"
	"github.com/quadev-ltd/qd-common/pkg/log"
	"golang.org/x/time/rate"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ImageAnalysisServiceServer implements the gRPC service for image analysis
type ImageAnalysisServiceServer struct {
	commonPB.UnimplementedImageAnalysisServiceServer
	imageAnalysisService ImageAnalysisServicer
	limiter              *rate.Limiter
}

// NewImageAnalysisServiceServer creates a new instance of the gRPC service server
func NewImageAnalysisServiceServer(imageAnalysisService ImageAnalysisServicer) *ImageAnalysisServiceServer {
	return &ImageAnalysisServiceServer{
		imageAnalysisService: imageAnalysisService,
		limiter:              rate.NewLimiter(rate.Limit(5), 10), // Allow 5 requests per second with burst of 10
	}
}

// ProcessImageAndPrompt handles the gRPC request to process an image with a prompt
func (server *ImageAnalysisServiceServer) ProcessImageAndPrompt(ctx context.Context, request *commonPB.ImagePromptRequest) (*commonPB.ImagePromptResponse, error) {
	logger, err := log.GetLoggerFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if !server.limiter.Allow() {
		logger.Error(nil, "Too many requests")
		return nil, status.Errorf(codes.ResourceExhausted, "Too many requests")
	}

	response, err := server.imageAnalysisService.ProcessImageAndPrompt(
		ctx,
		request.ImageData,
		request.Prompt,
	)
	if err != nil {
		logger.Error(err, "Error processing image and prompt")
		return nil, status.Errorf(codes.Internal, "Error processing image and prompt")
	}

	logger.Info("Image and prompt processed successfully")
	return &commonPB.ImagePromptResponse{
		ResponseToPrompt: response,
	}, nil
}
