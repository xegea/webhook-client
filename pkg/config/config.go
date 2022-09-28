package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env         string
	ApiKey      string
	AdminApiKey string
	ServerUrl   string
}

func LoadConfig(env *string) (*Config, error) {

	err := godotenv.Load(*env)
	if err != nil {
		log.Println(".env file not found")
	}

	environment := os.Getenv("ENV")

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("API_KEY cannot be empty")
	}

	adminApiKey := os.Getenv("ADMIN_API_KEY")

	serverUrl := os.Getenv("SERVER_URL")
	if apiKey == "" {
		return nil, fmt.Errorf("SERVER_URL cannot be empty")
	}

	return &Config{
		Env:         environment,
		ApiKey:      apiKey,
		AdminApiKey: adminApiKey,
		ServerUrl:   serverUrl,
	}, nil
}
