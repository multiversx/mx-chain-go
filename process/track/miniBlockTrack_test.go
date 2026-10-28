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
	var immunizedCacheID string
	var immunizedNonce uint64
	setOldestImmuneNonceCalled := false
	blockTransactionsPool := &testscommon.ShardedDataStub{
		SetOldestImmuneNonceCalled: func(cacheID string, nonce uint64) {
			setOldestImmuneNonceCalled = true
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

	// Immunization happens on metablock arrival.
	assert.Equal(t, txHashes, whitelistedKeys)
	assert.Equal(t, txHashes, immunizedKeys)
	assert.Equal(t, process.ShardCacherIdentifier(1, 0), immunizedCacheID)
	assert.Equal(t, uint64(7), immunizedNonce)
	// Regression guard: threshold advance is deferred to commit.
	assert.False(t, setOldestImmuneNonceCalled, "SetOldestImmuneNonce must not be called from metablock arrival path")
}

func TestMiniBlockTrack_CleanupConfirmedMiniBlocksBelow(t *testing.T) {
	t.Parallel()

	dataPool := createDataPool()
	blockTracker := &mock.BlockTrackerMock{}
	var finalMetachainHeadersHandler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)
	blockTracker.RegisterFinalMetachainHeadersHandlerCalled = func(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)) {
		finalMetachainHeadersHandler = handler
	}

	mbt, _ := track.NewMiniBlockTrack(dataPool, blockTracker, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	// Two confirmed miniblocks at different nonces, neither arriving in pool.
	finalMetachainHeadersHandler(core.MetachainShardId, []data.HeaderHandler{
		&block.MetaBlock{
			Nonce: 5,
			ShardInfo: []block.ShardData{
				{
					ShardID: 1,
					ShardMiniBlockHeaders: []block.MiniBlockHeader{
						{Hash: []byte("mb_old"), SenderShardID: 1, ReceiverShardID: 0, Type: block.TxBlock},
					},
				},
			},
		},
	}, nil)
	finalMetachainHeadersHandler(core.MetachainShardId, []data.HeaderHandler{
		&block.MetaBlock{
			Nonce: 10,
			ShardInfo: []block.ShardData{
				{
					ShardID: 1,
					ShardMiniBlockHeaders: []block.MiniBlockHeader{
						{Hash: []byte("mb_new"), SenderShardID: 1, ReceiverShardID: 0, Type: block.TxBlock},
					},
				},
			},
		},
	}, nil)

	// Cleanup with threshold 8 should drop the nonce-5 entry but keep nonce-10.
	mbt.CleanupConfirmedMiniBlocksBelow(8)

	_, _, hasOld := mbt.GetConfirmedMiniBlockInfo([]byte("mb_old"))
	_, _, hasNew := mbt.GetConfirmedMiniBlockInfo([]byte("mb_new"))
	assert.False(t, hasOld)
	assert.True(t, hasNew)
}

func TestMiniBlockTrack_ReleaseImmunityForCommittedMetaBlocks(t *testing.T) {
	t.Parallel()

	miniBlocksPool := cache.NewCacherStub()
	miniBlocksPool.PeekCalled = func(_ []byte) (interface{}, bool) { return nil, false }

	var blockPoolThreshold, rewardPoolThreshold, unsignedPoolThreshold uint64
	blockPool := &testscommon.ShardedDataStub{
		SetOldestImmuneNonceForAllCachesCalled: func(nonce uint64) { blockPoolThreshold = nonce },
	}
	rewardPool := &testscommon.ShardedDataStub{
		SetOldestImmuneNonceForAllCachesCalled: func(nonce uint64) { rewardPoolThreshold = nonce },
	}
	unsignedPool := &testscommon.ShardedDataStub{
		SetOldestImmuneNonceForAllCachesCalled: func(nonce uint64) { unsignedPoolThreshold = nonce },
	}

	dataPool := &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled:         func() dataRetriever.ShardedDataCacherNotifier { return blockPool },
		RewardTransactionsCalled:   func() dataRetriever.ShardedDataCacherNotifier { return rewardPool },
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier { return unsignedPool },
		MiniBlocksCalled:           func() storage.Cacher { return miniBlocksPool },
	}

	blockTracker := &mock.BlockTrackerMock{}
	var headersHandler func(uint32, []data.HeaderHandler, [][]byte)
	blockTracker.RegisterFinalMetachainHeadersHandlerCalled = func(handler func(uint32, []data.HeaderHandler, [][]byte)) {
		headersHandler = handler
	}

	mbt, _ := track.NewMiniBlockTrack(dataPool, blockTracker, mock.NewMultipleShardsCoordinatorMock(), &testscommon.WhiteListHandlerStub{})

	// Seed registry with entries at two different nonces.
	headersHandler(core.MetachainShardId, []data.HeaderHandler{
		&block.MetaBlock{Nonce: 5, ShardInfo: []block.ShardData{{ShardID: 1, ShardMiniBlockHeaders: []block.MiniBlockHeader{
			{Hash: []byte("mb_old"), SenderShardID: 1, ReceiverShardID: 0, Type: block.TxBlock},
		}}}},
	}, nil)
	headersHandler(core.MetachainShardId, []data.HeaderHandler{
		&block.MetaBlock{Nonce: 10, ShardInfo: []block.ShardData{{ShardID: 1, ShardMiniBlockHeaders: []block.MiniBlockHeader{
			{Hash: []byte("mb_new"), SenderShardID: 1, ReceiverShardID: 0, Type: block.TxBlock},
		}}}},
	}, nil)

	mbt.ReleaseImmunityForCommittedMetaBlocks(8)

	// All three pools should have received the threshold uniformly.
	assert.Equal(t, uint64(8), blockPoolThreshold)
	assert.Equal(t, uint64(8), rewardPoolThreshold)
	assert.Equal(t, uint64(8), unsignedPoolThreshold)

	// Registry pruned below threshold.
	_, _, hasOld := mbt.GetConfirmedMiniBlockInfo([]byte("mb_old"))
	_, _, hasNew := mbt.GetConfirmedMiniBlockInfo([]byte("mb_new"))
	assert.False(t, hasOld)
	assert.True(t, hasNew)
}

