package rpc

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
)

func SendLegacyTx(clientURL string) {
	client, err := ethclient.Dial(clientURL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	pKey := LoadPrivateKey()
	privateKey, err := crypto.HexToECDSA(pKey)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	log.Println("Sending transaction from: " + fromAddress.String())

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}

	value := big.NewInt(140000000000) // in wei (0.01 eth)
	gasLimit := uint64(50000)         // in units
	//gasPrice := big.NewInt(4000000000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	toAddress := fromAddress
	data := make([]byte, 4)
	_, err = rand.Read(data)
	if err != nil {
		log.Fatal(err)
	}

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
		fmt.Printf("Failed to send transaction: %v", err)
	}

	fmt.Printf("Transaction sent: %s\n", signedTx.Hash().Hex())
}
