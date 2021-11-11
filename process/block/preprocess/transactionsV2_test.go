package preprocess

import (
	"bytes"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionPreprocessor_SplitMiniBlockBasedOnTxTypeIfNeededShouldWork(t *testing.T) {
	t.Parallel()

	mb := block.MiniBlock{
		TxHashes: make([][]byte, 0),
	}

	mapSCTxs := make(map[string]struct{})
	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	mb.TxHashes = append(mb.TxHashes, txHash1)
	mb.TxHashes = append(mb.TxHashes, txHash2)
	mapSCTxs[string(txHash1)] = struct{}{}

	mbs := splitMiniBlockBasedOnTxTypeIfNeeded(&mb, mapSCTxs)
	require.Equal(t, 2, len(mbs))
	require.Equal(t, 1, len(mbs[0].TxHashes))
	require.Equal(t, 1, len(mbs[1].TxHashes))
	assert.Equal(t, txHash2, mbs[0].TxHashes[0])
	assert.Equal(t, txHash1, mbs[1].TxHashes[0])
}

func TestTransactions_ApplyVerifiedTransactionShouldWork(t *testing.T) {
	t.Parallel()

	numMiniBlocks := 0
	numTxs := 0
	dataPool := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	preprocessor, _ := NewTransactionPreprocessor(
		dataPool.Transactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&mock.FeeHandlerStub{},
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{
			AddNumTxsCalled: func(i int) {
				numTxs += i
			},
			AddNumMiniBlocksCalled: func(i int) {
				numMiniBlocks += i
			},
		},
		&testscommon.BalanceComputationStub{
			IsAddressSetCalled: func(address []byte) bool {
				return true
			},
			SubBalanceFromAddressCalled: func(address []byte, value *big.Int) bool {
				return false
			},
		},
		&epochNotifier.EpochNotifierStub{},
		2,
		2,
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
	)

	tx := &transaction.Transaction{}
	txHash := []byte("hash")
	mb := &block.MiniBlock{}
	receiverShardID := uint32(1)
	scheduledTxMbInfo := &scheduledTxAndMbInfo{
		isCrossShardScCallTx: true,
	}
	mapSCTxs := make(map[string]struct{})
	mbInfo := &createScheduledMiniBlocksInfo{
		mapCrossShardScCallTxs: make(map[uint32]int),
	}

	preprocessor.applyVerifiedTransaction(tx, txHash, mb, receiverShardID, scheduledTxMbInfo, mapSCTxs, mbInfo)

	assert.Equal(t, 2, numMiniBlocks)
	assert.Equal(t, 4, numTxs)
	assert.Equal(t, 1, len(mb.TxHashes))
	assert.True(t, mbInfo.firstCrossShardScCallTxFound)
	assert.Equal(t, 1, mbInfo.mapCrossShardScCallTxs[receiverShardID])
	assert.Equal(t, 1, mbInfo.maxCrossShardScCallTxsPerShard)
	assert.Equal(t, 1, mbInfo.schedulingInfo.numScheduledCrossShardScCalls)
	assert.Equal(t, 1, mbInfo.schedulingInfo.numScheduledTxsAdded)
}

func TestTransactions_GetScheduledTxAndMbInfoShouldWork(t *testing.T) {
	t.Parallel()

	dataPool := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	preprocessor, _ := NewTransactionPreprocessor(
		dataPool.Transactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&mock.FeeHandlerStub{},
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
		2,
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
	)

	receiverShardID := uint32(1)
	mbInfo := &createScheduledMiniBlocksInfo{
		mapCrossShardScCallTxs: make(map[uint32]int),
	}

	mbInfo.mapCrossShardScCallTxs[receiverShardID] = 1
	scheduledTxAndMbInfo := preprocessor.getScheduledTxAndMbInfo(true, receiverShardID, mbInfo)

	assert.True(t, scheduledTxAndMbInfo.isCrossShardScCallTx)
	assert.Equal(t, 2, scheduledTxAndMbInfo.numNewMiniBlocks)
	assert.Equal(t, 4, scheduledTxAndMbInfo.numNewTxs)
}

func TestTransactions_ShouldContinueProcessingScheduledTxShouldWork(t *testing.T) {
	t.Parallel()

	addressHasEnoughBalance := false
	dataPool := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	preprocessor, _ := NewTransactionPreprocessor(
		dataPool.Transactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&mock.FeeHandlerStub{},
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{
			IsAddressSetCalled: func(address []byte) bool {
				return true
			},
			AddressHasEnoughBalanceCalled: func(address []byte, value *big.Int) bool {
				return addressHasEnoughBalance
			},
		},
		&epochNotifier.EpochNotifierStub{},
		2,
		2,
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				if bytes.Equal(tx.GetRcvAddr(), []byte("smart contract address")) {
					return process.MoveBalance, process.SCInvoking
				}
				return process.MoveBalance, process.MoveBalance
			},
		},
		&testscommon.ScheduledTxsExecutionStub{},
	)

	isShardStuck := func(uint32) bool {
		return true
	}
	wrappedTx := &txcache.WrappedTransaction{}
	mapSCTxs := make(map[string]struct{})
	mbInfo := &createScheduledMiniBlocksInfo{}

	// should return false when tx is already added
	wrappedTx.TxHash = []byte("hash")
	mapSCTxs[string(wrappedTx.TxHash)] = struct{}{}
	tx, mb, shouldContinue := preprocessor.shouldContinueProcessingScheduledTx(isShardStuck, wrappedTx, mapSCTxs, mbInfo)
	assert.Nil(t, tx)
	assert.Nil(t, mb)
	assert.False(t, shouldContinue)

	// should return false when shard is stuck
	wrappedTx = &txcache.WrappedTransaction{}
	mapSCTxs = make(map[string]struct{})
	tx, mb, shouldContinue = preprocessor.shouldContinueProcessingScheduledTx(isShardStuck, wrappedTx, mapSCTxs, mbInfo)
	assert.Nil(t, tx)
	assert.Nil(t, mb)
	assert.False(t, shouldContinue)

	// should return false when wrong type assertion
	isShardStuck = func(uint32) bool {
		return false
	}
	wrappedTx = &txcache.WrappedTransaction{
		Tx: &smartContractResult.SmartContractResult{},
	}
	tx, mb, shouldContinue = preprocessor.shouldContinueProcessingScheduledTx(isShardStuck, wrappedTx, mapSCTxs, mbInfo)
	assert.Nil(t, tx)
	assert.Nil(t, mb)
	assert.False(t, shouldContinue)

	// should return false when sender must be skipped
	mbInfo.senderAddressToSkip = []byte("sender to skip")
	wrappedTx = &txcache.WrappedTransaction{
		Tx: &transaction.Transaction{
			SndAddr: mbInfo.senderAddressToSkip,
		},
	}
	tx, mb, shouldContinue = preprocessor.shouldContinueProcessingScheduledTx(isShardStuck, wrappedTx, mapSCTxs, mbInfo)
	assert.Nil(t, tx)
	assert.Nil(t, mb)
	assert.False(t, shouldContinue)
	assert.Equal(t, 1, mbInfo.schedulingInfo.numScheduledTxsSkipped)

	// should return false when receiver is not a smart contract address
	wrappedTx = &txcache.WrappedTransaction{
		Tx: &transaction.Transaction{
			SndAddr: []byte("sender"),
			RcvAddr: []byte("not a smart contract address"),
		},
	}
	mbInfo = &createScheduledMiniBlocksInfo{}
	tx, mb, shouldContinue = preprocessor.shouldContinueProcessingScheduledTx(isShardStuck, wrappedTx, mapSCTxs, mbInfo)
	assert.Nil(t, tx)
	assert.Nil(t, mb)
	assert.False(t, shouldContinue)
	assert.Equal(t, wrappedTx.Tx.GetSndAddr(), mbInfo.senderAddressToSkip)

	// should return false when mini block is not created
	wrappedTx = &txcache.WrappedTransaction{
		Tx: &transaction.Transaction{
			SndAddr: []byte("sender"),
			RcvAddr: []byte("smart contract address"),
		},
	}
	mbInfo = &createScheduledMiniBlocksInfo{
		mapMiniBlocks: make(map[uint32]*block.MiniBlock),
	}
	tx, mb, shouldContinue = preprocessor.shouldContinueProcessingScheduledTx(isShardStuck, wrappedTx, mapSCTxs, mbInfo)
	assert.Nil(t, tx)
	assert.Nil(t, mb)
	assert.False(t, shouldContinue)
	assert.Equal(t, wrappedTx.Tx.GetSndAddr(), mbInfo.senderAddressToSkip)

	// should return false when address has not enough initial balance
	wrappedTx = &txcache.WrappedTransaction{
		Tx: &transaction.Transaction{
			SndAddr: []byte("sender"),
			RcvAddr: []byte("smart contract address"),
		},
		ReceiverShardID: 1,
	}
	mbInfo = &createScheduledMiniBlocksInfo{
		mapMiniBlocks: make(map[uint32]*block.MiniBlock),
	}
	mbInfo.mapMiniBlocks[wrappedTx.ReceiverShardID] = &block.MiniBlock{}
	tx, mb, shouldContinue = preprocessor.shouldContinueProcessingScheduledTx(isShardStuck, wrappedTx, mapSCTxs, mbInfo)
	assert.Nil(t, tx)
	assert.Nil(t, mb)
	assert.False(t, shouldContinue)
	assert.Equal(t, wrappedTx.Tx.GetSndAddr(), mbInfo.senderAddressToSkip)
	assert.Equal(t, 1, mbInfo.schedulingInfo.numScheduledTxsWithInitialBalanceConsumed)

	// should return true
	addressHasEnoughBalance = true
	wrappedTx = &txcache.WrappedTransaction{
		Tx: &transaction.Transaction{
			SndAddr: []byte("sender"),
			RcvAddr: []byte("smart contract address"),
		},
		ReceiverShardID: 1,
	}
	mbInfo = &createScheduledMiniBlocksInfo{
		mapMiniBlocks: make(map[uint32]*block.MiniBlock),
	}
	mbInfo.mapMiniBlocks[wrappedTx.ReceiverShardID] = &block.MiniBlock{}
	tx, mb, shouldContinue = preprocessor.shouldContinueProcessingScheduledTx(isShardStuck, wrappedTx, mapSCTxs, mbInfo)
	assert.NotNil(t, tx)
	assert.NotNil(t, mb)
	assert.True(t, shouldContinue)
	assert.Equal(t, wrappedTx.Tx.GetSndAddr(), mbInfo.senderAddressToSkip)
}

