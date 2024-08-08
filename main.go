package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/shutter-network/nethermind-tests/config"
	"github.com/shutter-network/nethermind-tests/tests"
	"github.com/shutter-network/nethermind-tests/utils"
)

func main() {
	cfg := config.LoadConfig()
	fmt.Println(cfg.Mode)
	mode := cfg.Mode
	modes := strings.Split(mode, ",")

	utils.EnableExtLoggingFile()

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
