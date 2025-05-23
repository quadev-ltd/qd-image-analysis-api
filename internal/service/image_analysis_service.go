package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/quadev-ltd/qd-common/pkg/log"
)

// ImageAnalysisServiceConfig holds the configuration for the image analysis service
type ImageAnalysisServiceConfig struct {
	MockResponse string
	VertexAI     VertexAIConfig
}

type VertexAIConfig struct {
	ProjectID   string
	Location    string
	ModelName   string
	Enabled     bool
	Credentials string // Path to the credentials file or JSON content
}

// ImageAnalysisServicer defines the interface for image analysis operations
type ImageAnalysisServicer interface {
	ProcessImageAndPrompt(ctx context.Context, imageData []byte, prompt string) (string, error)
}

// ImageAnalysisService implements the ImageAnalysisServicer interface
type ImageAnalysisService struct {
	config ImageAnalysisServiceConfig
}

var _ ImageAnalysisServicer = &ImageAnalysisService{}

// NewImageAnalysisService creates a new instance of the image analysis service
func NewImageAnalysisService(config ImageAnalysisServiceConfig) *ImageAnalysisService {
	return &ImageAnalysisService{
		config: config,
	}
}

// ProcessImageAndPrompt processes an image with the given prompt and returns the analysis result
func (service *ImageAnalysisService) ProcessImageAndPrompt(ctx context.Context, imageData []byte, prompt string) (string, error) {
	logger, err := log.GetLoggerFromContext(ctx)
	if err != nil {
		return "", err
	}

	logger.Info(fmt.Sprintf("Processing image of size %d bytes with prompt: %s", len(imageData), prompt))

	if !service.config.VertexAI.Enabled {
		response := service.config.MockResponse
		if response == "" {
			response = fmt.Sprintf("Mock analysis result for prompt: %s. Image size: %d bytes.", prompt, len(imageData))
		}
		return response, nil
	}

	// Process with Vertex AI
	return service.processWithVertexAI(ctx, imageData, prompt, logger)
}

func (service *ImageAnalysisService) processWithVertexAI(ctx context.Context, imageData []byte, prompt string, logger log.Loggerer) (string, error) {
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role": "user",
				"parts": []map[string]interface{}{
					{
						"text": prompt,
					},
					{
						"inline_data": map[string]interface{}{
							"mime_type": "image/jpeg", // Assuming JPEG, adjust if needed
							"data":      base64Image,
						},
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.4,
			"topP":            1.0,
			"topK":            32,
			"maxOutputTokens": 2048,
		},
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		logger.Error(err, "Failed to marshal request body")
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	var authHeader string
	if service.config.VertexAI.Credentials != "" {
		credJSON, err := ioutil.ReadFile(service.config.VertexAI.Credentials)
		if err != nil {
			credJSON = []byte(service.config.VertexAI.Credentials)
		}
		
		var credMap map[string]interface{}
		if err := json.Unmarshal(credJSON, &credMap); err != nil {
			logger.Error(err, "Failed to parse credentials")
			return "", fmt.Errorf("failed to parse credentials: %v", err)
		}
		
		if token, ok := credMap["access_token"].(string); ok {
			authHeader = "Bearer " + token
		} else {
			logger.Error(nil, "No access token found in credentials")
			return "", fmt.Errorf("no access token found in credentials")
		}
	}

	url := fmt.Sprintf(
		"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:generateContent",
		service.config.VertexAI.Location,
		service.config.VertexAI.ProjectID,
		service.config.VertexAI.Location,
		service.config.VertexAI.ModelName,
	)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestJSON))
	if err != nil {
		logger.Error(err, "Failed to create request")
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err, "Failed to send request to Vertex AI")
		return "", fmt.Errorf("failed to send request to Vertex AI: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		logger.Error(nil, fmt.Sprintf("Vertex AI returned non-OK status: %d, body: %s", resp.StatusCode, string(body)))
		return "", fmt.Errorf("Vertex AI returned non-OK status: %d", resp.StatusCode)
	}

	var responseBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		logger.Error(err, "Failed to decode response")
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	candidates, ok := responseBody["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		logger.Error(nil, "Invalid response format or no candidates")
		return "", fmt.Errorf("invalid response format or no candidates")
	}

	content, ok := candidates[0].(map[string]interface{})["content"].(map[string]interface{})
	if !ok {
		logger.Error(nil, "Invalid content format")
		return "", fmt.Errorf("invalid content format")
	}

	parts, ok := content["parts"].([]interface{})
	if !ok || len(parts) == 0 {
		logger.Error(nil, "Invalid parts format or no parts")
		return "", fmt.Errorf("invalid parts format or no parts")
	}

	text, ok := parts[0].(map[string]interface{})["text"].(string)
	if !ok {
		logger.Error(nil, "Invalid text format")
		return "", fmt.Errorf("invalid text format")
	}

	return text, nil
}