func TestTransactions_IsFirstMiniBlockSplitForReceiverShardFoundShouldWork(t *testing.T) {
	t.Parallel()

	dataPool := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	preprocessor, _ := NewTransactionPreprocessor(
		dataPool.Transactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&mock.FeeHandlerStub{},
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
		2,
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
	)

	receiverShardID := uint32(1)
	txMbInfo := &txAndMbInfo{}
	mbInfo := &createAndProcessMiniBlocksInfo{
		mapTxsForShard: make(map[uint32]int),
		mapScsForShard: make(map[uint32]int),
	}

	// should return false when both num txs and num scs for shard are zero
	value := preprocessor.isFirstMiniBlockSplitForReceiverShardFound(receiverShardID, txMbInfo, mbInfo)
	assert.False(t, value)

	// should return false when both num txs and num scs for shard are NOT zero
	mbInfo.mapTxsForShard[receiverShardID] = 1
	mbInfo.mapScsForShard[receiverShardID] = 1
	value = preprocessor.isFirstMiniBlockSplitForReceiverShardFound(receiverShardID, txMbInfo, mbInfo)
	assert.False(t, value)

	// should return false when receiver is a smart contract address and num scs is NOT zero
	mbInfo.mapTxsForShard[receiverShardID] = 0
	mbInfo.mapScsForShard[receiverShardID] = 1
	txMbInfo.isReceiverSmartContractAddress = true
	value = preprocessor.isFirstMiniBlockSplitForReceiverShardFound(receiverShardID, txMbInfo, mbInfo)
	assert.False(t, value)

	// should return true when receiver is a smart contract address and num scs is zero
	mbInfo.mapTxsForShard[receiverShardID] = 1
	mbInfo.mapScsForShard[receiverShardID] = 0
	txMbInfo.isReceiverSmartContractAddress = true
	value = preprocessor.isFirstMiniBlockSplitForReceiverShardFound(receiverShardID, txMbInfo, mbInfo)
	assert.True(t, value)

	// should return false when receiver is NOT a smart contract address and num txs is NOT zero
	mbInfo.mapTxsForShard[receiverShardID] = 1
	mbInfo.mapScsForShard[receiverShardID] = 0
	txMbInfo.isReceiverSmartContractAddress = false
	value = preprocessor.isFirstMiniBlockSplitForReceiverShardFound(receiverShardID, txMbInfo, mbInfo)
	assert.False(t, value)

	// should return true when receiver is NOT a smart contract address and num txs is zero
	mbInfo.mapTxsForShard[receiverShardID] = 0
	mbInfo.mapScsForShard[receiverShardID] = 1
	txMbInfo.isReceiverSmartContractAddress = false
	value = preprocessor.isFirstMiniBlockSplitForReceiverShardFound(receiverShardID, txMbInfo, mbInfo)
	assert.True(t, value)
}

