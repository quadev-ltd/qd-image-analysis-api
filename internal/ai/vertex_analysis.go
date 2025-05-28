package ai

import (
	"context"
	"fmt"

	"cloud.google.com/go/vertexai/genai"
	"google.golang.org/api/option"

	configPkg "qd-image-analysis-api/internal/config"
)

// VertexAnalyzer is a concrete implementation of Analyzer using Vertex AI
type VertexAnalyzer struct {
	client *genai.Client
	config *configPkg.VertexAIConfig
}

// NewVertexAnalyzer creates a new instance of VertexAnalyzer with the provided configuration.
// It initializes a connection to the Vertex AI service using the specified credentials.
func NewVertexAnalyzer(config *configPkg.VertexAIConfig) (*VertexAnalyzer, error) {
	ctx := context.Background()
	cli, err := genai.NewClient(ctx, config.ProjectID, config.Location, option.WithCredentialsFile(config.ConfigPath))
	if err != nil {
		return nil, err
	}
	return &VertexAnalyzer{client: cli, config: config}, nil
}

// Analyze processes an image with a given prompt using Vertex AI's generative model.
// It returns the model's response as a string or an error if the analysis fails.
func (vertexAnalyzer *VertexAnalyzer) Analyze(ctx context.Context, imageData []byte, mimeType, prompt string) (string, error) {
	model := vertexAnalyzer.client.GenerativeModel(vertexAnalyzer.config.ModelName)
	model.SetMaxOutputTokens(vertexAnalyzer.config.MaxTokens)
	model.SetTemperature(vertexAnalyzer.config.Temperature)

	img := genai.ImageData(mimeType, imageData)
	txt := genai.Text(fmt.Sprintf("Please format your response as markdown. Here is the analysis request: %s", prompt))

	resp, err := model.GenerateContent(ctx, img, txt)
	if err != nil {
		return "", err
	}
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response candidates")
	}
	contentPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}
	return string(contentPart), nil
}

// Close closes the connection to the Vertex AI service.
// It should be called when the analyzer is no longer needed.
func (vertexAnalyzer *VertexAnalyzer) Close() error {
	return vertexAnalyzer.client.Close()
}
