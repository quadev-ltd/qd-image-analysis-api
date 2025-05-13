package service

import (
	"context"

	"github.com/quadev-ltd/qd-common/pkg/log"
	"golang.org/x/time/rate"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ImageAnalysisServiceServer struct {
	imageAnalysisService ImageAnalysisServicer
	limiter              *rate.Limiter
}

func NewImageAnalysisServiceServer(imageAnalysisService ImageAnalysisServicer) *ImageAnalysisServiceServer {
	return &ImageAnalysisServiceServer{
		imageAnalysisService: imageAnalysisService,
		limiter:              rate.NewLimiter(rate.Limit(5), 10), // Allow 5 requests per second with burst of 10
	}
}

func (server *ImageAnalysisServiceServer) ProcessImageAndPrompt(ctx context.Context, firebaseToken string, imageData []byte, prompt string) (string, error) {
	logger, err := log.GetLoggerFromContext(ctx)
	if err != nil {
		return "", err
	}

	if !server.limiter.Allow() {
		logger.Error(nil, "Too many requests")
		return "", status.Errorf(codes.ResourceExhausted, "Too many requests")
	}

	response, err := server.imageAnalysisService.ProcessImageAndPrompt(
		ctx,
		firebaseToken,
		imageData,
		prompt,
	)
	if err != nil {
		logger.Error(err, "Error processing image and prompt")
		return "", status.Errorf(codes.Internal, "Error processing image and prompt")
	}

	logger.Info("Image and prompt processed successfully")
	return response, nil
}
