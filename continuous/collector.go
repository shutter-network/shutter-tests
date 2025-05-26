package continuous

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgtype"
	"github.com/montanaflynn/stats"
	"github.com/shutter-network/nethermind-tests/utils"
	"github.com/shutter-network/shutter/shlib/shcrypto"
	"golang.org/x/sync/errgroup"
)

type ValidatorBlame struct {
	prefix            []byte
	triggerBlock      int64
	submitBlock       int64
	targetBlock       int64
	targetBlockTS     *pgtype.Date
	targetSlot        int64
	decryptedTxHash   common.Hash
	validatorIndex    int64
	decryptionKey     DecryptionKey
	sender            common.Address
	proposerPublicKey string
	credentials       common.Address
}

func (b ValidatorBlame) String() string {
	emptyHash := common.Hash(make([]byte, common.HashLength))
	if b.decryptedTxHash == emptyHash {
		return fmt.Sprintf(
			"validator id\t: %v\n"+
				"public key:\t %v\n"+
				"withdrawal:\t %v\n"+
				"triggered\t: %v\n"+
				"submitted\t: %v\n"+
				"target block\t: %v\n"+
				"target slot\t: %v\n"+
				"target ts\t: %v\n"+
				"NO DECRYPTION KEY SEEN\n"+
				"identity preimage:\n"+
				"prefix\t%v\n"+
				"sender\t%v\n",
			b.validatorIndex,
			b.proposerPublicKey,
			b.credentials.Hex(),
			b.triggerBlock,
			b.submitBlock,
			b.targetBlock,
			b.targetSlot,
			b.targetBlockTS.Time.UTC().Format("2006-01-01 15:04:05.000000"),
			hex.EncodeToString(b.prefix),
			hex.EncodeToString(b.sender.Bytes()),
		)
	} else {
		return fmt.Sprintf(
			"validator id\t: %v\n"+
				"public key:\t %v\n"+
				"withdrawal:\t %v\n"+
				"triggered\t: %v\n"+
				"submitted\t: %v\n"+
				"target block\t: %v\n"+
				"target slot\t: %v\n"+
				"target ts\t: %v\n"+
				"ts (key-target)\t: %vms\n"+
				"decrypted tx\t: %v\n"+
				"decryption key:\n%v\n",
			b.validatorIndex,
			b.proposerPublicKey,
			b.credentials.Hex(),
			b.triggerBlock,
			b.submitBlock,
			b.targetBlock,
			b.targetSlot,
			b.targetBlockTS.Time.UTC().Format("2006-01-01 15:04:05.000000"),
			b.decryptionKey.createdTs.Time.UnixMilli()-b.targetBlockTS.Time.UnixMilli(),
			b.decryptedTxHash.Hex(),
			b.decryptionKey,
		)
	}
}

type DecryptionKey struct {
	identityPreimage []byte
	txPointer        int
	eon              int
	createdTs        *pgtype.Date
}

func (d DecryptionKey) String() string {
	return fmt.Sprintf(
		"first seen\t: %v\n"+
			"tx pointer\t: %v\n"+
			"identity preimage:\n"+
			"prefix\t%v\n"+
			"sender\t%v",
		d.createdTs.Time.UTC().Format("2006-01-01 15:04:05.000000"),
		d.txPointer,
		hex.EncodeToString(d.identityPreimage[:32]),
		hex.EncodeToString(d.identityPreimage[32:]),
	)
}

type Submission struct {
	trigger   int64
	sequenced int64
}

func withdrawAddressForPublicKey(proposerPublicKey string, cfg *Configuration) (*common.Address, error) {
	fmt.Printf("proposer public key is %v\n", proposerPublicKey)
	proposerKeyBytes, err := hex.DecodeString(proposerPublicKey[2:])
	if err != nil {
		fmt.Println("could not decode %v", err)
		return nil, err
	}
	result, err := cfg.contracts.Depositcontract.ValidatorWithdrawalCredentials(&bind.CallOpts{
		Pending: false,
		Context: context.Background(),
	}, proposerKeyBytes)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Printf("got a result %v\n", result)
	credentials := common.BytesToAddress(result[:])
	fmt.Printf("credentials as hex %v\n", credentials.Hex())
	return &credentials, nil
}

