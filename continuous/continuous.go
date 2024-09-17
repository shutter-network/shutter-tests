package continuous

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"sync"
	"time"

	cryptorand "crypto/rand"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/montanaflynn/stats"
	"github.com/shutter-network/nethermind-tests/utils"
	"github.com/shutter-network/shutter/shlib/shcrypto"
)

const KeyperSetChangeLookAhead = 2
const NumFundedAccounts = 6
const MinimalFunding = int64(500000000000000000) // 0.5 ETH in wei

type Connection struct {
	db *pgxpool.Pool
}

type Status struct {
	statusModMutex *sync.Mutex
	lastShutterTS  pgtype.Date
	txInFlight     []*ShutterTx
	txDone         []*ShutterTx
}

func (s Status) TxCount() int {
	return len(s.txInFlight) + len(s.txDone)
}

func (s *Status) AddTxInFlight(t *ShutterTx) {
	s.statusModMutex.Lock()
	s.txInFlight = append(s.txInFlight, t)
	s.statusModMutex.Unlock()
}

type ShutterTx struct {
	innerTx         *types.Transaction
	outerTx         *types.Transaction
	sender          *utils.Account
	prefix          shcrypto.Block
	triggerBlock    int64
	submissionBlock int64
	inclusionBlock  int64
	cancelBlock     int64
	txStatus        TxStatus
	ctx             context.Context
	cancel          context.CancelFunc
}

func (tx *ShutterTx) String() string {
	var outerTxHash string
	var outerTxNonce string
	var innerTxHash string
	var innerTxNonce string
	if tx.innerTx == nil {
		innerTxHash = "nil"
		innerTxNonce = "nil"
	} else {
		innerTxHash = tx.innerTx.Hash().Hex()
		innerTxNonce = fmt.Sprint(tx.innerTx.Nonce())
	}
	if tx.outerTx == nil {
		outerTxHash = "nil"
		outerTxNonce = "nil"
	} else {
		outerTxHash = tx.outerTx.Hash().Hex()
		outerTxNonce = fmt.Sprint(tx.outerTx.Nonce())
	}
	return fmt.Sprintf("ShutterTx[%v]\t%v\n  trigger:\t%8d\n  submit :\t%8d\t%v n:%v\n  include:\t%8d\t%v n:%v\n  cancel :\t%8d",
		tx.txStatus,
		tx.sender.Address.Hex(),
		tx.triggerBlock,
		tx.submissionBlock,
		outerTxHash,
		outerTxNonce,
		tx.inclusionBlock,
		innerTxHash,
		innerTxNonce,
		tx.cancelBlock,
	)
}

type TxStatus int

const (
	Signed        TxStatus = iota + 1 // user transaction was encrypted and tx to the sequencer contract was signed and sent
	Sequenced                         // tx to sequencer contract was mined
	Included                          // next shutterized block was found and the tx was included
	NotSequenced                      // next shutterized block was found, but this tx was not sequenced
	NotIncluded                       // next shutterized block was found, but this tx was not part of it
	SystemFailure                     // we could not assess the status of this tx, e.g. because the client connection failed
)

func (ts TxStatus) String() string {
	return [...]string{"Signed", "Sequenced", "Included", "NotSequenced", "NotIncluded", "SystemFailure"}[ts-1]
}

func (ts TxStatus) EnumIndex() int {
	return int(ts)
}

type Configuration struct {
	accounts      []utils.Account
	submitAccount utils.Account
	client        *ethclient.Client
	status        Status
	contracts     utils.Contracts
	chainID       *big.Int
}

func (cfg *Configuration) NextAccount() *utils.Account {
	return &cfg.accounts[cfg.status.TxCount()%len(cfg.accounts)]
}

type ShutterBlock struct {
	Number int64
	Ts     pgtype.Date
}