func TestTransactions_ApplyExecutedTransactionShouldWork(t *testing.T) {
	t.Parallel()

	numMiniBlocks := 0
	numTxs := 0
	dataPool := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	preprocessor, _ := NewTransactionPreprocessor(
		dataPool.Transactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&mock.FeeHandlerStub{},
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{
			AddNumTxsCalled: func(i int) {
				numTxs += i
			},
			AddNumMiniBlocksCalled: func(i int) {
				numMiniBlocks += i
			},
		},
		&testscommon.BalanceComputationStub{
			IsAddressSetCalled: func(address []byte) bool {
				return true
			},
			SubBalanceFromAddressCalled: func(address []byte, value *big.Int) bool {
				return false
			},
		},
		&epochNotifier.EpochNotifierStub{},
		2,
		2,
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
	)

	tx := &transaction.Transaction{}
	txHash := []byte("hash")
	receiverShardID := uint32(1)

	// receiver is a smart contract address
	mb := &block.MiniBlock{}
	txMbInfo := &txAndMbInfo{
		isCrossShardScCallOrSpecialTx:  true,
		isReceiverSmartContractAddress: true,
	}
	mbInfo := &createAndProcessMiniBlocksInfo{
		mapCrossShardScCallsOrSpecialTxs: make(map[uint32]int),
		mapScsForShard:                   make(map[uint32]int),
		mapSCTxs:                         make(map[string]struct{}),
		mapTxsForShard:                   make(map[uint32]int),
	}
	preprocessor.applyExecutedTransaction(tx, txHash, mb, receiverShardID, txMbInfo, mbInfo)

	assert.Equal(t, 2, numMiniBlocks)
	assert.Equal(t, 4, numTxs)
	assert.Equal(t, 1, len(mb.TxHashes))
	assert.True(t, mbInfo.firstCrossShardScCallOrSpecialTxFound)
	assert.Equal(t, 1, mbInfo.mapCrossShardScCallsOrSpecialTxs[receiverShardID])
	assert.Equal(t, 1, mbInfo.maxCrossShardScCallsOrSpecialTxsPerShard)
	assert.Equal(t, 1, mbInfo.processingInfo.numCrossShardScCallsOrSpecialTxs)
	assert.Equal(t, 0, mbInfo.mapTxsForShard[receiverShardID])
	assert.Equal(t, 1, mbInfo.mapScsForShard[receiverShardID])
	_, ok := mbInfo.mapSCTxs[string(txHash)]
	assert.True(t, ok)

	// receiver is NOT a smart contract address
	numMiniBlocks = 0
	numTxs = 0
	mb = &block.MiniBlock{}
	txMbInfo = &txAndMbInfo{
		isCrossShardScCallOrSpecialTx:  true,
		isReceiverSmartContractAddress: false,
	}
	mbInfo = &createAndProcessMiniBlocksInfo{
		mapCrossShardScCallsOrSpecialTxs: make(map[uint32]int),
		mapScsForShard:                   make(map[uint32]int),
		mapSCTxs:                         make(map[string]struct{}),
		mapTxsForShard:                   make(map[uint32]int),
	}
	preprocessor.applyExecutedTransaction(tx, txHash, mb, receiverShardID, txMbInfo, mbInfo)

	assert.Equal(t, 2, numMiniBlocks)
	assert.Equal(t, 4, numTxs)
	assert.Equal(t, 1, len(mb.TxHashes))
	assert.True(t, mbInfo.firstCrossShardScCallOrSpecialTxFound)
	assert.Equal(t, 1, mbInfo.mapCrossShardScCallsOrSpecialTxs[receiverShardID])
	assert.Equal(t, 1, mbInfo.maxCrossShardScCallsOrSpecialTxsPerShard)
	assert.Equal(t, 1, mbInfo.processingInfo.numCrossShardScCallsOrSpecialTxs)
	assert.Equal(t, 1, mbInfo.mapTxsForShard[receiverShardID])
	assert.Equal(t, 0, mbInfo.mapScsForShard[receiverShardID])
	_, ok = mbInfo.mapSCTxs[string(txHash)]
	assert.False(t, ok)
}

