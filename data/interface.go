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

// PoolsHolder defines getters for data pools
type PoolsHolder interface {
	Transactions() ShardedDataCacherNotifier
	Headers() ShardedDataCacherNotifier
	HeadersNonces() Uint64Cacher
	MiniBlocks() storage.Cacher
	PeerChangesBlocks() storage.Cacher
	MetaBlocks() storage.Cacher
}

// MetaPoolsHolder defines getter for data pools for metachain
type MetaPoolsHolder interface {
	MetaChainBlocks() storage.Cacher
	MiniBlockHashes() ShardedDataCacherNotifier
	ShardHeaders() ShardedDataCacherNotifier
	MetaBlockNonces() Uint64Cacher
}

// HeaderHandler defines getters and setters for header data holder
type HeaderHandler interface {
	GetNonce() uint64
	GetEpoch() uint32
	GetRound() uint32
	GetRootHash() []byte
	GetPrevHash() []byte
	GetPubKeysBitmap() []byte
	GetSignature() []byte

	SetNonce(n uint64)
	SetEpoch(e uint32)
	SetRound(r uint32)
	SetTimeStamp(ts uint64)
	SetRootHash(rHash []byte)
	SetPrevHash(pvHash []byte)
	SetPubKeysBitmap(pkbm []byte)
	SetSignature(sg []byte)
	SetCommitment(commitment []byte)
}

// BodyHandler interface for a block body
type BodyHandler interface {
	// checks the integrity and validity of the block
	IntegrityAndValidity() bool
}
