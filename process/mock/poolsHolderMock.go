package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

type PoolsHolderMock struct {
	transactions         dataRetriever.ShardedDataCacherNotifier
	unsignedTransactions dataRetriever.ShardedDataCacherNotifier
	rewardTransactions   dataRetriever.ShardedDataCacherNotifier
	headers              storage.Cacher
	metaBlocks           storage.Cacher
	hdrNonces            dataRetriever.Uint64SyncMapCacher
	miniBlocks           storage.Cacher
	peerChangesBlocks    storage.Cacher
	trieNodes            storage.Cacher
	metaHdrNonces        dataRetriever.Uint64SyncMapCacher
	currBlockTxs         dataRetriever.TransactionCacher
}

func NewPoolsHolderMock() *PoolsHolderMock {
	phf := &PoolsHolderMock{}
	phf.transactions, _ = shardedData.NewShardedData(storageUnit.CacheConfig{Size: 10000, Type: storageUnit.LRUCache})
	phf.unsignedTransactions, _ = shardedData.NewShardedData(storageUnit.CacheConfig{Size: 10000, Type: storageUnit.LRUCache})
	phf.rewardTransactions, _ = shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache})
	phf.headers, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	phf.metaBlocks, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	phf.trieNodes, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	cacheHdrNonces, _ := storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	phf.hdrNonces, _ = dataPool.NewNonceSyncMapCacher(
		cacheHdrNonces,
		uint64ByteSlice.NewBigEndianConverter(),
	)
	cacheMetaHdrNonces, _ := storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	phf.metaHdrNonces, _ = dataPool.NewNonceSyncMapCacher(
		cacheMetaHdrNonces,
		uint64ByteSlice.NewBigEndianConverter(),
	)
	phf.miniBlocks, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	phf.peerChangesBlocks, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	phf.currBlockTxs, _ = dataPool.NewCurrentBlockPool()

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

func (phm *PoolsHolderMock) Headers() storage.Cacher {
	return phm.headers
}

func (phm *PoolsHolderMock) HeadersNonces() dataRetriever.Uint64SyncMapCacher {
	return phm.hdrNonces
}

func (phm *PoolsHolderMock) MiniBlocks() storage.Cacher {
	return phm.miniBlocks
}

func (phm *PoolsHolderMock) PeerChangesBlocks() storage.Cacher {
	return phm.peerChangesBlocks
}

func (phm *PoolsHolderMock) MetaBlocks() storage.Cacher {
	return phm.metaBlocks
}

func (phm *PoolsHolderMock) TrieNodes() storage.Cacher {
	return phm.trieNodes
}

func (phm *PoolsHolderMock) MetaHeadersNonces() dataRetriever.Uint64SyncMapCacher {
	return phm.metaHdrNonces
}

func (phm *PoolsHolderMock) SetTransactions(transactions dataRetriever.ShardedDataCacherNotifier) {
	phm.transactions = transactions
}

func (phm *PoolsHolderMock) SetUnsignedTransactions(scrs dataRetriever.ShardedDataCacherNotifier) {
	phm.unsignedTransactions = scrs
}

// IsInterfaceNil returns true if there is no value under the interface
func (phf *PoolsHolderMock) IsInterfaceNil() bool {
	if phf == nil {
		return true
	}
	return false
}
