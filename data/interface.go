package data

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// Notifier defines a way to register funcs that get called when something useful happens
type Notifier interface {
	RegisterHandler(func(key []byte))
}

// ShardedDataCacherNotifier defines what a sharded-data structure can perform
type ShardedDataCacherNotifier interface {
	Notifier

	ShardDataStore(shardID uint32) (c storage.Cacher)
	AddData(key []byte, data interface{}, destShardID uint32)
	//SearchData(key []byte) (shardValuesPairs map[uint32]interface{})
	SearchFirstData(key []byte) (value interface{}, ok bool)
	RemoveData(key []byte, destShardID uint32)
	RemoveSetOfDataFromPool(keys [][]byte, destShardID uint32)
	RemoveDataFromAllShards(key []byte)
	MergeShardStores(sourceShardID, destShardID uint32)
	MoveData(sourceShardID, destShardID uint32, key [][]byte)
	Clear()
	ClearShardStore(shardID uint32)
	CreateShardStore(destShardID uint32)
}

// Uint64Cacher defines a cacher-type struct that uses uint64 keys and []byte values (usually hashes)
type Uint64Cacher interface {
	Clear()
	Put(uint64, []byte) bool
	Get(uint64) ([]byte, bool)
	Has(uint64) bool
	Peek(uint64) ([]byte, bool)
	HasOrAdd(uint64, []byte) (bool, bool)
	Remove(uint64)
	RemoveOldest()
	Keys() []uint64
	Len() int
	RegisterHandler(handler func(nonce uint64))
}

// TransientDataHolder defines getters for a transient data holder
type TransientDataHolder interface {
	Transactions() ShardedDataCacherNotifier
	Headers() ShardedDataCacherNotifier
	HeadersNonces() Uint64Cacher
	TxBlocks() storage.Cacher
	PeerChangesBlocks() storage.Cacher
	StateBlocks() storage.Cacher
}
