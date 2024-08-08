package requests

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shutter-network/nethermind-tests/config"
)

func CancelTx(config config.Config, nonce uint64) error {
	log.Println("Cancelling transaction")
	client, err := ethclient.Dial(config.NodeURL)
	if err != nil {
		return fmt.Errorf("failed to connect to the Ethereum client: %w", err)
	}

	pKey := config.PrivateKey
	privateKey, err := crypto.HexToECDSA(pKey)
	if err != nil {
		return fmt.Errorf("failed to load private key: %w", err)
	}

	value := big.NewInt(0)    // in wei
	gasLimit := uint64(21000) // in units

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get suggested gas price: %w", err)
	}

	toAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")
	data := make([]byte, 0)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
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
		return fmt.Errorf("failed to sign transaction ID: %w", err)
	}

	rawTxBytes, err := signedTx.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal signed transaction ID: %w", err)
	}

	log.Printf("Signed transaction: %s\n", rawTxBytes)

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	log.Printf("Transaction sent: %s\n", signedTx.Hash().Hex())
	return nil
}
