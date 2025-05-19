package application

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	commonPB "github.com/quadev-ltd/qd-common/pb/gen/go/pb_image_analysis"
	commonConfig "github.com/quadev-ltd/qd-common/pkg/config"
	commonLog "github.com/quadev-ltd/qd-common/pkg/log"
	commonTLS "github.com/quadev-ltd/qd-common/pkg/tls"
	"github.com/stretchr/testify/assert"

	"qd-image-analysis-api/internal/config"
	grpcFactory "qd-image-analysis-api/internal/grpcserver"
	"qd-image-analysis-api/internal/service"
	"qd-image-analysis-api/internal/service/mock"
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
	MockImageAnalysisService *mock.MockImageAnalysisServicer
	Config                   *config.Config
	CentralConfig            *commonConfig.Config
	Application              Applicationer
	Controller               *gomock.Controller
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
	mockImageAnalysisService := mock.NewMockImageAnalysisServicer(controller)

	serviceFactory := &mockServiceFactory{
		mockService: mockImageAnalysisService,
	}

	// Create the application using the factory pattern similar to NewApplication
	application := createTestApplication(&testConfig, &mockCentralConfig, serviceFactory)

	go func() {
		t.Logf("Starting server on %s...\n", application.GetGRPCServerAddress())
		application.StartServer()
	}()

	waitForServerUp(t, application, mockCentralConfig.TLSEnabled)

	return &EnvironmentParams{
		MockImageAnalysisService: mockImageAnalysisService,
		Config:                   &testConfig,
		CentralConfig:            &mockCentralConfig,
		Application:              application,
		Controller:               controller,
	}
}

type mockServiceFactory struct {
	mockService service.ImageAnalysisServicer
}

func (f *mockServiceFactory) CreateService(config *config.Config, centralConfig *commonConfig.Config) (service.ImageAnalysisServicer, error) {
	return f.mockService, nil
}

func createTestApplication(config *config.Config, centralConfig *commonConfig.Config, serviceFactory service.Factoryer) Applicationer {
	logFactory := commonLog.NewLogFactory(config.Environment)
	logger := logFactory.NewLogger()

	imageAnalysisService, _ := serviceFactory.CreateService(config, centralConfig)

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
		defer envParams.Application.Close()
		defer envParams.Controller.Finish()

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
		expectedResponse := "Mock analysis result for prompt: What is in this image?. Image size: 16 bytes."

		envParams.MockImageAnalysisService.EXPECT().
			ProcessImageAndPrompt(gomock.Any(), testImageData, testPrompt).
			Return(expectedResponse, nil)

		response, err := grpcClient.ProcessImageAndPrompt(
			ctxWithCorrelationID,
			&commonPB.ImagePromptRequest{
				ImageData: testImageData,
				Prompt:    testPrompt,
			},
		)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, expectedResponse, response.ResponseToPrompt)
	})

	t.Run("ProcessImageAndPrompt_Error", func(t *testing.T) {
		envParams := setUpTestEnvironment(t)
		defer envParams.Application.Close()
		defer envParams.Controller.Finish()

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

		envParams.MockImageAnalysisService.EXPECT().
			ProcessImageAndPrompt(gomock.Any(), testImageData, testPrompt).
			Return("", errors.New("mock error"))

		response, err := grpcClient.ProcessImageAndPrompt(
			ctxWithCorrelationID,
			&commonPB.ImagePromptRequest{
				ImageData: testImageData,
				Prompt:    testPrompt,
			},
		)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "Error processing image and prompt")
	})

	t.Run("ProcessImageAndPrompt_LargeImage", func(t *testing.T) {
		envParams := setUpTestEnvironment(t)
		defer envParams.Application.Close()
		defer envParams.Controller.Finish()

		connection, err := commonTLS.CreateGRPCConnection(
			envParams.Application.GetGRPCServerAddress(),
			envParams.CentralConfig.TLSEnabled,
		)
		assert.NoError(t, err)
		defer connection.Close()

		ctxWithCorrelationID := commonLog.AddCorrelationIDToOutgoingContext(context.Background(), correlationID)
		grpcClient := commonPB.NewImageAnalysisServiceClient(connection)

		testImageData := make([]byte, 1024*1024) // 1MB image
		for i := range testImageData {
			testImageData[i] = byte(i % 256)
		}
		testPrompt := "Analyze this large image"
		expectedResponse := "Mock analysis result for prompt: Analyze this large image. Image size: 1048576 bytes."

		envParams.MockImageAnalysisService.EXPECT().
			ProcessImageAndPrompt(gomock.Any(), testImageData, testPrompt).
			Return(expectedResponse, nil)

		response, err := grpcClient.ProcessImageAndPrompt(
			ctxWithCorrelationID,
			&commonPB.ImagePromptRequest{
				ImageData: testImageData,
				Prompt:    testPrompt,
			},
		)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, expectedResponse, response.ResponseToPrompt)
	})
}
