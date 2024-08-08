package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/shutter-network/nethermind-tests/config"
	"github.com/shutter-network/nethermind-tests/tests"
)

func main() {
	cfg := config.LoadConfig()
	fmt.Println(cfg.Mode)
	mode := cfg.Mode
	modes := strings.Split(mode, ",")

	logFile, err := os.OpenFile("./logs/tests.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(fmt.Errorf("error opening file: %v", err))
	}

	log.SetOutput(logFile)

	var wg sync.WaitGroup
	for _, m := range modes {
		switch m {
		case "chiado":
			wg.Add(1)
			go func() {
				tests.RunChiadoTransactions(cfg)
				wg.Done()
			}()
		case "gnosis":
			wg.Add(1)
			go func() {
				tests.RunGnosisTransactions(cfg)
				wg.Done()
			}()
		case "send-wait":
			wg.Add(1)
			go func() {
				tests.RunSendAndWaitTest(cfg)
				wg.Done()
			}()
		default:
			log.Printf("Unknown mode: %s", m)
		}
	}
	wg.Wait()
}
