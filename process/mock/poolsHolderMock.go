package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/txpool"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

type PoolsHolderMock struct {
	transactions         dataRetriever.TxPool
	unsignedTransactions dataRetriever.TxPool
	rewardTransactions   dataRetriever.TxPool
	headers              storage.Cacher
	metaBlocks           storage.Cacher
	hdrNonces            dataRetriever.Uint64SyncMapCacher
	miniBlocks           storage.Cacher
	peerChangesBlocks    storage.Cacher
	metaHdrNonces        dataRetriever.Uint64SyncMapCacher
	currBlockTxs         dataRetriever.TransactionCacher
}

func NewPoolsHolderMock() *PoolsHolderMock {
	holder := &PoolsHolderMock{}

	holder.transactions = txpool.NewShardedTxPool(storageUnit.CacheConfig{Size: 10000, Type: storageUnit.LRUCache})
	holder.unsignedTransactions = txpool.NewShardedTxPool(storageUnit.CacheConfig{Size: 10000, Type: storageUnit.LRUCache})
	holder.rewardTransactions = txpool.NewShardedTxPool(storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache})

	holder.headers, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	holder.metaBlocks, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	cacheHdrNonces, _ := storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	holder.hdrNonces, _ = dataPool.NewNonceSyncMapCacher(
		cacheHdrNonces,
		uint64ByteSlice.NewBigEndianConverter(),
	)
	cacheMetaHdrNonces, _ := storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	holder.metaHdrNonces, _ = dataPool.NewNonceSyncMapCacher(
		cacheMetaHdrNonces,
		uint64ByteSlice.NewBigEndianConverter(),
	)
	holder.miniBlocks, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	holder.peerChangesBlocks, _ = storageUnit.NewCache(storageUnit.LRUCache, 10000, 1)
	holder.currBlockTxs, _ = dataPool.NewCurrentBlockPool()

	return holder
}

func (holder *PoolsHolderMock) CurrentBlockTxs() dataRetriever.TransactionCacher {
	return holder.currBlockTxs
}

func (holder *PoolsHolderMock) Transactions() dataRetriever.TxPool {
	return holder.transactions
}

func (holder *PoolsHolderMock) UnsignedTransactions() dataRetriever.TxPool {
	return holder.unsignedTransactions
}

func (holder *PoolsHolderMock) RewardTransactions() dataRetriever.TxPool {
	return holder.rewardTransactions
}

func (holder *PoolsHolderMock) Headers() storage.Cacher {
	return holder.headers
}

func (holder *PoolsHolderMock) HeadersNonces() dataRetriever.Uint64SyncMapCacher {
	return holder.hdrNonces
}

func (holder *PoolsHolderMock) MiniBlocks() storage.Cacher {
	return holder.miniBlocks
}

func (holder *PoolsHolderMock) PeerChangesBlocks() storage.Cacher {
	return holder.peerChangesBlocks
}

func (holder *PoolsHolderMock) MetaBlocks() storage.Cacher {
	return holder.metaBlocks
}

func (holder *PoolsHolderMock) MetaHeadersNonces() dataRetriever.Uint64SyncMapCacher {
	return holder.metaHdrNonces
}

func (holder *PoolsHolderMock) SetTransactions(txPool dataRetriever.TxPool) {
	holder.transactions = txPool
}

func (holder *PoolsHolderMock) SetUnsignedTransactions(txPool dataRetriever.TxPool) {
	holder.unsignedTransactions = txPool
}

// IsInterfaceNil returns true if there is no value under the interface
func (phf *PoolsHolderMock) IsInterfaceNil() bool {
	if phf == nil {
		return true
	}
	return false
}
