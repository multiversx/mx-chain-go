package mempool

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/config"
	testsChainSimulator "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/multiversx/mx-chain-go/txcache"
	"github.com/stretchr/testify/require"
)

var (
	oneEGLD                   = big.NewInt(1000000000000000000)
	oneQuarterOfEGLD          = big.NewInt(250000000000000000)
	durationWaitAfterSendMany = 3000 * time.Millisecond
	durationWaitAfterSendSome = 300 * time.Millisecond
	defaultBlockchainInfo     = holders.NewBlockchainInfo(nil, 0)
	gasLimit                  = 50_000
	gasPrice                  = 1_000_000_000
)

const maxNumBytesUpperBound = 1_073_741_824           // one GB
const maxNumBytesPerSenderUpperBoundTest = 33_554_432 // 32 MB

type accountInfo struct {
	balance *big.Int
	nonce   uint64
}

func startChainSimulator(t *testing.T, alterConfigsFunction func(cfg *config.Configs)) testsChainSimulator.ChainSimulator {
	simulator, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck: true,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    "../../../cmd/node/config/",
		NumOfShards:            1,
		GenesisTimestamp:       time.Now().Unix(),
		RoundDurationInMillis:  uint64(4000),
		RoundsPerEpoch: core.OptionalUint64{
			HasValue: true,
			Value:    10,
		},
		ApiInterface:             api.NewNoApiInterface(),
		MinNodesPerShard:         1,
		MetaChainMinNodes:        1,
		NumNodesWaitingListMeta:  0,
		NumNodesWaitingListShard: 0,
		AlterConfigsFunction:     alterConfigsFunction,
	})
	require.NoError(t, err)
	require.NotNil(t, simulator)

	err = simulator.GenerateBlocksUntilEpochIsReached(1)
	require.NoError(t, err)

	return simulator
}

type participantsHolder struct {
	sendersByShard  map[int][]dtos.WalletAddress
	relayerByShard  map[int]dtos.WalletAddress
	receiverByShard map[int]dtos.WalletAddress
}

func newParticipantsHolder() *participantsHolder {
	return &participantsHolder{
		sendersByShard:  make(map[int][]dtos.WalletAddress),
		relayerByShard:  make(map[int]dtos.WalletAddress),
		receiverByShard: make(map[int]dtos.WalletAddress),
	}
}

func createParticipants(t *testing.T, simulator testsChainSimulator.ChainSimulator, numSendersPerShard int) *participantsHolder {
	numShards := int(simulator.GetNodeHandler(0).GetShardCoordinator().NumberOfShards())
	participants := newParticipantsHolder()

	for shard := 0; shard < numShards; shard++ {
		senders := make([]dtos.WalletAddress, 0, numSendersPerShard)

		for i := 0; i < numSendersPerShard; i++ {
			sender, err := simulator.GenerateAndMintWalletAddress(uint32(shard), oneEGLD)
			require.NoError(t, err)

			senders = append(senders, sender)
		}

		relayer, err := simulator.GenerateAndMintWalletAddress(uint32(shard), oneEGLD)
		require.NoError(t, err)

		receiver, err := simulator.GenerateAndMintWalletAddress(uint32(shard), big.NewInt(0))
		require.NoError(t, err)

		participants.sendersByShard[shard] = senders
		participants.relayerByShard[shard] = relayer
		participants.receiverByShard[shard] = receiver
	}

	err := simulator.GenerateBlocks(1)
	require.Nil(t, err)

	return participants
}

type noncesTracker struct {
	nonceByAddress map[string]uint64
}

func newNoncesTracker() *noncesTracker {
	return &noncesTracker{
		nonceByAddress: make(map[string]uint64),
	}
}

func (tracker *noncesTracker) getThenIncrementNonce(address dtos.WalletAddress) uint64 {
	nonce, ok := tracker.nonceByAddress[address.Bech32]
	if !ok {
		tracker.nonceByAddress[address.Bech32] = 0
	}

	tracker.nonceByAddress[address.Bech32]++
	return nonce
}

