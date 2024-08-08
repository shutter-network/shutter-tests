package requests

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func SendLegacyTx(clientURL string, pKey string) (*types.Transaction, error) {
	client, err := ethclient.Dial(clientURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client: %w", err)
	}

	privateKey, err := crypto.HexToECDSA(pKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	log.Println("Sending transaction from: " + fromAddress.String())

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	value := big.NewInt(1)    // in wei
	gasLimit := uint64(21000) // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get suggested gas price: %w", err)
	}

	toAddress := fromAddress
	data := make([]byte, 0)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to chain ID: %w", err)
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
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	rawTxBytes, err := signedTx.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshall transaction: %w", err)
	}

	log.Printf("Signed transaction: %s\n", rawTxBytes)

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	txHash := signedTx.Hash().Hex()
	log.Printf("Transaction sent: %s\n", txHash)
	return signedTx, nil
}
