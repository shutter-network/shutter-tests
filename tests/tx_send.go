package tests

import (
	"github.com/shutter-network/nethermind-tests/config"
	"github.com/shutter-network/nethermind-tests/rpc_reqs"
	"log"
	"time"
)

func RunChiadoTransactions(cfg config.Config) {
	interval := cfg.ChiadoSendInterval
	log.Printf("Running Chiado transactions at an interval of [%d] seconds", interval)
	tick := time.NewTicker(interval)

	for range tick.C {
		rpc_reqs.SendLegacyTx(cfg.ChiadoURL, cfg.PrivateKey)
	}
}

func RunGnosisTransactions(cfg config.Config) {
	interval := cfg.GnosisSendInterval
	log.Printf("Running Gnosis transactions at an interval of [%d] seconds", interval)
	tick := time.NewTicker(interval)

	for range tick.C {
		rpc_reqs.SendLegacyTx(cfg.GnosisURL, cfg.PrivateKey)
	}
}
