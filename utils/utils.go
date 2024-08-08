package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

func EnableExtLoggingFile() {
	logFile, err := os.OpenFile(fmt.Sprintf("./logs/%s.log", time.Now().Format(time.RFC3339)), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(fmt.Errorf("error opening file: %v", err))
	}

	log.SetOutput(logFile)
}
