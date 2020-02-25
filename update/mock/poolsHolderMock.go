package mock

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

type PoolsHolderMock struct {
	transactions         dataRetriever.ShardedDataCacherNotifier
	unsignedTransactions dataRetriever.ShardedDataCacherNotifier
	rewardTransactions   dataRetriever.ShardedDataCacherNotifier
	headers              dataRetriever.HeadersPool
	miniBlocks           storage.Cacher
	peerChangesBlocks    storage.Cacher
	trieNodes            storage.Cacher
	currBlockTxs         dataRetriever.TransactionCacher
}

func NewPoolsHolderMock() *PoolsHolderMock {
	phf := &PoolsHolderMock{}
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

func (phm *PoolsHolderMock) CurrentBlockTxs() dataRetriever.TransactionCacher {
	return phm.currBlockTxs
}

func (phm *PoolsHolderMock) Transactions() dataRetriever.ShardedDataCacherNotifier {
	return phm.transactions
}

func (phm *PoolsHolderMock) UnsignedTransactions() dataRetriever.ShardedDataCacherNotifier {
	return phm.unsignedTransactions
}

func (phm *PoolsHolderMock) RewardTransactions() dataRetriever.ShardedDataCacherNotifier {
	return phm.rewardTransactions
}

func (phm *PoolsHolderMock) Headers() dataRetriever.HeadersPool {
	return phm.headers
}

func (phm *PoolsHolderMock) MiniBlocks() storage.Cacher {
	return phm.miniBlocks
}

func (phm *PoolsHolderMock) PeerChangesBlocks() storage.Cacher {
	return phm.peerChangesBlocks
}

func (phm *PoolsHolderMock) SetTransactions(transactions dataRetriever.ShardedDataCacherNotifier) {
	phm.transactions = transactions
}

func (phm *PoolsHolderMock) SetUnsignedTransactions(scrs dataRetriever.ShardedDataCacherNotifier) {
	phm.unsignedTransactions = scrs
}

func (phm *PoolsHolderMock) TrieNodes() storage.Cacher {
	return phm.trieNodes
}

// IsInterfaceNil returns true if there is no value under the interface
func (phm *PoolsHolderMock) IsInterfaceNil() bool {
	return phm == nil
}
