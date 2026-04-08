package config

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadEnvironment() error {
	err := godotenv.Load()
	if err != nil {
		log.Printf("warning: .env file not found, loading from injected environment variables")
	}

	return nil
}