func TestTransactions_GetTxAndMbInfoShouldWork(t *testing.T) {
	t.Parallel()

	numMiniBlocks := 0
	numTxs := 0
	dataPool := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	preprocessor, _ := NewTransactionPreprocessor(
		dataPool.Transactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&mock.FeeHandlerStub{},
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{
			AddNumTxsCalled: func(i int) {
				numTxs += i
			},
			AddNumMiniBlocksCalled: func(i int) {
				numMiniBlocks += i
			},
		},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
		2,
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				if bytes.Equal(tx.GetRcvAddr(), []byte("smart contract address")) {
					return process.MoveBalance, process.SCInvoking
				}
				return process.MoveBalance, process.MoveBalance
			},
		},
		&testscommon.ScheduledTxsExecutionStub{},
	)

	tx := &transaction.Transaction{
		RcvAddr: []byte("smart contract address"),
	}
	receiverShardID := uint32(1)
	mbInfo := &createAndProcessMiniBlocksInfo{}

	txMbInfo := preprocessor.getTxAndMbInfo(tx, true, receiverShardID, mbInfo)

	assert.Equal(t, 2, txMbInfo.numNewMiniBlocks)
	assert.Equal(t, 4, txMbInfo.numNewTxs)
	assert.True(t, txMbInfo.isReceiverSmartContractAddress)
	assert.True(t, txMbInfo.isCrossShardScCallOrSpecialTx)
	assert.Equal(t, scTx, txMbInfo.txType)
}

