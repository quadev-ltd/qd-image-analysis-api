package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Environment string `mapstructure:"environment"`
	Verbose     bool   `mapstructure:"verbose"`
	AWS         struct {
		Key    string `mapstructure:"key"`
		Secret string `mapstructure:"secret"`
	} `mapstructure:"aws"`
}

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
