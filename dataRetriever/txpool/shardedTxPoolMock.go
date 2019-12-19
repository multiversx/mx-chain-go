package txpool

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
)

// ShardedTxPoolMock is a mock
// Implements "dataRetriever.TxPool"
type ShardedTxPoolMock struct {
	implementation        dataRetriever.TxPool
	RegisterHandlerCalled func(func(key []byte))

	ShardDataStoreCalled          func(cacheID string) storage.Cacher
	AddDataCalled                 func(key []byte, data interface{}, cacheID string)
	SearchFirstDataCalled         func(key []byte) (value interface{}, ok bool)
	RemoveDataCalled              func(key []byte, cacheID string)
	RemoveDataFromAllShardsCalled func(key []byte)
	MoveDataCalled                func(sourceCacheID, destCacheID string, key [][]byte)
	RemoveSetOfDataFromPoolCalled func(keys [][]byte, destCacheId string)

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

// NewShardedTxPoolMock creates a new fake implementation
func NewShardedTxPoolMock() *ShardedTxPoolMock {
	mock := &ShardedTxPoolMock{
		implementation: NewShardedTxPool(storageUnit.CacheConfig{}),
	}

	return mock
}

// RegisterHandler fakes implementation
func (mock *ShardedTxPoolMock) RegisterHandler(handler func(key []byte)) {
	if mock.RegisterHandlerCalled != nil {
		mock.RegisterHandlerCalled(handler)
	}
}

// ShardDataStore fakes implementation
func (mock *ShardedTxPoolMock) ShardDataStore(cacheID string) (c storage.Cacher) {
	if mock.ShardDataStoreCalled != nil {
		return mock.ShardDataStoreCalled(cacheID)
	}

	return mock.implementation.ShardDataStore(cacheID)
}

// AddData fakes implementation
func (mock *ShardedTxPoolMock) AddData(key []byte, data interface{}, cacheID string) {
	if mock.AddDataCalled != nil {
		mock.AddDataCalled(key, data, cacheID)
	} else {
		mock.implementation.AddData(key, data, cacheID)
	}
}

// SearchFirstData fakes implementation
func (mock *ShardedTxPoolMock) SearchFirstData(key []byte) (value interface{}, ok bool) {
	if mock.SearchFirstDataCalled != nil {
		return mock.SearchFirstDataCalled(key)
	}

	return mock.implementation.SearchFirstData(key)
}

// RemoveData fakes implementation
func (mock *ShardedTxPoolMock) RemoveData(key []byte, cacheID string) {
	if mock.RemoveDataCalled != nil {
		mock.RemoveDataCalled(key, cacheID)
	} else {
		mock.implementation.RemoveData(key, cacheID)
	}
}

// RemoveDataFromAllShards fakes implementation
func (mock *ShardedTxPoolMock) RemoveDataFromAllShards(key []byte) {
	if mock.RemoveDataFromAllShardsCalled != nil {
		mock.RemoveDataFromAllShardsCalled(key)
	} else {
		mock.implementation.RemoveDataFromAllShards(key)
	}
}

// MergeShardStores fakes implementation
func (mock *ShardedTxPoolMock) MergeShardStores(sourceCacheID, destCacheID string) {
	if mock.MergeShardStoresCalled != nil {
		mock.MergeShardStoresCalled(sourceCacheID, destCacheID)
	} else {
		mock.implementation.MergeShardStores(sourceCacheID, destCacheID)
	}
}

// MoveData fakes implementation
func (mock *ShardedTxPoolMock) MoveData(sourceCacheID, destCacheID string, key [][]byte) {
	if mock.MoveDataCalled != nil {
		mock.MoveDataCalled(sourceCacheID, destCacheID, key)
	} else {
		mock.implementation.MoveData(sourceCacheID, destCacheID, key)
	}
}

// RemoveSetOfDataFromPool fakes implementation
func (mock *ShardedTxPoolMock) RemoveSetOfDataFromPool(keys [][]byte, cacheID string) {
	if mock.RemoveSetOfDataFromPoolCalled != nil {
		mock.RemoveSetOfDataFromPoolCalled(keys, cacheID)
	} else {
		mock.implementation.RemoveSetOfDataFromPool(keys, cacheID)
	}
}

// GetTxCache fakes implementation
func (mock *ShardedTxPoolMock) GetTxCache(cacheID string) *txcache.TxCache {
	if mock.GetTxCacheCalled == nil {
		panic("ShardedTxPoolMock.GetTxCacheCalled not set")
	}

	return mock.GetTxCacheCalled(cacheID)
}

// AddTx fakes implementation
func (mock *ShardedTxPoolMock) AddTx(txHash []byte, tx data.TransactionHandler, cacheID string) {
	if mock.AddTxCalled == nil {
		panic("ShardedTxPoolMock.AddTxCalled not set")
	}

	mock.AddTxCalled(txHash, tx, cacheID)
}

// SearchFirstTx fakes implementation
func (mock *ShardedTxPoolMock) SearchFirstTx(txHash []byte) (tx data.TransactionHandler, ok bool) {
	if mock.SearchFirstTxCalled == nil {
		panic("ShardedTxPoolMock.SearchFirstTxCalled not set")
	}

	return mock.SearchFirstTxCalled(txHash)
}

// RemoveTx fakes implementation
func (mock *ShardedTxPoolMock) RemoveTx(txHash []byte, cacheID string) {
	if mock.RemoveTxCalled == nil {
		panic("ShardedTxPoolMock.RemoveTxCalled not set")
	}

	mock.RemoveTxCalled(txHash, cacheID)
}

// RemoveTxBulk fakes implementation
func (mock *ShardedTxPoolMock) RemoveTxBulk(txHashes [][]byte, cacheID string) {
	if mock.RemoveTxBulkCalled == nil {
		panic("ShardedTxPoolMock.RemoveTxBulkCalled not set")
	}

	mock.RemoveTxBulkCalled(txHashes, cacheID)
}

// RemoveTxFromAllShards fakes implementation
func (mock *ShardedTxPoolMock) RemoveTxFromAllShards(txHash []byte) {
	if mock.RemoveTxFromAllShardsCalled == nil {
		panic("ShardedTxPoolMock.RemoveTxFromAllShardsCalled not set")
	}

	mock.RemoveTxFromAllShardsCalled(txHash)
}

// MoveTxs fakes implementation
func (mock *ShardedTxPoolMock) MoveTxs(sourceCacheID string, destCacheID string, txHashes [][]byte) {
	if mock.MoveTxsCalled == nil {
		panic("ShardedTxPoolMock.MoveTxsCalled not set")
	}

	mock.MoveTxsCalled(sourceCacheID, destCacheID, txHashes)
}

// Clear fakes implementation
func (mock *ShardedTxPoolMock) Clear() {
	if mock.ClearCalled == nil {
		panic("ShardedTxPoolMock.ClearCalled not set")
	}

	mock.ClearCalled()
}

// ClearShardStore fakes implementation
func (mock *ShardedTxPoolMock) ClearShardStore(cacheID string) {
	if mock.ClearShardStoreCalled == nil {
		panic("ShardedTxPoolMock.ClearShardStoreCalled not set")
	}

	mock.ClearShardStoreCalled(cacheID)
}

// CreateShardStore fakes implementation
func (mock *ShardedTxPoolMock) CreateShardStore(cacheID string) {
	if mock.CreateShardStoreCalled == nil {
		panic("ShardedTxPoolMock.CreateShardStoreCalled not set")
	}

	mock.CreateShardStoreCalled(cacheID)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mock *ShardedTxPoolMock) IsInterfaceNil() bool {
	if mock == nil {
		return true
	}
	return false
}
