package continuous

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shutter-network/nethermind-tests/utils"
)

type Configuration struct {
	accounts      []utils.Account
	submitAccount utils.Account
	client        *ethclient.Client
	status        Status
	contracts     utils.Contracts
	chainID       *big.Int
	DbUser        string
	DbPass        string
	DbAddr        string
	DbName        string
	PkFile        string
	blameFolder   string
	Connection
}

func (cfg *Configuration) NextAccount() *utils.Account {
	return &cfg.accounts[cfg.status.TxCount()%len(cfg.accounts)]
}
func createConfiguration() (Configuration, error) {
	cfg := Configuration{
		status: Status{
			statusModMutex: &sync.Mutex{},
		},
	}
	PkFile, err := utils.ReadStringFromEnv("CONTINUOUS_PK_FILE")
	if err != nil {
		return cfg, err
	}
	cfg.PkFile = PkFile

	RpcUrl, err := utils.ReadStringFromEnv("CONTINUOUS_TEST_RPC_URL")
	if err != nil {
		return cfg, err
	}
	client, err := ethclient.Dial(RpcUrl)
	if err != nil {
		return cfg, fmt.Errorf("could not create client %v", err)
	}

	cfg.client = client

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return cfg, fmt.Errorf("could not query chainId %v", err)
	}

	cfg.chainID = chainID
	signerForChain := types.LatestSignerForChainID(chainID)

	submitKeyHex, err := utils.ReadStringFromEnv("CONTINUOUS_TEST_PK")
	if err != nil {
		return cfg, err
	}
	submitPrivateKey, err := crypto.HexToECDSA(submitKeyHex)
	if err != nil {
		return cfg, err
	}
	submitAccount, err := utils.AccountFromPrivateKey(submitPrivateKey, signerForChain)
	if err != nil {
		return cfg, err
	}
	log.Printf("submit account is %v\n", submitAccount.Address.Hex())
	submitNonce, err := client.NonceAt(context.Background(), submitAccount.Address, nil)
	if err != nil {
		return cfg, err
	}
	submitAccount.Nonce = big.NewInt(int64(submitNonce))
	cfg.submitAccount = submitAccount
	accounts := retrieveAccounts(NumFundedAccounts, client, signerForChain, &cfg)
	createdAccounts, err := createAccounts(NumFundedAccounts-len(accounts), signerForChain)
	if err != nil {
		return cfg, err
	}
	for _, created := range createdAccounts {
		err = utils.StoreAccount(created)
		if err != nil {
			return cfg, err
		}
		accounts = append(accounts, created)
	}
	for i := range accounts {
		err = fundNewAccount(accounts[i], MinimalFunding, &submitAccount, client)
		if err != nil {
			return cfg, err
		}
	}
	cfg.accounts = accounts
	submitNonce, err = client.NonceAt(context.Background(), submitAccount.Address, nil)
	if err != nil {
		return cfg, err
	}
	cfg.submitAccount.Nonce = big.NewInt(int64(submitNonce))

	keyBroadcastAddress, err := utils.ReadStringFromEnv("CONTINUOUS_KEY_BROADCAST_CONTRACT_ADDRESS")
	if err != nil {
		return cfg, err
	}
	depositContractAddress, err := utils.ReadStringFromEnv("CONTINUOUS_DEPOSIT_CONTRACT_ADDRESS")
	if err != nil {
		return cfg, err
	}
	keyperSetAddress, err := utils.ReadStringFromEnv("CONTINUOUS_KEYPER_SET_CONTRACT_ADDRESS")
	if err != nil {
		return cfg, err
	}
	sequencerAddress, err := utils.ReadStringFromEnv("CONTINUOUS_SEQUENCER_ADDRESS")
	if err != nil {
		return cfg, err
	}
	contracts, err := utils.SetupContracts(client, keyBroadcastAddress, sequencerAddress, keyperSetAddress, depositContractAddress)
	if err != nil {
		return cfg, err
	}
	cfg.contracts = contracts
	DbName, err := utils.ReadStringFromEnv("CONTINUOUS_DB_NAME")
	if err != nil {
		return cfg, err
	}
	cfg.DbName = DbName
	DbUser, err := utils.ReadStringFromEnv("CONTINUOUS_DB_USER")
	if err != nil {
		return cfg, err
	}
	cfg.DbUser = DbUser
	DbAddr, err := utils.ReadStringFromEnv("CONTINUOUS_DB_ADDRESS")
	if err != nil {
		return cfg, err
	}
	log.Println("DbAddr is", DbAddr)
	cfg.DbAddr = DbAddr
	DbPass, err := utils.ReadStringFromEnv("CONTINUOUS_DB_PASS")
	if err != nil {
		return cfg, err
	}
	cfg.DbPass = DbPass
	blameFolder, err := utils.ReadStringFromEnv("CONTINUOUS_BLAME_FOLDER")
	if err != nil {
		return cfg, err
	}
	if len(blameFolder) == 0 {
		tmp, err := os.MkdirTemp("", "blame")
		if err != nil {
			return cfg, err
		}
		blameFolder = tmp
	}
	cfg.blameFolder = blameFolder
	return cfg, nil
}