func (tracker *noncesTracker) getThenIncrementNonceByStringAddress(address string) uint64 {
	nonce, ok := tracker.nonceByAddress[address]
	if !ok {
		tracker.nonceByAddress[address] = 0
	}

	tracker.nonceByAddress[address]++
	return nonce
}

func sendTransactions(t *testing.T, simulator testsChainSimulator.ChainSimulator, transactions []*transaction.Transaction) {
	transactionsBySenderShard := make(map[int][]*transaction.Transaction)
	shardCoordinator := simulator.GetNodeHandler(0).GetShardCoordinator()

	for _, tx := range transactions {
		shard := int(shardCoordinator.ComputeId(tx.SndAddr))
		transactionsBySenderShard[shard] = append(transactionsBySenderShard[shard], tx)
	}

	for shard, transactionsFromShard := range transactionsBySenderShard {
		node := simulator.GetNodeHandler(uint32(shard))

		for _, tx := range transactionsFromShard {
			err := node.GetFacadeHandler().ValidateTransaction(tx)
			require.NoError(t, err)
		}

		numSent, err := node.GetFacadeHandler().SendBulkTransactions(transactionsFromShard)

		require.NoError(t, err)
		require.Equal(t, len(transactionsFromShard), int(numSent))
	}
}

func sendTransaction(t *testing.T, simulator testsChainSimulator.ChainSimulator, tx *transaction.Transaction) {
	sendTransactions(t, simulator, []*transaction.Transaction{tx})
}

func selectTransactions(t *testing.T, simulator testsChainSimulator.ChainSimulator, shard int) ([]*txcache.WrappedTransaction, uint64) {
	shardAsString := strconv.Itoa(shard)
	node := simulator.GetNodeHandler(uint32(shard))
	accountsAdapter := node.GetStateComponents().AccountsAdapter()
	poolsHolder := node.GetDataComponents().Datapool().Transactions()

	selectionSession, err := preprocess.NewSelectionSession(preprocess.ArgsSelectionSession{
		AccountsAdapter:       accountsAdapter,
		TransactionsProcessor: &testscommon.TxProcessorStub{},
	})
	require.NoError(t, err)

	options := holders.NewTxSelectionOptions(
		10_000_000_000,
		30_000,
		250,
		10,
	)

	mempool := poolsHolder.ShardDataStore(shardAsString).(*txcache.TxCache)
	selectedTransactions, gas := mempool.SelectTransactions(selectionSession, options, defaultBlockchainInfo)
	return selectedTransactions, gas
}

func getNumTransactionsInPool(simulator testsChainSimulator.ChainSimulator, shard int) int {
	node := simulator.GetNodeHandler(uint32(shard))
	poolsHolder := node.GetDataComponents().Datapool().Transactions()
	return int(poolsHolder.GetCounts().GetTotal())
}

func getNumTransactionsInCurrentBlock(simulator testsChainSimulator.ChainSimulator, shard int) int {
	node := simulator.GetNodeHandler(uint32(shard))
	currentBlock := node.GetDataComponents().Blockchain().GetCurrentBlockHeader()
	return int(currentBlock.GetTxCount())
}

func getTransaction(t *testing.T, simulator testsChainSimulator.ChainSimulator, shard int, hash []byte) *transaction.ApiTransactionResult {
	hashAsHex := hex.EncodeToString(hash)
	transaction, err := simulator.GetNodeHandler(uint32(shard)).GetFacadeHandler().GetTransaction(hashAsHex, true)
	require.NoError(t, err)
	return transaction
}

func createMockSelectionSessionWithSpecificAccountInfo(accounts map[string]*accountInfo) *txcachemocks.SelectionSessionMock {
	sessionMock := txcachemocks.SelectionSessionMock{
		GetAccountStateCalled: func(address []byte) (state.UserAccountHandler, error) {
			return &stateMock.StateUserAccountHandlerStub{
				GetBalanceCalled: func() *big.Int {
					return accounts[string(address)].balance
				},
				GetNonceCalled: func() uint64 {
					return accounts[string(address)].nonce
				},
			}, nil
		},
	}

	return &sessionMock
}

