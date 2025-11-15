package server

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBURL      string
	Port       string
	ServerHost string
	ServerPort string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using default environment variables")
	}

	config := &Config{
		DBURL:      getEnv("DATABASE_URL", "postgres://guruorgoru:balakotalu77@localhost:5432/mmo"),
		Port:       getEnv("PORT", "8414"),
		ServerHost: getEnv("SERVER_HOST", "127.0.0.1"),
		ServerPort: getEnv("SERVER_PORT", "8414"),
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
