package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// config.go - for parsing of config yaml file
type Config struct {
	Env 		string `yaml:"env" env-default:"local"` // считываение параметров с yaml
	StoragePath string `yaml:"storage_path" env-required:"true"` 
	HTTPServer `yaml:"http_server"`
}

type HTTPServer struct {
	Address 		string 			`yaml:"address" env-default:"localhost:8080"`
	Timeout 		time.Duration 	`yaml:"timeout env-default:"4s"`
	IddleTimeout 	time.Duration 	`yaml:"iddle_timeout" env-default:"60s"`
}


func MustLoad() *Config{
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("env file is empty or not found")
	}
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}