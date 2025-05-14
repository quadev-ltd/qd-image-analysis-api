package main

import (
	"log"

	commontConfig "github.com/quadev-ltd/qd-common/pkg/config"

	"qd-image-analysis-api/internal/application"
	"qd-image-analysis-api/internal/config"
)

func main() {
	var configurations config.Config
	configLocation := "./internal/config"
	err := configurations.Load(configLocation)
	if err != nil {
		log.Fatalln("Failed loading the configurations", err)
	}

	var centralConfig commontConfig.Config
	centralConfig.Load(
		configurations.Environment,
		configurations.AWS.Key,
		configurations.AWS.Secret,
	)

	application := application.NewApplication(&configurations, &centralConfig)
	application.StartServer()

	defer application.Close()
}
