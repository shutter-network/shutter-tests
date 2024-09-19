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
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgtype"
	"github.com/montanaflynn/stats"
	"github.com/shutter-network/nethermind-tests/utils"
	"github.com/shutter-network/shutter/shlib/shcrypto"
)

type ValidatorBlame struct {
	prefix          []byte
	triggerBlock    int64
	submitBlock     int64
	targetBlock     int64
	targetBlockTS   *pgtype.Date
	targetSlot      int64
	decryptedTxHash common.Hash
	validatorIndex  int64
	decryptionKey   DecryptionKey
}

func (b ValidatorBlame) String() string {
	emptyHash := common.Hash(make([]byte, common.HashLength))
	if b.decryptedTxHash == emptyHash {
		return fmt.Sprintf(
			"validator id\t: %v\n"+
				"triggered\t: %v\n"+
				"submitted\t: %v\n"+
				"target block\t: %v\n"+
				"target slot\t: %v\n"+
				"target ts\t: %v\n"+
				"NO DECRYPTION KEY SEEN\n",
			b.validatorIndex,
			b.triggerBlock,
			b.submitBlock,
			b.targetBlock,
			b.targetSlot,
			b.targetBlockTS.Time.UTC().Format("2006-01-01 15:04:05.000000"),
		)
	} else {
		return fmt.Sprintf(
			"validator id\t: %v\n"+
				"triggered\t: %v\n"+
				"submitted\t: %v\n"+
				"target block\t: %v\n"+
				"target slot\t: %v\n"+
				"target ts\t: %v\n"+
				"ts (key-target)\t: %vms\n"+
				"decrypted tx\t: %v\n"+
				"decryption key:\n%v\n",
			b.validatorIndex,
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
			if submission.trigger >= int64(startBlock) {
				submissions = append(submissions, submission)
			} else {
				log.Printf("ignoring submission with prefix block %v", submission.trigger)
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

type BlockCache map[uint64][]Success

func collectSubmitIncomingTx(startBlock uint64, endBlock uint64, cache *BlockCache, cfg *Configuration) ([]Success, error) {
	var result []Success
	for blockNum := startBlock; blockNum <= endBlock; blockNum++ {
		if found, ok := (*cache)[blockNum]; ok {
			result = append(result, found...)
		} else {
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
			(*cache)[blockNum] = successForBlock
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
		to_timestamp(b.block_timestamp)
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
	for rows.Next() {
		rows.Scan(&targetBlock, &targetSlot, &validatorIndex, &targetTS)
		blame.targetBlock = targetBlock
		blame.targetSlot = targetSlot
		blame.targetBlockTS = targetTS
		blame.validatorIndex = validatorIndex
	}
	if rows.Err() != nil {
		log.Println("errors when finding validator to blame: ", rows.Err())
		return err
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
	}
	err := queryWhoToBlame(&blame, cfg)
	if err != nil {
		return blame, nil
	}
	err = queryDecryptionKeysBySlot(&blame, cfg)
	if err != nil {
		return blame, nil
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
			blame.decryptedTxHash = common.Hash(txHash)
		}
	}
	if rows.Err() != nil {
		log.Println("errors when finding validator to blame: ", rows.Err())
		return err
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
