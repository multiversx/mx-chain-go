package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

type ShardedDataCacheNotifierStub struct {
	RegisterHandlerStub         func(func(key []byte))
	ShardDataStoreStub          func(shardID uint32) (c storage.Cacher)
	AddDataStub                 func(key []byte, data interface{}, destShardID uint32)
	SearchDataStub              func(key []byte) (shardValuesPairs map[uint32]interface{})
	RemoveDataStub              func(key []byte, destShardID uint32)
	RemoveSetOfDataFromPoolStub func(keys [][]byte, destShardID uint32)
	RemoveDataFromAllShardsStub func(key []byte)
	MergeShardStoresStub        func(sourceShardID, destShardID uint32)
	MoveDataStub                func(sourceShardID, destShardID uint32, key [][]byte)
	ClearStub                   func()
	ClearShardStoreStub         func(shardID uint32)
}

func (sd *ShardedDataCacheNotifierStub) RegisterHandler(f func(key []byte)) {
	sd.RegisterHandlerStub(f)
}

func (sd *ShardedDataCacheNotifierStub) ShardDataStore(shardID uint32) (c storage.Cacher) {
	return sd.ShardDataStoreStub(shardID)
}

func (sd *ShardedDataCacheNotifierStub) AddData(key []byte, data interface{}, destShardID uint32) {
	sd.AddDataStub(key, data, destShardID)
}

func (sd *ShardedDataCacheNotifierStub) SearchData(key []byte) (shardValuesPairs map[uint32]interface{}) {
	return sd.SearchDataStub(key)
}

func (sd *ShardedDataCacheNotifierStub) RemoveData(key []byte, destShardID uint32) {
	sd.RemoveData(key, destShardID)
}

func (sd *ShardedDataCacheNotifierStub) RemoveSetOfDataFromPool(keys [][]byte, destShardID uint32) {
	sd.RemoveSetOfDataFromPoolStub(keys, destShardID)
}

func (sd *ShardedDataCacheNotifierStub) RemoveDataFromAllShards(key []byte) {
	sd.RemoveDataFromAllShardsStub(key)
}

func (sd *ShardedDataCacheNotifierStub) MergeShardStores(sourceShardID, destShardID uint32) {
	sd.MergeShardStoresStub(sourceShardID, destShardID)
}

func (sd *ShardedDataCacheNotifierStub) MoveData(sourceShardID, destShardID uint32, key [][]byte) {
	sd.MoveDataStub(sourceShardID, destShardID, key)
}

func (sd *ShardedDataCacheNotifierStub) Clear() {
	sd.ClearStub()
}

func (sd *ShardedDataCacheNotifierStub) ClearShardStore(shardID uint32) {
	sd.ClearShardStoreStub(shardID)
}
