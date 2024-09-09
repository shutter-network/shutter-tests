package stress

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	sequencerBindings "github.com/shutter-network/contracts/v2/bindings/sequencer"
	"github.com/shutter-network/shutter/shlib/shcrypto"
	"gotest.tools/assert"
)

func skipCI(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}
}

const KeyperSetChangeLookAhead = 2

func createSetup(fundNewAccount bool) (StressSetup, error) {
	setup := new(StressSetup)
	RpcUrl, err := ReadStringFromEnv("STRESS_TEST_RPC_URL")
	if err != nil {
		return *setup, err
	}
	client, err := ethclient.Dial(RpcUrl)
	if err != nil {
		return *setup, fmt.Errorf("could not create client %v", err)
	}

	setup.Client = client

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return *setup, fmt.Errorf("could not query chainId %v", err)
	}
	setup.ChainID = chainID

	signerForChain := types.LatestSignerForChainID(chainID)
	setup.SignerForChain = signerForChain

	submitKeyHex, err := ReadStringFromEnv("STRESS_TEST_PK")
	if err != nil {
		return *setup, err
	}
	submitPrivateKey, err := crypto.HexToECDSA(submitKeyHex)
	if err != nil {
		return *setup, err
	}

	submitAccount, err := AccountFromPrivateKey(submitPrivateKey, signerForChain)
	if err != nil {
		return *setup, err
	}
	setup.SubmitAccount = &submitAccount

	// TODO: allow multiple transacting accounts in StressEnvironment.TransactAccounts
	transactPrivateKey, err := crypto.GenerateKey()

	if err != nil {
		return *setup, err
	}
	transactAccount, err := AccountFromPrivateKey(transactPrivateKey, signerForChain)
	if err != nil {
		return *setup, err
	}

	setup.TransactAccount = &transactAccount
	err = StoreAccount(transactAccount)
	if err != nil {
		return *setup, err
	}
	if fundNewAccount {
		err = fund(*setup)
		if err != nil {
			return *setup, err
		}
		log.Println("Funding complete")
	}
	KeyperSetManagerContractAddress, err := ReadStringFromEnv("STRESS_TEST_KEYPER_SET_MANAGER_CONTRACT_ADDRESS")
	if err != nil {
		return *setup, err
	}

	KeyBroadcastContractAddress, err := ReadStringFromEnv("STRESS_TEST_KEY_BROADCAST_CONTRACT_ADDRESS")
	if err != nil {
		return *setup, err
	}

	SequencerContractAddress, err := ReadStringFromEnv("STRESS_TEST_SEQUENCER_CONTRACT_ADDRESS")
	if err != nil {
		return *setup, err
	}

	contracts, err := SetupContracts(client, KeyBroadcastContractAddress, SequencerContractAddress, KeyperSetManagerContractAddress)
	if err != nil {
		return *setup, err
	}
	setup.KeyBroadcastContract = *contracts.KeyBroadcastContract
	setup.KeyperSetManager = *contracts.KeyperSetManager
	setup.Sequencer = *contracts.Sequencer
	setup.SequencerContractAddress = contracts.SequencerContractAddress

	return *setup, nil
}

func fund(setup StressSetup) error {
	value := big.NewInt(100000000000000000) // 0.1 ETH in wei
	gasLimit := uint64(21000)
	gasPrice, err := setup.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}
	var data []byte
	nonce, err := setup.Client.NonceAt(context.Background(), setup.SubmitAccount.Address, nil)
	if err != nil {
		return err
	}
	log.Println("HeadNonce", nonce)
	tx := types.NewTransaction(nonce, setup.TransactAccount.Address, value, gasLimit, gasPrice, data)
	signedTx, err := setup.SubmitAccount.Sign(setup.SubmitAccount.Address, tx)
	if err != nil {
		return err
	}
	err = setup.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return err
	}
	log.Println("sent funding tx", signedTx.Hash().Hex(), "to", setup.TransactAccount.Address)
	_, err = bind.WaitMined(context.Background(), setup.Client, signedTx)
	return err
}

//lint:ignore U1000 Ignore unused function.
func increasingGasPriceFn(suggestedGasTipCap *big.Int, suggestedGasPrice *big.Int, i int, count int) (GasFeeCap, GasTipCap) {
	feeCapAndTipCap := big.NewInt(0).Add(suggestedGasPrice, suggestedGasTipCap)

	gasFloat, _ := suggestedGasPrice.Float64()
	x := int64(gasFloat * (2. / float64(count)) * float64(i+1)) // higher delta for higher nonces
	log.Println("delta is ", x)
	delta := big.NewInt(x)
	gasFeeCap := big.NewInt(0).Add(feeCapAndTipCap, delta)
	return gasFeeCap, suggestedGasTipCap
}

