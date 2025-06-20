package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	commonPB "github.com/quadev-ltd/qd-common/pb/gen/go/pb_image_analysis"
	commonConfig "github.com/quadev-ltd/qd-common/pkg/config"
	commonLog "github.com/quadev-ltd/qd-common/pkg/log"
	commonTLS "github.com/quadev-ltd/qd-common/pkg/tls"
	"github.com/stretchr/testify/assert"

	aiMock "qd-image-analysis-api/internal/ai/mock"
	"qd-image-analysis-api/internal/config"
	grpcFactory "qd-image-analysis-api/internal/grpcserver"
	"qd-image-analysis-api/internal/service"
)

func isServerUp(address string, tlsEnabled bool) bool {
	conn, err := commonTLS.CreateGRPCConnection(address, tlsEnabled)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func waitForServerUp(t *testing.T, app Applicationer, tlsEnabled bool) {
	t.Helper()

	maxRetries := 10
	retryInterval := 100 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		if isServerUp(app.GetGRPCServerAddress(), tlsEnabled) {
			return
		}
		time.Sleep(retryInterval)
	}
	t.Fatal("Server failed to start")
}

type EnvironmentParams struct {
	MockAIAnalyser       *aiMock.MockAnalyzer
	ImageAnalysisService *service.ImageAnalysisService
	Config               *config.Config
	CentralConfig        *commonConfig.Config
	Application          Applicationer
	Controller           *gomock.Controller
}

func setUpTestEnvironment(t *testing.T) *EnvironmentParams {
	var testConfig config.Config
	testConfig.Environment = "test"
	testConfig.Verbose = true

	var mockCentralConfig commonConfig.Config
	mockCentralConfig.ImageAnalysisService.Host = "localhost"
	mockCentralConfig.ImageAnalysisService.Port = "50053"
	mockCentralConfig.TLSEnabled = false

	controller := gomock.NewController(t)
	mockAiAnalyser := aiMock.NewMockAnalyzer(controller)
	imageAnalysisService := service.NewImageAnalysisService(mockAiAnalyser)

	// Create the application using the factory pattern similar to NewApplication
	application := createTestApplication(&testConfig, &mockCentralConfig, imageAnalysisService)

	go func() {
		t.Logf("Starting server on %s...\n", application.GetGRPCServerAddress())
		application.StartServer()
	}()

	waitForServerUp(t, application, mockCentralConfig.TLSEnabled)

	return &EnvironmentParams{
		MockAIAnalyser:       mockAiAnalyser,
		ImageAnalysisService: imageAnalysisService,
		Config:               &testConfig,
		CentralConfig:        &mockCentralConfig,
		Application:          application,
		Controller:           controller,
	}
}

func createTestApplication(config *config.Config, centralConfig *commonConfig.Config, imageAnalysisService *service.ImageAnalysisService) Applicationer {
	logFactory := commonLog.NewLogFactory(config.Environment)
	logger := logFactory.NewLogger()

	grpcServerAddress := fmt.Sprintf(
		"%s:%s",
		centralConfig.ImageAnalysisService.Host,
		centralConfig.ImageAnalysisService.Port,
	)

	grpcServiceServer, _ := (&grpcFactory.Factory{}).Create(
		grpcServerAddress,
		imageAnalysisService,
		logFactory,
		centralConfig.TLSEnabled,
	)

	return New(grpcServiceServer, grpcServerAddress, imageAnalysisService, logger)
}

