package continuous

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgtype"
)

const KeyperSetChangeLookAhead = 2
const NumFundedAccounts = 6
const MinimalFunding = int64(500000000000000000) // 0.5 ETH in wei

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

type ShutterBlock struct {
	Number int64
	Ts     pgtype.Date
}

func QueryAllShutterBlocks(out chan<- ShutterBlock, cfg *Configuration) {
	waitBetweenQueries := 1 * time.Second
	status := Status{lastShutterTS: pgtype.Date{}}
	connection := NewConnection(cfg)
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
		newShutterBlock := queryNewestShutterBlock(status.lastShutterTS, cfg)
		if !newShutterBlock.Ts.Time.IsZero() {
			status.lastShutterTS = newShutterBlock.Ts
			// send event (block number, timestamp) to out channel
			out <- newShutterBlock
		}
	}
}

func queryNewestShutterBlock(lastBlockTS pgtype.Date, cfg *Configuration) ShutterBlock {
	connection := NewConnection(cfg)
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
	rows, err := connection.db.Query(context.Background(), query, lastBlockTS.Time.Unix())
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
			// cancel signal: another included tx with inclusion block > submission block
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