func TestTransactions_ShouldContinueProcessingTxShouldWork(t *testing.T) {
	t.Parallel()

	addressHasEnoughBalance := false
	dataPool := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	preprocessor, _ := NewTransactionPreprocessor(
		dataPool.Transactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&mock.FeeHandlerStub{},
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{
			IsAddressSetCalled: func(address []byte) bool {
				return true
			},
			AddressHasEnoughBalanceCalled: func(address []byte, value *big.Int) bool {
				return addressHasEnoughBalance
			},
		},
		&epochNotifier.EpochNotifierStub{},
		2,
		2,
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				if bytes.Equal(tx.GetRcvAddr(), []byte("smart contract address")) {
					return process.MoveBalance, process.SCInvoking
				}
				return process.MoveBalance, process.MoveBalance
			},
		},
		&testscommon.ScheduledTxsExecutionStub{},
	)

	isShardStuck := func(uint32) bool {
		return true
	}
	wrappedTx := &txcache.WrappedTransaction{}
	mbInfo := &createAndProcessMiniBlocksInfo{}

	// should return false when shard is stuck
	tx, mb, shouldContinue := preprocessor.shouldContinueProcessingTx(isShardStuck, wrappedTx, mbInfo)
	assert.Nil(t, tx)
	assert.Nil(t, mb)
	assert.False(t, shouldContinue)

	// should return false when wrong type assertion
	isShardStuck = func(uint32) bool {
		return false
	}
	wrappedTx = &txcache.WrappedTransaction{
		Tx: &smartContractResult.SmartContractResult{},
	}
	tx, mb, shouldContinue = preprocessor.shouldContinueProcessingTx(isShardStuck, wrappedTx, mbInfo)
	assert.Nil(t, tx)
	assert.Nil(t, mb)
	assert.False(t, shouldContinue)

	// should return false when sender must be skipped
	mbInfo.senderAddressToSkip = []byte("sender to skip")
	wrappedTx = &txcache.WrappedTransaction{
		Tx: &transaction.Transaction{
			SndAddr: mbInfo.senderAddressToSkip,
		},
	}
	tx, mb, shouldContinue = preprocessor.shouldContinueProcessingTx(isShardStuck, wrappedTx, mbInfo)
	assert.Nil(t, tx)
	assert.Nil(t, mb)
	assert.False(t, shouldContinue)
	assert.Equal(t, 1, mbInfo.processingInfo.numTxsSkipped)

	// should return false when mini block is NOT created
	wrappedTx = &txcache.WrappedTransaction{
		Tx: &transaction.Transaction{
			SndAddr: []byte("sender"),
		},
	}
	mbInfo = &createAndProcessMiniBlocksInfo{
		mapMiniBlocks: make(map[uint32]*block.MiniBlock),
	}
	tx, mb, shouldContinue = preprocessor.shouldContinueProcessingTx(isShardStuck, wrappedTx, mbInfo)
	assert.Nil(t, tx)
	assert.Nil(t, mb)
	assert.False(t, shouldContinue)

	// should return false when address has not enough initial balance
	wrappedTx = &txcache.WrappedTransaction{
		Tx: &transaction.Transaction{
			SndAddr: []byte("sender"),
		},
		ReceiverShardID: 1,
	}
	mbInfo = &createAndProcessMiniBlocksInfo{
		mapMiniBlocks: make(map[uint32]*block.MiniBlock),
	}
	mbInfo.mapMiniBlocks[wrappedTx.ReceiverShardID] = &block.MiniBlock{}
	tx, mb, shouldContinue = preprocessor.shouldContinueProcessingTx(isShardStuck, wrappedTx, mbInfo)
	assert.Nil(t, tx)
	assert.Nil(t, mb)
	assert.False(t, shouldContinue)
	assert.Equal(t, 1, mbInfo.processingInfo.numTxsWithInitialBalanceConsumed)

	// should return true
	addressHasEnoughBalance = true
	wrappedTx = &txcache.WrappedTransaction{
		Tx: &transaction.Transaction{
			SndAddr: []byte("sender"),
		},
		ReceiverShardID: 1,
	}
	mbInfo = &createAndProcessMiniBlocksInfo{
		mapMiniBlocks: make(map[uint32]*block.MiniBlock),
	}
	mbInfo.mapMiniBlocks[wrappedTx.ReceiverShardID] = &block.MiniBlock{}
	tx, mb, shouldContinue = preprocessor.shouldContinueProcessingTx(isShardStuck, wrappedTx, mbInfo)
	assert.NotNil(t, tx)
	assert.NotNil(t, mb)
	assert.True(t, shouldContinue)
}

func TestTransactions_VerifyTransactionShouldWork(t *testing.T) {
	t.Parallel()

	var gasConsumedByTx uint64
	var verifyTransactionErr error
	dataPool := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	preprocessor, _ := NewTransactionPreprocessor(
		dataPool.Transactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{
			VerifyTransactionCalled: func(tx *transaction.Transaction) error {
				return verifyTransactionErr
			},
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&mock.FeeHandlerStub{
			MaxGasLimitPerMiniBlockForSafeCrossShardCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
			MaxGasLimitPerBlockForSafeCrossShardCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
			MaxGasLimitPerBlockCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
		},
		&mock.GasHandlerMock{
			ComputeGasConsumedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, gasConsumedByTx, nil
			},
		},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
		2,
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
	)

	tx := &transaction.Transaction{}
	txHash := []byte("hash")
	senderShardID := uint32(0)
	receiverShardID := uint32(1)

	// should err when computeGasConsumed method returns err
	mbInfo := &createScheduledMiniBlocksInfo{}
	gasConsumedByTx = MaxGasLimitPerBlock + 1
	err := preprocessor.verifyTransaction(tx, scTx, txHash, senderShardID, receiverShardID, mbInfo)
	assert.Equal(t, process.ErrMaxGasLimitPerOneTxInReceiverShardIsReached, err)
	assert.Equal(t, 1, mbInfo.schedulingInfo.numCrossShardTxsWithTooMuchGas)

	// should err when VerifyTransaction method returns err
	mbInfo = &createScheduledMiniBlocksInfo{
		mapGasConsumedByMiniBlockInReceiverShard: make(map[uint32]map[txType]uint64),
	}
	gasConsumedByTx = MaxGasLimitPerBlock - 1
	verifyTransactionErr = process.ErrLowerNonceInTransaction
	mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = make(map[txType]uint64)
	err = preprocessor.verifyTransaction(tx, scTx, txHash, senderShardID, receiverShardID, mbInfo)
	assert.Equal(t, process.ErrLowerNonceInTransaction, err)
	assert.Equal(t, 1, mbInfo.schedulingInfo.numScheduledBadTxs)

	// should work
	mbInfo = &createScheduledMiniBlocksInfo{
		mapGasConsumedByMiniBlockInReceiverShard: make(map[uint32]map[txType]uint64),
	}
	gasConsumedByTx = MaxGasLimitPerBlock - 1
	verifyTransactionErr = nil
	mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = make(map[txType]uint64)
	err = preprocessor.verifyTransaction(tx, scTx, txHash, senderShardID, receiverShardID, mbInfo)
	assert.Nil(t, err)
	_, ok := preprocessor.txsForCurrBlock.txHashAndInfo[string(txHash)]
	assert.True(t, ok)
}

