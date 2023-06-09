package dataRetriever

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/dataPool"
	"github.com/multiversx/mx-chain-go/dataRetriever/dataPool/headersCache"
	"github.com/multiversx/mx-chain-go/dataRetriever/shardedData"
	"github.com/multiversx/mx-chain-go/dataRetriever/txpool"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
)

// PoolsHolderMock -
type PoolsHolderMock struct {
	transactions                   dataRetriever.ShardedDataCacherNotifier
	unsignedTransactions           dataRetriever.ShardedDataCacherNotifier
	rewardTransactions             dataRetriever.ShardedDataCacherNotifier
	headers                        dataRetriever.HeadersPool
	miniBlocks                     storage.Cacher
	peerChangesBlocks              storage.Cacher
	trieNodes                      storage.Cacher
	trieNodesChunks                storage.Cacher
	smartContracts                 storage.Cacher
	currBlockTxs                   dataRetriever.TransactionCacher
	currEpochValidatorInfo         dataRetriever.ValidatorInfoCacher
	mainPeerAuthentications        storage.Cacher
	mainHeartbeats                 storage.Cacher
	fullArchivePeerAuthentications storage.Cacher
	fullArchiveHeartbeats          storage.Cacher
	validatorsInfo                 dataRetriever.ShardedDataCacherNotifier
}

// NewPoolsHolderMock -
func NewPoolsHolderMock() *PoolsHolderMock {
	var err error
	holder := &PoolsHolderMock{}

	holder.transactions, err = txpool.NewShardedTxPool(
		txpool.ArgShardedTxPool{
			Config: storageunit.CacheConfig{
				Capacity:             100000,
				SizePerSender:        1000,
				SizeInBytes:          1000000000,
				SizeInBytesPerSender: 10000000,
				Shards:               16,
			},
			TxGasHandler: &txcachemocks.TxGasHandlerMock{
				MinimumGasMove:       50000,
				MinimumGasPrice:      200000000000,
				GasProcessingDivisor: 100,
			},
			NumberOfShards: 1,
		},
	)
	panicIfError("NewPoolsHolderMock", err)

	holder.unsignedTransactions, err = shardedData.NewShardedData("unsignedTxPool", storageunit.CacheConfig{
		Capacity:    10000,
		SizeInBytes: 1000000000,
		Shards:      1,
	})
	panicIfError("NewPoolsHolderMock", err)

	holder.rewardTransactions, err = shardedData.NewShardedData("rewardsTxPool", storageunit.CacheConfig{
		Capacity:    100,
		SizeInBytes: 100000,
		Shards:      1,
	})
	panicIfError("NewPoolsHolderMock", err)

	holder.headers, err = headersCache.NewHeadersPool(config.HeadersPoolConfig{MaxHeadersPerShard: 1000, NumElementsToRemoveOnEviction: 100})
	panicIfError("NewPoolsHolderMock", err)

	holder.miniBlocks, err = storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: 10000, Shards: 1, SizeInBytes: 0})
	panicIfError("NewPoolsHolderMock", err)

	holder.peerChangesBlocks, err = storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: 10000, Shards: 1, SizeInBytes: 0})
	panicIfError("NewPoolsHolderMock", err)

	holder.currBlockTxs = dataPool.NewCurrentBlockTransactionsPool()
	holder.currEpochValidatorInfo = dataPool.NewCurrentEpochValidatorInfoPool()

	holder.trieNodes, err = storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.SizeLRUCache, Capacity: 900000, Shards: 1, SizeInBytes: 314572800})
	panicIfError("NewPoolsHolderMock", err)

	holder.trieNodesChunks, err = storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.SizeLRUCache, Capacity: 900000, Shards: 1, SizeInBytes: 314572800})
	panicIfError("NewPoolsHolderMock", err)

	holder.smartContracts, err = storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: 10000, Shards: 1, SizeInBytes: 0})
	panicIfError("NewPoolsHolderMock", err)

	holder.mainPeerAuthentications, err = cache.NewTimeCacher(cache.ArgTimeCacher{
		DefaultSpan: 10 * time.Second,
		CacheExpiry: 10 * time.Second,
	})
	panicIfError("NewPoolsHolderMock", err)

	holder.mainHeartbeats, err = storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: 10000, Shards: 1, SizeInBytes: 0})
	panicIfError("NewPoolsHolderMock", err)

	holder.fullArchivePeerAuthentications, err = cache.NewTimeCacher(cache.ArgTimeCacher{
		DefaultSpan: 10 * time.Second,
		CacheExpiry: 10 * time.Second,
	})
	panicIfError("NewPoolsHolderMock", err)

	holder.fullArchiveHeartbeats, err = storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: 10000, Shards: 1, SizeInBytes: 0})
	panicIfError("NewPoolsHolderMock", err)

	holder.validatorsInfo, err = shardedData.NewShardedData("validatorsInfoPool", storageunit.CacheConfig{
		Capacity:    100,
		SizeInBytes: 100000,
		Shards:      1,
	})
	panicIfError("NewPoolsHolderMock", err)

	return holder
}

