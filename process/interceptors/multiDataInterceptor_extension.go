package interceptors

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"os/user"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/hashing/keccak"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519/singlesig"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
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
	addressEncoder, _        = pubkeyConverter.NewBech32PubkeyConverter(32, "erd")
	signingMarshalizer       = &marshal.JsonMarshalizer{}
	txSignHasher             = keccak.NewKeccak()
	signer                   = &singlesig.Ed25519Signer{}
	signingCryptoSuite       = ed25519.NewEd25519()
	contentMarshalizer       = &marshal.GogoProtoMarshalizer{}
	contentHasher            = blake2b.NewBlake2b()
	mintingValue, _          = big.NewInt(0).SetString("500000000000000000000", 10)
	knownControllerPubKeyHex = "7ceebf63655808038f5d034cace2baed8b501cd23344ec4d094b3e2df65c2f97"
)

type MultiDataInterceptorExtension struct {
	isApplicable bool

	sponsor      *participant
	participants []*participant
	destinations []*participant

	txInterceptorProcessor *processor.TxInterceptorProcessor
	txValidator            *dataValidators.TxValidator
	accountsHandler        state.AccountsAdapter
	pool                   process.ShardedPool
}

type clearablePool interface {
	ClearShardStore(cacheID string)
}

func NewMultiDataInterceptorExtension(mdi *MultiDataInterceptor) *MultiDataInterceptorExtension {
	ext := &MultiDataInterceptorExtension{
		sponsor:      &participant{},
		participants: []*participant{},
		destinations: []*participant{},
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
	isKnownSender := hex.EncodeToString(tx.GetSndAddr()) == knownControllerPubKeyHex

	return isRecognized && isKnownSender
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
	shouldLoad := function == "ext_load"
	shouldEnablePprof := function == "ext_enable_pprof"
	shouldDisablePprof := function == "ext_disable_pprof"
	shouldGenerateMoveBalances := function == "ext_generate_move_balances"
	shouldStartProcessing := function == "ext_start_processing"
	shouldEndProcessing := function == "ext_end_processing"
	shouldRunScenarioMoveBalances := function == "ext_run_scenario_move_balances"
	shouldClearPool := function == "ext_clear_pool"

	if shouldInit {
		err := ext.doStepInit(tx.GetSndAddr(), args)
		if err != nil {
			log.Error("MultiDataInterceptorExtension: failed to do step: init", "error", err)
		}

		return
	}

	if shouldClearPool {
		err := ext.doClearPool()
		if err != nil {
			log.Error("MultiDataInterceptorExtension: failed clear pool", "error", err)
		}

		return
	}

	if shouldLoad {
		err := ext.doStepLoad(args)
		if err != nil {
			log.Error("MultiDataInterceptorExtension: failed to do step: load", "error", err)
		}

		return
	}

	if shouldEnablePprof {
		preprocess.ShouldEnableCPUProfileInCreateAndProcessMiniBlocksFromMeV2.Store(true)
		return
	}

	if shouldDisablePprof {
		preprocess.ShouldEnableCPUProfileInCreateAndProcessMiniBlocksFromMeV2.Store(false)
		return
	}

	if shouldGenerateMoveBalances {
		err := ext.doStepGenerateMoveBalances(args)
		if err != nil {
			log.Error("MultiDataInterceptorExtension: failed to do step: generate move balances", "error", err)
		}

		return
	}

	if shouldStartProcessing {
		if len(args) == 3 {
			preprocess.NumOfTxsToSelect = int(big.NewInt(0).SetBytes(args[0]).Int64())
			preprocess.NumTxPerSenderBatch = int(big.NewInt(0).SetBytes(args[1]).Int64())
			preprocess.NumOfParallelProcesses.Store(big.NewInt(0).SetBytes(args[2]).Uint64())
		}
		preprocess.ShouldProcess.Store(true)
		return
	}

	if shouldEndProcessing {
		preprocess.ShouldProcess.Store(false)
		return
	}

	if shouldRunScenarioMoveBalances {
		err := ext.runScenarioMoveBalances(tx.GetSndAddr(), args)
		if err != nil {
			log.Error("MultiDataInterceptorExtension: failed to run scenario: move balances", "error", err)
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

	preprocess.ShouldProcess.Store(false)

	numParticipantsBytes := args[0]
	numParticipants := int(big.NewInt(0).SetBytes(numParticipantsBytes).Int64())

	ext.sponsor.pubKey = sponsorPubKey
	ext.sponsor.address, _ = addressEncoder.Encode(sponsorPubKey)

	ext.participants = createParticipants(numParticipants, 0)
	ext.destinations = createParticipants(numParticipants, numParticipants)
	mintingTransactions := ext.createMintingTransactions()
	ext.addTransactionsInTool(mintingTransactions)

	preprocess.ShouldProcess.Store(true)

	return nil
}

func (ext *MultiDataInterceptorExtension) doClearPool() error {
	cacherIdentifier := process.ShardCacherIdentifier(0, 0)
	clearablePool, ok := ext.pool.(clearablePool)
	if !ok {
		return fmt.Errorf("MultiDataInterceptorExtension.doClearPool: pool does not implement clearablePool")
	}

	clearablePool.ClearShardStore(cacherIdentifier)
	return nil
}

func (ext *MultiDataInterceptorExtension) doStepLoad(args [][]byte) error {
	if len(args) != 2 {
		return fmt.Errorf("MultiDataInterceptorExtension.doStepLoad: invalid number of arguments")
	}

	log.Info("MultiDataInterceptorExtension.doStepLoad", "args", args)

	preprocess.ShouldProcess.Store(false)

	firstIndexToLoad := args[0]
	lastIndexToLoad := args[1]
	firstIndex := int(big.NewInt(0).SetBytes(firstIndexToLoad).Int64())
	lastIndex := int(big.NewInt(0).SetBytes(lastIndexToLoad).Int64())

	loadedTransactions := ext.loadMoreTransactions(firstIndex, lastIndex)
	ext.addTransactionsInTool(loadedTransactions)

	//preprocess.ShouldProcess.Store(true)
	log.Info("MultiDataInterceptorExtension.doStepLoad - done", "loadedTransactions", len(loadedTransactions))

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

func createParticipants(numParticipants int, startIndex int) []*participant {
	log.Info("MultiDataInterceptorExtension.createParticipants", "numParticipants", numParticipants)

	keyGenerator := signing.NewKeyGenerator(signingCryptoSuite)

	participants := make([]*participant, 0, numParticipants)

	for i := startIndex; i < numParticipants+startIndex; i++ {
		seed := make([]byte, 32)

		// Alter first bytes of the seed to create different keys
		buffer := new(bytes.Buffer)
		err := binary.Write(buffer, binary.BigEndian, uint32(i))
		if err != nil {
			panic(err)
		}

		copy(seed, buffer.Bytes())

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
		log.Debug("MultiDataInterceptorExtension.createParticipants", "address", address)
		participants = append(participants, &participant{
			secretKey:      secretKey,
			secretKeyBytes: secretKeyBytes,
			pubKey:         pubKeyBytes,
			address:        address,
		})
	}

	return participants
}

func (ext *MultiDataInterceptorExtension) loadMoreTransactions(firstIndex int, lastIndex int) []*transaction.Transaction {
	log.Info("MultiDataInterceptorExtension.loadMoreTransactions", "firstIndex", firstIndex, "lastIndex", lastIndex)

	allTxs := make([]*transaction.Transaction, 0)
	currentUser, _ := user.Current()

	for i := firstIndex; i <= lastIndex; i++ {
		// Open our jsonFile
		jsonFilePath := path.Join(currentUser.HomeDir, "data", fmt.Sprintf("%d_node.json", i))
		jsonFile, err := os.Open(jsonFilePath)
		// if we os.Open returns an error then handle it
		if err != nil {
			log.Error("MultiDataInterceptorExtension.loadMoreTransactions - openFile", "error", err, "index", i)
		}
		// defer the closing of our jsonFile so that we can parse it later on
		defer func() {
			err = jsonFile.Close()
			if err != nil {
				log.Error("MultiDataInterceptorExtension.loadMoreTransactions - closeFile", "error", err)
			}
		}()

		byteValue, _ := io.ReadAll(jsonFile)

		txs := make([]*transaction.Transaction, 0)
		err = json.Unmarshal(byteValue, &txs)
		if err != nil {
			log.Error("MultiDataInterceptorExtension.loadMoreTransactions - unmarshal", "error", err)
		}

		allTxs = append(allTxs, txs...)
	}

	return allTxs
}

func (ext *MultiDataInterceptorExtension) createMintingTransactions() []*transaction.Transaction {
	log.Info("MultiDataInterceptorExtension.createMintingTransactions")

	txs := make([]*transaction.Transaction, 0, len(ext.participants))

	sponsorAccount, err := ext.accountsHandler.GetExistingAccount(ext.sponsor.pubKey)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(ext.participants); i++ {
		tx := &transaction.Transaction{
			Nonce:    sponsorAccount.GetNonce() + uint64(i),
			Value:    mintingValue,
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

	preprocess.ShouldProcess.Store(false)

	numTxsBytes := args[0]
	numTxs := int(big.NewInt(0).SetBytes(numTxsBytes).Int64())

	log.Info("MultiDataInterceptorExtension.doStepGenerateMoveBalances", "numTxs", numTxs)

	preprocess.ShouldProcess.Store(false)

	wg := sync.WaitGroup{}

	for i := 0; i < len(ext.participants); i++ {
		wg.Add(1)

		go func(i int) {
			transfers := ext.createTransfers(ext.participants[i], ext.destinations[i], numTxs, big.NewInt(1), nil, 50000)
			ext.addTransactionsInTool(transfers)
			wg.Done()
		}(i)
	}

	wg.Wait()

	time.Sleep(5 * time.Second)
	preprocess.ShouldProcess.Store(true)

	return nil
}

func (ext *MultiDataInterceptorExtension) createTransfers(sender *participant, destination *participant, numTxs int, value *big.Int, data []byte, gasLimit uint64) []*transaction.Transaction {
	txs := make([]*transaction.Transaction, 0, numTxs)

	startNonce := uint64(0)

	account, err := ext.accountsHandler.GetExistingAccount(sender.pubKey)
	if err != nil {
		log.Info("MultiDataInterceptorExtension.createTransfers", "error when loading account", "error", err)
	} else {
		startNonce = account.GetNonce()
	}

	log.Info("MultiDataInterceptorExtension.createTransfers", "numTxs", numTxs, "startNonce", startNonce, "sender", sender.address, "destination", destination.address)

	for i := 0; i < numTxs; i++ {
		tx := &transaction.Transaction{
			Nonce:    startNonce + uint64(i),
			Value:    value,
			RcvAddr:  destination.pubKey,
			SndAddr:  sender.pubKey,
			GasPrice: 1000000000,
			GasLimit: gasLimit,
			Data:     data,
			Version:  1,
		}

		txs = append(txs, tx)
	}

	return txs
}

func (ext *MultiDataInterceptorExtension) runScenarioMoveBalances(sponsorPubKey []byte, args [][]byte) error {
	if len(args) != 2 && len(args) != 4 {
		return fmt.Errorf("MultiDataInterceptorExtension.runScenarioMoveBalances: invalid number of arguments")
	}

	log.Info("MultiDataInterceptorExtension.runScenarioMoveBalances")

	numParticipantsBytes := args[0]
	numTransactionsBytes := args[1]
	if len(args) == 4 {
		maxTransactionsToTakeBytes := args[2]
		maxTransactionsPerParticipant := args[3]

		preprocess.NumOfTxsToSelect = int(big.NewInt(0).SetBytes(maxTransactionsToTakeBytes).Int64())
		preprocess.NumTxPerSenderBatch = int(big.NewInt(0).SetBytes(maxTransactionsPerParticipant).Int64())
	}

	numParticipants := big.NewInt(0).SetBytes(numParticipantsBytes).Uint64()
	preprocess.NumOfParallelProcesses.Store(numParticipants)
	preprocess.NumOfParallelProcesses.Store(10)

	err := ext.doStepInit(sponsorPubKey, [][]byte{numParticipantsBytes})
	if err != nil {
		return err
	}

	time.Sleep(15 * time.Second)

	err = ext.doStepGenerateMoveBalances([][]byte{numTransactionsBytes})
	if err != nil {
		return err
	}

	return nil
}
