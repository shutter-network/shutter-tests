package main

import (
	"fmt"
	"github.com/shutter-network/nethermind-tests/rpc"
	"log"
	"strings"
	"sync"
	"time"
)

func main() {
	mode := rpc.LoadMode()
	modes := strings.Split(mode, ",")

	var wg sync.WaitGroup
	for _, m := range modes {
		switch m {
		case "chiado":
			wg.Add(1)
			go func() {
				runChiadoTransactions()
				wg.Done()
			}()
		case "gnosis":
			wg.Add(1)
			go func() {
				runGnosisTransactions()
				wg.Done()
			}()
		default:
			log.Printf("Unknown mode: %s", m)
		}
	}
	wg.Wait()
}

func runChiadoTransactions() {
	fmt.Println("Running Chiado transactions...")
	tick := time.NewTicker(5 * time.Second)

	for range tick.C {
		rpc.SendLegacyTx("https://erpc.chiado.staging.shutter.network")
	}
}

func runGnosisTransactions() {
	fmt.Println("Running Gnosis transactions...")
	tick := time.NewTicker(1 * time.Minute)

	for range tick.C {
		rpc.SendLegacyTx("https://erpc.gnosis.shutter.network")
	}
}
