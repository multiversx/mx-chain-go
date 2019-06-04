package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/storageUnit"
)

type MetaPoolsHolderFake struct {
	metaChainBlocks storage.Cacher
	miniBlockHashes dataRetriever.ShardedDataCacherNotifier
	shardHeaders    storage.Cacher
	metaBlockNonces dataRetriever.Uint64Cacher
}

func NewMetaPoolsHolderFake() *MetaPoolsHolderFake {
	mphf := &MetaPoolsHolderFake{}
	mphf.miniBlockHashes, _ = shardedData.NewShardedData(storageUnit.CacheConfig{Size: 10000, Type: storageUnit.LRUCache})
	mphf.metaChainBlocks, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	mphf.shardHeaders, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	cacheHdrNonces, _ := storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	mphf.metaBlockNonces, _ = dataPool.NewNonceToHashCacher(
		cacheHdrNonces,
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

func (mphf *MetaPoolsHolderFake) MetaBlockNonces() dataRetriever.Uint64Cacher {
	return mphf.metaBlockNonces
}