func TestImageAnalysisEndpoints(t *testing.T) {
	const correlationID = "test-correlation-id"

	t.Run("ProcessImageAndPrompt_Success", func(t *testing.T) {
		envParams := setUpTestEnvironment(t)

		connection, err := commonTLS.CreateGRPCConnection(
			envParams.Application.GetGRPCServerAddress(),
			envParams.CentralConfig.TLSEnabled,
		)
		assert.NoError(t, err)
		defer connection.Close()

		ctxWithCorrelationID := commonLog.AddCorrelationIDToOutgoingContext(context.Background(), correlationID)
		grpcClient := commonPB.NewImageAnalysisServiceClient(connection)

		testImageData := []byte("test-image-data")
		testPrompt := "What is in this image?"
		testMimeType := "image/png"
		expectedResponse := "# Image Analysis\n\nThis is a test image containing sample data."

		envParams.MockAIAnalyser.EXPECT().
			Analyze(gomock.Any(), testImageData, testMimeType, testPrompt).
			Return(expectedResponse, nil)

		envParams.MockAIAnalyser.EXPECT().
			Close().
			Return(nil)

		response, err := grpcClient.ProcessImageAndPrompt(
			ctxWithCorrelationID,
			&commonPB.ImagePromptRequest{
				ImageData: testImageData,
				Prompt:    testPrompt,
				MimeType:  testMimeType,
			},
		)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, expectedResponse, response.ResponseToPrompt)

		envParams.Application.Close()
		envParams.Controller.Finish()
	})

	t.Run("ProcessImageAndPrompt_ImageError", func(t *testing.T) {
		envParams := setUpTestEnvironment(t)

		connection, err := commonTLS.CreateGRPCConnection(
			envParams.Application.GetGRPCServerAddress(),
			envParams.CentralConfig.TLSEnabled,
		)
		assert.NoError(t, err)
		defer connection.Close()

		ctxWithCorrelationID := commonLog.AddCorrelationIDToOutgoingContext(context.Background(), correlationID)
		grpcClient := commonPB.NewImageAnalysisServiceClient(connection)

		testImageData := []byte("")
		testMimeType := "image/png"
		testPrompt := "What is in this image?"

		envParams.MockAIAnalyser.EXPECT().
			Close().
			Return(nil)

		response, err := grpcClient.ProcessImageAndPrompt(
			ctxWithCorrelationID,
			&commonPB.ImagePromptRequest{
				ImageData: testImageData,
				Prompt:    testPrompt,
				MimeType:  testMimeType,
			},
		)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "no image provided")

		envParams.Application.Close()
		envParams.Controller.Finish()
	})

	t.Run("ProcessImageAndPrompt_InvalidMimeType", func(t *testing.T) {
		envParams := setUpTestEnvironment(t)

		connection, err := commonTLS.CreateGRPCConnection(
			envParams.Application.GetGRPCServerAddress(),
			envParams.CentralConfig.TLSEnabled,
		)
		assert.NoError(t, err)
		defer connection.Close()

		ctxWithCorrelationID := commonLog.AddCorrelationIDToOutgoingContext(context.Background(), correlationID)
		grpcClient := commonPB.NewImageAnalysisServiceClient(connection)

		testImageData := []byte("test-image-data")
		testMimeType := "image/gif" // Invalid mime type
		testPrompt := "What is in this image?"

		envParams.MockAIAnalyser.EXPECT().
			Close().
			Return(nil)

		response, err := grpcClient.ProcessImageAndPrompt(
			ctxWithCorrelationID,
			&commonPB.ImagePromptRequest{
				ImageData: testImageData,
				Prompt:    testPrompt,
				MimeType:  testMimeType,
			},
		)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "unsupported mime type")

		envParams.Application.Close()
		envParams.Controller.Finish()

	})

	t.Run("ProcessImageAndPrompt_EmptyPrompt", func(t *testing.T) {
		envParams := setUpTestEnvironment(t)

		connection, err := commonTLS.CreateGRPCConnection(
			envParams.Application.GetGRPCServerAddress(),
			envParams.CentralConfig.TLSEnabled,
		)
		assert.NoError(t, err)
		defer connection.Close()

		ctxWithCorrelationID := commonLog.AddCorrelationIDToOutgoingContext(context.Background(), correlationID)
		grpcClient := commonPB.NewImageAnalysisServiceClient(connection)

		testImageData := []byte("test-image-data")
		testMimeType := "image/png"
		testPrompt := ""

		envParams.MockAIAnalyser.EXPECT().
			Close().
			Return(nil)

		response, err := grpcClient.ProcessImageAndPrompt(
			ctxWithCorrelationID,
			&commonPB.ImagePromptRequest{
				ImageData: testImageData,
				Prompt:    testPrompt,
				MimeType:  testMimeType,
			},
		)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "no prompt provided")

		envParams.Application.Close()
		envParams.Controller.Finish()
	})
}
