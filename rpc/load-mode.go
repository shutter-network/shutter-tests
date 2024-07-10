package rpc

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

func LoadMode() string {
	// Load environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Access the variables
	mode := os.Getenv("MODE")

	return mode
}
