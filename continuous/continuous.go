package continuous

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	cryptorand "crypto/rand"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shutter-network/nethermind-tests/stress"
	"github.com/shutter-network/shutter/shlib/shcrypto"
)

const KeyperSetChangeLookAhead = 2
const NumFundedAccounts = 6
const MinimalFunding = int64(500000000000000000) // 0.5 ETH in wei

type Connection struct {
	db *pgxpool.Pool
}

type Status struct {
	lastShutterTS pgtype.Date
	txInFlight    []ShutterTx
	txDone        []ShutterTx
}

func (s Status) TxCount() int {
	return len(s.txInFlight) + len(s.txDone)
}

type ShutterTx struct {
	innerTx      *types.Transaction
	outerTx      *types.Transaction
	sender       stress.Account
	prefix       shcrypto.Block
	triggerBlock int64
	txStatus     TxStatus
	ctx          context.Context
}

type TxStatus int

const (
	Signed        TxStatus = iota + 1 // user transaction was encrypted and tx to the sequencer contract was signed and sent
	Sequenced                         // tx to sequencer contract was mined
	Included                          // next shutterized block was found and the tx was included
	NotIncluded                       // next shutterized block was found, but this tx was not part of it
	SystemFailure                     // we could not assess the status of this tx, e.g. because the client connection failed
)

func (ts TxStatus) String() string {
	return [...]string{"Signed", "Sequenced", "NotIncluded", "Included"}[ts-1]
}

func (ts TxStatus) EnumIndex() int {
	return int(ts)
}

type Configuration struct {
	accounts      []stress.Account
	submitAccount stress.Account
	client        *ethclient.Client
	status        Status
	contracts     stress.Contracts
	chainID       *big.Int
}

func (cfg Configuration) NextAccount() stress.Account {
	return cfg.accounts[cfg.status.TxCount()%len(cfg.accounts)]
}

type ShutterBlock struct {
	Number int64
	Ts     pgtype.Date
}

func PrefixFromBlockNumber(blockNumber int64) shcrypto.Block {
	bytes := make([]byte, shcrypto.BlockSize)
	if blockNumber > 0 {
		binary.LittleEndian.PutUint64(bytes, uint64(blockNumber))
	}
	return shcrypto.Block(bytes)
}

func retrieveAccounts(num int, client *ethclient.Client, signerForChain types.Signer) []stress.Account {
	var result []stress.Account
	fd, err := os.Open("pk.hex")
	if err != nil {
		fmt.Println("could not open pk.hex")
	}
	defer fd.Close()
	pks, err := stress.ReadPks(fd)
	if err != nil {
		fmt.Printf("error when reading private keys %v\n", err)
	}
	for _, pk := range pks {
		acc, err := stress.AccountFromPrivateKey(pk, signerForChain)
		if err != nil {
			fmt.Printf("could not retrieve account %v\n", err)
		}
		result = append(result, acc)
		if len(result) == num {
			break
		}
	}
	for _, account := range result {
		accNonce, err := client.NonceAt(context.Background(), account.Address, nil)
		if err != nil {
			fmt.Printf("failed to get nonce for %v: %v\n", account.Address, err)
		}
		account.Nonce = *big.NewInt(int64(accNonce))
	}
	return result
}

func fundNewAccount(account stress.Account, amount int64, submitAccount *stress.Account, client *ethclient.Client) error {
	target := big.NewInt(amount)
	current, err := client.BalanceAt(context.Background(), account.Address, nil)
	if err != nil {
		return err
	}
	value := big.NewInt(0).Sub(target, current)
	if value.Int64() <= 0 {
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
	tx := types.NewTransaction(nonce.Uint64(), account.Address, value, gasLimit, gasPrice, data)
	signedTx, err := submitAccount.Sign(submitAccount.Address, tx)
	if err != nil {
		return err
	}
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return err
	}
	log.Println("sent funding tx", signedTx.Hash().Hex(), "to", signedTx.To)
	_, err = bind.WaitMined(context.Background(), client, signedTx)
	return err
}

