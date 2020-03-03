package disabled

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool/headersCache"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/txpool"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// PoolsHolder -
type PoolsHolder struct {
	transactions         dataRetriever.ShardedDataCacherNotifier
	unsignedTransactions dataRetriever.ShardedDataCacherNotifier
	rewardTransactions   dataRetriever.ShardedDataCacherNotifier
	headers              dataRetriever.HeadersPool
	miniBlocks           storage.Cacher
	peerChangesBlocks    storage.Cacher
	trieNodes            storage.Cacher
	currBlockTxs         dataRetriever.TransactionCacher
}

// NewDisabledPoolsHolder -
func NewDisabledPoolsHolder() *PoolsHolder {
	phf := &PoolsHolder{}

	phf.transactions, _ = txpool.NewShardedTxPool(
		txpool.ArgShardedTxPool{
			Config: storageUnit.CacheConfig{
				Size:        10000,
				SizeInBytes: 1000000000,
				Shards:      16,
			},
			MinGasPrice:    100000000000000,
			NumberOfShards: 1,
		},
	)

	phf.unsignedTransactions, _ = shardedData.NewShardedData(storageUnit.CacheConfig{Size: 10000, Type: storageUnit.LRUCache})
	phf.rewardTransactions, _ = shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache})
	phf.headers, _ = headersCache.NewHeadersPool(config.HeadersPoolConfig{MaxHeadersPerShard: 1000, NumElementsToRemoveOnEviction: 100})
	phf.miniBlocks, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	phf.peerChangesBlocks, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	phf.currBlockTxs, _ = dataPool.NewCurrentBlockPool()
	phf.trieNodes, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)

	return phf
}

// CurrentBlockTxs -
func (phm *PoolsHolder) CurrentBlockTxs() dataRetriever.TransactionCacher {
	return phm.currBlockTxs
}

// Transactions -
func (phm *PoolsHolder) Transactions() dataRetriever.ShardedDataCacherNotifier {
	return phm.transactions
}

// UnsignedTransactions -
func (phm *PoolsHolder) UnsignedTransactions() dataRetriever.ShardedDataCacherNotifier {
	return phm.unsignedTransactions
}

// RewardTransactions -
func (phm *PoolsHolder) RewardTransactions() dataRetriever.ShardedDataCacherNotifier {
	return phm.rewardTransactions
}

// Headers -
func (phm *PoolsHolder) Headers() dataRetriever.HeadersPool {
	return phm.headers
}

// MiniBlocks -
func (phm *PoolsHolder) MiniBlocks() storage.Cacher {
	return phm.miniBlocks
}

// PeerChangesBlocks -
func (phm *PoolsHolder) PeerChangesBlocks() storage.Cacher {
	return phm.peerChangesBlocks
}

// SetTransactions -
func (phm *PoolsHolder) SetTransactions(transactions dataRetriever.ShardedDataCacherNotifier) {
	phm.transactions = transactions
}

// SetUnsignedTransactions -
func (phm *PoolsHolder) SetUnsignedTransactions(scrs dataRetriever.ShardedDataCacherNotifier) {
	phm.unsignedTransactions = scrs
}

// TrieNodes -
func (phm *PoolsHolder) TrieNodes() storage.Cacher {
	return phm.trieNodes
}

// IsInterfaceNil returns true if there is no value under the interface
func (phm *PoolsHolder) IsInterfaceNil() bool {
	return phm == nil
}