//lint:ignore U1000 Ignore unused function.
func decreasingGasPriceFn(suggestedGasTipCap *big.Int, suggestedGasPrice *big.Int, i int, count int) (GasFeeCap, GasTipCap) {
	feeCapAndTipCap := big.NewInt(0).Add(suggestedGasPrice, suggestedGasTipCap)

	gasFloat, _ := suggestedGasPrice.Float64()
	x := int64(gasFloat * (2. / float64(count)) * float64(count-i)) // lower delta for higher nonces to test cut off
	log.Println("delta is ", x)
	delta := big.NewInt(x)
	gasFeeCap := big.NewInt(0).Add(feeCapAndTipCap, delta)
	return gasFeeCap, suggestedGasTipCap
}

func defaultGasLimitFn(data []byte, toAddress *common.Address, i int, count int) uint64 {
	return uint64(21000)
}

func createStressEnvironment(ctx context.Context, setup StressSetup) (StressEnvironment, error) {
	eon, eonKey, err := getEonKey(ctx, setup)

	environment := StressEnvironment{
		TransacterOpts: bind.TransactOpts{
			From:   setup.TransactAccount.Address,
			Signer: setup.TransactAccount.Sign,
		},
		TransactGasPriceFn:   DefaultGasPriceFn,
		TransactGasLimitFn:   defaultGasLimitFn,
		InclusionWaitTimeout: time.Duration(time.Minute * 2),
		InclusionConstraints: func(inclusions []*types.Receipt) error { return nil },
		SubmitterOpts: bind.TransactOpts{
			From:   setup.SubmitAccount.Address,
			Signer: setup.SubmitAccount.Sign,
		},
		SubmissionWaitTimeout: time.Duration(time.Second * 30),
		Eon:                   eon,
		EonPublicKey:          eonKey,
		WaitOnEverySubmit:     false,
		RandomIdentitySuffix:  false,
	}
	if err != nil {
		return environment, fmt.Errorf("could not get eonKey %v", err)
	}
	submitterNonce, err := setup.Client.PendingNonceAt(context.Background(), setup.SubmitAccount.Address)
	log.Println("Current submitter nonce is", submitterNonce)
	if err != nil {
		return environment, fmt.Errorf("could not query starting nonce %v", err)
	}
	setup.SubmitAccount.Nonce = *big.NewInt(int64(submitterNonce))

	transactNonce, err := setup.Client.PendingNonceAt(context.Background(), setup.TransactAccount.Address)
	if err != nil {
		return environment, fmt.Errorf("could not query starting nonce %v", err)
	}
	setup.TransactAccount.Nonce = *big.NewInt(int64(transactNonce))

	log.Println("eon is ", eon)
	return environment, nil
}

func getEonKey(ctx context.Context, setup StressSetup) (uint64, *shcrypto.EonPublicKey, error) {
	return GetEonKey(ctx, setup.Client, &setup.KeyperSetManager, &setup.KeyBroadcastContract, KeyperSetChangeLookAhead)
}

func createIdentityPrefix() (shcrypto.Block, error) {
	identityPrefix, err := shcrypto.RandomSigma(cryptorand.Reader)
	if err != nil {
		return shcrypto.Block{}, fmt.Errorf("could not get random identityPrefix %v", err)
	}
	return identityPrefix, nil
}

func encrypt(ctx context.Context, tx types.Transaction, env *StressEnvironment, submitter common.Address, i int) (*shcrypto.EncryptedMessage, shcrypto.Block, error) {

	sigma, err := shcrypto.RandomSigma(cryptorand.Reader)
	if err != nil {
		return nil, shcrypto.Block{}, fmt.Errorf("could not get sigma bytes %s", err)
	}

	var identityPrefix shcrypto.Block
	if i < len(env.IdentityPrefixes) {
		identityPrefix = env.IdentityPrefixes[i]
	} else {
		identityPrefix, err = createIdentityPrefix()

		if err != nil {
			return nil, identityPrefix, err
		}
	}
	if env.RandomIdentitySuffix {
		submitter, err = createRandomAddress()
		if err != nil {
			return nil, identityPrefix, err
		}
	}

	identity := ComputeIdentity(identityPrefix[:], submitter)

	var buff bytes.Buffer
	err = tx.EncodeRLP(&buff)

	if err != nil {
		return nil, identityPrefix, fmt.Errorf("failed encode RLP %v", err)
	}
	j, err := tx.MarshalJSON()
	if err != nil {
		return nil, identityPrefix, fmt.Errorf("failed to marshal json %v", err)
	}
	log.Println("tx to be encrypted", string(j[:]))
	encryptedTx := shcrypto.Encrypt(buff.Bytes(), (*shcrypto.EonPublicKey)(env.EonPublicKey), identity, sigma)
	return encryptedTx, identityPrefix, nil
}

