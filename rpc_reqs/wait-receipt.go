package rpc_reqs

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"time"
)

func WaitForReceipt(clientURL string, txHash string, timeout time.Duration) (bool, error) {
	client, err := ethclient.Dial(clientURL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
		return false, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		receipt, err := client.TransactionReceipt(ctx, common.HexToHash(txHash))
		if err == nil {
			if receipt.Status == types.ReceiptStatusFailed {
				log.Printf("Transaction failed: %s", txHash)
				return false, nil
			}
			log.Printf("Transaction succeeded: %s", txHash)
			return true, nil
		}

		select {
		case <-ctx.Done():
			log.Printf("Timeout waiting for transaction receipt: %s", txHash)
			return false, ctx.Err()
		case <-time.After(1 * time.Second):
			// Wait before retrying
		}
	}
}
