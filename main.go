package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/shutter-network/nethermind-tests/config"
	"github.com/shutter-network/nethermind-tests/tests"
	"github.com/shutter-network/nethermind-tests/utils"

	"github.com/shutter-network/nethermind-tests/continuous"
)

func main() {
	var modes []string
	var cfg config.Config
	if len(os.Args[1:]) == 0 {
		cfg := config.LoadConfig()
		log.Println(cfg.Mode)
		mode := cfg.Mode

		modes = strings.Split(mode, ",")
	} else {
		modes = []string{os.Args[1]}
	}
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
			if len(os.Args[2:]) != 2 {
				log.Fatalf("Usage: %v %v start-block end-block", os.Args[0], os.Args[1])
			}
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
	lastStats := time.Now().Unix()
	cache := make(continuous.BlockCache)
	startBlock := uint64(0)
	blocks := make(chan continuous.ShutterBlock)
	go continuous.QueryAllShutterBlocks(blocks, &cfg)
	for block := range blocks {
		if startBlock == 0 {
			startBlock = uint64(block.Number)
		}
		continuous.CheckTxInFlight(block.Number, &cfg)
		continuous.SendShutterizedTX(block.Number, block.Ts, &cfg)
		now := time.Now().Unix()
		if now-lastStats > 120 {
			log.Println("running stats")
			lastStats = now
			err = continuous.CollectContinuousTestStats(startBlock, uint64(block.Number), &cache, &cfg)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func runCollector() {
	start, end := utils.CollectBlockRangeFromArgs()
	cfg, err := continuous.Setup()
	if err != nil {
		log.Fatal(err)
	}
	cache := make(continuous.BlockCache)
	err = continuous.CollectContinuousTestStats(start, end, &cache, &cfg)
	if err != nil {
		log.Fatal(err)
	}
}
