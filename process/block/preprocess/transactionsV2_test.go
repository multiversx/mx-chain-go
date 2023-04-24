package preprocess

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/storage/txcache"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
)

func createTransactionPreprocessor() *transactions {
	dataPool := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txPreProcArgs := ArgsTransactionPreProcessor{
		TxDataPool:           dataPool.Transactions(),
		Store:                &storageStubs.ChainStorerStub{},
		Hasher:               &hashingMocks.HasherMock{},
		Marshalizer:          &mock.MarshalizerMock{},
		TxProcessor:          &testscommon.TxProcessorMock{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(3),
		Accounts:             &stateMock.AccountsStub{},
		OnRequestTransaction: requestTransaction,
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			MaxGasLimitPerMiniBlockForSafeCrossShardCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
			MaxGasLimitPerBlockForSafeCrossShardCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
			MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
				return MaxGasLimitPerBlock
			},
			MaxGasLimitPerTxCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
		},
		GasHandler:           &mock.GasHandlerMock{},
		BlockTracker:         &mock.BlockTrackerMock{},
		BlockType:            block.TxBlock,
		PubkeyConverter:      createMockPubkeyConverter(),
		BlockSizeComputation: &testscommon.BlockSizeComputationStub{},
		BalanceComputation: &testscommon.BalanceComputationStub{
			IsAddressSetCalled: func(address []byte) bool {
				return true
			},
			SubBalanceFromAddressCalled: func(address []byte, value *big.Int) bool {
				return false
			},
		},
		EnableEpochsHandler: &testscommon.EnableEpochsHandlerStub{},
		TxTypeHandler: &testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				if bytes.Equal(tx.GetRcvAddr(), []byte("smart contract address")) {
					return process.MoveBalance, process.SCInvoking
				}
				return process.MoveBalance, process.MoveBalance
			},
		},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}

	preprocessor, _ := NewTransactionPreprocessor(txPreProcArgs)

	return preprocessor
}

