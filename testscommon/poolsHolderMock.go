package testscommon

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool/headersCache"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/txpool"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	 "github.com/ElrondNetwork/elrond-go/testscommon/txcachemocks"
)

// PoolsHolderMock -
type PoolsHolderMock struct {
	transactions         dataRetriever.ShardedDataCacherNotifier
	unsignedTransactions dataRetriever.ShardedDataCacherNotifier
	rewardTransactions   dataRetriever.ShardedDataCacherNotifier
	headers              dataRetriever.HeadersPool
	miniBlocks           storage.Cacher
	peerChangesBlocks    storage.Cacher
	trieNodes            storage.Cacher
	smartContracts       storage.Cacher
	currBlockTxs         dataRetriever.TransactionCacher
}

// NewPoolsHolderMock -
func NewPoolsHolderMock() *PoolsHolderMock {
	var err error
	holder := &PoolsHolderMock{}

	holder.transactions, err = txpool.NewShardedTxPool(
		txpool.ArgShardedTxPool{
			Config: storageUnit.CacheConfig{
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

	holder.unsignedTransactions, err = shardedData.NewShardedData("unsignedTxPool", storageUnit.CacheConfig{
		Capacity:    10000,
		SizeInBytes: 1000000000,
		Shards:      1,
	})
	panicIfError("NewPoolsHolderMock", err)

	holder.rewardTransactions, err = shardedData.NewShardedData("rewardsTxPool", storageUnit.CacheConfig{
		Capacity:    100,
		SizeInBytes: 100000,
		Shards:      1,
	})
	panicIfError("NewPoolsHolderMock", err)

	holder.headers, err = headersCache.NewHeadersPool(config.HeadersPoolConfig{MaxHeadersPerShard: 1000, NumElementsToRemoveOnEviction: 100})
	panicIfError("NewPoolsHolderMock", err)

	holder.miniBlocks, err = storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 10000, Shards: 1, SizeInBytes: 0})
	panicIfError("NewPoolsHolderMock", err)

	holder.peerChangesBlocks, err = storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 10000, Shards: 1, SizeInBytes: 0})
	panicIfError("NewPoolsHolderMock", err)

	holder.currBlockTxs, err = dataPool.NewCurrentBlockPool()
	panicIfError("NewPoolsHolderMock", err)

	holder.trieNodes, err = storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 10000, Shards: 1, SizeInBytes: 0})
	panicIfError("NewPoolsHolderMock", err)

	holder.smartContracts, err = storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 10000, Shards: 1, SizeInBytes: 0})
	panicIfError("NewPoolsHolderMock", err)

	return holder
}

// CurrentBlockTxs -
func (holder *PoolsHolderMock) CurrentBlockTxs() dataRetriever.TransactionCacher {
	return holder.currBlockTxs
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

// SmartContracts -
func (holder *PoolsHolderMock) SmartContracts() storage.Cacher {
	return holder.smartContracts
}

// IsInterfaceNil returns true if there is no value under the interface
func (holder *PoolsHolderMock) IsInterfaceNil() bool {
	return holder == nil
}
