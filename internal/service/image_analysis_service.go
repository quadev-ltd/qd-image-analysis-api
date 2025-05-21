package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"cloud.google.com/go/vertexai/genai"
	aiplatform "cloud.google.com/go/vertexai"
	"github.com/quadev-ltd/qd-common/pkg/log"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"
)

type ModelProvider string

const (
	VertexAI ModelProvider = "vertex"
	// MockProvider is a mock provider for testing and development
	MockProvider ModelProvider = "mock"
)

// ModelConfig holds configuration for an AI model
type ModelConfig struct {
	Provider ModelProvider
	// ProjectID is the Google Cloud project ID
	ProjectID string
	Location string
	ModelID string
	APIKey string
	// Parameters contains additional model-specific parameters
	Parameters map[string]interface{}
}

// ImageAnalysisServiceConfig holds the configuration for the image analysis service
type ImageAnalysisServiceConfig struct {
	// MockResponse is the response to return when using the mock provider
	MockResponse string
	// ModelConfig contains the configuration for the AI model
	ModelConfig ModelConfig
}

// ImageAnalysisServicer defines the interface for image analysis operations
type ImageAnalysisServicer interface {
	// ProcessImageAndPrompt processes an image with the given prompt and returns the analysis result
	ProcessImageAndPrompt(ctx context.Context, firebaseToken string, imageData []byte, prompt string) (string, error)
}

// ImageAnalysisService implements the ImageAnalysisServicer interface
type ImageAnalysisService struct {
	config ImageAnalysisServiceConfig
}

var _ ImageAnalysisServicer = &ImageAnalysisService{}

// NewImageAnalysisService creates a new instance of the image analysis service
func NewImageAnalysisService(config ImageAnalysisServiceConfig) *ImageAnalysisService {
	if config.ModelConfig.Provider == "" {
		config.ModelConfig.Provider = MockProvider
	}

	return &ImageAnalysisService{
		config: config,
	}
}

// ProcessImageAndPrompt processes an image with the given prompt and returns the analysis result
func (service *ImageAnalysisService) ProcessImageAndPrompt(ctx context.Context, firebaseToken string, imageData []byte, prompt string) (string, error) {
	logger, err := log.GetLoggerFromContext(ctx)
	if err != nil {
		return "", err
	}

	logger.Info(fmt.Sprintf("Processing image of size %d bytes with prompt: %s", len(imageData), prompt))

	if firebaseToken == "" {
		return "", fmt.Errorf("firebase token is required")
	}

	// Process the image based on the configured model provider
	switch service.config.ModelConfig.Provider {
	case VertexAI:
		return service.processWithVertexAI(ctx, imageData, prompt)
	case MockProvider:
		return service.processWithMockProvider(ctx, imageData, prompt)
	default:
		return "", fmt.Errorf("unsupported model provider: %s", service.config.ModelConfig.Provider)
	}
}

// and returns the analysis result
func (service *ImageAnalysisService) processWithVertexAI(ctx context.Context, imageData []byte, prompt string) (string, error) {
	config := service.config.ModelConfig
	
	client, err := aiplatform.NewClient(ctx, config.ProjectID, config.Location, option.WithAPIKey(config.APIKey))
	if err != nil {
		return "", fmt.Errorf("failed to create Vertex AI client: %v", err)
	}
	
	modelID := config.ModelID
	if modelID == "" {
		modelID = "gemini-pro-vision"
	}
	
	geminiClient, err := client.GetGenerativeModel(modelID)
	if err != nil {
		return "", fmt.Errorf("failed to get Gemini model: %v", err)
	}
	
	if config.Parameters != nil {
		if temp, ok := config.Parameters["temperature"].(float32); ok {
			geminiClient.SetTemperature(temp)
		}
		if maxTokens, ok := config.Parameters["maxOutputTokens"].(int32); ok {
			geminiClient.SetMaxOutputTokens(maxTokens)
		}
	}
	
	mimeType := "image/jpeg"
	imagePart := genai.ImageData{
		MIMEType: mimeType,
		Data:     imageData,
	}
	
	promptPart := genai.Text(prompt)
	
	resp, err := geminiClient.GenerateContent(ctx, imagePart, promptPart)
	if err != nil {
		return "", fmt.Errorf("content generation failed: %v", err)
	}
	
	// Validate response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Gemini model")
	}
	
	responseText, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", fmt.Errorf("unexpected response format from Gemini model")
	}
	
	return string(responseText), nil
}

func (service *ImageAnalysisService) processWithMockProvider(ctx context.Context, imageData []byte, prompt string) (string, error) {
	time.Sleep(500 * time.Millisecond)

	response := service.config.MockResponse
	if response == "" {
		response = fmt.Sprintf("Mock analysis result for prompt: %s. Image size: %d bytes.\n", prompt, len(imageData))
		response += "Detected objects: computer, keyboard, mouse, coffee mug\n"
		response += "Scene classification: office workspace\n"
		response += "Dominant colors: gray, black, white\n"
		response += "Text detected: 'HELLO WORLD' (confidence: 0.92)"
	}

	return response, nil
}