func TestTransactions_ApplyVerifiedTransactionShouldWork(t *testing.T) {
	t.Parallel()

	numMiniBlocks := 0
	numTxs := 0
	preprocessor := createTransactionPreprocessor()
	preprocessor.blockSizeComputation = &testscommon.BlockSizeComputationStub{
		AddNumTxsCalled: func(i int) {
			numTxs += i
		},
		AddNumMiniBlocksCalled: func(i int) {
			numMiniBlocks += i
		},
	}

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

	preprocessor := createTransactionPreprocessor()

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
	preprocessor := createTransactionPreprocessor()
	preprocessor.balanceComputation = &testscommon.BalanceComputationStub{
		IsAddressSetCalled: func(address []byte) bool {
			return true
		},
		AddressHasEnoughBalanceCalled: func(address []byte, value *big.Int) bool {
			return addressHasEnoughBalance
		},
	}

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

func TestTransactions_ApplyExecutedTransactionShouldWork(t *testing.T) {
	t.Parallel()

	numMiniBlocks := 0
	numTxs := 0
	preprocessor := createTransactionPreprocessor()
	preprocessor.blockSizeComputation = &testscommon.BlockSizeComputationStub{
		AddNumTxsCalled: func(i int) {
			numTxs += i
		},
		AddNumMiniBlocksCalled: func(i int) {
			numMiniBlocks += i
		},
	}

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
	preprocessor := createTransactionPreprocessor()
	preprocessor.blockSizeComputation = &testscommon.BlockSizeComputationStub{
		AddNumTxsCalled: func(i int) {
			numTxs += i
		},
		AddNumMiniBlocksCalled: func(i int) {
			numMiniBlocks += i
		},
	}

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
}

func TestTransactions_ShouldContinueProcessingTxShouldWork(t *testing.T) {
	t.Parallel()

	addressHasEnoughBalance := false
	preprocessor := createTransactionPreprocessor()
	preprocessor.balanceComputation = &testscommon.BalanceComputationStub{
		IsAddressSetCalled: func(address []byte) bool {
			return true
		},
		AddressHasEnoughBalanceCalled: func(address []byte, value *big.Int) bool {
			return addressHasEnoughBalance
		},
	}

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

	var gasProvidedByTx uint64
	var verifyTransactionErr error
	preprocessor := createTransactionPreprocessor()
	preprocessor.txProcessor = &testscommon.TxProcessorMock{
		VerifyTransactionCalled: func(tx *transaction.Transaction) error {
			return verifyTransactionErr
		},
	}
	preprocessor.gasHandler = &mock.GasHandlerMock{
		ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return 0, gasProvidedByTx, nil
		},
	}

	tx := &transaction.Transaction{}
	txHash := []byte("hash")
	senderShardID := uint32(0)
	receiverShardID := uint32(1)

	// should err when computeGasProvided method returns err
	mbInfo := &createScheduledMiniBlocksInfo{}
	gasProvidedByTx = MaxGasLimitPerBlock + 1
	err := preprocessor.verifyTransaction(tx, txHash, senderShardID, receiverShardID, mbInfo)
	assert.Equal(t, process.ErrMaxGasLimitPerOneTxInReceiverShardIsReached, err)
	assert.Equal(t, 1, mbInfo.schedulingInfo.numCrossShardTxsWithTooMuchGas)

	// should err when VerifyTransaction method returns err
	mbInfo = &createScheduledMiniBlocksInfo{
		mapGasConsumedByMiniBlockInReceiverShard: make(map[uint32]uint64),
	}
	gasProvidedByTx = MaxGasLimitPerBlock - 1
	verifyTransactionErr = process.ErrLowerNonceInTransaction
	mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = 0
	err = preprocessor.verifyTransaction(tx, txHash, senderShardID, receiverShardID, mbInfo)
	assert.Equal(t, process.ErrLowerNonceInTransaction, err)
	assert.Equal(t, 1, mbInfo.schedulingInfo.numScheduledBadTxs)

	// should work
	mbInfo = &createScheduledMiniBlocksInfo{
		mapGasConsumedByMiniBlockInReceiverShard: make(map[uint32]uint64),
	}
	gasProvidedByTx = MaxGasLimitPerBlock - 1
	verifyTransactionErr = nil
	mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = 0
	err = preprocessor.verifyTransaction(tx, txHash, senderShardID, receiverShardID, mbInfo)
	assert.Nil(t, err)
	_, ok := preprocessor.txsForCurrBlock.txHashAndInfo[string(txHash)]
	assert.True(t, ok)
}

func TestTransactions_CreateScheduledMiniBlocksShouldWork(t *testing.T) {
	t.Parallel()

	var gasProvidedByTx uint64
	preprocessor := createTransactionPreprocessor()
	preprocessor.gasHandler = &mock.GasHandlerMock{
		ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return 0, gasProvidedByTx, nil
		},
	}

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
	gasProvidedByTx = MaxGasLimitPerBlock + 1
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
	gasProvidedByTx = MaxGasLimitPerBlock
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

	preprocessor := createTransactionPreprocessor()

	mapMiniBlocks := make(map[uint32]*block.MiniBlock)

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
	mbs := preprocessor.getMiniBlockSliceFromMapV2(mapMiniBlocks)
	assert.Equal(t, 4, len(mbs))
}

