package conf

import (
	"os"
)

var root string

func GetRunnerRoot() string {
	if root != "" {
		return root
	}
	if os.Getenv("DEV_ROOT") != "" {
		root = os.Getenv("DEV_ROOT")
		return root
	}

	root = os.Getenv("RUNNER_ROOT")
	return root
}

func IsDev() bool {
	return os.Getenv("DEV_ROOT") != ""
}

type Config struct {
}

func NewConfig() *Config {
	return &Config{}
}