func TestMiniBlockTrack_ReleaseImmunityForCommittedShardBlocks(t *testing.T) {
	t.Parallel()

	miniBlocksPool := cache.NewCacherStub()
	miniBlocksPool.PeekCalled = func(_ []byte) (interface{}, bool) { return nil, false }

	type call struct {
		cacheID string
		nonce   uint64
	}
	var blockCalls, rewardCalls, unsignedCalls []call
	blockPool := &testscommon.ShardedDataStub{
		SetOldestImmuneNonceCalled: func(c string, n uint64) { blockCalls = append(blockCalls, call{c, n}) },
	}
	rewardPool := &testscommon.ShardedDataStub{
		SetOldestImmuneNonceCalled: func(c string, n uint64) { rewardCalls = append(rewardCalls, call{c, n}) },
	}
	unsignedPool := &testscommon.ShardedDataStub{
		SetOldestImmuneNonceCalled: func(c string, n uint64) { unsignedCalls = append(unsignedCalls, call{c, n}) },
	}

	dataPool := &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled:         func() dataRetriever.ShardedDataCacherNotifier { return blockPool },
		RewardTransactionsCalled:   func() dataRetriever.ShardedDataCacherNotifier { return rewardPool },
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier { return unsignedPool },
		MiniBlocksCalled:           func() storage.Cacher { return miniBlocksPool },
	}

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.CurrentShard = core.MetachainShardId

	blockTracker := &mock.BlockTrackerMock{}
	var crossHeadersHandler func(uint32, []data.HeaderHandler, [][]byte)
	blockTracker.RegisterCrossNotarizedHeadersHandlerCalled = func(handler func(uint32, []data.HeaderHandler, [][]byte)) {
		crossHeadersHandler = handler
	}

	mbt, _ := track.NewMiniBlockTrack(dataPool, blockTracker, shardCoordinator, &testscommon.WhiteListHandlerStub{})

	// Seed registry: SCR from shard 1 to meta at nonce 5, and unrelated entry from shard 2 to meta at nonce 5.
	crossHeadersHandler(0, []data.HeaderHandler{
		&block.Header{Nonce: 5, ShardID: 1, MiniBlockHeaders: []block.MiniBlockHeader{
			{Hash: []byte("scr_shard1"), SenderShardID: 1, ReceiverShardID: core.MetachainShardId, Type: block.SmartContractResultBlock},
		}},
		&block.Header{Nonce: 5, ShardID: 2, MiniBlockHeaders: []block.MiniBlockHeader{
			{Hash: []byte("scr_shard2"), SenderShardID: 2, ReceiverShardID: core.MetachainShardId, Type: block.SmartContractResultBlock},
		}},
	}, nil)

	mbt.ReleaseImmunityForCommittedShardBlocks(1, 6)

	expectedCacheID := process.ShardCacherIdentifier(1, core.MetachainShardId)
	assert.Equal(t, []call{{expectedCacheID, 6}}, blockCalls)
	assert.Equal(t, []call{{expectedCacheID, 6}}, rewardCalls)
	assert.Equal(t, []call{{expectedCacheID, 6}}, unsignedCalls)

	// Only the shard-1 registry entry should be pruned.
	_, _, hasShard1 := mbt.GetConfirmedMiniBlockInfo([]byte("scr_shard1"))
	_, _, hasShard2 := mbt.GetConfirmedMiniBlockInfo([]byte("scr_shard2"))
	assert.False(t, hasShard1)
	assert.True(t, hasShard2)
}