func collectSequencerEvents(startBlock uint64, endBlock uint64, cfg *Configuration) ([]Submission, error) {
	var submissions []Submission
	ctx := context.Background()
	opts := &bind.FilterOpts{
		Start:   startBlock,
		End:     &endBlock,
		Context: ctx,
	}
	it, err := cfg.contracts.Sequencer.FilterTransactionSubmitted(opts)
	if err != nil {
		return submissions, err
	}
	for {
		if it.Next() {
			ev := *it.Event
			submission := Submission{
				trigger:   utils.BlockNumberFromPrefix(ev.IdentityPrefix),
				sequenced: int64(ev.Raw.BlockNumber),
			}
			if submission.trigger >= int64(startBlock) && submission.trigger <= (int64(endBlock)+1000) {
				submissions = append(submissions, submission)
			}
		} else {
			err = it.Error()
			break
		}
	}
	return submissions, err
}

type Success struct {
	trigger  int64
	included int64
}

type BlockCache struct {
	store sync.Map
}

func (b *BlockCache) Load(key uint64) ([]Success, bool) {
	v, ok := b.store.Load(key)
	if ok {
		return v.([]Success), ok
	} else {
		return nil, false
	}
}

func (b *BlockCache) Store(key uint64, value []Success) {
	b.store.Store(key, value)
}

func (b *BlockCache) MaxKey() uint64 {
	m := uint64(0)
	b.store.Range(func(k, v interface{}) bool {
		m = max(m, k.(uint64))
		return true
	})
	return m
}

type NewBlockNumber struct {
	uint64
}

func WatchHeads(ctx context.Context, client *ethclient.Client, blocksChannel chan *NewBlockNumber) error {
	log.Println("START watching heads")
	newHeads := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(ctx, newHeads)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case head := <-newHeads:
			ev := &NewBlockNumber{
				head.Number.Uint64(),
			}
			blocksChannel <- ev
		case err := <-sub.Err():
			log.Println("error when watching heads:", err)
			return err
		}
	}
}

func PrimeBlockCache(cache *BlockCache, cfg *Configuration) error {
	blocks := make(chan *NewBlockNumber)
	group, ctx := errgroup.WithContext(context.Background())
	group.Go(func() error { return WatchHeads(ctx, cfg.client, blocks) })
	for block := range blocks {
		maxCached := cache.MaxKey()
		if maxCached > 0 && maxCached < block.uint64 {
			collectSubmitIncomingTx(maxCached, block.uint64, cache, cfg)
		}
	}
	go func() {
		group.Wait()
		close(blocks)
	}()

	log.Println("DONE watching heads")
	return group.Wait()
}

func collectSubmitIncomingTx(startBlock uint64, endBlock uint64, cache *BlockCache, cfg *Configuration) ([]Success, error) {
	var result []Success
	for blockNum := startBlock; blockNum <= endBlock; blockNum++ {
		if found, ok := cache.Load(blockNum); ok {
			result = append(result, found...)
		} else {
			if endBlock-startBlock > 1 {
				log.Println("cache miss", blockNum)
			}
			num := big.NewInt(int64(blockNum))
			block, err := cfg.client.BlockByNumber(context.Background(), num)
			if err != nil {
				return result, nil
			}
			txs := block.Transactions()
			var successForBlock []Success
			for _, tx := range txs {
				if tx.To() != nil && tx.To().Hex() == cfg.submitAccount.Address.Hex() {
					success := Success{tx.Value().Int64(), block.Number().Int64()}
					result = append(result, success)
					successForBlock = append(successForBlock, success)
				}
			}
			cache.Store(blockNum, successForBlock)
		}
	}
	return result, nil
}
func queryBlockTriggers(startBlock uint64, endBlock uint64, cfg *Configuration) ([]int64, error) {
	var blocks []int64
	var block int64
	query := `
		SELECT
			b.block_number
		FROM block AS b
			LEFT JOIN proposer_duties AS p
			ON p.slot = b.slot
			LEFT JOIN validator_status AS s
			ON p.validator_index = s.validator_index
		WHERE b.block_number >= $1
		AND b.block_number <= $2
		AND s.status = 'active_ongoing';
		`
	connection := GetConnection(cfg)
	rows, err := connection.db.Query(context.Background(), query, startBlock, endBlock)
	if err != nil {
		return blocks, err
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&block)
		blocks = append(blocks, block)
	}
	if rows.Err() != nil {
		log.Println("errors when finding shutterized blocks: ", rows.Err())
		return blocks, err
	}
	return blocks, nil
}

