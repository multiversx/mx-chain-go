package track_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	dataRetrieverMock "github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
	"github.com/stretchr/testify/assert"
)

func TestNewMiniBlockTrack_NilDataPoolHolderErr(t *testing.T) {
	t.Parallel()

	mbt, err := track.NewMiniBlockTrack(nil, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	assert.Nil(t, mbt)
	assert.Equal(t, process.ErrNilPoolsHolder, err)
}

func TestNewMiniBlockTrack_NilTxsPoolErr(t *testing.T) {
	t.Parallel()

	dataPool := &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return nil
		},
	}
	mbt, err := track.NewMiniBlockTrack(dataPool, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	assert.Nil(t, mbt)
	assert.Equal(t, process.ErrNilTransactionPool, err)
}

func TestNewMiniBlockTrack_NilRewardTxsPoolErr(t *testing.T) {
	t.Parallel()

	dataPool := &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return nil
		},
	}
	mbt, err := track.NewMiniBlockTrack(dataPool, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	assert.Nil(t, mbt)
	assert.Equal(t, process.ErrNilRewardTxDataPool, err)
}

func TestNewMiniBlockTrack_NilUnsignedTxsPoolErr(t *testing.T) {
	t.Parallel()

	dataPool := &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return nil
		},
	}
	mbt, err := track.NewMiniBlockTrack(dataPool, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	assert.Nil(t, mbt)
	assert.Equal(t, process.ErrNilUnsignedTxDataPool, err)
}

func TestNewMiniBlockTrack_NilMiniBlockPoolShouldErr(t *testing.T) {
	t.Parallel()

	dataPool := &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		MiniBlocksCalled: func() storage.Cacher {
			return nil
		},
	}
	mbt, err := track.NewMiniBlockTrack(dataPool, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	assert.Nil(t, mbt)
	assert.Equal(t, process.ErrNilMiniBlockPool, err)
}

