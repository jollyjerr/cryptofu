package bittrex

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadCredentials() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("BIT_KEY")
	// secretKey := os.Getenv("BIT_SECRET")

	return apiKey
}