func TestTransactions_CreateScheduledMiniBlocksShouldWork(t *testing.T) {
	t.Parallel()

	var gasConsumedByTx uint64
	dataPool := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	preprocessor, _ := NewTransactionPreprocessor(
		dataPool.Transactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&mock.FeeHandlerStub{
			MaxGasLimitPerMiniBlockForSafeCrossShardCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
			MaxGasLimitPerBlockForSafeCrossShardCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
			MaxGasLimitPerBlockCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
		},
		&mock.GasHandlerMock{
			ComputeGasConsumedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, gasConsumedByTx, nil
			},
		},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
		2,
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				if bytes.Equal(tx.GetRcvAddr(), []byte("smart contract address")) {
					return process.MoveBalance, process.SCInvoking
				}
				return process.MoveBalance, process.MoveBalance
			},
		},
		&testscommon.ScheduledTxsExecutionStub{},
	)

	var haveTimeMethodReturn bool
	var haveAdditionalTimeMethodReturn bool
	var isShardStuckMethodReturn bool
	var isMaxBlockSizeReachedMethodReturn bool

	haveTimeMethod := func() bool {
		return haveTimeMethodReturn
	}
	haveAdditionalTimeMethod := func() bool {
		return haveAdditionalTimeMethodReturn
	}
	isShardStuckMethod := func(uint32) bool {
		return isShardStuckMethodReturn
	}
	isMaxBlockSizeReachedMethod := func(int, int) bool {
		return isMaxBlockSizeReachedMethodReturn
	}

	// should not create scheduled mini blocks when time is out
	haveTimeMethodReturn = false
	haveAdditionalTimeMethodReturn = false
	isShardStuckMethodReturn = false
	isMaxBlockSizeReachedMethodReturn = false
	sortedTxs := make([]*txcache.WrappedTransaction, 0)
	mapSCTxs := make(map[string]struct{})
	tx := &txcache.WrappedTransaction{}
	sortedTxs = append(sortedTxs, tx)

	mbs := preprocessor.createScheduledMiniBlocks(haveTimeMethod, haveAdditionalTimeMethod, isShardStuckMethod, isMaxBlockSizeReachedMethod, sortedTxs, mapSCTxs)
	assert.Equal(t, 0, len(mbs))

	// should not create scheduled mini blocks when max block size is reached
	haveTimeMethodReturn = true
	haveAdditionalTimeMethodReturn = true
	isShardStuckMethodReturn = false
	isMaxBlockSizeReachedMethodReturn = true
	sortedTxs = make([]*txcache.WrappedTransaction, 0)
	mapSCTxs = make(map[string]struct{})
	tx = &txcache.WrappedTransaction{
		Tx: &transaction.Transaction{RcvAddr: []byte("smart contract address")},
	}
	sortedTxs = append(sortedTxs, tx)

	mbs = preprocessor.createScheduledMiniBlocks(haveTimeMethod, haveAdditionalTimeMethod, isShardStuckMethod, isMaxBlockSizeReachedMethod, sortedTxs, mapSCTxs)
	assert.Equal(t, 0, len(mbs))

	// should not create scheduled mini blocks when verifyTransaction returns error
	gasConsumedByTx = MaxGasLimitPerBlock + 1
	haveTimeMethodReturn = true
	haveAdditionalTimeMethodReturn = true
	isShardStuckMethodReturn = false
	isMaxBlockSizeReachedMethodReturn = false
	sortedTxs = make([]*txcache.WrappedTransaction, 0)
	mapSCTxs = make(map[string]struct{})
	tx = &txcache.WrappedTransaction{
		Tx: &transaction.Transaction{RcvAddr: []byte("smart contract address")},
	}
	sortedTxs = append(sortedTxs, tx)

	mbs = preprocessor.createScheduledMiniBlocks(haveTimeMethod, haveAdditionalTimeMethod, isShardStuckMethod, isMaxBlockSizeReachedMethod, sortedTxs, mapSCTxs)
	assert.Equal(t, 0, len(mbs))

	// should create two scheduled mini blocks
	gasConsumedByTx = MaxGasLimitPerBlock
	haveTimeMethodReturn = true
	haveAdditionalTimeMethodReturn = true
	isShardStuckMethodReturn = false
	isMaxBlockSizeReachedMethodReturn = false
	sortedTxs = make([]*txcache.WrappedTransaction, 0)
	mapSCTxs = make(map[string]struct{})
	tx1 := &txcache.WrappedTransaction{
		ReceiverShardID: 0,
		Tx:              &transaction.Transaction{Nonce: 1},
		TxHash:          []byte("hash1"),
	}
	tx2 := &txcache.WrappedTransaction{
		ReceiverShardID: 1,
		Tx:              &transaction.Transaction{Nonce: 2, RcvAddr: []byte("smart contract address")},
		TxHash:          []byte("hash2"),
	}
	tx3 := &txcache.WrappedTransaction{
		ReceiverShardID: 2,
		Tx:              &transaction.Transaction{Nonce: 3, RcvAddr: []byte("smart contract address")},
		TxHash:          []byte("hash3"),
	}
	sortedTxs = append(sortedTxs, tx1)
	sortedTxs = append(sortedTxs, tx2)
	sortedTxs = append(sortedTxs, tx3)
	mapSCTxs["hash1"] = struct{}{}

	mbs = preprocessor.createScheduledMiniBlocks(haveTimeMethod, haveAdditionalTimeMethod, isShardStuckMethod, isMaxBlockSizeReachedMethod, sortedTxs, mapSCTxs)
	assert.Equal(t, 2, len(mbs))
}