func TestTransactions_CreateAndProcessMiniBlocksFromMeV2ShouldWork(t *testing.T) {
	t.Parallel()

	var gasProvidedByTx uint64
	preprocessor := createTransactionPreprocessor()
	preprocessor.gasHandler = &mock.GasHandlerMock{
		ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return 0, gasProvidedByTx, nil
		},
	}

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

	mbs, _, mapSCTxs, _ := preprocessor.createAndProcessMiniBlocksFromMeV2(haveTimeMethod, isShardStuckMethod, isMaxBlockSizeReachedMethod, sortedTxs)
	assert.Equal(t, 0, len(mbs))
	assert.Equal(t, 0, len(mapSCTxs))

	// should not create and process mini blocks when max block size is reached
	haveTimeMethodReturn = true
	isShardStuckMethodReturn = false
	isMaxBlockSizeReachedMethodReturn = true
	sortedTxs = make([]*txcache.WrappedTransaction, 0)
	tx = &txcache.WrappedTransaction{
		Tx: &transaction.Transaction{RcvAddr: []byte("smart contract address")},
	}
	sortedTxs = append(sortedTxs, tx)

	mbs, _, mapSCTxs, _ = preprocessor.createAndProcessMiniBlocksFromMeV2(haveTimeMethod, isShardStuckMethod, isMaxBlockSizeReachedMethod, sortedTxs)
	assert.Equal(t, 0, len(mbs))
	assert.Equal(t, 0, len(mapSCTxs))

	// should not create and process mini blocks when processTransaction returns error
	gasProvidedByTx = MaxGasLimitPerBlock + 1
	haveTimeMethodReturn = true
	isShardStuckMethodReturn = false
	isMaxBlockSizeReachedMethodReturn = false
	sortedTxs = make([]*txcache.WrappedTransaction, 0)
	tx = &txcache.WrappedTransaction{
		Tx: &transaction.Transaction{RcvAddr: []byte("smart contract address")},
	}
	sortedTxs = append(sortedTxs, tx)

	mbs, _, mapSCTxs, _ = preprocessor.createAndProcessMiniBlocksFromMeV2(haveTimeMethod, isShardStuckMethod, isMaxBlockSizeReachedMethod, sortedTxs)
	assert.Equal(t, 0, len(mbs))
	assert.Equal(t, 0, len(mapSCTxs))

	// should create and process two mini blocks
	gasProvidedByTx = MaxGasLimitPerBlock
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

	mbs, _, mapSCTxs, _ = preprocessor.createAndProcessMiniBlocksFromMeV2(haveTimeMethod, isShardStuckMethod, isMaxBlockSizeReachedMethod, sortedTxs)
	assert.Equal(t, 3, len(mbs))
	assert.Equal(t, 2, len(mapSCTxs))
}

func TestTransactions_ProcessTransactionShouldWork(t *testing.T) {
	t.Parallel()

	var gasProvidedByTx uint64
	var processTransactionErr error
	preprocessor := createTransactionPreprocessor()
	preprocessor.txProcessor = &testscommon.TxProcessorMock{
		ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
			return vmcommon.UserError, processTransactionErr
		},
	}
	preprocessor.gasHandler = &mock.GasHandlerMock{
		ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return 0, gasProvidedByTx, nil
		},
	}

	tx := &transaction.Transaction{}
	txHash := []byte("txHash")
	senderShardID := uint32(0)
	receiverShardID := uint32(0)

	// should not process transaction when computeGasProvided method returns error
	mbInfo := &createAndProcessMiniBlocksInfo{
		mapGasConsumedByMiniBlockInReceiverShard: make(map[uint32]uint64),
	}
	mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = 0
	gasProvidedByTx = MaxGasLimitPerBlock + 1
	_, err := preprocessor.processTransaction(tx, txHash, senderShardID, receiverShardID, mbInfo)
	assert.Equal(t, process.ErrMaxGasLimitPerOneTxInReceiverShardIsReached, err)

	// should not process transaction when processAndRemoveBadTransaction method returns error
	mbInfo = &createAndProcessMiniBlocksInfo{
		mapGasConsumedByMiniBlockInReceiverShard: make(map[uint32]uint64),
	}
	mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = 0
	gasProvidedByTx = MaxGasLimitPerBlock - 1
	processTransactionErr = process.ErrHigherNonceInTransaction
	_, err = preprocessor.processTransaction(tx, txHash, senderShardID, receiverShardID, mbInfo)
	assert.Equal(t, process.ErrHigherNonceInTransaction, err)

	// should process transaction bat should return ErrFailedTransaction
	mbInfo = &createAndProcessMiniBlocksInfo{
		mapGasConsumedByMiniBlockInReceiverShard: make(map[uint32]uint64),
	}
	mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = 0
	processTransactionErr = process.ErrFailedTransaction
	_, err = preprocessor.processTransaction(tx, txHash, senderShardID, receiverShardID, mbInfo)
	assert.Equal(t, process.ErrFailedTransaction, err)

	// should process transaction
	mbInfo = &createAndProcessMiniBlocksInfo{
		mapGasConsumedByMiniBlockInReceiverShard: make(map[uint32]uint64),
	}
	mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = 0
	processTransactionErr = nil
	_, err = preprocessor.processTransaction(tx, txHash, senderShardID, receiverShardID, mbInfo)
	assert.Nil(t, err)
}
