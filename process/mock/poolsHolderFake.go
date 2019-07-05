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
	unsignedtransactions dataRetriever.ShardedDataCacherNotifier
	headers              storage.Cacher
	metaBlocks           storage.Cacher
	hdrNonces            dataRetriever.Uint64Cacher
	miniBlocks           storage.Cacher
	peerChangesBlocks    storage.Cacher
	metaHdrNonces        dataRetriever.Uint64Cacher
}

func NewPoolsHolderFake() *PoolsHolderFake {
	phf := &PoolsHolderFake{}
	phf.transactions, _ = shardedData.NewShardedData(storageUnit.CacheConfig{Size: 10000, Type: storageUnit.LRUCache})
	phf.unsignedtransactions, _ = shardedData.NewShardedData(storageUnit.CacheConfig{Size: 10000, Type: storageUnit.LRUCache})
	phf.headers, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	phf.metaBlocks, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	cacheHdrNonces, _ := storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	phf.hdrNonces, _ = dataPool.NewNonceToHashCacher(
		cacheHdrNonces,
		uint64ByteSlice.NewBigEndianConverter(),
	)
	cacheMetaHdrNonces, _ := storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	phf.metaHdrNonces, _ = dataPool.NewNonceToHashCacher(
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
	return phf.unsignedtransactions
}

func (phf *PoolsHolderFake) Headers() storage.Cacher {
	return phf.headers
}

func (phf *PoolsHolderFake) HeadersNonces() dataRetriever.Uint64Cacher {
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

func (phf *PoolsHolderFake) MetaHeadersNonces() dataRetriever.Uint64Cacher {
	return phf.metaHdrNonces
}

func (phf *PoolsHolderFake) SetTransactions(transactions dataRetriever.ShardedDataCacherNotifier) {
	phf.transactions = transactions
}

func (phf *PoolsHolderFake) SetUnsignedTransactions(scrs dataRetriever.ShardedDataCacherNotifier) {
	phf.unsignedtransactions = scrs
}
