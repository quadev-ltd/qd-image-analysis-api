package ai

import "context"

// Analyzer knows how to take an image and a prompt and return text
type Analyzer interface {
	Analyze(ctx context.Context, imageData []byte, mimeType, prompt string) (string, error)
	Close() error
}
