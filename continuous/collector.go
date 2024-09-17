package continuous

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/jackc/pgtype"
	"github.com/shutter-network/contracts/v2/bindings/sequencer"
	"github.com/shutter-network/nethermind-tests/utils"
)

type Submission struct {
	trigger   int64
	sequenced int64
}

func collectSequencerEvents(startBlock uint64, endBlock uint64, cfg *Configuration) ([]Submission, error) {
	var evs []sequencer.SequencerTransactionSubmitted
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
			evs = append(evs, ev)
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

func collectSubmitIncomingTx(startBlock uint64, endBlock uint64, cfg *Configuration) ([]Success, error) {
	var result []Success
	for blockNum := startBlock; blockNum <= endBlock; blockNum++ {
		num := big.NewInt(int64(blockNum))
		block, err := cfg.client.BlockByNumber(context.Background(), num)
		if err != nil {
			return result, nil
		}
		txs := block.Transactions()
		for _, tx := range txs {
			if tx.To() != nil && tx.To().Hex() == cfg.submitAccount.Address.Hex() {
				success := Success{tx.Value().Int64(), block.Number().Int64()}
				result = append(result, success)
			}
		}

	}
	return result, nil
}
func queryBlockTriggers(startBlock uint64, endBlock uint64, cfg *Configuration) ([]int64, error) {
	var blocks []int64
	var block int64
	query := `
		SELECT b.block_number 
		FROM block AS b 
			LEFT JOIN proposer_duties AS p 
			ON p.slot = b.slot 
			LEFT JOIN validator_status AS s 
			ON p.validator_index = s.validator_index 
		WHERE b.block_number >= $1 
		AND b.block_number <= $2 
		AND s.status = 'active_ongoing';
		`
	connection := NewConnection()
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

func queryWhoToBlame(blame *ValidatorBlame) error {
	var targetBlock, targetSlot, validatorIndex int64

	queryWhoToBlame := `
	SELECT b.block_number, b.slot, v.validator_index
	FROM block AS b 
		LEFT JOIN proposer_duties AS p ON p.slot = b.slot 
		LEFT JOIN validator_status AS v ON v.validator_index = p.validator_index 
	WHERE v.status = 'active_ongoing' 
	AND b.block_number > $1 
	ORDER BY b.block_number ASC 
	LIMIT 1;`
	connection := NewConnection()
	rows, err := connection.db.Query(context.Background(), queryWhoToBlame, blame.submitBlock)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&targetBlock, &targetSlot, &validatorIndex)
		blame.targetBlock = targetBlock
		blame.targetSlot = targetSlot
		blame.validatorIndex = validatorIndex
	}
	if rows.Err() != nil {
		log.Println("errors when finding validator to blame: ", rows.Err())
		return err
	}
	return nil
}

type DecryptionKey struct {
	identityPreimage []byte
	txPointer        int
	eon              int
	createdTs        pgtype.Date
}

func queryDecryptionKeysBySlot(blame ValidatorBlame) error {
	var identityPreimage []byte
	var txPointer, eon int
	var createdTs pgtype.Date

	queryDecryptionKeysBySlot := `
	SELECT k.identity_preimage, d.tx_pointer, d.eon, d.created_at 
	FROM decryption_key AS k 
		LEFT JOIN decryption_keys_message AS d 
		ON d.created_at=k.created_at 
	WHERE d.slot=$1;`
	connection := NewConnection()
	rows, err := connection.db.Query(context.Background(), queryDecryptionKeysBySlot, blame.targetSlot)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&identityPreimage, &txPointer, &eon, &createdTs)
		log.Println(identityPreimage, txPointer, eon, createdTs)
	}
	if rows.Err() != nil {
		log.Println("errors when finding validator to blame: ", rows.Err())
		return err
	}
	return nil
}
