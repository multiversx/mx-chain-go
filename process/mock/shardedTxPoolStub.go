package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
)

// ShardedTxPoolStub is a stub
// Implements "dataRetriever.TxPool"
type ShardedTxPoolStub struct {
	RegisterHandlerCalled func(func(key []byte))

	GetTxCacheCalled            func(cacheID string) *txcache.TxCache
	AddTxCalled                 func(txHash []byte, tx data.TransactionHandler, cacheID string)
	SearchFirstTxCalled         func(txHash []byte) (tx data.TransactionHandler, ok bool)
	RemoveTxCalled              func(txHash []byte, cacheID string)
	RemoveTxBulkCalled          func(txHashes [][]byte, cacheID string)
	RemoveTxFromAllShardsCalled func(txHash []byte)
	MergeShardStoresCalled      func(sourceCacheID, destCacheID string)
	MoveTxsCalled               func(sourceCacheID, destCacheID string, txHashes [][]byte)
	ClearCalled                 func()
	ClearShardStoreCalled       func(cacheID string)
	CreateShardStoreCalled      func(cacheID string)
}

// RegisterHandler fakes implementation
func (stub *ShardedTxPoolStub) RegisterHandler(handler func(key []byte)) {
	if stub.RegisterHandlerCalled != nil {
		stub.RegisterHandlerCalled(handler)
	}
}

// GetTxCache fakes implementation
func (stub *ShardedTxPoolStub) GetTxCache(cacheID string) *txcache.TxCache {
	if stub.GetTxCacheCalled == nil {
		panic("ShardedTxPoolStub.GetTxCacheCalled not set")
	}

	return stub.GetTxCacheCalled(cacheID)
}

// AddTx fakes implementation
func (stub *ShardedTxPoolStub) AddTx(txHash []byte, tx data.TransactionHandler, cacheID string) {
	if stub.AddTxCalled == nil {
		panic("ShardedTxPoolStub.AddTxCalled not set")
	}

	stub.AddTxCalled(txHash, tx, cacheID)
}

// SearchFirstTx fakes implementation
func (stub *ShardedTxPoolStub) SearchFirstTx(txHash []byte) (tx data.TransactionHandler, ok bool) {
	if stub.SearchFirstTxCalled == nil {
		panic("ShardedTxPoolStub.SearchFirstTxCalled not set")
	}

	return stub.SearchFirstTxCalled(txHash)
}

// RemoveTx fakes implementation
func (stub *ShardedTxPoolStub) RemoveTx(txHash []byte, cacheID string) {
	if stub.RemoveTxCalled == nil {
		panic("ShardedTxPoolStub.RemoveTxCalled not set")
	}

	stub.RemoveTxCalled(txHash, cacheID)
}

// RemoveTxBulk fakes implementation
func (stub *ShardedTxPoolStub) RemoveTxBulk(txHashes [][]byte, cacheID string) {
	if stub.RemoveTxBulkCalled == nil {
		panic("ShardedTxPoolStub.RemoveTxBulkCalled not set")
	}

	stub.RemoveTxBulkCalled(txHashes, cacheID)
}

// RemoveTxFromAllShards fakes implementation
func (stub *ShardedTxPoolStub) RemoveTxFromAllShards(txHash []byte) {
	if stub.RemoveTxFromAllShardsCalled == nil {
		panic("ShardedTxPoolStub.RemoveTxFromAllShardsCalled not set")
	}

	stub.RemoveTxFromAllShardsCalled(txHash)
}

// MergeShardStores fakes implementation
func (stub *ShardedTxPoolStub) MergeShardStores(sourceCacheID string, destCacheID string) {
	if stub.MergeShardStoresCalled == nil {
		panic("ShardedTxPoolStub.MergeShardStoresCalled not set")
	}

	stub.MergeShardStoresCalled(sourceCacheID, destCacheID)
}

// MoveTxs fakes implementation
func (stub *ShardedTxPoolStub) MoveTxs(sourceCacheID string, destCacheID string, txHashes [][]byte) {
	if stub.MoveTxsCalled == nil {
		panic("ShardedTxPoolStub.MoveTxsCalled not set")
	}

	stub.MoveTxsCalled(sourceCacheID, destCacheID, txHashes)
}

// Clear fakes implementation
func (stub *ShardedTxPoolStub) Clear() {
	if stub.ClearCalled == nil {
		panic("ShardedTxPoolStub.ClearCalled not set")
	}

	stub.ClearCalled()
}

// ClearShardStore fakes implementation
func (stub *ShardedTxPoolStub) ClearShardStore(cacheID string) {
	if stub.ClearShardStoreCalled == nil {
		panic("ShardedTxPoolStub.ClearShardStoreCalled not set")
	}

	stub.ClearShardStoreCalled(cacheID)
}

// CreateShardStore fakes implementation
func (stub *ShardedTxPoolStub) CreateShardStore(cacheID string) {
	if stub.CreateShardStoreCalled == nil {
		panic("ShardedTxPoolStub.CreateShardStoreCalled not set")
	}

	stub.CreateShardStoreCalled(cacheID)
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *ShardedTxPoolStub) IsInterfaceNil() bool {
	if stub == nil {
		return true
	}
	return false
}
