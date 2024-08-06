package tests

import (
	"context"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shutter-network/nethermind-tests/config"
	"github.com/shutter-network/nethermind-tests/rpc_reqs"
	"log"
	"time"
)

func SendAndCheckTransaction(cfg config.Config) bool {
	signedTx, err := rpc_reqs.SendLegacyTx(cfg.NodeURL, cfg.PrivateKey)
	if err != nil {
		log.Fatalf("Failed to send transaction %s", err)
	}

	nonce := signedTx.Nonce()
	result := rpc_reqs.WaitForReceipt(cfg.NodeURL, signedTx.Hash().Hex(), cfg.Timeout)

	if result == false { // we didn't receive the transaction within the timeout
		log.Printf("Transaction not received within timeout: %s. Cancelling transaction.", signedTx.Hash().Hex())
		err := rpc_reqs.CancelTx(cfg, nonce)
		if err != nil {
			log.Printf("Cancelling transaction failed with error: %s. "+
				"Checking if transaction got confirmed in the meantime.", err)

			client, err := ethclient.Dial(cfg.NodeURL)
			if err != nil {
				log.Fatalf("Failed to connect to the Ethereum client: %v", err)
			}

			privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
			if err != nil {
				log.Fatalf("Failed to load private key: %v", err)
			}

			fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
			log.Println("Sending transaction from: " + fromAddress.String())

			pendingNonce, err := client.PendingNonceAt(context.Background(), fromAddress)
			if err != nil {
				log.Fatalf("Failed to get pending nonce: %v", err)
			}

			if pendingNonce == nonce+1 {
				log.Printf("Transaction confirmed in the meantime. "+
					"Pending nonce [%d] == nonce [%d] + 1", pendingNonce, nonce)
				return false
			}
		}
	}
	return true
}

func RunSendAndWaitTest(cfg config.Config) {
	endTime := time.Now().Add(cfg.TestDuration)
	successCount := 0
	failCount := 0

	log.Printf("Running Send And Wait transactions")
	log.Printf("Test Duration [%s]", cfg.TestDuration)
	log.Printf("Wait Timeout [%s]", cfg.Timeout)
	log.Printf("Node URL %s", cfg.NodeURL)

	for time.Now().Before(endTime) {
		success := SendAndCheckTransaction(cfg)
		if success {
			successCount++
		} else {
			failCount++
		}
	}

	// Calculate execution percentage
	totalAttempts := successCount + failCount
	successPercentage := (float64(successCount) / float64(totalAttempts)) * 100
	failurePercentage := (float64(failCount) / float64(totalAttempts)) * 100

	log.Printf("Test Duration: %s, Wait Timeout: %s,  Successes: %d, Failures: %d, Success Percentage: %.2f%%, Failure Percentage: %.2f%%",
		cfg.TestDuration, cfg.Timeout, successCount, failCount, successPercentage, failurePercentage)

}
