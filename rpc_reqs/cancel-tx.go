package rpc_reqs

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shutter-network/nethermind-tests/config"
	"log"
	"math/big"
)

func CancelTx(config config.Config, nonce uint64) error {
	fmt.Println("Cancelling transaction")
	client, err := ethclient.Dial(config.NodeURL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	pKey := config.PrivateKey
	privateKey, err := crypto.HexToECDSA(pKey)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	value := big.NewInt(0)    // in wei
	gasLimit := uint64(21000) // in units

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("Failed to get suggested gas price: %v", err)
	}

	toAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")
	data := make([]byte, 0)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &toAddress,
		Value:    value,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     data,
	})
	signer := types.NewLondonSigner(chainID)

	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}

	rawTxBytes, err := signedTx.MarshalJSON()
	if err != nil {
		log.Fatalf("Failed to marshal signed transaction: %v", err)
	}

	fmt.Printf("Signed transaction: %s\n", rawTxBytes)

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Printf("Failed to send transaction: %v", err)
		return err
	}

	fmt.Printf("Transaction sent: %s\n", signedTx.Hash().Hex())
	return nil
}
