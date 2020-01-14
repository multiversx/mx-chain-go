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

type MetaPoolsHolderFake struct {
	miniBlocks   storage.Cacher
	trieNodes     storage.Cacher
	shardHeaders dataRetriever.HeadersPool
	transactions dataRetriever.ShardedDataCacherNotifier
	unsigned     dataRetriever.ShardedDataCacherNotifier
	currTxs      dataRetriever.TransactionCacher

	ShardHeadersCalled func() dataRetriever.HeadersPool
}

func NewMetaPoolsHolderFake() *MetaPoolsHolderFake {
	mphf := &MetaPoolsHolderFake{}
	mphf.miniBlocks, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	mphf.transactions, _ = txpool.NewShardedTxPool(storageUnit.CacheConfig{Size: 10000, Shards: 1})
	mphf.unsigned, _ = shardedData.NewShardedData(storageUnit.CacheConfig{Size: 10000, Type: storageUnit.LRUCache})
	mphf.currTxs, _ = dataPool.NewCurrentBlockPool()
	mphf.shardHeaders, _ = headersCache.NewHeadersPool(config.HeadersPoolConfig{MaxHeadersPerShard: 1000, NumElementsToRemoveOnEviction: 100})
	mphf.trieNodes, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)

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

func (mphf *MetaPoolsHolderFake) MiniBlocks() storage.Cacher {
	return mphf.miniBlocks
}

func (mphf *MetaPoolsHolderFake) TrieNodes() storage.Cacher {
	return mphf.trieNodes
}

func (mphf *MetaPoolsHolderFake) Headers() dataRetriever.HeadersPool {
	if mphf.ShardHeadersCalled != nil {
		return mphf.ShardHeadersCalled()
	}
	return mphf.shardHeaders
}

// IsInterfaceNil returns true if there is no value under the interface
func (mphf *MetaPoolsHolderFake) IsInterfaceNil() bool {
	if mphf == nil {
		return true
	}
	return false
}
