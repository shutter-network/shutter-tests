package stress

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	keybroadcastcontract "github.com/shutter-network/contracts/v2/bindings/keybroadcastcontract"
	keypersetmanager "github.com/shutter-network/contracts/v2/bindings/keypersetmanager"
	sequencerBindings "github.com/shutter-network/contracts/v2/bindings/sequencer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/shutter/shlib/shcrypto"
	"golang.org/x/exp/maps"
)

type Account struct {
	Address    common.Address
	privateKey *ecdsa.PrivateKey
	Sign       bind.SignerFn
	Nonce      big.Int
}

// Returns the currently stored account nonce and increases the stored value
func (acc Account) UseNonce() big.Int {
	one := big.NewInt(1)
	result := acc.Nonce
	acc.Nonce.Add(&result, one)
	return result
}

// contains all the setup required to interact with the chain
type StressSetup struct {
	Client                   *ethclient.Client
	SignerForChain           types.Signer
	ChainID                  *big.Int
	SubmitAccount            *Account
	TransactAccount          *Account
	Sequencer                sequencerBindings.Sequencer
	SequencerContractAddress common.Address
	KeyperSetManager         keypersetmanager.Keypersetmanager
	KeyBroadcastContract     keybroadcastcontract.Keybroadcastcontract
}

// contains the context for the current stress test to create transactions
type StressEnvironment struct {
	TransacterOpts        bind.TransactOpts
	TransactGasPriceFn    GasPriceFn
	TransactGasLimitFn    GasLimitFn
	InclusionWaitTimeout  time.Duration
	InclusionConstraints  ConstraintFn
	SubmitterOpts         bind.TransactOpts
	SubmissionWaitTimeout time.Duration
	Eon                   uint64
	EonPublicKey          *shcrypto.EonPublicKey
	WaitOnEverySubmit     bool
	IdentityPrefixes      []shcrypto.Block
	RandomIdentitySuffix  bool
}

type GasFeeCap *big.Int
type GasTipCap *big.Int

type GasLimitFn func(data []byte, toAddress *common.Address, i int, count int) uint64

type GasPriceFn func(suggestedGasTipCap *big.Int, suggestedGasPrice *big.Int, i int, count int) (GasFeeCap, GasTipCap)

func DefaultGasPriceFn(suggestedGasTipCap *big.Int, suggestedGasPrice *big.Int, i int, count int) (GasFeeCap, GasTipCap) {
	feeCapAndTipCap := big.NewInt(0).Add(suggestedGasPrice, suggestedGasTipCap)

	gasFloat, _ := suggestedGasPrice.Float64()
	x := int64(gasFloat * 1.5) // fixed delta
	log.Println("delta is ", x)
	delta := big.NewInt(x)
	gasFeeCap := big.NewInt(0).Add(feeCapAndTipCap, delta)
	return gasFeeCap, suggestedGasTipCap
}

type ConstraintFn func(inclusions []*types.Receipt) error

func waitForTx(tx types.Transaction, description string, timeout time.Duration, client *ethclient.Client) (*types.Receipt, error) {
	log.Println("waiting for "+description+" ", tx.Hash().Hex())
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	receipt, err := bind.WaitMined(ctx, client, &tx)
	if err != nil {
		return nil, fmt.Errorf("error on WaitMined %s", err)
	}
	log.Println("status", receipt.Status, "block", receipt.BlockNumber)
	if receipt.Status != 1 {
		return nil, fmt.Errorf("included tx failed")
	}
	return receipt, nil
}

func ReadStringFromEnv(envName string) (string, error) {
	value := os.Getenv(envName)
	if len(value) < 2 {
		return "", fmt.Errorf("could not read %v from environment. See README for details!", envName)
	}
	return value, nil
}

func AccountFromPrivateKey(privateKey *ecdsa.PrivateKey, signerForChain types.Signer) (Account, error) {
	account := Account{privateKey: privateKey}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return account, errors.New("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	account.Address = fromAddress
	account.Sign = func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		if address != fromAddress {
			return nil, errors.New("not authorized")
		}
		signature, err := crypto.Sign(signerForChain.Hash(tx).Bytes(), privateKey)
		if err != nil {
			return nil, err
		}
		return tx.WithSignature(signerForChain, signature)
	}
	account.Nonce = *big.NewInt(0)
	return account, nil
}