// CurrentBlockTxs -
func (holder *PoolsHolderMock) CurrentBlockTxs() dataRetriever.TransactionCacher {
	return holder.currBlockTxs
}

// CurrentEpochValidatorInfo -
func (holder *PoolsHolderMock) CurrentEpochValidatorInfo() dataRetriever.ValidatorInfoCacher {
	return holder.currEpochValidatorInfo
}

// Transactions -
func (holder *PoolsHolderMock) Transactions() dataRetriever.ShardedDataCacherNotifier {
	return holder.transactions
}

// UnsignedTransactions -
func (holder *PoolsHolderMock) UnsignedTransactions() dataRetriever.ShardedDataCacherNotifier {
	return holder.unsignedTransactions
}

// RewardTransactions -
func (holder *PoolsHolderMock) RewardTransactions() dataRetriever.ShardedDataCacherNotifier {
	return holder.rewardTransactions
}

// Headers -
func (holder *PoolsHolderMock) Headers() dataRetriever.HeadersPool {
	return holder.headers
}

// MiniBlocks -
func (holder *PoolsHolderMock) MiniBlocks() storage.Cacher {
	return holder.miniBlocks
}

// PeerChangesBlocks -
func (holder *PoolsHolderMock) PeerChangesBlocks() storage.Cacher {
	return holder.peerChangesBlocks
}

// SetTransactions -
func (holder *PoolsHolderMock) SetTransactions(pool dataRetriever.ShardedDataCacherNotifier) {
	holder.transactions = pool
}

// SetUnsignedTransactions -
func (holder *PoolsHolderMock) SetUnsignedTransactions(pool dataRetriever.ShardedDataCacherNotifier) {
	holder.unsignedTransactions = pool
}

// TrieNodes -
func (holder *PoolsHolderMock) TrieNodes() storage.Cacher {
	return holder.trieNodes
}

// TrieNodesChunks -
func (holder *PoolsHolderMock) TrieNodesChunks() storage.Cacher {
	return holder.trieNodesChunks
}

// SmartContracts -
func (holder *PoolsHolderMock) SmartContracts() storage.Cacher {
	return holder.smartContracts
}

// PeerAuthentications -
func (holder *PoolsHolderMock) PeerAuthentications() storage.Cacher {
	return holder.mainPeerAuthentications
}

// Heartbeats -
func (holder *PoolsHolderMock) Heartbeats() storage.Cacher {
	return holder.mainHeartbeats
}

// FullArchivePeerAuthentications -
func (holder *PoolsHolderMock) FullArchivePeerAuthentications() storage.Cacher {
	return holder.fullArchivePeerAuthentications
}

// FullArchiveHeartbeats -
func (holder *PoolsHolderMock) FullArchiveHeartbeats() storage.Cacher {
	return holder.fullArchiveHeartbeats
}

// ValidatorsInfo -
func (holder *PoolsHolderMock) ValidatorsInfo() dataRetriever.ShardedDataCacherNotifier {
	return holder.validatorsInfo
}

// Close -
func (holder *PoolsHolderMock) Close() error {
	var lastError error
	if !check.IfNil(holder.trieNodes) {
		err := holder.trieNodes.Close()
		if err != nil {
			lastError = err
		}
	}

	if !check.IfNil(holder.mainPeerAuthentications) {
		err := holder.mainPeerAuthentications.Close()
		if err != nil {
			lastError = err
		}
	}

	if !check.IfNil(holder.fullArchivePeerAuthentications) {
		err := holder.fullArchivePeerAuthentications.Close()
		if err != nil {
			lastError = err
		}
	}

	return lastError
}

// IsInterfaceNil returns true if there is no value under the interface
func (holder *PoolsHolderMock) IsInterfaceNil() bool {
	return holder == nil
}