func queryWhoToBlame(blame *ValidatorBlame, cfg *Configuration) error {
	var targetBlock, targetSlot, validatorIndex int64
	var targetTS *pgtype.Date

	queryWhoToBlame := `
	SELECT
		b.block_number,
		b.slot,
		v.validator_index,
		to_timestamp(b.block_timestamp),
		p.public_key
	FROM block AS b
		LEFT JOIN proposer_duties AS p ON p.slot = b.slot
		LEFT JOIN validator_status AS v ON v.validator_index = p.validator_index
	WHERE v.status = 'active_ongoing'
	AND b.block_number > $1
	ORDER BY b.block_number ASC
	LIMIT 1;`
	connection := GetConnection(cfg)
	rows, err := connection.db.Query(context.Background(), queryWhoToBlame, blame.submitBlock)
	if err != nil {
		return err
	}
	defer rows.Close()
	var publicKey string
	for rows.Next() {
		rows.Scan(&targetBlock, &targetSlot, &validatorIndex, &targetTS, &publicKey)
		blame.targetBlock = targetBlock
		blame.targetSlot = targetSlot
		blame.targetBlockTS = targetTS
		blame.validatorIndex = validatorIndex
		blame.proposerPublicKey = publicKey
		credentials, err := withdrawAddressForPublicKey(publicKey, cfg)
		if err != nil {
			fmt.Println("error getting withdrawal credentials %v", err)
		}
		blame.credentials = *credentials

	}
	if rows.Err() != nil {
		log.Println("errors when finding validator to blame: ", rows.Err())
		return err
	}
	return nil
}

func checkSlotMismatch(blame *ValidatorBlame, cfg *Configuration) error {
	identityPreimage := utils.PrefixFromBlockNumber(blame.triggerBlock)
	var seenSlot int64

	queryDecryptionKeysByPreimage := `
	SELECT decryption_keys_message_slot
	FROM decryption_key AS k
		LEFT JOIN decryption_keys_message_decryption_key AS dkmdk
			ON k.id=dkmdk.decryption_key_id
	WHERE k.identity_preimage=decode($1, 'hex')
	`
	connection := GetConnection(cfg)
	rows, err := connection.db.Query(context.Background(), queryDecryptionKeysByPreimage, hex.EncodeToString(identityPreimage[:]))
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&identityPreimage, &seenSlot)
		log.Println("seen at", seenSlot)
	}
	return nil
}

func blameValidator(submission Submission, cfg *Configuration) (ValidatorBlame, error) {
	prefix := utils.PrefixFromBlockNumber(submission.trigger)
	prefixBytes := prefix[:]
	blame := ValidatorBlame{
		triggerBlock: submission.trigger,
		submitBlock:  submission.sequenced,
		prefix:       prefixBytes,
		sender:       cfg.submitAccount.Address,
	}
	err := queryWhoToBlame(&blame, cfg)
	if err != nil {
		return blame, err
	}
	err = queryDecryptionKeysBySlot(&blame, cfg)
	if err != nil {
		return blame, err
	}
	if true || blame.decryptionKey.identityPreimage == nil {
		err = checkSlotMismatch(&blame, cfg)
		if err != nil {
			return blame, err
		}
	}
	return blame, nil
}

func queryDecryptionKeysBySlot(blame *ValidatorBlame, cfg *Configuration) error {
	var identityPreimage, txHash []byte
	var txPointer, eon int
	var createdTs pgtype.Date

	queryDecryptionKeysBySlot := `
		SELECT
			k.identity_preimage,
			d.tx_pointer,
			d.eon,
			d.created_at,
			t.tx_hash
        FROM decryption_keys_message_decryption_key AS dkmdk
                LEFT JOIN decryption_key AS k
                ON dkmdk.decryption_key_id=k.id
                LEFT JOIN decryption_keys_message AS d
                ON d.slot=dkmdk.decryption_keys_message_slot
				LEFT JOIN decrypted_tx AS t
				ON t.decryption_key_id=k.id
        WHERE dkmdk.decryption_keys_message_slot=$1;`
	connection := GetConnection(cfg)
	rows, err := connection.db.Query(context.Background(), queryDecryptionKeysBySlot, blame.targetSlot)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&identityPreimage, &txPointer, &eon, &createdTs, &txHash)
		blockNumber := utils.BlockNumberFromPrefix(shcrypto.Block(identityPreimage[0:32]))
		address := common.BytesToAddress(identityPreimage[32:])
		if address.Hex() == cfg.submitAccount.Address.Hex() && blockNumber == blame.triggerBlock {
			blame.decryptionKey = DecryptionKey{
				createdTs:        &createdTs,
				txPointer:        txPointer,
				eon:              eon,
				identityPreimage: identityPreimage,
			}
			if len(txHash) < 32 {
				log.Println("received empty txhash", hex.EncodeToString(identityPreimage), txPointer, createdTs)
			} else {
				blame.decryptedTxHash = common.Hash(txHash)
			}
		}
	}
	// panic: runtime error: cannot convert slice with length 0 to array or pointer to array with length 32
	if rows.Err() != nil {
		log.Println("errors when finding validator to blame: ", rows.Err())
		return err
	}
	return nil
}