func submitEncryptedTx(ctx context.Context, setup StressSetup, env *StressEnvironment, tx types.Transaction, i int) (*types.Transaction, error) {

	opts := env.SubmitterOpts
	log.Println("submit nonce", opts.Nonce)

	opts.Value = big.NewInt(0).Sub(tx.Cost(), tx.Value())

	encryptedTx, identityPrefix, err := encrypt(ctx, tx, env, setup.SubmitAccount.Address, i)
	if err != nil {
		return nil, fmt.Errorf("could not encrypt %v", err)
	}

	submitTx, err := setup.Sequencer.SubmitEncryptedTransaction(&opts, env.Eon, identityPrefix, encryptedTx.Marshal(), new(big.Int).SetUint64(tx.Gas()))
	if err != nil {
		return nil, fmt.Errorf("Could not submit %s", err)
	}
	log.Println("submitted identityPrefix ", hex.EncodeToString(identityPrefix[:]))
	return submitTx, nil

}

func transact(setup *StressSetup, env *StressEnvironment, count int) error {

	value := big.NewInt(1) // in wei

	toAddress := setup.SubmitAccount.Address
	var data []byte
	var submissions []types.Transaction
	var innerTxs []types.Transaction

	suggestedGasTipCap, err := setup.Client.SuggestGasTipCap(context.Background())
	if err != nil {
		return err
	}
	suggestedGasPrice, err := setup.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	identityPrefixes := env.IdentityPrefixes
	for i := len(identityPrefixes); i < count; i++ {
		identity, err := createIdentityPrefix()
		if err != nil {
			return err
		}
		identityPrefixes = append(identityPrefixes, identity)
	}

	env.IdentityPrefixes = identityPrefixes

	for i := 0; i < count; i++ {
		gasFeeCap, suggestedGasTipCap := env.TransactGasPriceFn(suggestedGasTipCap, suggestedGasPrice, i, count)
		gasLimit := env.TransactGasLimitFn(data, &toAddress, i, count)
		innerNonceP := setup.TransactAccount.UseNonce()
		innerNonce := innerNonceP.Uint64()
		log.Printf("inner nonce: %v", innerNonce)
		tx := types.NewTx(
			&types.DynamicFeeTx{
				ChainID:   setup.ChainID,
				Nonce:     innerNonce,
				GasFeeCap: gasFeeCap,
				GasTipCap: suggestedGasTipCap,
				Gas:       gasLimit,
				To:        &toAddress,
				Value:     value,
				Data:      data,
			},
		)

		signedTx, err := setup.TransactAccount.Sign(setup.TransactAccount.Address, tx)
		if err != nil {
			return err
		}
		innerTxs = append(innerTxs, *signedTx)
		log.Println("used nonce", signedTx.Nonce())
	}
	for i := range innerTxs {
		signedTx := innerTxs[i]
		submitNonce := setup.SubmitAccount.UseNonce()
		env.SubmitterOpts.Nonce = &submitNonce
		submitTx, err := submitEncryptedTx(context.Background(), *setup, env, signedTx, i)
		if err != nil {
			return err
		}
		submissions = append(submissions, *submitTx)
		if env.WaitOnEverySubmit {
			_, err = waitForTx(*submitTx, "submission", env.SubmissionWaitTimeout, setup.Client)
			if err != nil {
				return err
			}
		}
		log.Println("Submit tx hash", submitTx.Hash().Hex(), "Encrypted tx hash", signedTx.Hash().Hex())
	}
	for _, submitTx := range submissions {
		_, err = waitForTx(submitTx, "submission", env.SubmissionWaitTimeout, setup.Client)
		if err != nil {
			return err
		}
	}
	var receipts []*types.Receipt
	for _, innerTx := range innerTxs {
		receipt, err := waitForTx(innerTx, "inclusion", env.InclusionWaitTimeout, setup.Client)
		if err != nil {
			return err
		}
		receipts = append(receipts, receipt)
	}
	err = env.InclusionConstraints(receipts)
	if err != nil {
		return err
	}
	err = countAndLog(receipts)
	return err
}

