package config

import (
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env  string     `yaml:"env" env-defoult:"local"`
	GRPC GRPCConfig `yaml:"grpc"`
}

type GRPCConfig struct {
	Port int `yaml:"port"`
}

func MustLoad() *Config {
	path := fetchConfigPath()

	if path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config path is empty")
	}

	var config Config

	if err := cleanenv.ReadConfig(path, &config); err != nil {
		panic("failed to read config")
	}

	return &config
}

func fetchConfigPath() string {
	var res string

	// --config="path/to/config.yaml"

	// CONFIG_PATH=./path/to/config/file.yaml myApp
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