func createProposedBlock(selectedTransactions []*txcache.WrappedTransaction) *block.Body {
	// extract the tx hashes from the selected transactions
	proposedTxs := make([][]byte, 0, len(selectedTransactions))
	for _, tx := range selectedTransactions {
		proposedTxs = append(proposedTxs, tx.TxHash)
	}

	return &block.Body{MiniBlocks: []*block.MiniBlock{
		{
			TxHashes: proposedTxs,
		},
	}}
}

func createDefaultSelectionSessionMockWithInitialAmount(initialAmountPerAccount *big.Int) *txcachemocks.SelectionSessionMock {
	sessionMock := txcachemocks.SelectionSessionMock{
		GetAccountStateCalled: func(address []byte) (state.UserAccountHandler, error) {
			return &stateMock.StateUserAccountHandlerStub{
				GetBalanceCalled: func() *big.Int {
					return initialAmountPerAccount
				},
				GetNonceCalled: func() uint64 {
					return 0
				},
			}, nil
		},
	}

	return &sessionMock
}

func createFakeAddresses(numAddresses int) []string {
	base := "sender"
	addresses := make([]string, numAddresses)
	for i := 0; i < numAddresses; i++ {
		addresses[i] = fmt.Sprintf("%s:%d", base, i)
	}

	return addresses
}

func createRandomTx(nonceTracker *noncesTracker, accounts []string) *transaction.Transaction {
	sender := rand.Intn(len(accounts))
	receiver := rand.Intn(len(accounts))
	for sender == receiver {
		receiver = rand.Intn(len(accounts))
	}

	return &transaction.Transaction{
		Nonce:     nonceTracker.getThenIncrementNonceByStringAddress(accounts[sender]),
		Value:     big.NewInt(1),
		SndAddr:   []byte(accounts[sender]),
		RcvAddr:   []byte(accounts[receiver]),
		Data:      []byte{},
		GasLimit:  50_000,
		GasPrice:  1_000_000_000,
		ChainID:   []byte(configs.ChainID),
		Version:   2,
		Signature: []byte("signature"),
	}
}

func createRandomTxs(txpool *txcache.TxCache, numTxs int, nonceTracker *noncesTracker, accounts []string) {
	for i := 0; i < numTxs; i++ {
		tx := createRandomTx(nonceTracker, accounts)
		txHash := []byte(fmt.Sprintf("txHash%d", i))
		wtx := &txcache.WrappedTransaction{
			Tx:               tx,
			TxHash:           txHash,
			SenderShardID:    0,
			ReceiverShardID:  0,
			Size:             0,
			Fee:              core.SafeMul(tx.GasLimit, tx.GasPrice),
			PricePerUnit:     0,
			TransferredValue: tx.Value,
			FeePayer:         tx.SndAddr,
		}
		txpool.AddTx(wtx)
	}
}

func testOnProposed(t *testing.T, sw *core.StopWatch, numTxs int, numAddresses int) {
	// create some fake address for each account
	accounts := createFakeAddresses(numAddresses)

	host := txcachemocks.NewMempoolHostMock()
	txpool, err := txcache.NewTxCache(configSourceMe, host)

	require.Nil(t, err)
	require.NotNil(t, txpool)

	initialAmount := big.NewInt(0)
	numTxsAsBigInt := big.NewInt(int64(numTxs))

	// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
	_ = initialAmount.Mul(numTxsAsBigInt, core.SafeMul(uint64(gasLimit), uint64(gasPrice)))
	_ = initialAmount.Add(initialAmount, core.SafeMul(uint64(numTxs), uint64(transferredValue)))

	selectionSession := createDefaultSelectionSessionMockWithInitialAmount(initialAmount)
	options := holders.NewTxSelectionOptions(
		10_000_000_000,
		numTxs,
		selectionLoopMaximumDuration,
		10,
	)

	nonceTracker := newNoncesTracker()
	createRandomTxs(txpool, numTxs, nonceTracker, accounts)

	require.Equal(t, numTxs, int(txpool.CountTx()))

	blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)
	selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
	require.Equal(t, numTxs, len(selectedTransactions))

	proposedBlock1 := createProposedBlock(selectedTransactions)

	// start profiling
	fileName := strings.ReplaceAll(t.Name(), "/", "_")
	f, err := os.Create(fmt.Sprintf("TestOnProposed%s.pprof", fileName))
	require.Nil(t, err)
	err = pprof.StartCPUProfile(f)
	require.Nil(t, err)

	sw.Start(t.Name())
	// measure the time spent
	err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock1, &block.Header{
		Nonce:    0,
		PrevHash: []byte("blockHash0"),
		RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
	},
		selectionSession,
		defaultBlockchainInfo,
	)
	sw.Stop(t.Name())
	require.Nil(t, err)

	// stop profiling
	pprof.StopCPUProfile()
	err = f.Close()
	require.NoError(t, err)
}

