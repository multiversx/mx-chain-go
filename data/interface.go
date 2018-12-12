package data

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

type CacheNotifier interface {
	RegisterAddedDataHandler(func(key []byte))
}

type ShardedDataCacherNotifier interface {
	CacheNotifier

	ShardDataStore(shardID uint32) (c storage.Cacher)
	AddData(key []byte, data interface{}, destShardID uint32)
	SearchData(key []byte) (shardValuesPairs map[uint32]interface{})
	RemoveData(key []byte, destShardID uint32)
	RemoveDataFromAllShards(key []byte)
	MergeShardStores(sourceShardID, destShardID uint32)
	MoveData(sourceShardID, destShardID uint32, key [][]byte)
	Clear()
	ClearMiniPool(shardID uint32)
}
