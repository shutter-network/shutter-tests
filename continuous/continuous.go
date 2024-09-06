package continuous

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shutter-network/nethermind-tests/stress"
	"github.com/shutter-network/shutter/shlib/shcrypto"
)

type Connection struct {
	db *pgxpool.Pool
}

type Status struct {
	lastShutterTS *pgtype.Date
	txInFlight    []ShutterTx
}

type ShutterTx struct {
	sender       stress.Account
	prefix       shcrypto.Block
	triggerBlock int64
	txStatus     TxStatus
}

type TxStatus int

const (
	Signed      TxStatus = iota + 1 // user transaction was encrypted and tx to the sequencer contract was signed and sent
	Sequenced                       // tx to sequencer contract was mined
	Included                        // next shutterized block was found and the tx was included
	NotIncluded                     // next shutterized block was found, but this tx was not part of it
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
}

type ShutterBlock struct {
	Number int64
	Ts     *pgtype.Date
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
	cfg.submitAccount = submitAccount
	accounts, err := createAccounts(6, signerForChain)
	if err != nil {
		return cfg, err
	}
	for _, account := range accounts {
		stress.StoreAccount(account)
	}
	cfg.accounts = accounts
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
		accounts = append(accounts, account)
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
	status := Status{lastShutterTS: nil}
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
			fmt.Println(ts.Time, count)
			status.lastShutterTS = &ts
		}
	}
	if rows.Err() != nil {
		fmt.Println("errors when finding shutterized blocks: ", rows.Err())
	}
	for {
		fmt.Printf(".")
		newShutterBlock := queryNewestShutterBlock(*status.lastShutterTS, *connection.db)
		time.Sleep(waitBetweenQueries)
		if !newShutterBlock.Ts.Time.IsZero() {
			fmt.Println(newShutterBlock)
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
			fmt.Println("FOUND NEW SHUTTER BLOCK:", ts.Time, count)
		}
	}
	if rows.Err() != nil {
		fmt.Println("errors when finding shutterized blocks: ", rows.Err())
	}
	res := ShutterBlock{}
	res.Number = block
	res.Ts = &ts
	return res
}

func SendShutterizedTX(blockNumber int64, lastTimestamp pgtype.Date, cfg Configuration) {
	// get available account from cfg
	// create prefix from trigger data
	// encrypt tx
	// send to sequencer
	// add to txInFlight
	fmt.Printf("SENDING NEW TX FOR %v", blockNumber)
}

func Setup() (Configuration, error) {
	return createConfiguration()
}
