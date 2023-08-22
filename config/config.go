package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type envConfig struct {
	MongoDBUrl  string
	MongoDBname string
	Port        string
}

var envCfg envConfig

func init() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	envPath := filepath.Join(dir, ".env")

	if err := godotenv.Load(envPath); err != nil {
		panic(err)
	}

	envCfg = envConfig{
		MongoDBUrl:  os.Getenv("MONGODB_URI"),
		MongoDBname: os.Getenv("MONGODB_NAME"),
		Port:        os.Getenv("PORT"),
	}
}

func GetConfig() *envConfig {
	return &envCfg
}