func queryStatusRatios(w *bufio.Writer, startBlock, endBlock uint64, cfg *Configuration) error {
	queryStatusRatios := `
	SELECT
        COUNT(*) AS known_tx, 
        SUM(CASE WHEN dt.tx_status='shielded inclusion' THEN 1.0 END)/COUNT(*) * 100 AS shielded_ratio,    
        SUM(CASE WHEN dt.tx_status='shielded inclusion' THEN 1 END) AS shielded_amount,    
        SUM(CASE WHEN dt.tx_status='unshielded inclusion' THEN 1.0 END)/COUNT(*) * 100 AS unshielded_ratio,
        SUM(CASE WHEN dt.tx_status='unshielded inclusion' THEN 1 END) AS unshielded_amount,
        SUM(CASE WHEN dt.tx_status='not included' THEN 1.0 END)/COUNT(*) * 100 AS not_included_ratio, 
        SUM(CASE WHEN dt.tx_status='not included' THEN 1 END) not_included_amount, 
        SUM(CASE WHEN dt.tx_status='pending' THEN 1.0 END)/COUNT(*) * 100 AS pending_ratio,
        SUM(CASE WHEN dt.tx_status='pending' THEN 1 END) AS pending_amount
        FROM decryption_key AS dk 
                LEFT JOIN decrypted_tx AS dt
                        ON dt.decryption_key_id=dk.id
                LEFT JOIN block AS b
                        ON b.slot=dt.slot 
	WHERE 
        SUBSTRING(
                ENCODE(dk.identity_preimage, 'hex'),  --- encode preimage as hex string
                65  --- match only sender suffix of identity_preimage
        ) = $1  --- address of tester account
	AND 
        b.block_number BETWEEN $2 AND $3;`
	connection := GetConnection(cfg)
	rows, err := connection.db.Query(context.Background(), queryStatusRatios, strings.ToLower(cfg.submitAccount.Address.Hex())[2:], startBlock, endBlock)
	if err != nil {
		return err
	}
	var count uint64
	var shielded, unshielded, notIncluded, pending float64
	var shieldedAmount, unshieldedAmount, notIncludedAmount, pendingAmount int64
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&count, &shielded, &shieldedAmount, &unshielded, &unshieldedAmount, &notIncluded, &notIncludedAmount, &pending, &pendingAmount)

		_, err = fmt.Fprintf(w,
			`%v tx found by observer
%3.2f%% shielded (%v/%v)
%3.2f%% unshielded (%v/%v)
%3.2f%% not included (%v/%v)
%3.2f%% still pending (%v/%v)
`,
			count,
			shielded, shieldedAmount, count,
			unshielded, unshieldedAmount, count,
			notIncluded, notIncludedAmount, count,
			pending, pendingAmount, count)
	}
	return nil
}

func CollectContinuousTestStats(startBlock uint64, endBlock uint64, cache *BlockCache, cfg *Configuration) error {
	failCnt := 0
	var failed []Submission
	var delays []float64
	success, err := collectSubmitIncomingTx(startBlock, endBlock, cache, cfg)
	if err != nil {
		return err
	}
	log.Printf("found %v successful.", len(success))
	successByTrigger := make(map[int64]Success)
	for i := range success {
		successByTrigger[success[i].trigger] = success[i]
	}

	submit, err := collectSequencerEvents(startBlock, endBlock, cfg)
	if err != nil {
		return err
	}
	log.Printf("found %v submissions.", len(submit))
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
	var blames []ValidatorBlame
	for _, f := range failed {
		blame, err := blameValidator(f, cfg)
		if err != nil {
			log.Println(err)
		}
		blames = append(blames, blame)
	}

	blameFile := path.Join(cfg.blameFolder, fmt.Sprint(time.Now().Unix())+".blame")
	log.Println("writing blame to ", blameFile)
	f, err := os.Create(blameFile)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	_, err = fmt.Fprintf(w, "found %v shutter test tx in block range[%v:%v] (%v triggers)\n", len(submit), startBlock, endBlock, len(triggers))
	if err != nil {
		return err
	}
	err = queryStatusRatios(w, startBlock, endBlock, cfg)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "shutterized blocks %3.2f%%\n", shutterizedPct)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "fails %v (%3.2f%%)\n", failCnt, failPct)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "missed triggers %v: %v\n", len(triggers)-len(submit), utils.Difference(triggers, submitTriggers))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "delay max %0.0f min %0.0f avg %3.2f median %3.2f\n", maxDelay, minDelay, avgDelay, medianDelay)
	if err != nil {
		return err
	}
	for _, blame := range blames {
		_, err = fmt.Fprintln(w, blame)
		if err != nil {
			return err
		}
	}
	w.Flush()
	return err
}