func TestNewMiniBlockTrack_NilShardCoordinatorErr(t *testing.T) {
	t.Parallel()

	dataPool := createDataPool()
	miniBlockTrack, err := track.NewMiniBlockTrack(dataPool, nil, &testscommon.WhiteListHandlerStub{})

	assert.Nil(t, miniBlockTrack)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewMiniBlockTrack_NilWhitelistHandlerErr(t *testing.T) {
	t.Parallel()

	dataPool := createDataPool()
	miniBlockTrack, err := track.NewMiniBlockTrack(dataPool, mock.NewMultipleShardsCoordinatorMock(), nil)

	assert.Nil(t, miniBlockTrack)
	assert.Equal(t, process.ErrNilWhiteListHandler, err)
}

func TestNewMiniBlockTrack_ShouldWork(t *testing.T) {
	t.Parallel()

	dataPool := createDataPool()
	mbt, err := track.NewMiniBlockTrack(dataPool, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	assert.Nil(t, err)
	assert.NotNil(t, mbt)
}

func TestReceivedMiniBlock_ShouldReturnIfKeyIsNil(t *testing.T) {
	t.Parallel()

	dataPool := createDataPool()
	mbt, _ := track.NewMiniBlockTrack(dataPool, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	wasCalled := false
	blockTransactionsPool := &testscommon.ShardedDataStub{
		ImmunizeSetOfDataAgainstEvictionCalled: func(keys [][]byte, destCacheId string) {
			wasCalled = true
		},
	}
	mbt.SetBlockTransactionsPool(blockTransactionsPool)
	mbt.ReceivedMiniBlock(nil, nil)

	assert.False(t, wasCalled)
}

func TestReceivedMiniBlock_ShouldReturnIfWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	dataPool := createDataPool()
	mbt, _ := track.NewMiniBlockTrack(dataPool, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	wasCalled := false
	blockTransactionsPool := &testscommon.ShardedDataStub{
		ImmunizeSetOfDataAgainstEvictionCalled: func(keys [][]byte, destCacheId string) {
			wasCalled = true
		},
	}
	mbt.SetBlockTransactionsPool(blockTransactionsPool)
	mbt.ReceivedMiniBlock([]byte("mb_hash"), nil)

	assert.False(t, wasCalled)
}

func TestReceivedMiniBlock_ShouldReturnIfMiniBlockIsNotCrossShardDestMe(t *testing.T) {
	t.Parallel()

	dataPool := createDataPool()
	mbt, _ := track.NewMiniBlockTrack(dataPool, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	wasCalled := false
	blockTransactionsPool := &testscommon.ShardedDataStub{
		ImmunizeSetOfDataAgainstEvictionCalled: func(keys [][]byte, destCacheId string) {
			wasCalled = true
		},
	}
	mbt.SetBlockTransactionsPool(blockTransactionsPool)
	mbt.ReceivedMiniBlock([]byte("mb_hash"), &block.MiniBlock{})

	assert.False(t, wasCalled)
}

func TestReceivedMiniBlock_ShouldReturnIfMiniBlockTypeIsWrong(t *testing.T) {
	t.Parallel()

	dataPool := createDataPool()
	mbt, _ := track.NewMiniBlockTrack(dataPool, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	wasCalled := false
	blockTransactionsPool := &testscommon.ShardedDataStub{
		ImmunizeSetOfDataAgainstEvictionCalled: func(keys [][]byte, destCacheId string) {
			wasCalled = true
		},
	}
	mbt.SetBlockTransactionsPool(blockTransactionsPool)
	mbt.ReceivedMiniBlock(
		[]byte("mb_hash"),
		&block.MiniBlock{
			SenderShardID: 1,
			Type:          block.PeerBlock,
		})

	assert.False(t, wasCalled)
}

func TestReceivedMiniBlock_ShouldWork(t *testing.T) {
	t.Parallel()

	dataPool := createDataPool()
	mbt, _ := track.NewMiniBlockTrack(dataPool, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	wasCalled := false
	blockTransactionsPool := &testscommon.ShardedDataStub{
		ImmunizeSetOfDataAgainstEvictionCalled: func(keys [][]byte, destCacheId string) {
			wasCalled = true
		},
	}
	mbt.SetBlockTransactionsPool(blockTransactionsPool)
	mbt.ReceivedMiniBlock(
		[]byte("mb_hash"),
		&block.MiniBlock{
			SenderShardID: 1,
			Type:          block.TxBlock,
		})

	assert.True(t, wasCalled)
}

func TestGetTransactionPool_ShouldWork(t *testing.T) {
	t.Parallel()

	blockTransactionsPool := &testscommon.ShardedDataStub{
		SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
			return &block.MiniBlock{Type: block.TxBlock}, true
		},
	}
	rewardTransactionsPool := &testscommon.ShardedDataStub{
		SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
			return &block.MiniBlock{Type: block.RewardsBlock}, true
		},
	}
	unsignedTransactionsPool := &testscommon.ShardedDataStub{
		SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
			return &block.MiniBlock{Type: block.SmartContractResultBlock}, true
		},
	}
	dataPool := &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return blockTransactionsPool
		},
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return rewardTransactionsPool
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return unsignedTransactionsPool
		},
		MiniBlocksCalled: func() storage.Cacher {
			return testscommon.NewCacherStub()
		},
	}
	mbt, _ := track.NewMiniBlockTrack(dataPool, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	tp := mbt.GetTransactionPool(block.TxBlock)
	assert.Equal(t, blockTransactionsPool, tp)

	tp = mbt.GetTransactionPool(block.RewardsBlock)
	assert.Equal(t, rewardTransactionsPool, tp)

	tp = mbt.GetTransactionPool(block.SmartContractResultBlock)
	assert.Equal(t, unsignedTransactionsPool, tp)

	tp = mbt.GetTransactionPool(block.PeerBlock)
	assert.Nil(t, tp)
}

func createDataPool() dataRetriever.PoolsHolder {
	return &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		MiniBlocksCalled: func() storage.Cacher {
			return testscommon.NewCacherStub()
		},
	}
}
