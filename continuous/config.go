package continuous

import (
	"context"
	"encoding/json"
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
	GraffitiSet   map[string]bool
	Connection
}

type GraffitiList struct {
	Graffitis []string `json:"graffitis"`
}

func (cfg *Configuration) NextAccount() *utils.Account {
	return &cfg.accounts[cfg.status.TxCount()%len(cfg.accounts)]
}
func createConfiguration(mode string) (Configuration, error) {
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
	keyperSetAddress, err := utils.ReadStringFromEnv("CONTINUOUS_KEYPER_SET_CONTRACT_ADDRESS")
	if err != nil {
		return cfg, err
	}
	sequencerAddress, err := utils.ReadStringFromEnv("CONTINUOUS_SEQUENCER_ADDRESS")
	if err != nil {
		return cfg, err
	}
	contracts, err := utils.SetupContracts(client, keyBroadcastAddress, sequencerAddress, keyperSetAddress, chainID)
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

	// Only load graffiti JSON when running in graffiti mode
	if mode == "graffiti" {
		graffitiSet, err := loadGraffitiJSON()
		if err != nil {
			return cfg, err
		}
		cfg.GraffitiSet = graffitiSet
	} else {
		// Initialize empty map for non-graffiti mode
		cfg.GraffitiSet = make(map[string]bool)
	}

	return cfg, nil
}

func loadGraffitiJSON() (map[string]bool, error) {
	graffitiFilePath, err := utils.ReadStringFromEnv("GRAFFITI_FILE_PATH")
	if err != nil {
		return nil, fmt.Errorf("graffiti file path variable not set: %w", err)
	}

	data, err := os.ReadFile(graffitiFilePath)
	if err != nil {
		return nil, fmt.Errorf("graffiti file not found: %w", err)
	}

	var gl GraffitiList
	if err := json.Unmarshal(data, &gl); err != nil {
		return nil, fmt.Errorf("invalid graffitis: %w", err)
	}

	graffitiSet := make(map[string]bool)
	for _, g := range gl.Graffitis {
		graffitiSet[g] = true
	}

	return graffitiSet, nil
}