func StoreAccount(account Account) error {
	// we're going to store the privatekey of the secondary address in a file 'pk.hex'
	// this will allow us to recover funds, in case the clean up step fails
	transactPrivateKeyBytes := crypto.FromECDSA(account.privateKey)
	f, err := os.OpenFile("pk.hex", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := hex.NewEncoder(f)
	_, err = encoder.Write(transactPrivateKeyBytes)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(" "))
	if err != nil {
		return err
	}
	_, err = encoder.Write(account.Address.Bytes())
	if err != nil {
		return err
	}
	_, err = f.Write([]byte("\n"))
	return err
}

func countAndLog(receipts []*types.Receipt) error {
	c := map[string]uint16{}
	g := map[string]uint64{}
	for _, receipt := range receipts {
		n := receipt.BlockNumber.Text(10)
		c[n]++
		g[n] += receipt.GasUsed
	}
	log.Println("block\ttxs\tgas used")
	keys := maps.Keys(c)
	sort.Strings(keys)
	for _, n := range keys {
		log.Println(n, "\t", c[n], "\t", g[n])
	}
	return nil
}

func ReadPks(r io.Reader) ([]*ecdsa.PrivateKey, error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords)
	var result []*ecdsa.PrivateKey
	for scanner.Scan() {
		x := scanner.Text()
		if len(x) == 64 {
			pk, err := crypto.HexToECDSA(x)
			if err != nil {
				return result, err
			}
			result = append(result, pk)
		}
	}
	return result, scanner.Err()
}

func fixNonce(client *ethclient.Client, account Account) error {
	value := big.NewInt(1) // 1 wei
	gasLimit := uint64(21000)

	var data []byte
	headNonce, err := client.NonceAt(context.Background(), account.Address, nil)
	if err != nil {
		return err
	}
	log.Println("HeadNonce", headNonce)

	pendingNonce, err := client.PendingNonceAt(context.Background(), account.Address)
	if err != nil {
		return err
	}
	log.Println("PendingNonce", pendingNonce)
	var txs []types.Transaction
	for i := uint64(0); i < pendingNonce-headNonce; i++ {
		headNonce, err := client.NonceAt(context.Background(), account.Address, nil)
		if err != nil {
			return err
		}
		log.Println("HeadNonce", headNonce, "Pending", pendingNonce, "current", headNonce+i, "i", i)

		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			return err
		}
		gasPrice = gasPrice.Add(gasPrice, gasPrice)
		tx := types.NewTransaction(headNonce+i, account.Address, value, gasLimit, gasPrice, data)
		signedTx, err := account.Sign(account.Address, tx)
		if err != nil {
			return err
		}
		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			log.Println("error on send", err)
		}
		log.Println("sent nonce fix tx", signedTx.Hash().Hex(), "to", account.Address)
		txs = append(txs, *signedTx)
	}

	log.Println("waiting for tx")
	for _, signedTx := range txs {
		_, err = bind.WaitMined(context.Background(), client, &signedTx)
		if err != nil {
			log.Println("error on wait", err)
		}
		headNonce, err := client.NonceAt(context.Background(), account.Address, nil)
		if err != nil {
			return err
		}
		log.Println("HeadNonce", headNonce, "Pending", pendingNonce)
	}
	return err
}