func createConfiguration() (Configuration, error) {
	cfg := Configuration{}
	RpcUrl, err := stress.ReadStringFromEnv("CONTINUOUS_TEST_RPC_URL")
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

	submitKeyHex, err := stress.ReadStringFromEnv("CONTINUOUS_TEST_PK")
	if err != nil {
		return cfg, err
	}
	submitPrivateKey, err := crypto.HexToECDSA(submitKeyHex)
	if err != nil {
		return cfg, err
	}
	submitAccount, err := stress.AccountFromPrivateKey(submitPrivateKey, signerForChain)
	if err != nil {
		return cfg, err
	}
	fmt.Printf("submit account is %v\n", submitAccount.Address.Hex())
	submitNonce, err := client.NonceAt(context.Background(), submitAccount.Address, nil)
	if err != nil {
		return cfg, err
	}
	submitAccount.Nonce = *big.NewInt(int64(submitNonce))
	cfg.submitAccount = submitAccount
	accounts := retrieveAccounts(NumFundedAccounts, client, signerForChain)
	createdAccounts, err := createAccounts(NumFundedAccounts-len(accounts), signerForChain)
	if err != nil {
		return cfg, err
	}
	for _, created := range createdAccounts {
		err = stress.StoreAccount(created)
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

	keyBroadcastAddress, err := stress.ReadStringFromEnv("CONTINUOUS_KEY_BROADCAST_CONTRACT_ADDRESS")
	if err != nil {
		return cfg, err
	}
	keyperSetAddress, err := stress.ReadStringFromEnv("CONTINUOUS_KEYPER_SET_CONTRACT_ADDRESS")
	if err != nil {
		return cfg, err
	}
	sequencerAddress, err := stress.ReadStringFromEnv("CONTINUOUS_SEQUENCER_ADDRESS")
	if err != nil {
		return cfg, err
	}
	contracts, err := stress.SetupContracts(client, keyBroadcastAddress, sequencerAddress, keyperSetAddress)
	if err != nil {
		return cfg, err
	}
	cfg.contracts = contracts

	return cfg, nil
}

func createAccounts(num int, signerForChain types.Signer) ([]stress.Account, error) {
	accounts := make([]stress.Account, num)
	for i := 0; i < num; i++ {
		pk, err := crypto.GenerateKey()
		if err != nil {
			return accounts, err
		}
		account, err := stress.AccountFromPrivateKey(pk, signerForChain)
		if err != nil {
			return accounts, err
		}
		accounts[i] = account
	}
	return accounts, nil

}

func NewConnection() Connection {
	DbUser := "postgres"
	DbPass := "test"
	dbAddr := "localhost:5432"
	DbName := "shutter_metrics"

	ctx := context.Background()
	db, err := pgxpool.New(ctx, fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", DbUser, DbPass, dbAddr, DbName))
	if err != nil {
		panic("db connection failed")
	}
	connection := Connection{db: db}
	return connection
}

func QueryAllShutterBlocks(out chan<- ShutterBlock) {
	waitBetweenQueries := 5 * time.Second
	status := Status{lastShutterTS: pgtype.Date{}}
	connection := NewConnection()
	query := `
	SELECT to_timestamp(max(b.block_timestamp)) as ts,
	COUNT(d.*) as count
	FROM decryption_keys_message_decryption_key d
		LEFT JOIN block b ON d.decryption_keys_message_slot = b.slot
		GROUP BY d.decryption_keys_message_slot
		ORDER BY d.decryption_keys_message_slot;
	`
	rows, err := connection.db.Query(context.Background(), query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var ts pgtype.Date
	var count int
	for rows.Next() {
		rows.Scan(&ts, &count)
		if !ts.Time.IsZero() {
			status.lastShutterTS = ts
		}
	}
	if rows.Err() != nil {
		fmt.Println("errors when finding shutterized blocks: ", rows.Err())
	}
	for {
		time.Sleep(waitBetweenQueries)
		fmt.Printf(".")
		newShutterBlock := queryNewestShutterBlock(status.lastShutterTS, *connection.db)
		if !newShutterBlock.Ts.Time.IsZero() {
			status.lastShutterTS = newShutterBlock.Ts
			// send event (block number, timestamp) to out channel
			out <- newShutterBlock
		}
	}
}

func queryNewestShutterBlock(lastBlockTS pgtype.Date, db pgxpool.Pool) ShutterBlock {

	var block int64
	var ts pgtype.Date
	var count int
	query := `
	SELECT b.block_number,
	to_timestamp(max(b.block_timestamp)) as ts,
	COUNT(d.*) as count
	FROM decryption_keys_message_decryption_key d
		LEFT JOIN block b ON d.decryption_keys_message_slot = b.slot
		WHERE b.block_timestamp > $1
		GROUP BY d.decryption_keys_message_slot, b.block_number
		ORDER BY d.decryption_keys_message_slot;
	`
	rows, err := db.Query(context.Background(), query, lastBlockTS.Time.Unix())
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&block, &ts, &count)
		if !ts.Time.IsZero() {
			fmt.Printf("\nFOUND NEW SHUTTER BLOCK %v: %v [%v]", block, ts.Time, count)
		}
		if count > 1 {
			fmt.Printf("missed some blocks: %v", count-1)
		}
	}
	if rows.Err() != nil {
		fmt.Println("errors when finding shutterized blocks: ", rows.Err())
	}
	res := ShutterBlock{}
	res.Number = block
	res.Ts = ts
	return res
}

func SendShutterizedTX(blockNumber int64, lastTimestamp pgtype.Date, cfg *Configuration) {
	// [x] fund accounts in cfg
	// [x] get available account from cfg
	// [x] create prefix from trigger data
	// [x] encrypt tx
	// [x] send to sequencer
	// [x] add to txInFlight
	fmt.Printf("\nSENDING NEW TX FOR %v", blockNumber)
	account := cfg.NextAccount()
	fmt.Printf("\nUsing %v\n", account.Address.Hex())
	var data []byte
	gasLimit := uint64(21000)
	gas, err := stress.GasCalculationFromClient(context.Background(), cfg.client)
	if err != nil {
		panic(err)
	}
	innerNonceP := account.UseNonce()
	innerTx := types.NewTx(
		&types.DynamicFeeTx{
			ChainID:   cfg.chainID,
			Nonce:     innerNonceP.Uint64(),
			GasFeeCap: gas.Fee,
			GasTipCap: gas.Tip,
			Gas:       gasLimit,
			To:        &cfg.submitAccount.Address,
			Value:     big.NewInt(1),
			Data:      data,
		},
	)

	signedInnerTx, err := cfg.submitAccount.Sign(cfg.submitAccount.Address, innerTx)
	if err != nil {
		panic(err)
	}
	sigma, err := shcrypto.RandomSigma(cryptorand.Reader)
	if err != nil {
		panic("could not get random sigma")
	}
	identityPrefix := PrefixFromBlockNumber(blockNumber)
	identity := stress.ComputeIdentity(identityPrefix[:], cfg.submitAccount.Address)

	var buff bytes.Buffer
	err = signedInnerTx.EncodeRLP(&buff)
	if err != nil {
		panic(err)
	}

	eon, eonKey, err := stress.GetEonKey(context.Background(), cfg.client, cfg.contracts.KeyperSetManager, cfg.contracts.KeyBroadcastContract, KeyperSetChangeLookAhead)
	if err != nil {
		panic(err)
	}
	encrypted := shcrypto.Encrypt(buff.Bytes(), (*shcrypto.EonPublicKey)(eonKey), identity, sigma)
	opts := cfg.submitAccount.Opts()

	opts.Value = big.NewInt(0).Sub(signedInnerTx.Cost(), signedInnerTx.Value())

	fmt.Println(opts)
	outerTx, err := cfg.contracts.Sequencer.SubmitEncryptedTransaction(
		opts, eon, identityPrefix, encrypted.Marshal(), new(big.Int).SetUint64(signedInnerTx.Gas()),
	)
	if err != nil {
		panic(err)
	}

	tx := ShutterTx{
		outerTx:      outerTx,
		innerTx:      signedInnerTx,
		sender:       account,
		prefix:       identityPrefix,
		triggerBlock: blockNumber,
		txStatus:     TxStatus(Signed),
	}
	cfg.status.txInFlight = append(cfg.status.txInFlight, tx)
	go WatchTx(&tx, cfg.client)
}

func WatchTx(tx *ShutterTx, client *ethclient.Client) {
	submissionReceipt, err := stress.WaitForTx(*tx.outerTx, "submission", time.Hour, client)
	if err != nil {
		tx.txStatus = TxStatus(SystemFailure)
	}
	if submissionReceipt.Status == 1 {
		tx.txStatus = TxStatus(Sequenced)
	} else {
		tx.txStatus = TxStatus(SystemFailure)
	}
	if tx.txStatus == SystemFailure {
		// TODO: forfeit nonce
		return
	}
	includedReceipt, err := stress.WaitForTx(*tx.innerTx, "inclusion", time.Hour, client)
	if err != nil {
		tx.txStatus = TxStatus(SystemFailure)
	}
	if includedReceipt.Status == 1 {
		tx.txStatus = TxStatus(Included)
	} else {
		tx.txStatus = TxStatus(NotIncluded)
	}
	// [ ] wait for submit
	// [ ] wait for failure signal
	// [ ] on failure: forfeit account nonce
	// [ ] move to txDone
	fmt.Println(tx)
}

func forfeitNonce(account stress.Account, client *ethclient.Client) error {
	nonce := account.Nonce
	chainId, err := client.ChainID(context.Background())
	if err != nil {
		return err
	}
	gasLimit := uint64(21000)
	var data []byte
	gas, err := stress.GasCalculationFromClient(context.Background(), client)
	if err != nil {
		return err
	}
	tx := types.NewTx(
		&types.DynamicFeeTx{
			ChainID:   chainId,
			Nonce:     nonce.Uint64(),
			GasFeeCap: gas.Fee,
			GasTipCap: gas.Tip,
			Gas:       gasLimit,
			To:        &account.Address,
			Value:     big.NewInt(1),
			Data:      data,
		},
	)

	signed, err := account.Sign(account.Address, tx)
	if err != nil {
		return err
	}
	err = client.SendTransaction(context.Background(), signed)
	return err
}

func Setup() (Configuration, error) {
	return createConfiguration()
}
