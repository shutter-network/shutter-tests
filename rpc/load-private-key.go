package rpc

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

func LoadPrivateKey() string {
	// Load environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Access the variables
	privateKey := os.Getenv("PRIVATE_KEY")

	return privateKey
}
