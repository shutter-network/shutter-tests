package rpc_reqs

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
)

func SendLegacyTx(clientURL string, pKey string) (*types.Transaction, error) {
	client, err := ethclient.Dial(clientURL)
	if err != nil {
		return nil, err
	}

	privateKey, err := crypto.HexToECDSA(pKey)
	if err != nil {
		return nil, err
	}

	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	log.Println("Sending transaction from: " + fromAddress.String())

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, err
	}

	value := big.NewInt(1)    // in wei
	gasLimit := uint64(21000) // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	toAddress := fromAddress
	data := make([]byte, 0)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, err
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
		return nil, err
	}

	rawTxBytes, err := signedTx.MarshalJSON()
	if err != nil {
		return nil, err
	}

	fmt.Printf("Signed transaction: %s\n", rawTxBytes)

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}

	txHash := signedTx.Hash().Hex()
	fmt.Printf("Transaction sent: %s\n", txHash)
	return signedTx, nil
}