// send a single transaction
func TestStressSingle(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(true)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	env, err := createStressEnvironment(context.Background(), setup)
	if err != nil {
		log.Fatal("could not set up environment", err)
	}
	err = transact(&setup, &env, 1)
	assert.NilError(t, err, "not included")
}

// send two transactions but wait for each submission to the sequencer possible
func TestStressDualWait(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(true)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	env, err := createStressEnvironment(context.Background(), setup)
	if err != nil {
		log.Fatal("could not set up environment", err)
	}
	env.WaitOnEverySubmit = true

	err = transact(&setup, &env, 2)
	assert.NilError(t, err, "not included")
}

// send two transactions as quickly as possible
func TestStressDualNoWait(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(true)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	env, err := createStressEnvironment(context.Background(), setup)
	if err != nil {
		log.Fatal("could not set up environment", err)
	}

	err = transact(&setup, &env, 2)
	assert.NilError(t, err, "not included")
}

// send two transactions in the same block by the same sender with the same identityPrefix
func TestStressDualDuplicatePrefix(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(true)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	env, err := createStressEnvironment(context.Background(), setup)
	if err != nil {
		log.Fatal("could not set up environment", err)
	}
	prefix, err := createIdentityPrefix()
	if err != nil {
		log.Fatal("error creating prefix", err)
	}
	var prefixes []shcrypto.Block
	prefixes = append(prefixes, prefix)
	prefixes = append(prefixes, prefix)
	env.IdentityPrefixes = prefixes

	err = transact(&setup, &env, 2)
	assert.NilError(t, err, "not included")
}

// send many transactions as quickly as possible.
func TestStressManyNoWait(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(true)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	env, err := createStressEnvironment(context.Background(), setup)
	if err != nil {
		log.Fatal("could not set up environment", err)
	}

	err = transact(&setup, &env, 47)
	assert.NilError(t, err, "not included")
}

// test that tx using together more than ENCYRPTED_GAS_LIMIT end up in different blocks
func TestStressExceedEncryptedGasLimit(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(true)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	env, err := createStressEnvironment(context.Background(), setup)
	if err != nil {
		log.Fatal("could not set up environment", err)
	}

	env.TransactGasLimitFn = func(data []byte, toAddress *common.Address, i, count int) uint64 {
		// last consumes over the limit
		if count-i == 1 {
			return uint64(1_000_000 - (i * 21_000) + 1)
		}
		return uint64(21000)
	}
	env.InclusionConstraints = func(receipts []*types.Receipt) error {
		sort.Slice(receipts, func(a, b int) bool {
			return receipts[a].BlockNumber.Uint64() < receipts[b].BlockNumber.Uint64()
		})
		if receipts[0].BlockNumber.Uint64() == receipts[len(receipts)-1].BlockNumber.Uint64() {
			return fmt.Errorf("tx must not be all in the same block")
		}
		return nil
	}
	err = transact(&setup, &env, 2)
	assert.NilError(t, err, "failed")
}