func drain(ctx context.Context, account Account, balance uint64, target common.Address, client *ethclient.Client) {
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Println("could not query gasPrice")
	}
	gasLimit := uint64(21000)
	remaining := balance - gasLimit*gasPrice.Uint64()
	data := make([]byte, 0)

	nonce, err := client.PendingNonceAt(ctx, account.Address)
	if err != nil {
		log.Println("could not query nonce", err)
	}
	tx := types.NewTransaction(nonce, target, big.NewInt(int64(remaining)), gasLimit, gasPrice, data)

	signed, err := account.Sign(account.Address, tx)
	if err != nil {
		log.Println("could not sign transaction", err)
	}
	err = client.SendTransaction(ctx, signed)
	if err != nil {
		log.Println("failed to send", err)
	}
	receipt, err := bind.WaitMined(ctx, client, signed)
	if err != nil {
		log.Println("failed to wait for tx", err)
	}
	log.Println("status", receipt.Status)
}

func createRandomAddress() (common.Address, error) {
	var address common.Address
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return address, fmt.Errorf("could not generate random key %v", err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return address, fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	address = crypto.PubkeyToAddress(*publicKeyECDSA)
	return address, nil
}

func ComputeIdentity(prefix []byte, sender common.Address) *shcrypto.EpochID {
	imageBytes := append(prefix, sender.Bytes()...)
	return shcrypto.ComputeEpochID(identitypreimage.IdentityPreimage(imageBytes).Bytes())
}

type Contracts struct {
	KeyperSetManager         *keypersetmanager.Keypersetmanager
	KeyBroadcastContract     *keybroadcastcontract.Keybroadcastcontract
	SequencerContractAddress common.Address
	Sequencer                *sequencerBindings.Sequencer
}

func SetupContracts(client *ethclient.Client, KeyBroadcastContractAddress, SequencerContractAddress, KeyperSetManagerContractAddress string) (Contracts, error) {
	var setup Contracts
	keyperSetManagerContract, err := keypersetmanager.NewKeypersetmanager(common.HexToAddress(KeyperSetManagerContractAddress), client)
	if err != nil {
		return setup, fmt.Errorf("can not get KeyperSetManager %v", err)
	}
	setup.KeyperSetManager = keyperSetManagerContract

	keyBroadcastContract, err := keybroadcastcontract.NewKeybroadcastcontract(common.HexToAddress(KeyBroadcastContractAddress), client)
	if err != nil {
		return setup, fmt.Errorf("can not get KeyBrodcastContract %v", err)
	}

	setup.KeyBroadcastContract = keyBroadcastContract

	setup.SequencerContractAddress = common.HexToAddress(SequencerContractAddress)
	sequencerContract, err := sequencerBindings.NewSequencer(common.HexToAddress(SequencerContractAddress), client)
	if err != nil {
		return setup, fmt.Errorf("can not get SequencerContract %v", err)
	}

	setup.Sequencer = sequencerContract
	blockNumber, err := client.BlockNumber(context.Background())
	if err != nil {
		panic(err)
	}
	eon, err := keyperSetManagerContract.GetKeyperSetIndexByBlock(nil, blockNumber+uint64(2))
	if err != nil {
		log.Fatal(err, ": ", KeyperSetManagerContractAddress)
	}
	log.Println(eon)
	return setup, nil
}

func GetEonKey(
	ctx context.Context,
	client *ethclient.Client,
	keyperSetManager *keypersetmanager.Keypersetmanager,
	keyBroadcastContract *keybroadcastcontract.Keybroadcastcontract,
	KeyperSetChangeLookAhead int) (uint64, *shcrypto.EonPublicKey, error) {
	blockNumber, err := client.BlockNumber(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("could not query blockNumber %v", err)
	}

	// could not get eon no contract code at given address
	eon, err := keyperSetManager.GetKeyperSetIndexByBlock(nil, blockNumber+uint64(KeyperSetChangeLookAhead))
	if err != nil {
		return 0, nil, fmt.Errorf("could not get eon %v", err)
	}

	eonKeyBytes, err := keyBroadcastContract.GetEonKey(nil, eon)
	if err != nil {
		return 0, nil, fmt.Errorf("could not get eonKeyBytes %v", err)
	}

	eonKey := &shcrypto.EonPublicKey{}
	if err := eonKey.Unmarshal(eonKeyBytes); err != nil {
		return 0, nil, fmt.Errorf("could not unmarshal eonKeyBytes %v", err)
	}
	return eon, eonKey, nil
}
