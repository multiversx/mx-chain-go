package interceptors

import (
	"crypto"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/hashing/keccak"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519/singlesig"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/dataValidators"
	"github.com/multiversx/mx-chain-go/process/interceptors/processor"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/state"
)

type participant struct {
	secretKey      crypto.PrivateKey
	secretKeyBytes []byte
	pubKey         []byte
	address        string
}

var (
	addressEncoder, _  = pubkeyConverter.NewBech32PubkeyConverter(32, "erd")
	signingMarshalizer = &marshal.JsonMarshalizer{}
	txSignHasher       = keccak.NewKeccak()
	signer             = &singlesig.Ed25519Signer{}
	signingCryptoSuite = ed25519.NewEd25519()
	contentMarshalizer = &marshal.GogoProtoMarshalizer{}
	contentHasher      = blake2b.NewBlake2b()
)

type MultiDataInterceptorExtension struct {
	isApplicable bool

	sponsor      *participant
	participants []*participant

	txInterceptorProcessor *processor.TxInterceptorProcessor
	txValidator            *dataValidators.TxValidator
	accountsHandler        state.AccountsAdapter
	pool                   process.ShardedPool
}

func NewMultiDataInterceptorExtension(mdi *MultiDataInterceptor) *MultiDataInterceptorExtension {
	ext := &MultiDataInterceptorExtension{
		sponsor:      &participant{},
		participants: []*participant{},
	}
	initExtension(ext, mdi)
	return ext
}

func initExtension(ext *MultiDataInterceptorExtension, mdi *MultiDataInterceptor) {
	txInterceptorProcessor, ok := mdi.processor.(*processor.TxInterceptorProcessor)
	if !ok {
		return
	}

	validator := txInterceptorProcessor.GetValidator()
	txValidator, ok := validator.(*dataValidators.TxValidator)
	if !ok {
		return
	}

	accountsHandler := txValidator.GetAccounts()
	shardedTxPool := txInterceptorProcessor.GetShardedData()

	log.Info("MultiDataInterceptorExtension.initExtension (applicable)")

	ext.txInterceptorProcessor = txInterceptorProcessor
	ext.txValidator = txValidator
	ext.accountsHandler = accountsHandler
	ext.pool = shardedTxPool
	ext.isApplicable = true
}

func (ext *MultiDataInterceptorExtension) isRecognizedTransaction(interceptedData process.InterceptedData) bool {
	interceptedTx, ok := interceptedData.(process.InterceptedTransactionHandler)
	if !ok {
		return false
	}

	tx := interceptedTx.Transaction()
	txData := string(tx.GetData())
	isRecognized := strings.HasPrefix(txData, "ext_")

	log.Info("MultiDataInterceptorExtension.isRecognizedTransaction?", "txData", txData, "isRecognized", isRecognized)

	return isRecognized
}

func (ext *MultiDataInterceptorExtension) doProcess(interceptedData process.InterceptedData) {
	log.Info("MultiDataInterceptorExtension.doProcess")

	tx := interceptedData.(process.InterceptedTransactionHandler).Transaction()
	function, args, err := parseCall(string(tx.GetData()))
	if err != nil {
		log.Error("failed to parse call", "error", err)
		return
	}

	shouldInit := function == "ext_init"
	shouldGenerateMoveBalances := function == "ext_generate_move_balances"

	if shouldInit {
		err := ext.doStepInit(tx.GetSndAddr(), args)
		if err != nil {
			log.Error("MultiDataInterceptorExtension: failed to do step: init", "error", err)
		}

		return
	}

	if shouldGenerateMoveBalances {
		err := ext.doStepGenerateMoveBalances(args)
		if err != nil {
			log.Error("MultiDataInterceptorExtension: failed to do step: generate move balances", "error", err)
		}

		return
	}

	log.Error("MultiDataInterceptorExtension: unrecognized function", "function", function)
}

func (ext *MultiDataInterceptorExtension) doStepInit(sponsorPubKey []byte, args [][]byte) error {
	if len(args) != 1 {
		return fmt.Errorf("MultiDataInterceptorExtension.doStepInit: invalid number of arguments")
	}

	log.Info("MultiDataInterceptorExtension.doStepInit", "sponsorPubKey", sponsorPubKey, "args", args)

	numParticipantsBytes := args[0]
	numParticipants := int(big.NewInt(0).SetBytes(numParticipantsBytes).Int64())

	ext.sponsor.pubKey = sponsorPubKey
	ext.sponsor.address, _ = addressEncoder.Encode(sponsorPubKey)

	ext.participants = createParticipants(numParticipants)
	mintingTransactions := ext.createMintingTransactions()
	ext.addTransactionsInTool(mintingTransactions)

	return nil
}

