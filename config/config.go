package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Mode               string
	PrivateKey         string
	ChiadoURL          string
	ChiadoSendInterval time.Duration
	GnosisURL          string
	GnosisSendInterval time.Duration
	Interval           time.Duration
	Timeout            time.Duration
	TestDuration       time.Duration
	NodeURL            string
}

func LoadConfig() Config {
	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	config := Config{
		Mode:               os.Getenv("MODE"),
		PrivateKey:         os.Getenv("PRIVATE_KEY"),
		ChiadoURL:          os.Getenv("CHIADO_URL"),
		ChiadoSendInterval: time.Duration(GetEnvAsInt("CHIADO_SEND_INTERVAL")) * time.Second,
		GnosisURL:          os.Getenv("GNOSIS_URL"),
		GnosisSendInterval: time.Duration(GetEnvAsInt("GNOSIS_SEND_INTERVAL")) * time.Second,
		Timeout:            time.Duration(GetEnvAsInt("WAIT_TX_TIMEOUT")) * time.Second,
		TestDuration:       time.Duration(GetEnvAsInt("TEST_DURATION")) * time.Second,
		NodeURL:            os.Getenv("NODE_URL"),
	}

	return config
}

func GetEnvAsInt(name string) int {
	valueStr := os.Getenv(name)
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Fatalf("Invalid value for %s: %v.", name, err)
	}
	return value
}
