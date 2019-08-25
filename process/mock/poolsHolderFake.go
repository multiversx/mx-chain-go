package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

type PoolsHolderFake struct {
	transactions         dataRetriever.ShardedDataCacherNotifier
	unsignedTransactions dataRetriever.ShardedDataCacherNotifier
	rewardTransactions   dataRetriever.ShardedDataCacherNotifier
	headers              storage.Cacher
	metaBlocks           storage.Cacher
	hdrNonces            dataRetriever.Uint64SyncMapCacher
	miniBlocks           storage.Cacher
	peerChangesBlocks    storage.Cacher
	metaHdrNonces        dataRetriever.Uint64SyncMapCacher
}

func NewPoolsHolderFake() *PoolsHolderFake {
	phf := &PoolsHolderFake{}
	phf.transactions, _ = shardedData.NewShardedData(storageUnit.CacheConfig{Size: 10000, Type: storageUnit.LRUCache})
	phf.unsignedTransactions, _ = shardedData.NewShardedData(storageUnit.CacheConfig{Size: 10000, Type: storageUnit.LRUCache})
	phf.rewardTransactions, _ = shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache})
	phf.headers, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	phf.metaBlocks, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
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
	return phf
}

func (phf *PoolsHolderFake) Transactions() dataRetriever.ShardedDataCacherNotifier {
	return phf.transactions
}

func (phf *PoolsHolderFake) UnsignedTransactions() dataRetriever.ShardedDataCacherNotifier {
	return phf.unsignedTransactions
}

func (phf *PoolsHolderFake) RewardTransactions() dataRetriever.ShardedDataCacherNotifier {
	return phf.rewardTransactions
}

func (phf *PoolsHolderFake) Headers() storage.Cacher {
	return phf.headers
}

func (phf *PoolsHolderFake) HeadersNonces() dataRetriever.Uint64SyncMapCacher {
	return phf.hdrNonces
}

func (phf *PoolsHolderFake) MiniBlocks() storage.Cacher {
	return phf.miniBlocks
}

func (phf *PoolsHolderFake) PeerChangesBlocks() storage.Cacher {
	return phf.peerChangesBlocks
}

func (phf *PoolsHolderFake) MetaBlocks() storage.Cacher {
	return phf.metaBlocks
}

func (phf *PoolsHolderFake) MetaHeadersNonces() dataRetriever.Uint64SyncMapCacher {
	return phf.metaHdrNonces
}

func (phf *PoolsHolderFake) SetTransactions(transactions dataRetriever.ShardedDataCacherNotifier) {
	phf.transactions = transactions
}

func (phf *PoolsHolderFake) SetUnsignedTransactions(scrs dataRetriever.ShardedDataCacherNotifier) {
	phf.unsignedTransactions = scrs
}
