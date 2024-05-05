package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
)

type Config struct {
	Env     string        `yaml:"env" env-required:"true"`
	GRPC    GRPCConfig    `yaml:"grpc"`
	Storage StorageConfig `yaml:"storage"`
	Cloud   CloudConfig   `yaml:"cloud"`
}

type GRPCConfig struct {
	Port int `yaml:"port"`
}

type StorageConfig struct {
	TmpPath       string `yaml:"tmp_path"`
	CompletedPath string `yaml:"completed_path"`
}

type CloudConfig struct {
	MaxImageSize int                 `yaml:"max_image_size"`
	AvailableExt map[string]struct{} `yaml:"available_ext"`
	LimitUD      int                 `yaml:"limit_ud"`
	LimitList    int                 `yaml:"limit_list"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
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
