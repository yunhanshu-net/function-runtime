package conf

import "os"

func GetRunnerRoot() string {
	return os.Getenv("RUNNER_ROOT")
}

type Config struct {
}

func NewConfig() *Config {
	return &Config{}
}
