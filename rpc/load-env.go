package rpc

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

func InitEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func LoadMode() string {
	return os.Getenv("MODE")
}

func LoadPrivateKey() string {
	return os.Getenv("PRIVATE_KEY")
}

func GetEnvAsInt(name string, defaultVal int) int {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		return defaultVal
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Invalid value for %s: %v. Using default: %d", name, err, defaultVal)
		return defaultVal
	}
	return value
}