func testFirstSelection(t *testing.T, sw *core.StopWatch, numTxs int, numTxsToBeSelected, numAddresses int) {
	// create some fake address for each account
	accounts := createFakeAddresses(numAddresses)

	host := txcachemocks.NewMempoolHostMock()
	txpool, err := txcache.NewTxCache(configSourceMe, host)

	require.Nil(t, err)
	require.NotNil(t, txpool)

	// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
	initialAmount := big.NewInt(0)
	numTxsAsBigInt := big.NewInt(int64(numTxs))

	_ = initialAmount.Mul(numTxsAsBigInt, core.SafeMul(uint64(gasLimit), uint64(gasPrice)))
	_ = initialAmount.Add(initialAmount, big.NewInt(int64(numTxs)))

	selectionSession := createDefaultSelectionSessionMockWithInitialAmount(initialAmount)
	options := holders.NewTxSelectionOptions(
		10_000_000_000*10, // in case of 1_000_000 txs
		numTxsToBeSelected,
		selectionLoopMaximumDuration*3, // in case of 1_000_000 txs
		10,
	)

	nonceTracker := newNoncesTracker()
	createRandomTxs(txpool, numTxs, nonceTracker, accounts)

	require.Equal(t, numTxs, int(txpool.CountTx()))

	blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)

	// start profiling
	fileName := strings.ReplaceAll(t.Name(), "/", "_")
	f, err := os.Create(fmt.Sprintf("TestOnProposed%s.pprof", fileName))
	require.Nil(t, err)
	err = pprof.StartCPUProfile(f)
	require.Nil(t, err)

	sw.Start(t.Name())
	selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
	sw.Stop(t.Name())

	// stop profiling
	pprof.StopCPUProfile()
	err = f.Close()
	require.NoError(t, err)

	require.Equal(t, numTxsToBeSelected, len(selectedTransactions))
}

