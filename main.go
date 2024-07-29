package main

import (
	"github.com/shutter-network/nethermind-tests/rpc"
	"log"
	"strings"
	"sync"
	"time"
)

func main() {
	rpc.InitEnv()
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
	interval := rpc.GetEnvAsInt("CHIADO_SEND_INTERVAL", 5)
	log.Printf("Running Chiado transactions at an interval of [%d] seconds", interval)
	tick := time.NewTicker(time.Duration(interval) * time.Second)

	for range tick.C {
		rpc.SendLegacyTx("https://erpc.chiado.staging.shutter.network")
	}
}

func runGnosisTransactions() {
	interval := rpc.GetEnvAsInt("GNOSIS_SEND_INTERVAL", 60)
	log.Printf("Running Gnosis transactions at an interval of [%d] seconds", interval)
	tick := time.NewTicker(time.Duration(interval) * time.Second)

	for range tick.C {
		rpc.SendLegacyTx("https://erpc.gnosis.shutter.network")
	}
}