func retrieveAccounts(num int, client *ethclient.Client, signerForChain types.Signer) []utils.Account {
	var result []utils.Account
	fd, err := os.Open("pk.hex")
	if err != nil {
		log.Println("could not open pk.hex")
	}
	defer fd.Close()
	pks, err := utils.ReadPks(fd)
	if err != nil {
		log.Printf("error when reading private keys %v\n", err)
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

func createConfiguration() (Configuration, error) {
	cfg := Configuration{
		status: Status{
			statusModMutex: &sync.Mutex{},
		},
	}
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
	accounts := retrieveAccounts(NumFundedAccounts, client, signerForChain)
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
	contracts, err := utils.SetupContracts(client, keyBroadcastAddress, sequencerAddress, keyperSetAddress)
	if err != nil {
		return cfg, err
	}
	cfg.contracts = contracts

	return cfg, nil
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

// FIXME: maybe https://github.com/shutter-network/rolling-shutter/blob/fe4ea5fe73a3416b4df9c710d27f764505f89a1f/rolling-shutter/keyperimpl/gnosis/newslot.go#L129 is a better way to do this!
func QueryAllShutterBlocks(out chan<- ShutterBlock) {
	waitBetweenQueries := 1 * time.Second
	status := Status{lastShutterTS: pgtype.Date{}}
	connection := NewConnection()
	query := `
		SELECT
		to_timestamp(b.block_timestamp)
		FROM validator_status AS v
		LEFT JOIN proposer_duties AS p
		ON p.validator_index = v.validator_index
		LEFT JOIN block AS b
		ON b.slot=p.slot
		WHERE v.status = 'active_ongoing'
		AND b.slot = p.slot
		ORDER BY b.block_number DESC
		LIMIT 1;
	`
	rows, err := connection.db.Query(context.Background(), query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var ts pgtype.Date
	for rows.Next() {
		rows.Scan(&ts)
		if !ts.Time.IsZero() {
			status.lastShutterTS = ts
		}
	}
	if rows.Err() != nil {
		log.Println("errors when finding shutterized blocks: ", rows.Err())
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
	block := int64(0)
	var ts pgtype.Date
	query := `
		SELECT
		b.block_number,
		to_timestamp(b.block_timestamp)
		FROM validator_status AS v
		LEFT JOIN proposer_duties AS p
		ON p.validator_index = v.validator_index
		LEFT JOIN block AS b
		ON b.slot=p.slot
		WHERE v.status = 'active_ongoing'
		AND b.slot = p.slot
		AND b.block_timestamp > $1;
	`
	rows, err := db.Query(context.Background(), query, lastBlockTS.Time.Unix())
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		if block != 0 {
			log.Fatal("Finding multiple blocks")
		}
		rows.Scan(&block, &ts)
		if !ts.Time.IsZero() {
			log.Printf("FOUND NEW SHUTTER BLOCK %v: %v", block, ts.Time)
		}
	}
	if rows.Err() != nil {
		log.Println("errors when finding shutterized blocks: ", rows.Err())
	}
	res := ShutterBlock{}
	res.Number = block
	res.Ts = ts
	return res
}

func SendShutterizedTX(blockNumber int64, lastTimestamp pgtype.Date, cfg *Configuration) {
	account := cfg.NextAccount()
	log.Printf("SENDING NEW TX FOR %v from %v", blockNumber, account.Address.Hex())
	gasLimit := uint64(21000)
	var data []byte
	gas, err := utils.GasCalculationFromClient(context.Background(), cfg.client, utils.DefaultGasPriceFn)
	if err != nil {
		panic(err)
	}
	identityPrefix := utils.PrefixFromBlockNumber(blockNumber)
	identity := utils.ComputeIdentity(identityPrefix[:], cfg.submitAccount.Address)
	innerNonceP := account.UseNonce()
	innerTx := types.NewTx(
		&types.DynamicFeeTx{
			ChainID:   cfg.chainID,
			Nonce:     innerNonceP.Uint64(),
			GasFeeCap: gas.Fee,
			GasTipCap: gas.Tip,
			Gas:       gasLimit,
			To:        &cfg.submitAccount.Address,
			Value:     big.NewInt(blockNumber),
			Data:      data,
		},
	)

	signedInnerTx, err := account.Sign(account.Address, innerTx)
	if err != nil {
		panic(err)
	}
	sigma, err := shcrypto.RandomSigma(cryptorand.Reader)
	if err != nil {
		panic("could not get random sigma")
	}

	buff, err := signedInnerTx.MarshalBinary()
	if err != nil {
		panic(err)
	}

	eon, eonKey, err := utils.GetEonKey(context.Background(), cfg.client, cfg.contracts.KeyperSetManager, cfg.contracts.KeyBroadcastContract, KeyperSetChangeLookAhead)
	if err != nil {
		panic(err)
	}
	encrypted := shcrypto.Encrypt(buff, (*shcrypto.EonPublicKey)(eonKey), identity, sigma)
	opts := cfg.submitAccount.Opts()

	opts.Value = big.NewInt(0).Sub(signedInnerTx.Cost(), signedInnerTx.Value())

	submitGas, err := utils.GasCalculationFromClient(context.Background(), cfg.client, utils.HighPriorityGasPriceFn)
	if err != nil {
		panic(err)
	}
	opts.GasFeeCap = submitGas.Fee
	opts.GasTipCap = submitGas.Tip
	log.Printf("submit nonce: %v\n", opts.Nonce)
	outerTx, err := cfg.contracts.Sequencer.SubmitEncryptedTransaction(
		opts, eon, identityPrefix, encrypted.Marshal(), new(big.Int).SetUint64(signedInnerTx.Gas()),
	)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)

	tx := ShutterTx{
		outerTx:      outerTx,
		innerTx:      signedInnerTx,
		sender:       account,
		prefix:       identityPrefix,
		triggerBlock: blockNumber,
		txStatus:     TxStatus(Signed),
		ctx:          ctx,
		cancel:       cancel,
	}
	cfg.status.AddTxInFlight(&tx)
	log.Println(signedInnerTx.Hash())
	go WatchTx(&tx, cfg.client)
}

func WatchTx(tx *ShutterTx, client *ethclient.Client) {
	defer tx.cancel()
	submissionReceipt, err := utils.WaitForTxCtx(tx.ctx, *tx.outerTx, fmt.Sprintf("submission[%v]", tx.triggerBlock), client)
	select {
	case <-tx.ctx.Done():
		switch tx.ctx.Err() {
		case context.Canceled:
			tx.txStatus = TxStatus(NotSequenced)
			log.Println(tx)
			return
		case context.DeadlineExceeded:
			tx.txStatus = TxStatus(SystemFailure)
			log.Println(tx)
			return
		default:
			fmt.Println("something else")
		}
	default:
	}
	if err != nil {
		tx.txStatus = TxStatus(SystemFailure)
	}
	if submissionReceipt.Status == 1 {
		tx.txStatus = TxStatus(Sequenced)
		tx.submissionBlock = submissionReceipt.BlockNumber.Int64()
	} else {
		tx.txStatus = TxStatus(SystemFailure)
	}
	if tx.txStatus != Sequenced {
		log.Println(tx)
		err = forfeitNonce(tx.innerTx.Nonce(), *tx.sender, client)
		if err != nil {
			log.Println("could not reset nonce", err)
		}
		return
	}
	includedReceipt, err := utils.WaitForTxCtx(tx.ctx, *tx.innerTx, fmt.Sprintf("inclusion[%v]", tx.triggerBlock), client)
	select {
	case <-tx.ctx.Done():
		switch tx.ctx.Err() {
		case context.Canceled:
			err = forfeitNonce(tx.innerTx.Nonce(), *tx.sender, client)
			if err != nil {
				if err.Error()[0:8] == "OldNonce" {
					// FIXME: the error message is rpc endpoint implementation specific (in this case Nethermind)
					// ...but at this point there is a very high chance, that the tx
					// was included before we could send the cancellation.
					fmt.Println("OOOOOOLD NONCE")
				}
				log.Println("could not reset nonce", err)
			}
			tx.txStatus = TxStatus(NotIncluded)
			log.Println(tx)
			return
		case context.DeadlineExceeded:
			err = forfeitNonce(tx.innerTx.Nonce(), *tx.sender, client)
			if err != nil {
				if err.Error()[0:8] == "OldNonce" {
					// FIXME: the error message is rpc endpoint implementation specific (in this case Nethermind)
					// ...but at this point there is a very high chance, that the tx
					// was included before we could send the cancellation.
					fmt.Println("OOOOOOLD NONCE")
				}
				log.Println("could not reset nonce", err)
			}
			tx.txStatus = TxStatus(SystemFailure)
			log.Println(tx)
			return
		default:
		}
	default:
	}
	if err != nil {
		tx.txStatus = TxStatus(SystemFailure)
		log.Println(err)
	}
	if includedReceipt != nil {
		tx.txStatus = TxStatus(Included)
		tx.inclusionBlock = includedReceipt.BlockNumber.Int64()
		log.Printf("INCLUDED!!! %v\n", tx.innerTx.Hash())
	}
	log.Println(tx)
}

func forfeitNonce(nonce uint64, account utils.Account, client *ethclient.Client) error {
	chainId, err := client.ChainID(context.Background())
	if err != nil {
		return err
	}
	gasLimit := uint64(21000)
	var data []byte
	gas, err := utils.GasCalculationFromClient(context.Background(), client, utils.HighPriorityGasPriceFn)
	if err != nil {
		return err
	}
	tx := types.NewTx(
		&types.DynamicFeeTx{
			ChainID:   chainId,
			Nonce:     nonce,
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
	if err != nil {
		return err
	}
	receipt, err := bind.WaitMined(context.Background(), client, signed)
	if err != nil {
		return err
	}
	if receipt.Status != 1 {
		return fmt.Errorf("forfeit tx not accepted")
	}
	return err
}

func CheckTxInFlight(blockNumber int64, cfg *Configuration) {
	cfg.status.statusModMutex.Lock()
	var newInflight []*ShutterTx
	newDone := cfg.status.txDone[:]
	highestInclusion := int64(0)
	for _, tx := range cfg.status.txInFlight {
		if tx.inclusionBlock > highestInclusion {
			highestInclusion = tx.inclusionBlock
		}
	}
	for _, tx := range cfg.status.txInFlight {
		done := false
		switch s := tx.txStatus; s {
		case Sequenced:
			// cancel signal should be: another included tx with inclusion block > submission block
			if highestInclusion > tx.submissionBlock {
				tx.cancel()
				tx.cancelBlock = blockNumber
				done = true
			}
		case Included:
			done = true
		case SystemFailure:
			done = true
		default:
		}
		if done {
			newDone = append(newDone, tx)
		} else {
			newInflight = append(newInflight, tx)
		}
	}
	cfg.status.txInFlight = newInflight
	cfg.status.txDone = newDone
	cfg.status.statusModMutex.Unlock()
}

func PrintAllTx(cfg *Configuration) {
	fmt.Println("INFLIGHT")
	for _, tx := range cfg.status.txInFlight {
		fmt.Println(tx)
	}
	fmt.Println("DONE")
	for _, tx := range cfg.status.txDone {
		fmt.Println(tx)
	}
}

func Setup() (Configuration, error) {
	return createConfiguration()
}

func CollectContinuousTestStats(startBlock uint64, endBlock uint64, cfg *Configuration) error {
	failCnt := 0
	var failed []Submission
	var delays []float64
	success, err := collectSubmitIncomingTx(startBlock, endBlock, cfg)
	if err != nil {
		return err
	}
	submit, err := collectSequencerEvents(startBlock, endBlock, cfg)
	if err != nil {
		return err
	}
	successByTrigger := make(map[int64]Success)
	for i := range success {
		successByTrigger[success[i].trigger] = success[i]
	}

	for i := range submit {
		trigger := submit[i].trigger
		included, ok := successByTrigger[trigger]
		if ok {
			delay := float64(included.included - submit[i].sequenced)
			delays = append(delays, delay)
		} else {
			failed = append(failed, submit[i])
			failCnt++
		}
	}
	failPct := (float64(failCnt) / float64(len(submit)) * 100)
	avgDelay, err := stats.Mean(delays)
	if err != nil {
		return err
	}
	maxDelay, err := stats.Max(delays)
	if err != nil {
		return err
	}
	minDelay, err := stats.Min(delays)
	if err != nil {
		return err
	}
	medianDelay, err := stats.Median(delays)
	if err != nil {
		return err
	}
	lastValidTrigger := endBlock - 1
	triggers, err := queryBlockTriggers(startBlock, lastValidTrigger, cfg)
	if err != nil {
		return err
	}
	submitTriggers := make([]int64, len(submit))
	for i, s := range submit {
		submitTriggers[i] = s.trigger
	}

	shutterizedPct := float64(len(triggers)) / float64(endBlock-startBlock) * 100
	log.Printf("found %v shutter test tx in block range[%v:%v] (%v triggers)\n", len(submit), startBlock, endBlock, len(triggers))
	log.Printf("shutterized blocks %3.2f%%", shutterizedPct)
	log.Printf("fail percentage %3.2f", failPct)
	log.Printf("missed triggers %v: %v", len(triggers)-len(submit), utils.Difference(triggers, submitTriggers))
	log.Printf("delay max %0.0f min %0.0f avg %3.2f median %3.2f", maxDelay, minDelay, avgDelay, medianDelay)
	for _, f := range failed {
		blame, err := blameValidator(f, cfg)
		if err != nil {
			log.Println(err)
		}
		log.Println(blame)
	}
	return err
}

type ValidatorBlame struct {
	prefix         []byte
	triggerBlock   int64
	submitBlock    int64
	targetBlock    int64
	targetSlot     int64
	validatorIndex int64
}

func blameValidator(submission Submission, cfg *Configuration) (ValidatorBlame, error) {
	prefix := utils.PrefixFromBlockNumber(submission.trigger)
	prefixBytes := prefix[:]
	blame := ValidatorBlame{
		triggerBlock: submission.trigger,
		submitBlock:  submission.sequenced,
		prefix:       prefixBytes,
	}
	err := queryWhoToBlame(&blame)
	if err != nil {
		return blame, nil
	}
	err = queryDecryptionKeysBySlot(blame)
	if err != nil {
		return blame, nil
	}
	log.Println(blame)

	return blame, nil
}