func TestTransactions_GetMiniBlockSliceFromMapV2ShouldWork(t *testing.T) {
	t.Parallel()

	dataPool := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	preprocessor, _ := NewTransactionPreprocessor(
		dataPool.Transactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&mock.FeeHandlerStub{},
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
		2,
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
	)

	mapMiniBlocks := make(map[uint32]*block.MiniBlock)
	mapSCTxs := make(map[string]struct{})

	mapMiniBlocks[0] = &block.MiniBlock{
		TxHashes:        [][]byte{[]byte("txHash1")},
		ReceiverShardID: 0}
	mapMiniBlocks[1] = &block.MiniBlock{
		TxHashes:        [][]byte{[]byte("txHash1"), []byte("txHash2")},
		ReceiverShardID: 1}
	mapMiniBlocks[2] = &block.MiniBlock{
		TxHashes:        [][]byte{[]byte("txHash1"), []byte("txHash2"), []byte("txHash3")},
		ReceiverShardID: 2}
	mapMiniBlocks[core.MetachainShardId] = &block.MiniBlock{
		TxHashes:        [][]byte{[]byte("txHash1"), []byte("txHash2"), []byte("txHash3"), []byte("txHash4")},
		ReceiverShardID: core.MetachainShardId}
	mbs := preprocessor.getMiniBlockSliceFromMapV2(mapMiniBlocks, mapSCTxs)
	assert.Equal(t, 4, len(mbs))
}

func TestTransactions_CreateAndProcessMiniBlocksFromMeV2ShouldWork(t *testing.T) {
	t.Parallel()

	var gasConsumedByTx uint64
	dataPool := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	preprocessor, _ := NewTransactionPreprocessor(
		dataPool.Transactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&mock.FeeHandlerStub{
			MaxGasLimitPerMiniBlockForSafeCrossShardCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
			MaxGasLimitPerBlockForSafeCrossShardCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
			MaxGasLimitPerBlockCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
		},
		&mock.GasHandlerMock{
			ComputeGasConsumedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, gasConsumedByTx, nil
			},
		},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
		2,
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				if bytes.Equal(tx.GetRcvAddr(), []byte("smart contract address")) {
					return process.MoveBalance, process.SCInvoking
				}
				return process.MoveBalance, process.MoveBalance
			},
		},
		&testscommon.ScheduledTxsExecutionStub{},
	)

	var haveTimeMethodReturn bool
	var isShardStuckMethodReturn bool
	var isMaxBlockSizeReachedMethodReturn bool

	haveTimeMethod := func() bool {
		return haveTimeMethodReturn
	}
	isShardStuckMethod := func(uint32) bool {
		return isShardStuckMethodReturn
	}
	isMaxBlockSizeReachedMethod := func(int, int) bool {
		return isMaxBlockSizeReachedMethodReturn
	}

	// should not create and process mini blocks when time is out
	haveTimeMethodReturn = false
	isShardStuckMethodReturn = false
	isMaxBlockSizeReachedMethodReturn = false
	sortedTxs := make([]*txcache.WrappedTransaction, 0)
	tx := &txcache.WrappedTransaction{}
	sortedTxs = append(sortedTxs, tx)

	mbs, mapSCTxs, _ := preprocessor.createAndProcessMiniBlocksFromMeV2(haveTimeMethod, isShardStuckMethod, isMaxBlockSizeReachedMethod, sortedTxs)
	assert.Equal(t, 0, len(mbs))
	assert.Equal(t, 0, len(mapSCTxs))

	// should not create and process mini blocks when max block size is reached
	haveTimeMethodReturn = true
	isShardStuckMethodReturn = false
	isMaxBlockSizeReachedMethodReturn = true
	sortedTxs = make([]*txcache.WrappedTransaction, 0)
	mapSCTxs = make(map[string]struct{})
	tx = &txcache.WrappedTransaction{
		Tx: &transaction.Transaction{RcvAddr: []byte("smart contract address")},
	}
	sortedTxs = append(sortedTxs, tx)

	mbs, mapSCTxs, _ = preprocessor.createAndProcessMiniBlocksFromMeV2(haveTimeMethod, isShardStuckMethod, isMaxBlockSizeReachedMethod, sortedTxs)
	assert.Equal(t, 0, len(mbs))
	assert.Equal(t, 0, len(mapSCTxs))

	// should not create and process mini blocks when processTransaction returns error
	gasConsumedByTx = MaxGasLimitPerBlock + 1
	haveTimeMethodReturn = true
	isShardStuckMethodReturn = false
	isMaxBlockSizeReachedMethodReturn = false
	sortedTxs = make([]*txcache.WrappedTransaction, 0)
	mapSCTxs = make(map[string]struct{})
	tx = &txcache.WrappedTransaction{
		Tx: &transaction.Transaction{RcvAddr: []byte("smart contract address")},
	}
	sortedTxs = append(sortedTxs, tx)

	mbs, mapSCTxs, _ = preprocessor.createAndProcessMiniBlocksFromMeV2(haveTimeMethod, isShardStuckMethod, isMaxBlockSizeReachedMethod, sortedTxs)
	assert.Equal(t, 0, len(mbs))
	assert.Equal(t, 0, len(mapSCTxs))

	// should create and process two mini blocks
	gasConsumedByTx = MaxGasLimitPerBlock
	haveTimeMethodReturn = true
	isShardStuckMethodReturn = false
	isMaxBlockSizeReachedMethodReturn = false
	sortedTxs = make([]*txcache.WrappedTransaction, 0)
	mapSCTxs = make(map[string]struct{})
	tx1 := &txcache.WrappedTransaction{
		ReceiverShardID: 0,
		Tx:              &transaction.Transaction{Nonce: 1},
		TxHash:          []byte("hash1"),
	}
	tx2 := &txcache.WrappedTransaction{
		ReceiverShardID: 1,
		Tx:              &transaction.Transaction{Nonce: 2, RcvAddr: []byte("smart contract address")},
		TxHash:          []byte("hash2"),
	}
	tx3 := &txcache.WrappedTransaction{
		ReceiverShardID: 2,
		Tx:              &transaction.Transaction{Nonce: 3, RcvAddr: []byte("smart contract address")},
		TxHash:          []byte("hash3"),
	}
	sortedTxs = append(sortedTxs, tx1)
	sortedTxs = append(sortedTxs, tx2)
	sortedTxs = append(sortedTxs, tx3)
	mapSCTxs["hash1"] = struct{}{}

	mbs, mapSCTxs, _ = preprocessor.createAndProcessMiniBlocksFromMeV2(haveTimeMethod, isShardStuckMethod, isMaxBlockSizeReachedMethod, sortedTxs)
	assert.Equal(t, 3, len(mbs))
	assert.Equal(t, 2, len(mapSCTxs))
}

