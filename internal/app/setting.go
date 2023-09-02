package app

import (
	"os"
)

type appConfig struct {
	RootPath string
}

var (
	Config appConfig = appConfig{
		RootPath: setConfig(),
	}
)

func setConfig() string {
	var config, exists = os.LookupEnv("APP_ROOT_PATH")
	if !exists {
		config = "."
	}
	return config
}
