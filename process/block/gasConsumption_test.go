package block_test

// import (
// 	"testing"
//
// 	"github.com/multiversx/mx-chain-core-go/data"
// 	"github.com/multiversx/mx-chain-core-go/data/transaction"
// 	"github.com/multiversx/mx-chain-go/process/block"
// 	"github.com/multiversx/mx-chain-go/process/mock"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
//
// 	"github.com/multiversx/mx-chain-go/process"
// 	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
// )
//
// const (
// 	maxGasLimitPerBlock     = uint64(400)
// 	maxGasLimitPerMiniBlock = uint64(100)
// 	maxGasLimitPerTx        = uint64(10)
// )
//
// func getTxs(
// 	gasLimitPerTx uint64,
// 	numTxs int,
// ) []data.TransactionHandler {
// 	txs := make([]data.TransactionHandler, 0, numTxs)
// 	for i := 0; i < numTxs; i++ {
// 		txs = append(txs, &transaction.Transaction{
// 			GasLimit: gasLimitPerTx,
// 		})
// 	}
//
// 	return txs
// }
//
// func TestNewGasConsumption(t *testing.T) {
// 	t.Parallel()
//
// 	t.Run("nil economics handler should error", func(t *testing.T) {
// 		t.Parallel()
//
// 		gc, err := block.NewGasConsumption(nil, &mock.ShardCoordinatorStub{})
// 		assert.Nil(t, gc)
// 		assert.Equal(t, process.ErrNilEconomicsData, err)
// 	})
// 	t.Run("nil economics handler should error", func(t *testing.T) {
// 		t.Parallel()
//
// 		gc, err := block.NewGasConsumption(&economicsmocks.EconomicsHandlerMock{}, nil)
// 		assert.Nil(t, gc)
// 		assert.Equal(t, process.ErrNilShardCoordinator, err)
// 	})
// 	t.Run("should work", func(t *testing.T) {
// 		t.Parallel()
//
// 		gc, err := block.NewGasConsumption(&economicsmocks.EconomicsHandlerMock{}, &mock.ShardCoordinatorStub{})
// 		assert.NotNil(t, gc)
// 		assert.NoError(t, err)
// 	})
// }
//
// func TestGasConsumption_IsInterfaceNil(t *testing.T) {
// 	t.Parallel()
//
// 	gc, _ := block.NewGasConsumption(nil, &mock.ShardCoordinatorStub{})
// 	assert.True(t, gc.IsInterfaceNil())
//
// 	gc, _ = block.NewGasConsumption(&economicsmocks.EconomicsHandlerMock{}, &mock.ShardCoordinatorStub{})
// 	assert.False(t, gc.IsInterfaceNil())
// }
//
// func TestGasConsumption_CheckTransactionsForMiniBlock(t *testing.T) {
// 	t.Parallel()
//
// 	economicsFee := &economicsmocks.EconomicsHandlerMock{
// 		MaxGasLimitPerBlockCalled: func(shardID uint32) uint64 {
// 			return maxGasLimitPerBlock
// 		},
// 		MaxGasLimitPerMiniBlockCalled: func(shardID uint32) uint64 {
// 			return maxGasLimitPerMiniBlock
// 		},
// 		MaxGasLimitPerTxCalled: func() uint64 {
// 			return maxGasLimitPerTx
// 		},
// 	}
//
// 	shardCoordinator := &mock.ShardCoordinatorStub{
// 		NumberOfShardsCalled: func() uint32 {
// 			return 2
// 		},
// 	}
// 	gc, _ := block.NewGasConsumption(economicsFee, shardCoordinator)
// 	require.NotNil(t, gc)
//
// 	// empty slice of txs, coverage only
// 	lastIndex, err := gc.CheckTransactionsForMiniBlock([]byte("hash"), getTxs(0, 0))
// 	require.NoError(t, err)
// 	require.Equal(t, int32(-1), lastIndex)
//
// 	// gas exceeded per tx
// 	txs := getTxs(maxGasLimitPerTx+1, 1)
// 	lastIndex, err = gc.CheckTransactionsForMiniBlock([]byte("hash"), txs)
// 	require.Equal(t, process.ErrInvalidMaxGasLimitPerTx, err)
// 	require.Equal(t, int32(-1), lastIndex)
//
// 	// gas exceeded per mb
// 	txsToExceedMb := getTxs(maxGasLimitPerTx, 11)
// 	lastIndex, err = gc.CheckTransactionsForMiniBlock([]byte("hash"), txsToExceedMb)
// 	require.Equal(t, process.ErrMaxGasLimitPerMiniBlockIsReached, err)
// 	require.Equal(t, int32(9), lastIndex)
//
// 	// reset state
// 	gc.Reset()
//
// 	// gas exceeded intra shard
// 	txsMb1 := getTxs(maxGasLimitPerTx, 10)
// 	txsMb2 := getTxs(maxGasLimitPerTx, 9)
// 	txsMb3 := getTxs(maxGasLimitPerTx, 2)
// 	txsMb4 := getTxs(maxGasLimitPerTx, 1)
// 	lastIndex, err = gc.CheckTransactionsForMiniBlock([]byte("mb1"), txsMb1)
// 	require.NoError(t, err)
// 	require.Equal(t, int32(9), lastIndex) // added all
// 	lastIndex, err = gc.CheckTransactionsForMiniBlock([]byte("mb2"), txsMb2)
// 	require.NoError(t, err)
// 	require.Equal(t, int32(8), lastIndex) // added all
// 	lastIndex, err = gc.CheckTransactionsForMiniBlock([]byte("mb3"), txsMb3)
// 	require.Equal(t, process.ErrMaxGasLimitPerBlockIsReached, err)
// 	require.Equal(t, int32(0), lastIndex) // limit reached, added one
// 	lastIndex, err = gc.CheckTransactionsForMiniBlock([]byte("mb4"), txsMb4)
// 	require.Equal(t, process.ErrMaxGasLimitPerBlockIsReached, err)
// 	require.Equal(t, int32(-1), lastIndex) // limit reached, nothing will be added
//
// 	// reset state
// 	gc.Reset()
//
// 	// different shard should work
// 	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
// 		return 0
// 	}
// 	txsMb1Sh0 := getTxs(maxGasLimitPerTx, 10)
// 	txsMb2Sh0 := getTxs(maxGasLimitPerTx, 10)
// 	lastIndex, err = gc.CheckTransactionsForMiniBlock([]byte("mb1sh0"), txsMb1Sh0)
// 	require.NoError(t, err)
// 	require.Equal(t, int32(9), lastIndex) // added all
// 	lastIndex, err = gc.CheckTransactionsForMiniBlock([]byte("mb2sh0"), txsMb2Sh0)
// 	require.NoError(t, err)
// 	require.Equal(t, int32(9), lastIndex) // added all
//
// 	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
// 		return 1
// 	}
// 	txsMb1Sh1 := getTxs(maxGasLimitPerTx, 10)
// 	lastIndex, err = gc.CheckTransactionsForMiniBlock([]byte("mb1sh1"), txsMb1Sh1)
// 	require.NoError(t, err)
// 	require.Equal(t, int32(9), lastIndex) // added all, different shard
//
// 	// reset state
// 	gc.Reset()
//
// 	// should allow more cross shard if intra is not filled
// 	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
// 		return 0
// 	}
// 	txsMb1Sh0 = getTxs(maxGasLimitPerTx, 10)
// 	lastIndex, err = gc.CheckTransactionsForMiniBlock([]byte("mb1sh0"), txsMb1Sh0)
// 	require.NoError(t, err)
// 	require.Equal(t, int32(9), lastIndex) // added all
//
// 	// only 100/400 consumed so far, remaining will be used for cross
//
// 	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
// 		return 1
// 	}
// 	txsMb1Sh1 = getTxs(maxGasLimitPerTx, 10)
// 	lastIndex, err = gc.CheckTransactionsForMiniBlock([]byte("mb1sh1"), txsMb1Sh1)
// 	require.NoError(t, err)
// 	require.Equal(t, int32(9), lastIndex) // added all
//
// 	txsMb2Sh1 := getTxs(maxGasLimitPerTx, 10)
// 	lastIndex, err = gc.CheckTransactionsForMiniBlock([]byte("mb2sh1"), txsMb2Sh1)
// 	require.NoError(t, err)
// 	require.Equal(t, int32(9), lastIndex) // added all
//
// 	txsMb3Sh1 := getTxs(maxGasLimitPerTx, 10)
// 	lastIndex, err = gc.CheckTransactionsForMiniBlock([]byte("mb3sh1"), txsMb3Sh1)
// 	require.NoError(t, err)
// 	require.Equal(t, int32(9), lastIndex) // added all
// }
//
// func TestGasConsumption_DecreaseLimits(t *testing.T) {
// 	t.Parallel()
//
// 	economicsFee := &economicsmocks.EconomicsHandlerMock{
// 		MaxGasLimitPerBlockCalled: func(shardID uint32) uint64 {
// 			return maxGasLimitPerBlock
// 		},
// 		MaxGasLimitPerMiniBlockCalled: func(shardID uint32) uint64 {
// 			return maxGasLimitPerMiniBlock
// 		},
// 		MaxGasLimitPerTxCalled: func() uint64 {
// 			return maxGasLimitPerTx
// 		},
// 	}
//
// 	shardCoordinator := &mock.ShardCoordinatorStub{
// 		NumberOfShardsCalled: func() uint32 {
// 			return 2
// 		},
// 	}
// 	gc, _ := block.NewGasConsumption(economicsFee, shardCoordinator)
// 	require.NotNil(t, gc)
//
// 	// decrease limits 5 times, so they will be half
// 	gc.DecreaseLimits()
// 	gc.DecreaseLimits()
// 	gc.DecreaseLimits()
// 	gc.DecreaseLimits()
// 	gc.DecreaseLimits()
//
// 	// gas exceeded per mb
// 	txsToExceedMb := getTxs(maxGasLimitPerTx, 6) // 6 should be enough as the limit is now 50 / mb
// 	lastIndex, err := gc.CheckTransactionsForMiniBlock([]byte("hash"), txsToExceedMb)
// 	require.Equal(t, process.ErrMaxGasLimitPerMiniBlockIsReached, err)
// 	require.Equal(t, int32(4), lastIndex)
//
// 	// reset state
// 	gc.Reset()
//
// 	// go over the limit, should stay at min value
// 	for i := 0; i < 20; i++ {
// 		gc.DecreaseLimits()
// 	}
//
// 	lastIndex, err = gc.CheckTransactionsForMiniBlock([]byte("hash"), getTxs(maxGasLimitPerTx, 2))
// 	require.Equal(t, process.ErrMaxGasLimitPerMiniBlockIsReached, err)
// 	require.Equal(t, int32(0), lastIndex)
// }
//
// func TestGasConsumption_GetGasConsumedByMiniBlock(t *testing.T) {
// 	t.Parallel()
//
// 	economicsFee := &economicsmocks.EconomicsHandlerMock{
// 		MaxGasLimitPerBlockCalled: func(shardID uint32) uint64 {
// 			return maxGasLimitPerBlock
// 		},
// 		MaxGasLimitPerMiniBlockCalled: func(shardID uint32) uint64 {
// 			return maxGasLimitPerMiniBlock
// 		},
// 		MaxGasLimitPerTxCalled: func() uint64 {
// 			return maxGasLimitPerTx
// 		},
// 	}
//
// 	shardCoordinator := &mock.ShardCoordinatorStub{
// 		NumberOfShardsCalled: func() uint32 {
// 			return 2
// 		},
// 	}
// 	gc, _ := block.NewGasConsumption(economicsFee, shardCoordinator)
// 	require.NotNil(t, gc)
//
// 	// gas exceeded intra shard
// 	txsMb1 := getTxs(maxGasLimitPerTx, 10)
// 	lastIndex, err := gc.CheckTransactionsForMiniBlock([]byte("mb1"), txsMb1)
// 	require.NoError(t, err)
// 	require.Equal(t, int32(9), lastIndex)
//
// 	gasConsumed := gc.GetGasConsumedByMiniBlock([]byte("mb1"))
// 	require.Equal(t, uint64(100), gasConsumed)
//
// 	gasConsumed = gc.GetGasConsumedByMiniBlock([]byte("mb2"))
// 	require.Zero(t, gasConsumed)
// }
