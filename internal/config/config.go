package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string `yaml:"env" env-default:"local"`
	Version  string `yaml:"version" env-default:"unknown"`
	URL      string `yaml:"url"`
	Requests struct {
		Amount    int64 `yaml:"amount"`
		PerSecond int64 `yaml:"per_second"`
	} `yaml:"requests"`
}

func MustLoad() *Config {
	configPath := os.Getenv("SENDER_CONFIG_PATH")
	if configPath == "" {
		configPath = "config/local.yaml"
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}

func GetMockPort() string {
	port := os.Getenv("MOCK_SERVER_PORT")
	if port == "" {
		port = "8091"
	}

	return port
}

func GetMockEnv() string {
	env := os.Getenv("MOCK_SERVER_ENV")
	if env == "" {
		env = "local"
	}

	return env
}
