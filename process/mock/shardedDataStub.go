package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

type ShardedDataStub struct {
	RegisterHandlerCalled         func(func(key []byte))
	ShardDataStoreCalled          func(shardID uint32) (c storage.Cacher)
	AddDataCalled                 func(key []byte, data interface{}, destShardID uint32)
	SearchFirstDataCalled         func(key []byte) (value interface{}, ok bool)
	RemoveDataCalled              func(key []byte, destShardID uint32)
	RemoveDataFromAllShardsCalled func(key []byte)
	MergeShardStoresCalled        func(sourceShardID, destShardID uint32)
	MoveDataCalled                func(sourceShardID, destShardID uint32, key [][]byte)
	ClearCalled                   func()
	ClearShardStoreCalled         func(shardID uint32)
	RemoveSetOfDataFromPoolCalled func(keys [][]byte, destShardID uint32)
	CreateShardStoreCalled        func(destShardID uint32)
}

func (sd *ShardedDataStub) RegisterHandler(handler func(key []byte)) {
	sd.RegisterHandlerCalled(handler)
}

func (sd *ShardedDataStub) ShardDataStore(shardID uint32) (c storage.Cacher) {
	return sd.ShardDataStoreCalled(shardID)
}

func (sd *ShardedDataStub) AddData(key []byte, data interface{}, destShardID uint32) {
	sd.AddDataCalled(key, data, destShardID)
}

func (sd *ShardedDataStub) SearchFirstData(key []byte) (value interface{}, ok bool) {
	return sd.SearchFirstDataCalled(key)
}

func (sd *ShardedDataStub) RemoveData(key []byte, destShardID uint32) {
	sd.RemoveDataCalled(key, destShardID)
}

func (sd *ShardedDataStub) RemoveDataFromAllShards(key []byte) {
	sd.RemoveDataFromAllShardsCalled(key)
}

func (sd *ShardedDataStub) MergeShardStores(sourceShardID, destShardID uint32) {
	sd.MergeShardStoresCalled(sourceShardID, destShardID)
}

func (sd *ShardedDataStub) CreateShardStore(destShardID uint32) {
	sd.CreateShardStoreCalled(destShardID)
}

func (sd *ShardedDataStub) MoveData(sourceShardID, destShardID uint32, key [][]byte) {
	sd.MoveDataCalled(sourceShardID, destShardID, key)
}

func (sd *ShardedDataStub) Clear() {
	sd.ClearCalled()
}

func (sd *ShardedDataStub) ClearShardStore(shardID uint32) {
	sd.ClearShardStoreCalled(shardID)
}

func (sd *ShardedDataStub) RemoveSetOfDataFromPool(keys [][]byte, destShardID uint32) {
	sd.RemoveSetOfDataFromPoolCalled(keys, destShardID)
}
