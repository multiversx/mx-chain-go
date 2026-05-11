package track_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
)

func TestNewMiniBlockTrack_NilDataPoolHolderErr(t *testing.T) {
	t.Parallel()

	mbt, err := track.NewMiniBlockTrack(nil, &mock.BlockTrackerMock{}, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

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
	mbt, err := track.NewMiniBlockTrack(dataPool, &mock.BlockTrackerMock{}, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

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
	mbt, err := track.NewMiniBlockTrack(dataPool, &mock.BlockTrackerMock{}, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

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
	mbt, err := track.NewMiniBlockTrack(dataPool, &mock.BlockTrackerMock{}, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

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
	mbt, err := track.NewMiniBlockTrack(dataPool, &mock.BlockTrackerMock{}, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	assert.Nil(t, mbt)
	assert.Equal(t, process.ErrNilMiniBlockPool, err)
}

func TestNewMiniBlockTrack_NilBlockTrackerErr(t *testing.T) {
	t.Parallel()

	dataPool := createDataPool()
	miniBlockTrack, err := track.NewMiniBlockTrack(dataPool, nil, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	assert.Nil(t, miniBlockTrack)
	assert.Equal(t, process.ErrNilBlockTracker, err)
}

func TestNewMiniBlockTrack_NilShardCoordinatorErr(t *testing.T) {
	t.Parallel()

	dataPool := createDataPool()
	miniBlockTrack, err := track.NewMiniBlockTrack(dataPool, &mock.BlockTrackerMock{}, nil, &testscommon.WhiteListHandlerStub{})

	assert.Nil(t, miniBlockTrack)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewMiniBlockTrack_NilWhitelistHandlerErr(t *testing.T) {
	t.Parallel()

	dataPool := createDataPool()
	miniBlockTrack, err := track.NewMiniBlockTrack(dataPool, &mock.BlockTrackerMock{}, mock.NewMultipleShardsCoordinatorMock(), nil)

	assert.Nil(t, miniBlockTrack)
	assert.Equal(t, process.ErrNilWhiteListHandler, err)
}

func TestNewMiniBlockTrack_ShouldWork(t *testing.T) {
	t.Parallel()

	dataPool := createDataPool()
	mbt, err := track.NewMiniBlockTrack(dataPool, &mock.BlockTrackerMock{}, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	assert.Nil(t, err)
	assert.NotNil(t, mbt)
}

func TestReceivedMiniBlock_ShouldReturnIfKeyIsNil(t *testing.T) {
	t.Parallel()

	dataPool := createDataPool()
	mbt, _ := track.NewMiniBlockTrack(dataPool, &mock.BlockTrackerMock{}, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	wasCalled := false
	blockTransactionsPool := &testscommon.ShardedDataStub{
		ImmunizeSetOfDataAgainstEvictionCalled: func(keys [][]byte, destCacheId string, nonce uint64) {
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
	mbt, _ := track.NewMiniBlockTrack(dataPool, &mock.BlockTrackerMock{}, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	wasCalled := false
	blockTransactionsPool := &testscommon.ShardedDataStub{
		ImmunizeSetOfDataAgainstEvictionCalled: func(keys [][]byte, destCacheId string, nonce uint64) {
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
	mbt, _ := track.NewMiniBlockTrack(dataPool, &mock.BlockTrackerMock{}, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	wasCalled := false
	blockTransactionsPool := &testscommon.ShardedDataStub{
		ImmunizeSetOfDataAgainstEvictionCalled: func(keys [][]byte, destCacheId string, nonce uint64) {
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
	mbt, _ := track.NewMiniBlockTrack(dataPool, &mock.BlockTrackerMock{}, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	wasCalled := false
	blockTransactionsPool := &testscommon.ShardedDataStub{
		ImmunizeSetOfDataAgainstEvictionCalled: func(keys [][]byte, destCacheId string, nonce uint64) {
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

func TestReceivedMiniBlock_ShouldNotImmunizeUnconfirmedMiniBlock(t *testing.T) {
	t.Parallel()

	dataPool := createDataPool()
	mbt, _ := track.NewMiniBlockTrack(dataPool, &mock.BlockTrackerMock{}, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	wasCalled := false
	blockTransactionsPool := &testscommon.ShardedDataStub{
		ImmunizeSetOfDataAgainstEvictionCalled: func(keys [][]byte, destCacheId string, nonce uint64) {
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

	assert.False(t, wasCalled)
}

func TestReceivedMiniBlock_ShouldImmunizeConfirmedMiniBlock(t *testing.T) {
	t.Parallel()

	dataPool := createDataPool()
	blockTracker := &mock.BlockTrackerMock{}

	var finalMetachainHeadersHandler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)
	blockTracker.RegisterFinalMetachainHeadersHandlerCalled = func(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)) {
		finalMetachainHeadersHandler = handler
	}

	whitelistCalled := false
	whiteListHandler := &testscommon.WhiteListHandlerStub{
		AddCalled: func(keys [][]byte) {
			whitelistCalled = true
		},
	}
	mbt, _ := track.NewMiniBlockTrack(dataPool, blockTracker, mock.NewMultipleShardsCoordinatorMock(), whiteListHandler)

	var cacheID string
	var nonce uint64
	wasCalled := false
	blockTransactionsPool := &testscommon.ShardedDataStub{
		ImmunizeSetOfDataAgainstEvictionCalled: func(keys [][]byte, destCacheId string, providedNonce uint64) {
			wasCalled = true
			cacheID = destCacheId
			nonce = providedNonce
		},
	}
	mbt.SetBlockTransactionsPool(blockTransactionsPool)

	finalMetachainHeadersHandler(core.MetachainShardId, []data.HeaderHandler{
		&block.MetaBlock{
			Nonce: 7,
			ShardInfo: []block.ShardData{
				{
					ShardID: 1,
					ShardMiniBlockHeaders: []block.MiniBlockHeader{
						{
							Hash:            []byte("mb_hash"),
							SenderShardID:   1,
							ReceiverShardID: 0,
							Type:            block.TxBlock,
						},
					},
				},
			},
		},
	}, nil)

	mbt.ReceivedMiniBlock(
		[]byte("mb_hash"),
		&block.MiniBlock{
			SenderShardID:   1,
			ReceiverShardID: 0,
			Type:            block.TxBlock,
			TxHashes:        [][]byte{[]byte("txHash")},
		},
	)

	assert.True(t, wasCalled)
	assert.True(t, whitelistCalled)
	assert.Equal(t, process.ShardCacherIdentifier(1, 0), cacheID)
	assert.Equal(t, uint64(7), nonce)
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
			return cache.NewCacherStub()
		},
	}
	mbt, _ := track.NewMiniBlockTrack(dataPool, &mock.BlockTrackerMock{}, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

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
			return cache.NewCacherStub()
		},
	}
}

func TestRegisterConfirmedMiniBlocksForHeader_ShouldImmunizeStoredMiniBlock(t *testing.T) {
	t.Parallel()

	miniBlockHash := []byte("mb_hash")
	txHashes := [][]byte{[]byte("txHash")}
	storedMiniBlock := &block.MiniBlock{
		SenderShardID:   1,
		ReceiverShardID: 0,
		Type:            block.TxBlock,
		TxHashes:        txHashes,
	}

	miniBlocksPool := cache.NewCacherStub()
	miniBlocksPool.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		if string(key) != string(miniBlockHash) {
			return nil, false
		}

		return storedMiniBlock, true
	}

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
			return miniBlocksPool
		},
	}

	blockTracker := &mock.BlockTrackerMock{}
	var finalMetachainHeadersHandler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)
	blockTracker.RegisterFinalMetachainHeadersHandlerCalled = func(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)) {
		finalMetachainHeadersHandler = handler
	}

	var whitelistedKeys [][]byte
	whiteListHandler := &testscommon.WhiteListHandlerStub{
		AddCalled: func(keys [][]byte) {
			whitelistedKeys = keys
		},
	}
	var immunizedKeys [][]byte
	var setOldestImmuneNonceCacheID string
	var immunizedCacheID string
	var setOldestImmuneNonceNonce uint64
	var immunizedNonce uint64
	blockTransactionsPool := &testscommon.ShardedDataStub{
		SetOldestImmuneNonceCalled: func(cacheID string, nonce uint64) {
			setOldestImmuneNonceCacheID = cacheID
			setOldestImmuneNonceNonce = nonce
		},
		ImmunizeSetOfDataAgainstEvictionCalled: func(keys [][]byte, destCacheID string, nonce uint64) {
			immunizedKeys = keys
			immunizedCacheID = destCacheID
			immunizedNonce = nonce
		},
	}

	mbt, _ := track.NewMiniBlockTrack(dataPool, blockTracker, mock.NewMultipleShardsCoordinatorMock(), whiteListHandler)
	mbt.SetBlockTransactionsPool(blockTransactionsPool)

	finalMetachainHeadersHandler(core.MetachainShardId, []data.HeaderHandler{
		&block.MetaBlock{
			Nonce: 7,
			ShardInfo: []block.ShardData{
				{
					ShardID: 1,
					ShardMiniBlockHeaders: []block.MiniBlockHeader{
						{
							Hash:            miniBlockHash,
							SenderShardID:   1,
							ReceiverShardID: 0,
							Type:            block.TxBlock,
						},
					},
				},
			},
		},
	}, nil)

	assert.Equal(t, txHashes, whitelistedKeys)
	assert.Equal(t, txHashes, immunizedKeys)
	assert.Equal(t, process.ShardCacherIdentifier(1, 0), setOldestImmuneNonceCacheID)
	assert.Equal(t, process.ShardCacherIdentifier(1, 0), immunizedCacheID)
	assert.Equal(t, uint64(7), setOldestImmuneNonceNonce)
	assert.Equal(t, uint64(7), immunizedNonce)
}