func testSecondSelection(t *testing.T, sw *core.StopWatch, numTxs int, numTxsToBeSelected int, numAddresses int) {
	// create some fake address for each account
	accounts := createFakeAddresses(numAddresses)

	host := txcachemocks.NewMempoolHostMock()
	txpool, err := txcache.NewTxCache(configSourceMe, host)

	require.Nil(t, err)
	require.NotNil(t, txpool)

	// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
	initialAmount := big.NewInt(0)
	numTxsAsBigInt := big.NewInt(int64(numTxs))

	_ = initialAmount.Mul(numTxsAsBigInt, core.SafeMul(uint64(gasLimit), uint64(gasPrice)))
	_ = initialAmount.Add(initialAmount, core.SafeMul(uint64(numTxs), uint64(transferredValue)))

	selectionSession := createDefaultSelectionSessionMockWithInitialAmount(initialAmount)
	options := holders.NewTxSelectionOptions(
		10_000_000_000*10,
		numTxsToBeSelected,
		selectionLoopMaximumDuration*10,
		10,
	)

	nonceTracker := newNoncesTracker()
	createRandomTxs(txpool, numTxs, nonceTracker, accounts)

	require.Equal(t, numTxs, int(txpool.CountTx()))

	blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)
	selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
	require.Equal(t, numTxsToBeSelected, len(selectedTransactions))

	proposedBlock := createProposedBlock(selectedTransactions)
	// propose those txs in order to track them (create the breadcrumbs used for the virtual records)
	err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock, &block.Header{
		Nonce:    0,
		PrevHash: []byte("blockHash0"),
		RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
	},
		selectionSession,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// measure the time for the second selection (now we use the breadcrumbs to create the virtual records)
	sw.Start(t.Name())
	selectedTransactions, _ = txpool.SelectTransactions(selectionSession, options, blockchainInfo)
	sw.Stop(t.Name())

	require.Equal(t, numTxsToBeSelected, len(selectedTransactions))

	// propose the block and make sure the selection works well
	proposedBlock = createProposedBlock(selectedTransactions)
	err = txpool.OnProposedBlock([]byte("blockHash2"), proposedBlock, &block.Header{
		Nonce:    0,
		PrevHash: []byte("blockHash1"),
		RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
	},
		selectionSession,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// start profiling
	fileName := strings.ReplaceAll(t.Name(), "/", "_")
	f, err := os.Create(fmt.Sprintf("TestOnProposed%s.pprof", fileName))
	require.Nil(t, err)
	err = pprof.StartCPUProfile(f)
	require.Nil(t, err)

	selectedTransactions, _ = txpool.SelectTransactions(selectionSession, options, blockchainInfo)

	// stop profiling
	pprof.StopCPUProfile()
	err = f.Close()
	require.NoError(t, err)

	require.Equal(t, 0, len(selectedTransactions))
}

func testSecondSelectionWithManyTxsInPool(t *testing.T, sw *core.StopWatch, numTxs int, numTxsToBeSelected int, numAddresses int) {
	accounts := createFakeAddresses(numAddresses)

	host := txcachemocks.NewMempoolHostMock()
	txpool, err := txcache.NewTxCache(configSourceMe, host)

	require.Nil(t, err)
	require.NotNil(t, txpool)

	// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
	initialAmount := big.NewInt(0)
	numTxsAsBigInt := big.NewInt(int64(numTxs))

	_ = initialAmount.Mul(numTxsAsBigInt, core.SafeMul(uint64(gasLimit), uint64(gasPrice)))
	_ = initialAmount.Add(initialAmount, core.SafeMul(uint64(numTxs), uint64(transferredValue)))

	selectionSession := createDefaultSelectionSessionMockWithInitialAmount(initialAmount)
	options := holders.NewTxSelectionOptions(
		10_000_000_000,
		numTxsToBeSelected,
		selectionLoopMaximumDuration,
		10,
	)

	nonceTracker := newNoncesTracker()
	createRandomTxs(txpool, numTxs, nonceTracker, accounts)

	require.Equal(t, numTxs, int(txpool.CountTx()))

	blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)
	selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
	require.Equal(t, numTxsToBeSelected, len(selectedTransactions))

	proposedBlock := createProposedBlock(selectedTransactions)
	// propose those txs in order to track them (create the breadcrumbs used for the virtual records)
	err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock, &block.Header{
		Nonce:    0,
		PrevHash: []byte("blockHash0"),
		RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
	},
		selectionSession,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// start profiling
	fileName := strings.ReplaceAll(t.Name(), "/", "_")
	f, err := os.Create(fmt.Sprintf("TestOnProposed%s.pprof", fileName))
	require.Nil(t, err)
	err = pprof.StartCPUProfile(f)
	require.Nil(t, err)

	// measure the time for the second selection (now we use the breadcrumbs to create the virtual records)
	sw.Start(t.Name())
	selectedTransactions, _ = txpool.SelectTransactions(selectionSession, options, blockchainInfo)
	sw.Stop(t.Name())

	// stop profiling
	pprof.StopCPUProfile()
	err = f.Close()
	require.NoError(t, err)

	require.Equal(t, numTxsToBeSelected, len(selectedTransactions))
}
