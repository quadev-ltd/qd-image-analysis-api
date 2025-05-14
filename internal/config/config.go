package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config holds the application configuration settings
type Config struct {
	Environment string `mapstructure:"environment"`
	Verbose     bool   `mapstructure:"verbose"`
	AWS         struct {
		Key    string `mapstructure:"key"`
		Secret string `mapstructure:"secret"`
	} `mapstructure:"aws"`
}

// Load reads and parses the configuration file from the specified location
func (config *Config) Load(configLocation string) error {
	environment := os.Getenv("APP_ENVIRONMENT")
	if environment == "" {
		environment = "local"
	}

	viper.SetConfigName(fmt.Sprintf("config.%s", environment))
	viper.SetConfigType("yml")
	viper.AddConfigPath(configLocation)

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	err = viper.Unmarshal(config)
	if err != nil {
		return err
	}

	config.Environment = environment
	return nil
}
