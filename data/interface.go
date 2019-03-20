package data

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

const (
	// TransactionUnit is the transactions storage unit identifier
	TransactionUnit UnitType = 0
	// MiniBlockUnit is the transaction block body storage unit identifier
	MiniBlockUnit UnitType = 1
	// PeerChangesUnit is the peer change block body storage unit identifier
	PeerChangesUnit UnitType = 2
	// BlockHeaderUnit is the Block Headers Storage unit identifier
	BlockHeaderUnit UnitType = 3
	// MetaBlockUnit is the metachain blocks storage unit identifier
	MetaBlockUnit UnitType = 4
	// MetaShardDataUnit is the metachain shard data unit identifier
	MetaShardDataUnit UnitType = 5
	// MetaPeerDataUnit is the metachain peer data unit identifier
	MetaPeerDataUnit UnitType = 6
)

// UnitType is the type for Storage unit identifiers
type UnitType uint8

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
	GetPrevRandSeed() []byte
	GetRandSeed() []byte
	GetPubKeysBitmap() []byte
	GetSignature() []byte

	SetNonce(n uint64)
	SetEpoch(e uint32)
	SetRound(r uint32)
	SetTimeStamp(ts uint64)
	SetRootHash(rHash []byte)
	SetPrevHash(pvHash []byte)
	SetPrevRandSeed(pvRandSeed []byte)
	SetRandSeed(randSeed []byte)
	SetPubKeysBitmap(pkbm []byte)
	SetSignature(sg []byte)
}

// BodyHandler interface for a block body
type BodyHandler interface {
	// IntegrityAndValidity checks the integrity and validity of the block
	IntegrityAndValidity() error
}

// StorageService is the interface for blockChain storage unit provided services
type StorageService interface {
	// GetStorer returns the storer from the chain map
	GetStorer(unitType UnitType) storage.Storer
	// AddStorer will add a new storer to the chain map
	AddStorer(key UnitType, s storage.Storer)
	// Has returns true if the key is found in the selected Unit or false otherwise
	Has(unitType UnitType, key []byte) (bool, error)
	// Get returns the value for the given key if found in the selected storage unit, nil otherwise
	Get(unitType UnitType, key []byte) ([]byte, error)
	// Put stores the key, value pair in the selected storage unit
	Put(unitType UnitType, key []byte, value []byte) error
	// GetAll gets all the elements with keys in the keys array, from the selected storage unit
	// If there is a missing key in the unit, it returns an error
	GetAll(unitType UnitType, keys [][]byte) (map[string][]byte, error)
	// Destroy removes the underlying files/resources used by the storage service
	Destroy() error
}

// ChainHandler is the interface defining the functionality a blockchain should implement
type ChainHandler interface {
	StorageService
	GetGenesisHeader() HeaderHandler
	SetGenesisHeader(gb HeaderHandler) error
	GetGenesisHeaderHash() []byte
	SetGenesisHeaderHash(hash []byte)
	GetCurrentBlockHeader() HeaderHandler
	SetCurrentBlockHeader(bh HeaderHandler) error
	GetCurrentBlockHeaderHash() []byte
	SetCurrentBlockHeaderHash(hash []byte)
	GetCurrentBlockBody() BodyHandler
	SetCurrentBlockBody(body BodyHandler) error
	GetLocalHeight() int64
	SetLocalHeight(height int64)
	GetNetworkHeight() int64
	SetNetworkHeight(height int64)
	IsBadBlock(blockHash []byte) bool
	PutBadBlock(blockHash []byte)
}
