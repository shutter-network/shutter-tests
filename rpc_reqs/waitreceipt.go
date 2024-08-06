package rpc_reqs

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"time"
)

func WaitForReceipt(clientURL string, txHash string, timeout time.Duration) (result bool) {
	client, err := ethclient.Dial(clientURL)
	if err != nil {
		log.Fatalf("Failed to connect to client %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		receipt, err := client.TransactionReceipt(ctx, common.HexToHash(txHash))
		if err == nil {
			if receipt.Status == types.ReceiptStatusFailed {
				log.Printf("Transaction failed: %s", txHash)
				return false
			}
			log.Printf("Transaction succeeded: %s", txHash)
			return true
		}

		if !errors.Is(err, ethereum.NotFound) {
			log.Fatalf("Receipt retrieval failed %s", err)
		}

		select {
		case <-ctx.Done():
			log.Printf("Timeout waiting for transaction receipt: %s", txHash)
			return false
		case <-time.After(1 * time.Second):
			// Wait before retrying
		}
	}
}
