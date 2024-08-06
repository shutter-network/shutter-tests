package tests

import (
	"github.com/shutter-network/nethermind-tests/config"
	"github.com/shutter-network/nethermind-tests/requests"
	"log"
	"time"
)

func RunChiadoTransactions(cfg config.Config) {
	interval := cfg.ChiadoSendInterval
	log.Printf("Running Chiado transactions at an interval of [%d] seconds", interval)
	tick := time.NewTicker(interval)

	for range tick.C {
		_, err := requests.SendLegacyTx(cfg.ChiadoURL, cfg.PrivateKey)
		if err != nil {
			log.Fatalf("Failed to send transaction %s", err)
		}
	}
}

func RunGnosisTransactions(cfg config.Config) {
	interval := cfg.GnosisSendInterval
	log.Printf("Running Gnosis transactions at an interval of [%d] seconds", interval)
	tick := time.NewTicker(interval)

	for range tick.C {
		_, err := requests.SendLegacyTx(cfg.GnosisURL, cfg.PrivateKey)
		if err != nil {
			log.Fatalf("Failed to send transaction %s", err)
		}
	}
}
