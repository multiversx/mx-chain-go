package mock

import (
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ShardedDataStub -
type ShardedDataStub struct {
	RegisterHandlerCalled         func(func(key []byte))
	ShardDataStoreCalled          func(cacheId string) (c storage.Cacher)
	AddDataCalled                 func(key []byte, data interface{}, cacheId string)
	SearchFirstDataCalled         func(key []byte) (value interface{}, ok bool)
	RemoveDataCalled              func(key []byte, cacheId string)
	RemoveDataFromAllShardsCalled func(key []byte)
	MergeShardStoresCalled        func(sourceCacheId, destCacheId string)
	MoveDataCalled                func(sourceCacheId, destCacheId string, key [][]byte)
	ClearCalled                   func()
	ClearShardStoreCalled         func(cacheId string)
	RemoveSetOfDataFromPoolCalled func(keys [][]byte, destCacheId string)
	CreateShardStoreCalled        func(destCacheId string)
}

// RegisterHandler -
func (sd *ShardedDataStub) RegisterHandler(handler func(key []byte)) {
	sd.RegisterHandlerCalled(handler)
}

// ShardDataStore -
func (sd *ShardedDataStub) ShardDataStore(cacheId string) (c storage.Cacher) {
	return sd.ShardDataStoreCalled(cacheId)
}

// AddData -
func (sd *ShardedDataStub) AddData(key []byte, data interface{}, cacheId string) {
	sd.AddDataCalled(key, data, cacheId)
}

// SearchFirstData -
func (sd *ShardedDataStub) SearchFirstData(key []byte) (value interface{}, ok bool) {
	return sd.SearchFirstDataCalled(key)
}

// RemoveData -
func (sd *ShardedDataStub) RemoveData(key []byte, cacheId string) {
	sd.RemoveDataCalled(key, cacheId)
}

// RemoveDataFromAllShards -
func (sd *ShardedDataStub) RemoveDataFromAllShards(key []byte) {
	sd.RemoveDataFromAllShardsCalled(key)
}

// MergeShardStores -
func (sd *ShardedDataStub) MergeShardStores(sourceCacheId, destCacheId string) {
	sd.MergeShardStoresCalled(sourceCacheId, destCacheId)
}

// Clear -
func (sd *ShardedDataStub) Clear() {
	sd.ClearCalled()
}

// ClearShardStore -
func (sd *ShardedDataStub) ClearShardStore(cacheId string) {
	sd.ClearShardStoreCalled(cacheId)
}

// RemoveSetOfDataFromPool -
func (sd *ShardedDataStub) RemoveSetOfDataFromPool(keys [][]byte, cacheId string) {
	sd.RemoveSetOfDataFromPoolCalled(keys, cacheId)
}

// CreateShardStore -
func (sd *ShardedDataStub) CreateShardStore(cacheId string) {
	sd.CreateShardStoreCalled(cacheId)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sd *ShardedDataStub) IsInterfaceNil() bool {
	if sd == nil {
		return true
	}
	return false
}
