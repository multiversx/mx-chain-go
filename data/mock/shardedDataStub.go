package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

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

func (sd *ShardedDataStub) RegisterHandler(handler func(key []byte)) {
	sd.RegisterHandlerCalled(handler)
}

func (sd *ShardedDataStub) ShardDataStore(cacheId string) (c storage.Cacher) {
	return sd.ShardDataStoreCalled(cacheId)
}

func (sd *ShardedDataStub) AddData(key []byte, data interface{}, cacheId string) {
	sd.AddDataCalled(key, data, cacheId)
}

func (sd *ShardedDataStub) SearchFirstData(key []byte) (value interface{}, ok bool) {
	return sd.SearchFirstDataCalled(key)
}

func (sd *ShardedDataStub) RemoveData(key []byte, cacheId string) {
	sd.RemoveDataCalled(key, cacheId)
}

func (sd *ShardedDataStub) RemoveDataFromAllShards(key []byte) {
	sd.RemoveDataFromAllShardsCalled(key)
}

func (sd *ShardedDataStub) MergeShardStores(sourceCacheId, destCacheId string) {
	sd.MergeShardStoresCalled(sourceCacheId, destCacheId)
}

func (sd *ShardedDataStub) MoveData(sourceCacheId, destCacheId string, key [][]byte) {
	sd.MoveDataCalled(sourceCacheId, destCacheId, key)
}

func (sd *ShardedDataStub) Clear() {
	sd.ClearCalled()
}

func (sd *ShardedDataStub) ClearShardStore(cacheId string) {
	sd.ClearShardStoreCalled(cacheId)
}

func (sd *ShardedDataStub) RemoveSetOfDataFromPool(keys [][]byte, cacheId string) {
	sd.RemoveSetOfDataFromPoolCalled(keys, cacheId)
}

func (sd *ShardedDataStub) CreateShardStore(cacheId string) {
	sd.CreateShardStoreCalled(cacheId)
}
