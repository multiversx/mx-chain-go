package preprocess

import (
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
)

const testTxHash = "tx1_hash"

func TestNewRewardTxPreprocessor_NilRewardTxDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	rtp, err := NewRewardTxPreprocessor(
		nil,
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilRewardTxDataPool, err)
}

func TestNewRewardTxPreprocessor_NilStoreShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.Transactions(),
		nil,
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestNewRewardTxPreprocessor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.Transactions(),
		&storageStubs.ChainStorerStub{},
		nil,
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewRewardTxPreprocessor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		nil,
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewRewardTxPreprocessor_NilRewardTxProcessorShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		nil,
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilRewardsTxProcessor, err)
}

func TestNewRewardTxPreprocessor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		nil,
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewRewardTxPreprocessor_NilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		nil,
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewRewardTxPreprocessor_NilRequestHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		nil,
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilRequestHandler, err)
}

func TestNewRewardTxPreprocessor_NilGasHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		nil,
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilGasHandler, err)
}

func TestNewRewardTxPreprocessor_NilPubkeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		nil,
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestNewRewardTxPreprocessor_NilBlockSizeComputationHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		nil,
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilBlockSizeComputationHandler, err)
}

func TestNewRewardTxPreprocessor_NilBalanceComputationHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		nil,
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilBalanceComputationHandler, err)
}

func TestNewRewardTxPreprocessor_NilProcessedMiniBlocksTrackerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		nil,
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilProcessedMiniBlocksTracker, err)
}

func TestNewRewardTxPreprocessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)
	assert.Nil(t, err)
	assert.NotNil(t, rtp)
}

func TestRewardTxPreprocessor_CreateMarshalizedDataShouldWork(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	txHashes := [][]byte{[]byte(txHash)}
	txs := []data.TransactionHandler{&rewardTx.RewardTx{}}
	rtp.AddTxs(txHashes, txs)

	res, err := rtp.CreateMarshalledData(txHashes)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(res))
}

func TestRewardTxPreprocessor_ProcessMiniBlockInvalidMiniBlockTypeShouldErr(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	txHashes := [][]byte{[]byte(txHash)}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            0,
	}

	preProcessorExecutionInfoHandlerMock := &testscommon.PreProcessorExecutionInfoHandlerMock{
		GetNumOfCrossInterMbsAndTxsCalled: getNumOfCrossInterMbsAndTxsZero,
	}

	_, _, _, err := rtp.ProcessMiniBlock(&mb1, haveTimeTrue, haveAdditionalTimeFalse, false, false, -1, preProcessorExecutionInfoHandlerMock)
	assert.Equal(t, process.ErrWrongTypeInMiniBlock, err)
}

func TestRewardTxPreprocessor_ProcessMiniBlockShouldWork(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	txHashes := [][]byte{[]byte(txHash)}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   core.MetachainShardId,
		Type:            block.RewardsBlock,
	}

	txs := []data.TransactionHandler{&rewardTx.RewardTx{}}
	rtp.AddTxs(txHashes, txs)

	preProcessorExecutionInfoHandlerMock := &testscommon.PreProcessorExecutionInfoHandlerMock{
		GetNumOfCrossInterMbsAndTxsCalled: getNumOfCrossInterMbsAndTxsZero,
	}

	_, _, _, err := rtp.ProcessMiniBlock(&mb1, haveTimeTrue, haveAdditionalTimeFalse, false, false, -1, preProcessorExecutionInfoHandlerMock)
	assert.Nil(t, err)

	txsMap := rtp.GetAllCurrentUsedTxs()
	if _, ok := txsMap[txHash]; !ok {
		assert.Fail(t, "miniblock was not added")
	}
}

func TestRewardTxPreprocessor_ProcessMiniBlockNotFromMeta(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	txHashes := [][]byte{[]byte(txHash)}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.RewardsBlock,
	}

	txs := []data.TransactionHandler{&rewardTx.RewardTx{}}
	rtp.AddTxs(txHashes, txs)

	preProcessorExecutionInfoHandlerMock := &testscommon.PreProcessorExecutionInfoHandlerMock{
		GetNumOfCrossInterMbsAndTxsCalled: getNumOfCrossInterMbsAndTxsZero,
	}

	_, _, _, err := rtp.ProcessMiniBlock(&mb1, haveTimeTrue, haveAdditionalTimeFalse, false, false, -1, preProcessorExecutionInfoHandlerMock)
	assert.Equal(t, process.ErrRewardMiniBlockNotFromMeta, err)
}

func TestRewardTxPreprocessor_SaveTxsToStorageShouldWork(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	txHashes := [][]byte{[]byte(txHash)}
	txs := []data.TransactionHandler{&rewardTx.RewardTx{}}
	rtp.AddTxs(txHashes, txs)

	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.RewardsBlock,
	}
	mb2 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 0,
		SenderShardID:   1,
		Type:            block.RewardsBlock,
	}

	blockBody := &block.Body{}
	blockBody.MiniBlocks = append(blockBody.MiniBlocks, &mb1, &mb2)
	err := rtp.SaveTxsToStorage(blockBody)

	assert.Nil(t, err)
}

