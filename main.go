package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/shutter-network/nethermind-tests/config"
	"github.com/shutter-network/nethermind-tests/tests"
	"github.com/shutter-network/nethermind-tests/utils"

	"github.com/shutter-network/nethermind-tests/continuous"
)

func main() {
	cfg := config.LoadConfig()
	log.Println(cfg.Mode)
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
		case "continuous":
			wg.Add(1)
			go func() {
				runContinous()
				wg.Done()
			}()
		case "collect":
			wg.Add(1)
			go func() {
				runCollector()
				wg.Done()
			}()
		default:
			log.Printf("Unknown mode: %s", m)
		}
	}
	wg.Wait()
}

func runContinous() {
	cfg, err := continuous.Setup()
	if err != nil {
		panic(err)
	}
	fmt.Println("Running continous tx tests...")
	startBlock := uint64(0)
	blocks := make(chan continuous.ShutterBlock)
	go continuous.QueryAllShutterBlocks(blocks, &cfg)
	for block := range blocks {
		if startBlock == 0 {
			startBlock = uint64(block.Number)
		}
		continuous.CheckTxInFlight(block.Number, &cfg)
		continuous.SendShutterizedTX(block.Number, block.Ts, &cfg)
		if block.Number%10 == 0 {
			continuous.CollectContinuousTestStats(startBlock, uint64(block.Number), &cfg)
		}
	}
}

func runCollector() {
	cfg, err := continuous.Setup()
	if err != nil {
		panic(err)
	}
	continuous.CollectContinuousTestStats(11857846, 11859032, &cfg)
}