// test nested shutter transactions
func TestInception(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(true)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	env, err := createStressEnvironment(context.Background(), setup)
	if err != nil {
		log.Fatal("could not set up environment", err)
	}
	ctx := context.Background()

	innerGasLimit := 21000
	if err != nil {
		log.Fatal(err)
	}
	price, err := setup.Client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal(err)
	}
	tip, err := setup.Client.SuggestGasTipCap(ctx)
	if err != nil {
		log.Fatal(err)
	}
	gasFeeCap, gasTipCap := env.TransactGasPriceFn(price, tip, 0, 1)
	var data []byte
	innerTx := types.NewTx(
		&types.DynamicFeeTx{
			ChainID:   setup.ChainID,
			Nonce:     1,
			GasFeeCap: gasFeeCap,
			GasTipCap: gasTipCap,
			Gas:       uint64(innerGasLimit),
			To:        &setup.SubmitAccount.Address,
			Value:     big.NewInt(1),
			Data:      data,
		},
	)

	// TODO: we could start a loop here
	signedInnerTx, err := setup.TransactAccount.Sign(setup.TransactAccount.Address, innerTx)
	if err != nil {
		log.Fatal(err)
	}

	encryptedInnerTx, innerIdentityPrefix, err := encrypt(ctx, *signedInnerTx, &env, setup.TransactAccount.Address, 1)
	if err != nil {
		log.Fatal(err)
	}

	abi, err := sequencerBindings.SequencerMetaData.GetAbi()
	if err != nil {
		log.Fatal(err)
	}

	input, err := abi.Pack("submitEncryptedTransaction", env.Eon, innerIdentityPrefix, encryptedInnerTx.Marshal(), big.NewInt(int64(signedInnerTx.Gas())))
	if err != nil {
		log.Fatal(err)
	}
	price, err = setup.Client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal(err)
	}
	tip, err = setup.Client.SuggestGasTipCap(ctx)
	if err != nil {
		log.Fatal(err)
	}
	gasFeeCap, gasTipCap = env.TransactGasPriceFn(price, tip, 0, 1)

	middleCallMsg := ethereum.CallMsg{
		From:  setup.TransactAccount.Address,
		To:    &setup.SequencerContractAddress,
		Value: signedInnerTx.Cost(),
		Data:  input,
	}
	middleGasLimit, err := setup.Client.EstimateGas(ctx, middleCallMsg)
	if err != nil {
		log.Fatal(err)
	}

	middleTx := types.NewTx(
		&types.DynamicFeeTx{
			ChainID:   setup.ChainID,
			Nonce:     0,
			GasFeeCap: gasFeeCap,
			GasTipCap: gasTipCap,
			Gas:       middleGasLimit,
			To:        &setup.SequencerContractAddress,
			Value:     big.NewInt(0).Sub(signedInnerTx.Cost(), signedInnerTx.Value()),
			Data:      input,
		},
	)

	signedMiddleTx, err := setup.TransactAccount.Sign(setup.TransactAccount.Address, middleTx)
	if err != nil {
		log.Fatal(err)
	}

	middleTxEncryptedMsg, middleIdentityPrefix, err := encrypt(ctx, *signedMiddleTx, &env, setup.SubmitAccount.Address, 1)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("submitting outer. Gas", signedMiddleTx.Gas())
	opts := &env.SubmitterOpts
	log.Println("submit nonce", opts.Nonce)

	// and end the loop here?
	opts.Value = big.NewInt(0).Sub(signedMiddleTx.Cost(), signedMiddleTx.Value())
	submitTx, err := setup.Sequencer.SubmitEncryptedTransaction(opts, env.Eon, middleIdentityPrefix, middleTxEncryptedMsg.Marshal(), big.NewInt(int64(signedMiddleTx.Gas())))
	if err != nil {
		log.Fatal(err)
	}
	submitReceipt, err := waitForTx(*submitTx, "outer tx", env.SubmissionWaitTimeout, setup.Client)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(hex.EncodeToString(middleIdentityPrefix[:]))
	middleReceipt, err := waitForTx(*signedMiddleTx, "middle tx", env.InclusionWaitTimeout, setup.Client)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(hex.EncodeToString(innerIdentityPrefix[:]))
	innerReceipt, err := waitForTx(*signedInnerTx, "inner tx", env.InclusionWaitTimeout, setup.Client)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("inner gas", innerReceipt.GasUsed, "middle gas", middleReceipt.GasUsed, "outer gas", submitReceipt.GasUsed)
}

func TestIncorrectIdentitySuffix(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(true)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	env, err := createStressEnvironment(context.Background(), setup)
	if err != nil {
		log.Fatal("could not set up environment", err)
	}
	env.RandomIdentitySuffix = true

	err = transact(&setup, &env, 1)
	assert.Error(t, err, "error on WaitMined context deadline exceeded", "this must time out")
}

// not really a test, but useful to collect from previously funded test accounts
func TestEmptyAccounts(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(false)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	fd, err := os.Open("pk.hex")
	if err != nil {
		log.Fatal("Could not open pk.hex")
	}
	defer fd.Close()
	pks, err := ReadPks(fd)
	if err != nil {
		log.Fatal("error when reading private keys", err)
	}
	block, err := setup.Client.BlockNumber(context.Background())
	if err != nil {
		log.Fatal("could not query block number", err)
	}
	for i := range pks {
		account, err := AccountFromPrivateKey(pks[i], setup.SignerForChain)
		if err != nil {
			log.Fatal("could not create account from privatekey", err, pks[i])
		}
		balance, err := setup.Client.BalanceAt(context.Background(), account.Address, big.NewInt(int64(block)))
		if err == nil {
			log.Println(account.Address.Hex(), balance)
		}
		if balance.Uint64() > 0 {
			drain(context.Background(), account, balance.Uint64(), setup.SubmitAccount.Address, setup.Client)
		}
	}
}

// not really a test, but useful to fix the submit account's nonce, if an earlier test failed
func TestFixNonce(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(false)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	err = fixNonce(setup.Client, *setup.SubmitAccount)
	if err != nil {
		log.Fatal(err)
	}
}