func TestRewardTxPreprocessor_RequestBlockTransactionsNoMissingTxsShouldWork(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	txHashes := [][]byte{[]byte(txHash)}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.RewardsBlock,
	}
	mb2 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 0,
		SenderShardID:   1,
		Type:            block.RewardsBlock,
	}

	blockBody := &block.Body{}
	blockBody.MiniBlocks = append(blockBody.MiniBlocks, &mb1, &mb2)

	_ = rtp.SaveTxsToStorage(blockBody)

	res := rtp.RequestBlockTransactions(blockBody)
	assert.Equal(t, 0, res)
}

func TestRewardTxPreprocessor_RequestTransactionsForMiniBlockShouldWork(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	txHashes := [][]byte{[]byte(txHash)}
	mb1 := &block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.RewardsBlock,
	}

	res := rtp.RequestTransactionsForMiniBlock(mb1)
	assert.Equal(t, 0, res)
}

func TestRewardTxPreprocessor_ProcessBlockTransactions(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	txHashes := [][]byte{[]byte(txHash)}
	txs := []data.TransactionHandler{&rewardTx.RewardTx{}}
	rtp.AddTxs(txHashes, txs)

	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.RewardsBlock,
	}
	mb2 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 0,
		SenderShardID:   1,
		Type:            block.RewardsBlock,
	}

	mbHash1, _ := core.CalculateHash(rtp.marshalizer, rtp.hasher, &mb1)
	mbHash2, _ := core.CalculateHash(rtp.marshalizer, rtp.hasher, &mb2)

	var blockBody block.Body
	blockBody.MiniBlocks = append(blockBody.MiniBlocks, &mb1, &mb2)

	err := rtp.ProcessBlockTransactions(&block.Header{MiniBlockHeaders: []block.MiniBlockHeader{{TxCount: 1, Hash: mbHash1}, {TxCount: 1, Hash: mbHash2}}}, &blockBody, haveTimeTrue)
	assert.Nil(t, err)
}

func TestRewardTxPreprocessor_ProcessBlockTransactionsMissingTrieNode(t *testing.T) {
	t.Parallel()

	missingNodeErr := fmt.Errorf(core.GetNodeFromDBErrorString)
	txHash := testTxHash
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{
			GetExistingAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
				return nil, missingNodeErr
			},
		},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	txHashes := [][]byte{[]byte(txHash)}
	txs := []data.TransactionHandler{&rewardTx.RewardTx{}}
	rtp.AddTxs(txHashes, txs)

	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.RewardsBlock,
	}
	mb2 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 0,
		SenderShardID:   1,
		Type:            block.RewardsBlock,
	}

	mbHash1, _ := core.CalculateHash(rtp.marshalizer, rtp.hasher, &mb1)
	mbHash2, _ := core.CalculateHash(rtp.marshalizer, rtp.hasher, &mb2)

	var blockBody block.Body
	blockBody.MiniBlocks = append(blockBody.MiniBlocks, &mb1, &mb2)

	err := rtp.ProcessBlockTransactions(&block.Header{MiniBlockHeaders: []block.MiniBlockHeader{{TxCount: 1, Hash: mbHash1}, {TxCount: 1, Hash: mbHash2}}}, &blockBody, haveTimeTrue)
	assert.Equal(t, missingNodeErr, err)
}

func TestRewardTxPreprocessor_IsDataPreparedShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	err := rtp.IsDataPrepared(1, haveTime)

	assert.Equal(t, process.ErrTimeIsOut, err)
}

func TestRewardTxPreprocessor_IsDataPrepared(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	go func() {
		time.Sleep(50 * time.Millisecond)
		rtp.chReceivedAllRewardTxs <- true
	}()

	err := rtp.IsDataPrepared(1, haveTime)

	assert.Nil(t, err)
}

func TestRewardTxPreprocessor_RestoreBlockDataIntoPools(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	storer := storageStubs.ChainStorerStub{
		GetAllCalled: func(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
			retMap := map[string][]byte{
				"tx_hash1": []byte(`{"Round": 0}`),
			}

			return retMap, nil
		},
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				RemoveCalled: func(key []byte) error {
					return nil
				},
			}, nil
		},
	}
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storer,
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	txHashes := [][]byte{[]byte("tx_hash1")}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.RewardsBlock,
	}

	blockBody := &block.Body{}
	blockBody.MiniBlocks = append(blockBody.MiniBlocks, &mb1)
	miniBlockPool := testscommon.NewCacherMock()

	numRestoredTxs, err := rtp.RestoreBlockDataIntoPools(blockBody, miniBlockPool)
	assert.Equal(t, 1, numRestoredTxs)
	assert.Nil(t, err)
}

func TestRewardTxPreprocessor_CreateAndProcessMiniBlocksShouldWork(t *testing.T) {
	t.Parallel()

	totalGasProvided := uint64(0)
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{
			InitCalled: func() {
				totalGasProvided = 0
			},
			TotalGasProvidedCalled: func() uint64 {
				return totalGasProvided
			},
		},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	mBlocksSlice, err := rtp.CreateAndProcessMiniBlocks(haveTimeTrue, []byte("randomness"))
	assert.NotNil(t, mBlocksSlice)
	assert.Nil(t, err)
}

func TestRewardTxPreprocessor_CreateBlockStartedShouldCleanMap(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	rtp.CreateBlockStarted()
	assert.Equal(t, 0, len(rtp.rewardTxsForBlock.txHashAndInfo))
}
