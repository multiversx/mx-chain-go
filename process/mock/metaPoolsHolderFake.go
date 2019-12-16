package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

type MetaPoolsHolderFake struct {
	metaBlocks    storage.Cacher
	miniBlocks    storage.Cacher
	shardHeaders  storage.Cacher
	headersNonces dataRetriever.Uint64SyncMapCacher
	transactions  dataRetriever.ShardedDataCacherNotifier
	unsigned      dataRetriever.ShardedDataCacherNotifier
	currTxs       dataRetriever.TransactionCacher

	MetaBlocksCalled func() storage.Cacher
	ShardHeadersCalled func() storage.Cacher
}

func NewMetaPoolsHolderFake() *MetaPoolsHolderFake {
	mphf := &MetaPoolsHolderFake{}
	mphf.miniBlocks, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	mphf.transactions, _ = shardedData.NewShardedData(storageUnit.CacheConfig{Size: 10000, Type: storageUnit.LRUCache})
	mphf.unsigned, _ = shardedData.NewShardedData(storageUnit.CacheConfig{Size: 10000, Type: storageUnit.LRUCache})
	mphf.metaBlocks, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	mphf.shardHeaders, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)

	cacheShardHdrNonces, _ := storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	mphf.headersNonces, _ = dataPool.NewNonceSyncMapCacher(
		cacheShardHdrNonces,
		uint64ByteSlice.NewBigEndianConverter(),
	)
	mphf.currTxs, _ = dataPool.NewCurrentBlockPool()

	return mphf
}

func (mphf *MetaPoolsHolderFake) CurrentBlockTxs() dataRetriever.TransactionCacher {
	return mphf.currTxs
}

func (mphf *MetaPoolsHolderFake) Transactions() dataRetriever.ShardedDataCacherNotifier {
	return mphf.transactions
}

func (mphf *MetaPoolsHolderFake) UnsignedTransactions() dataRetriever.ShardedDataCacherNotifier {
	return mphf.unsigned
}

func (mphf *MetaPoolsHolderFake) MetaBlocks() storage.Cacher {
	if mphf.MetaBlocksCalled != nil {
		return mphf.MetaBlocksCalled()
	}
	return mphf.metaBlocks
}

func (mphf *MetaPoolsHolderFake) MiniBlocks() storage.Cacher {
	return mphf.miniBlocks
}

func (mphf *MetaPoolsHolderFake) ShardHeaders() storage.Cacher {
	if mphf.ShardHeadersCalled != nil {
		return mphf.ShardHeadersCalled()
	}
	return mphf.shardHeaders
}

func (mphf *MetaPoolsHolderFake) HeadersNonces() dataRetriever.Uint64SyncMapCacher {
	return mphf.headersNonces
}

// IsInterfaceNil returns true if there is no value under the interface
func (mphf *MetaPoolsHolderFake) IsInterfaceNil() bool {
	if mphf == nil {
		return true
	}
	return false
}
