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
	metaChainBlocks storage.Cacher
	miniBlockHashes dataRetriever.ShardedDataCacherNotifier
	shardHeaders    storage.Cacher
	headersNonces   dataRetriever.Uint64SyncMapCacher
}

func NewMetaPoolsHolderFake() *MetaPoolsHolderFake {
	mphf := &MetaPoolsHolderFake{}
	mphf.miniBlockHashes, _ = shardedData.NewShardedData(storageUnit.CacheConfig{Size: 10000, Type: storageUnit.LRUCache})
	mphf.metaChainBlocks, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	mphf.shardHeaders, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)

	cacheShardHdrNonces, _ := storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	mphf.headersNonces, _ = dataPool.NewNonceSyncMapCacher(
		cacheShardHdrNonces,
		uint64ByteSlice.NewBigEndianConverter(),
	)
	return mphf
}

func (mphf *MetaPoolsHolderFake) MetaChainBlocks() storage.Cacher {
	return mphf.metaChainBlocks
}

func (mphf *MetaPoolsHolderFake) MiniBlockHashes() dataRetriever.ShardedDataCacherNotifier {
	return mphf.miniBlockHashes
}

func (mphf *MetaPoolsHolderFake) ShardHeaders() storage.Cacher {
	return mphf.shardHeaders
}

func (mphf *MetaPoolsHolderFake) HeadersNonces() dataRetriever.Uint64SyncMapCacher {
	return mphf.headersNonces
}
