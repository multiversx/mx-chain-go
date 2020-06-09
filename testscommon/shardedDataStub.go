package testscommon

import (
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ShardedDataStub -
type ShardedDataStub struct {
	RegisterHandlerCalled                  func(func(key []byte, value interface{}))
	ShardDataStoreCalled                   func(cacheID string) storage.Cacher
	AddDataCalled                          func(key []byte, data interface{}, sizeInBytes int, cacheID string)
	SearchFirstDataCalled                  func(key []byte) (value interface{}, ok bool)
	RemoveDataCalled                       func(key []byte, cacheID string)
	RemoveDataFromAllShardsCalled          func(key []byte)
	MergeShardStoresCalled                 func(sourceCacheID, destCacheID string)
	MoveDataCalled                         func(sourceCacheID, destCacheID string, key [][]byte)
	ClearCalled                            func()
	ClearShardStoreCalled                  func(cacheID string)
	RemoveSetOfDataFromPoolCalled          func(keys [][]byte, destCacheID string)
	ImmunizeSetOfDataAgainstEvictionCalled func(keys [][]byte, cacheID string)
	CreateShardStoreCalled                 func(destCacheID string)
}

// NewShardedDataStub -
func NewShardedDataStub() *ShardedDataStub {
	return &ShardedDataStub{}
}

// RegisterHandler -
func (shardedData *ShardedDataStub) RegisterHandler(handler func(key []byte, value interface{})) {
	if shardedData.RegisterHandlerCalled != nil {
		shardedData.RegisterHandlerCalled(handler)
	}
}

// ShardDataStore -
func (shardedData *ShardedDataStub) ShardDataStore(cacheID string) storage.Cacher {
	return shardedData.ShardDataStoreCalled(cacheID)
}

// AddData -
func (shardedData *ShardedDataStub) AddData(key []byte, data interface{}, sizeInBytes int, cacheID string) {
	shardedData.AddDataCalled(key, data, sizeInBytes, cacheID)
}

// SearchFirstData -
func (shardedData *ShardedDataStub) SearchFirstData(key []byte) (value interface{}, ok bool) {
	return shardedData.SearchFirstDataCalled(key)
}

// RemoveData -
func (shardedData *ShardedDataStub) RemoveData(key []byte, cacheID string) {
	shardedData.RemoveDataCalled(key, cacheID)
}

// RemoveDataFromAllShards -
func (shardedData *ShardedDataStub) RemoveDataFromAllShards(key []byte) {
	shardedData.RemoveDataFromAllShardsCalled(key)
}

// MergeShardStores -
func (shardedData *ShardedDataStub) MergeShardStores(sourceCacheID, destCacheID string) {
	shardedData.MergeShardStoresCalled(sourceCacheID, destCacheID)
}

// Clear -
func (shardedData *ShardedDataStub) Clear() {
	shardedData.ClearCalled()
}

// ClearShardStore -
func (shardedData *ShardedDataStub) ClearShardStore(cacheID string) {
	shardedData.ClearShardStoreCalled(cacheID)
}

// RemoveSetOfDataFromPool -
func (shardedData *ShardedDataStub) RemoveSetOfDataFromPool(keys [][]byte, cacheID string) {
	shardedData.RemoveSetOfDataFromPoolCalled(keys, cacheID)
}

// ImmunizeSetOfDataAgainstEviction -
func (shardedData *ShardedDataStub) ImmunizeSetOfDataAgainstEviction(keys [][]byte, cacheID string) {
	if shardedData.ImmunizeSetOfDataAgainstEvictionCalled != nil {
		shardedData.ImmunizeSetOfDataAgainstEvictionCalled(keys, cacheID)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (shardedData *ShardedDataStub) IsInterfaceNil() bool {
	return shardedData == nil
}
