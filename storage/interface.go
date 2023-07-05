package storage

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/storage"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-storage-go/types"
)

// Cacher provides caching services
type Cacher interface {
	// Clear is used to completely clear the cache.
	Clear()
	// Put adds a value to the cache.  Returns true if an eviction occurred.
	Put(key []byte, value interface{}, sizeInBytes int) (evicted bool)
	// Get looks up a key's value from the cache.
	Get(key []byte) (value interface{}, ok bool)
	// Has checks if a key is in the cache, without updating the
	// recent-ness or deleting it for being stale.
	Has(key []byte) bool
	// Peek returns the key value (or undefined if not found) without updating
	// the "recently used"-ness of the key.
	Peek(key []byte) (value interface{}, ok bool)
	// HasOrAdd checks if a key is in the cache without updating the
	// recent-ness or deleting it for being stale, and if not adds the value.
	HasOrAdd(key []byte, value interface{}, sizeInBytes int) (has, added bool)
	// Remove removes the provided key from the cache.
	Remove(key []byte)
	// Keys returns a slice of the keys in the cache, from oldest to newest.
	Keys() [][]byte
	// Len returns the number of items in the cache.
	Len() int
	// SizeInBytesContained returns the size in bytes of all contained elements
	SizeInBytesContained() uint64
	// MaxSize returns the maximum number of items which can be stored in the cache.
	MaxSize() int
	// RegisterHandler registers a new handler to be called when a new data is added
	RegisterHandler(handler func(key []byte, value interface{}), id string)
	// UnRegisterHandler deletes the handler from the list
	UnRegisterHandler(id string)
	// Close closes the underlying temporary db if the cacher implementation has one,
	// otherwise it does nothing
	Close() error
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

// Persister provides storage of data services in a database like construct
type Persister = types.Persister

// Batcher allows to batch the data first then write the batch to the persister in one go
type Batcher interface {
	// Put inserts one entry - key, value pair - into the batch
	Put(key []byte, val []byte) error
	// Get returns the value from the batch
	Get(key []byte) []byte
	// Delete deletes the batch
	Delete(key []byte) error
	// Reset clears the contents of the batch
	Reset()
	// IsRemoved returns true if the provided key is marked for deletion
	IsRemoved(key []byte) bool
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

// Storer provides storage services in a two layered storage construct, where the first layer is
// represented by a cache and second layer by a persistent storage (DB-like)
type Storer interface {
	Put(key, data []byte) error
	PutInEpoch(key, data []byte, epoch uint32) error
	Get(key []byte) ([]byte, error)
	Has(key []byte) error
	SearchFirst(key []byte) ([]byte, error)
	RemoveFromCurrentEpoch(key []byte) error
	Remove(key []byte) error
	ClearCache()
	DestroyUnit() error
	GetFromEpoch(key []byte, epoch uint32) ([]byte, error)
	GetBulkFromEpoch(keys [][]byte, epoch uint32) ([]storage.KeyValuePair, error)
	GetOldestEpoch() (uint32, error)
	RangeKeys(handler func(key []byte, val []byte) bool)
	Close() error
	IsInterfaceNil() bool
}

// StorerWithPutInEpoch is an extended storer with the ability to set the epoch which will be used for put operations
type StorerWithPutInEpoch interface {
	Storer
	SetEpochForPutOperation(epoch uint32)
}

// PathManagerHandler defines which actions should be done for generating paths for databases directories
type PathManagerHandler interface {
	PathForEpoch(shardId string, epoch uint32, identifier string) string
	PathForStatic(shardId string, identifier string) string
	DatabasePath() string
	IsInterfaceNil() bool
}

// PersisterFactory defines which actions should be done for creating a persister
type PersisterFactory interface {
	Create(path string) (Persister, error)
	IsInterfaceNil() bool
}

// UnitOpenerHandler defines which actions should be done for opening storage units
type UnitOpenerHandler interface {
	OpenDB(dbConfig config.DBConfig, shardID uint32, epoch uint32) (Storer, error)
	GetMostRecentStorageUnit(config config.DBConfig) (Storer, error)
	IsInterfaceNil() bool
}

// DirectoryReaderHandler defines which actions should be done by a directory reader
type DirectoryReaderHandler interface {
	ListFilesAsString(directoryPath string) ([]string, error)
	ListDirectoriesAsString(directoryPath string) ([]string, error)
	ListAllAsString(directoryPath string) ([]string, error)
	IsInterfaceNil() bool
}

// LatestStorageDataProviderHandler defines which actions be done by a component who fetches the latest data from storage
type LatestStorageDataProviderHandler interface {
	GetParentDirectory() string
	GetParentDirAndLastEpoch() (string, uint32, error)
	Get() (LatestDataFromStorage, error)
	GetShardsFromDirectory(path string) ([]string, error)
	IsInterfaceNil() bool
}

// LatestDataFromStorage represents the DTO structure to return from storage
type LatestDataFromStorage = types.LatestDataFromStorage

// ShardCoordinator defines what a shard state coordinator should hold
type ShardCoordinator interface {
	NumberOfShards() uint32
	ComputeId(address []byte) uint32
	SelfId() uint32
	SameShard(firstAddress, secondAddress []byte) bool
	CommunicationIdentifier(destShardID uint32) string
	IsInterfaceNil() bool
}

// TimeCacher defines the cache that can keep a record for a bounded time
type TimeCacher interface {
	Add(key string) error
	Upsert(key string, span time.Duration) error
	Has(key string) bool
	Sweep()
	IsInterfaceNil() bool
}

// StoredDataFactory creates empty objects of the stored data type
type StoredDataFactory interface {
	CreateEmpty() interface{}
	IsInterfaceNil() bool
}

// CustomDatabaseRemoverHandler defines the behaviour of a component that should tell if a database is removable or not
type CustomDatabaseRemoverHandler interface {
	ShouldRemove(dbIdentifier string, epoch uint32) bool
	IsInterfaceNil() bool
}

// SizedLRUCacheHandler is the interface for size capable LRU cache.
type SizedLRUCacheHandler interface {
	AddSized(key, value interface{}, sizeInBytes int64) bool
	Get(key interface{}) (value interface{}, ok bool)
	Contains(key interface{}) (ok bool)
	AddSizedIfMissing(key, value interface{}, sizeInBytes int64) (ok, evicted bool)
	Peek(key interface{}) (value interface{}, ok bool)
	Remove(key interface{}) bool
	Keys() []interface{}
	Len() int
	SizeInBytesContained() uint64
	Purge()
}

// AdaptedSizedLRUCache defines a cache that returns the evicted value
type AdaptedSizedLRUCache interface {
	SizedLRUCacheHandler
	AddSizedAndReturnEvicted(key, value interface{}, sizeInBytes int64) map[interface{}]interface{}
	IsInterfaceNil() bool
}

// ShardIDProvider defines what a component which is able to provide persister id per key should do
type ShardIDProvider interface {
	ComputeId(key []byte) uint32
	NumberOfShards() uint32
	GetShardIDs() []uint32
	IsInterfaceNil() bool
}

// PersisterCreator defines the behavour of a component which is able to create a persister
type PersisterCreator = types.PersisterCreator

// DBConfigHandler defines the behaviour of a component that will handle db config
type DBConfigHandler interface {
	GetDBConfig(path string) (*config.DBConfig, error)
	SaveDBConfigToFilePath(path string, dbConfig *config.DBConfig) error
	IsInterfaceNil() bool
}

// ManagedPeersHolder defines the operations of an entity that holds managed identities for a node
type ManagedPeersHolder interface {
	IsMultiKeyMode() bool
	IsInterfaceNil() bool
}

type StateStatisticsHandler interface {
	Reset()
	AddNumCache(value int)
	NumCache() int
	AddNumPersister(value int)
	NumPersister() int
	AddNumTrie(value int)
	NumTrie() int
	IsInterfaceNil() bool
}