func parseCall(txData string) (string, [][]byte, error) {
	parser := smartContract.NewArgumentParser()
	function, args, err := parser.ParseCallData(txData)
	if err != nil {
		return "", nil, err
	}

	return function, args, nil
}

func createParticipants(numParticipants int) []*participant {
	log.Info("MultiDataInterceptorExtension.createParticipants", "numParticipants", numParticipants)

	keyGenerator := signing.NewKeyGenerator(signingCryptoSuite)

	participants := []*participant{}

	for i := 0; i < numParticipants; i++ {
		seed := make([]byte, 32)
		seed[0] = byte(i)
		secretKey, err := keyGenerator.PrivateKeyFromByteArray(seed)
		if err != nil {
			panic(err)
		}

		secretKeyBytes, err := secretKey.ToByteArray()
		if err != nil {
			panic(err)
		}

		pubKey := secretKey.GeneratePublic()
		pubKeyBytes, err := pubKey.ToByteArray()
		if err != nil {
			panic(err)
		}

		address, err := addressEncoder.Encode(pubKeyBytes)
		if err != nil {
			panic(err)
		}

		participants = append(participants, &participant{
			secretKey:      secretKey,
			secretKeyBytes: secretKeyBytes,
			pubKey:         pubKeyBytes,
			address:        address,
		})
	}

	return participants
}

func (ext *MultiDataInterceptorExtension) createMintingTransactions() []*transaction.Transaction {
	log.Info("MultiDataInterceptorExtension.createMintingTransactions")

	txs := []*transaction.Transaction{}

	sponsorAccount, err := ext.accountsHandler.GetExistingAccount(ext.sponsor.pubKey)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(ext.participants); i++ {
		tx := &transaction.Transaction{
			Nonce:    sponsorAccount.GetNonce() + uint64(i),
			Value:    big.NewInt(1000000000000000000),
			RcvAddr:  ext.participants[i].pubKey,
			SndAddr:  ext.sponsor.pubKey,
			GasPrice: 1000000000,
			GasLimit: 50000,
			Version:  1,
		}

		txs = append(txs, tx)
	}

	return txs
}

func (ext *MultiDataInterceptorExtension) addTransactionsInTool(transactions []*transaction.Transaction) {
	log.Info("MultiDataInterceptorExtension.addTransactionsInTool", "numTransactions", len(transactions))

	sw := core.NewStopWatch()
	sw.Start("default")

	cacherIdentifier := process.ShardCacherIdentifier(0, 0)

	for _, tx := range transactions {
		data, _ := contentMarshalizer.Marshal(tx)
		txHash := contentHasher.Compute(string(data))

		ext.pool.AddData(
			txHash,
			tx,
			350,
			cacherIdentifier,
		)
	}

	sw.Stop("default")

	measurement := sw.GetMeasurement("default")
	log.Info("MultiDataInterceptorExtension.addTransactionsInTool", "num txs", len(transactions), "duration", measurement.Milliseconds())
}

func (ext *MultiDataInterceptorExtension) doStepGenerateMoveBalances(args [][]byte) error {
	if len(args) != 1 {
		return fmt.Errorf("doStepGenerateMoveBalances: invalid number of arguments")
	}

	numTxsBytes := args[0]
	numTxs := int(big.NewInt(0).SetBytes(numTxsBytes).Int64())

	log.Info("MultiDataInterceptorExtension.doStepGenerateMoveBalances", "numTxs", numTxs)

	wg := sync.WaitGroup{}

	for i := 0; i < len(ext.participants); i++ {
		wg.Add(1)

		go func(i int) {
			transfers := ext.createTransfersToSelf(i, numTxs)
			ext.addTransactionsInTool(transfers)
			wg.Done()
		}(i)
	}

	wg.Wait()

	return nil
}

func (ext *MultiDataInterceptorExtension) createTransfersToSelf(participantIndex int, numTxs int) []*transaction.Transaction {
	txs := []*transaction.Transaction{}

	participant := ext.participants[participantIndex]
	nonce := uint64(0)

	account, err := ext.accountsHandler.GetExistingAccount(participant.pubKey)
	if err == nil {
		nonce = account.GetNonce()
	}

	for i := 0; i < numTxs; i++ {
		tx := &transaction.Transaction{
			Nonce:    nonce + uint64(i),
			Value:    big.NewInt(1),
			RcvAddr:  participant.pubKey,
			SndAddr:  participant.pubKey,
			GasPrice: 1000000000,
			GasLimit: 50000,
			Version:  1,
		}

		txs = append(txs, tx)
	}

	return txs
}
