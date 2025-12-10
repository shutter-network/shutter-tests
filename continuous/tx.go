package continuous

import (
	"context"
	cryptorand "crypto/rand"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgtype"
	"github.com/shutter-network/nethermind-tests/utils"
	"github.com/shutter-network/shutter/shlib/shcrypto"
)

type ShutterTx struct {
	innerTx         *types.Transaction
	outerTx         *types.Transaction
	sender          *utils.Account
	prefix          shcrypto.Block
	triggerBlock    int64
	submissionBlock int64
	inclusionBlock  int64
	cancelBlock     int64
	targetSlot      int64
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
	return fmt.Sprintf(
		"ShutterTx[%v]\t%v\n"+
			"\ttrigger:\t%8d\n"+
			"\tsubmit :\t%8d\t%v n:%v\n"+
			"\tinclude:\t%8d\t%v n:%v\n"+
			"\tcancel :\t%8d",
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
func SendShutterizedTX(blockNumber int64, lastTimestamp pgtype.Date, targetSlot int64, cfg *Configuration) {
	account := cfg.NextAccount()
	log.Printf("SENDING NEW TX FOR %v from %v", blockNumber, account.Address.Hex())
	gasLimit := uint64(21000)
	var data []byte
	gas, err := utils.GasCalculationFromClient(context.Background(), cfg.client, utils.Min1GweiGasPriceFn)
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*20)

	tx := ShutterTx{
		outerTx:      outerTx,
		innerTx:      signedInnerTx,
		sender:       account,
		prefix:       identityPrefix,
		triggerBlock: blockNumber,
		targetSlot:   targetSlot,
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
	submissionReceipt, err := utils.WaitForTxSubscribe(tx.ctx, *tx.outerTx, fmt.Sprintf("submission[%v]", tx.triggerBlock), client)
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
	includedReceipt, err := utils.WaitForTxSubscribe(tx.ctx, *tx.innerTx, fmt.Sprintf("inclusion[%v]", tx.triggerBlock), client)
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
	receipt, err := utils.WaitForTxSubscribe(
		context.Background(),
		*signed,
		fmt.Sprintf("forfeit nonce[%v] for %v", nonce, account.Address.Hex()),
		client,
	)
	if err != nil {
		return err
	}
	if receipt.Status != 1 {
		return fmt.Errorf("forfeit tx not accepted")
	}
	return err
}
