package continuous

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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
			if tx.To().Hex() == cfg.submitAccount.Address.Hex() {
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
