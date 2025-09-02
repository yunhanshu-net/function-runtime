package conf

import (
	"github.com/yunhanshu-net/function-runtime/pkg/config"
	"os"
)

var root string

func GetRunnerRoot() string {
	if root != "" {
		return root
	}
	if config.Get().ServerConfig.RunnerRoot != "" {
		return config.Get().ServerConfig.RunnerRoot
	}
	if os.Getenv("DEV_ROOT") != "" {
		root = os.Getenv("DEV_ROOT")
		return root
	}

	root = os.Getenv("RUNNER_ROOT")
	return root
}

func IsDev() bool {
	return config.Get().ServerConfig.IsDev
}

type Config struct {
}

func NewConfig() *Config {
	return &Config{}
}
