package config

import (
	"fmt"

	commonAWS "github.com/quadev-ltd/qd-common/pkg/aws"
	commonConfig "github.com/quadev-ltd/qd-common/pkg/config"
	"github.com/rs/zerolog/log"
)

// VertexAIConfig holds the Vertex AI configuration
type VertexAIConfig struct {
	ProjectID   string  `mapstructure:"project_id"`
	Location    string  `mapstructure:"location"`
	ModelName   string  `mapstructure:"model_name"`
	ConfigPath  string  `mapstructure:"config_path"`
	MaxTokens   int32   `mapstructure:"max_tokens"`
	Temperature float32 `mapstructure:"temperature"`
}

// Config is the configuration of the application
type Config struct {
	Verbose     bool
	Environment string
	AWS         commonAWS.Config
	VertexAI    VertexAIConfig `mapstructure:"vertex_ai"`
}

// Load reads and parses the configuration file from the specified location
func (config *Config) Load(path string) error {
	env := commonConfig.GetEnvironment()
	config.Environment = env
	config.Verbose = commonConfig.GetVerbose()

	log.Info().Msgf("Loading configuration for environment: %s", env)
	vip, err := commonConfig.SetupConfig(path, env)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}
	if err := vip.Unmarshal(&config); err != nil {
		return fmt.Errorf("Error unmarshaling configuration: %v", err)
	}

	return nil
}
