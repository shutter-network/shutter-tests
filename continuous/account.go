package continuous

import (
	"context"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shutter-network/nethermind-tests/utils"
)

func retrieveAccounts(num int, client *ethclient.Client, signerForChain types.Signer, cfg *Configuration) []utils.Account {
	var result []utils.Account
	fd, err := os.Open(cfg.PkFile)
	if err != nil {
		log.Println("could not open pk file", cfg.PkFile)
	}
	defer fd.Close()
	pks, err := utils.ReadPks(fd)
	if err != nil {
		log.Printf("error when reading private keys %v\n", err)
		panic(err)
	}
	for _, pk := range pks {
		acc, err := utils.AccountFromPrivateKey(pk, signerForChain)
		if err != nil {
			log.Printf("could not retrieve account %v\n", err)
		}
		result = append(result, acc)
		if len(result) == num {
			break
		}
	}
	for i := range result {
		accNonce, err := client.NonceAt(context.Background(), result[i].Address, nil)
		if err != nil {
			log.Printf("failed to get nonce for %v: %v\n", result[i].Address, err)
		}
		log.Printf("setting account nonce for %v to %v\n", result[i].Address.Hex(), accNonce)
		result[i].Nonce = big.NewInt(int64(accNonce))
	}
	return result
}

func fundNewAccount(account utils.Account, amount int64, submitAccount *utils.Account, client *ethclient.Client) error {
	target := big.NewInt(amount)
	current, err := client.BalanceAt(context.Background(), account.Address, nil)
	if err != nil {
		return err
	}
	missing := big.NewInt(0).Sub(target, current)

	half := big.NewInt(0).Div(target, big.NewInt(2))
	if missing.Int64() <= half.Int64() {
		return nil
	}
	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}
	var data []byte
	nonce := submitAccount.UseNonce()
	if err != nil {
		return err
	}
	log.Printf("Using submitter nonce %v\n", nonce)
	tx := types.NewTransaction(nonce.Uint64(), account.Address, missing, gasLimit, gasPrice, data)
	signedTx, err := submitAccount.Sign(submitAccount.Address, tx)
	if err != nil {
		return err
	}
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return err
	}
	log.Println("sent funding tx", signedTx.Hash().Hex(), "to", signedTx.To().Hex())
	_, err = bind.WaitMined(context.Background(), client, signedTx)
	return err
}

func createAccounts(num int, signerForChain types.Signer) ([]utils.Account, error) {
	accounts := make([]utils.Account, num)
	for i := 0; i < num; i++ {
		pk, err := crypto.GenerateKey()
		if err != nil {
			return accounts, err
		}
		account, err := utils.AccountFromPrivateKey(pk, signerForChain)
		if err != nil {
			return accounts, err
		}
		accounts[i] = account
	}
	return accounts, nil
}