func TestTransactions_ProcessTransactionShouldWork(t *testing.T) {
	t.Parallel()

	var gasConsumedByTx uint64
	var processTransactionErr error
	dataPool := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	preprocessor, _ := NewTransactionPreprocessor(
		dataPool.Transactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return vmcommon.UserError, processTransactionErr
			},
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&mock.FeeHandlerStub{
			MaxGasLimitPerMiniBlockForSafeCrossShardCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
			MaxGasLimitPerBlockForSafeCrossShardCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
			MaxGasLimitPerBlockCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
		},
		&mock.GasHandlerMock{
			ComputeGasConsumedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, gasConsumedByTx, nil
			},
		},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
		2,
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				if bytes.Equal(tx.GetRcvAddr(), []byte("smart contract address")) {
					return process.MoveBalance, process.SCInvoking
				}
				return process.MoveBalance, process.MoveBalance
			},
		},
		&testscommon.ScheduledTxsExecutionStub{},
	)

	tx := &transaction.Transaction{}
	txHash := []byte("txHash")
	senderShardID := uint32(0)
	receiverShardID := uint32(0)

	// should not process transaction when computeGasConsumed method returns error
	mbInfo := &createAndProcessMiniBlocksInfo{
		mapGasConsumedByMiniBlockInReceiverShard: make(map[uint32]map[txType]uint64),
	}
	mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = make(map[txType]uint64)
	gasConsumedByTx = MaxGasLimitPerBlock + 1
	err := preprocessor.processTransaction(tx, nonScTx, txHash, senderShardID, receiverShardID, mbInfo)
	assert.Equal(t, process.ErrMaxGasLimitPerOneTxInReceiverShardIsReached, err)

	// should not process transaction when processAndRemoveBadTransaction method returns error
	mbInfo = &createAndProcessMiniBlocksInfo{
		mapGasConsumedByMiniBlockInReceiverShard: make(map[uint32]map[txType]uint64),
	}
	mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = make(map[txType]uint64)
	gasConsumedByTx = MaxGasLimitPerBlock - 1
	processTransactionErr = process.ErrHigherNonceInTransaction
	err = preprocessor.processTransaction(tx, nonScTx, txHash, senderShardID, receiverShardID, mbInfo)
	assert.Equal(t, process.ErrHigherNonceInTransaction, err)

	// should process transaction bat should return ErrFailedTransaction
	mbInfo = &createAndProcessMiniBlocksInfo{
		mapGasConsumedByMiniBlockInReceiverShard: make(map[uint32]map[txType]uint64),
	}
	mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = make(map[txType]uint64)
	processTransactionErr = process.ErrFailedTransaction
	err = preprocessor.processTransaction(tx, nonScTx, txHash, senderShardID, receiverShardID, mbInfo)
	assert.Equal(t, process.ErrFailedTransaction, err)

	// should process transaction
	mbInfo = &createAndProcessMiniBlocksInfo{
		mapGasConsumedByMiniBlockInReceiverShard: make(map[uint32]map[txType]uint64),
	}
	mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = make(map[txType]uint64)
	processTransactionErr = nil
	err = preprocessor.processTransaction(tx, nonScTx, txHash, senderShardID, receiverShardID, mbInfo)
	assert.Nil(t, err)
}
