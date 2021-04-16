package storage

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// Persister provides storage of data services in a database like construct
type Persister interface {
	// Put add the value to the (key, val) persistence medium
	Put(key, val []byte) error
	// Get gets the value associated to the key
	Get(key []byte) ([]byte, error)
	// Has returns true if the given key is present in the persistence medium
	Has(key []byte) error
	// Init initializes the persistence medium and prepares it for usage
	Init() error
	// Close closes the files/resources associated to the persistence medium
	Close() error
	// Remove removes the data associated to the given key
	Remove(key []byte) error
	// Destroy removes the persistence medium stored data
	Destroy() error
	// DestroyClosed removes the already closed persistence medium stored data
	DestroyClosed() error
	RangeKeys(handler func(key []byte, val []byte) bool)
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

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
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

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
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

// BloomFilter provides services for filtering database requests
type BloomFilter interface {
	//Add adds the value to the bloom filter
	Add([]byte)
	// MayContain checks if the value is in in the set. If it returns 'false',
	//the item is definitely not in the DB
	MayContain([]byte) bool
	//Clear sets all the bits from the filter to 0
	Clear()
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

// Storer provides storage services in a two layered storage construct, where the first layer is
// represented by a cache and second layer by a persitent storage (DB-like)
type Storer interface {
	Put(key, data []byte) error
	PutInEpoch(key, data []byte, epoch uint32) error
	Get(key []byte) ([]byte, error)
	Has(key []byte) error
	SearchFirst(key []byte) ([]byte, error)
	Remove(key []byte) error
	ClearCache()
	DestroyUnit() error
	GetFromEpoch(key []byte, epoch uint32) ([]byte, error)
	GetBulkFromEpoch(keys [][]byte, epoch uint32) (map[string][]byte, error)
	HasInEpoch(key []byte, epoch uint32) error
	IsInterfaceNil() bool
	Close() error
	RangeKeys(handler func(key []byte, val []byte) bool)
}

// StorerWithPutInEpoch is an extended storer with the ability to set the epoch which will be used for put operations
type StorerWithPutInEpoch interface {
	Storer
	SetEpochForPutOperation(epoch uint32)
}

// EpochStartNotifier defines which actions should be done for handling new epoch's events
type EpochStartNotifier interface {
	RegisterHandler(handler epochStart.ActionHandler)
	IsInterfaceNil() bool
}

// PathManagerHandler defines which actions should be done for generating paths for databases directories
type PathManagerHandler interface {
	PathForEpoch(shardId string, epoch uint32, identifier string) string
	PathForStatic(shardId string, identifier string) string
	IsInterfaceNil() bool
}

// PersisterFactory defines which actions should be done for creating a persister
type PersisterFactory interface {
	Create(path string) (Persister, error)
	IsInterfaceNil() bool
}

// UnitOpenerHandler defines which actions should be done for opening storage units
type UnitOpenerHandler interface {
	GetMostRecentBootstrapStorageUnit() (Storer, error)
	IsInterfaceNil() bool
}

// DirectoryReaderHandler defines which actions should be done by a directory reader
type DirectoryReaderHandler interface {
	ListFilesAsString(directoryPath string) ([]string, error)
	ListDirectoriesAsString(directoryPath string) ([]string, error)
	ListAllAsString(directoryPath string) ([]string, error)
	IsInterfaceNil() bool
}

// LatestStorageDataProviderHandler defines which actions be done by a component who fetches latest data from storage
type LatestStorageDataProviderHandler interface {
	GetParentDirAndLastEpoch() (string, uint32, error)
	Get() (LatestDataFromStorage, error)
	GetShardsFromDirectory(path string) ([]string, error)
	IsInterfaceNil() bool
}

// LatestDataFromStorage represents the DTO structure to return from storage
type LatestDataFromStorage struct {
	Epoch           uint32
	ShardID         uint32
	LastRound       int64
	EpochStartRound uint64
}

// ShardCoordinator defines what a shard state coordinator should hold
type ShardCoordinator interface {
	NumberOfShards() uint32
	ComputeId(address []byte) uint32
	SelfId() uint32
	SameShard(firstAddress, secondAddress []byte) bool
	CommunicationIdentifier(destShardID uint32) string
	IsInterfaceNil() bool
}

// ForEachItem is an iterator callback
type ForEachItem func(key []byte, value interface{})

// LRUCacheHandler is the interface for LRU cache.
type LRUCacheHandler interface {
	Add(key, value interface{}) bool
	Get(key interface{}) (value interface{}, ok bool)
	Contains(key interface{}) (ok bool)
	ContainsOrAdd(key, value interface{}) (ok, evicted bool)
	Peek(key interface{}) (value interface{}, ok bool)
	Remove(key interface{}) bool
	Keys() []interface{}
	Len() int
	Purge()
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

// TimeCacher defines the cache that can keep a record for a bounded time
type TimeCacher interface {
	Upsert(key string, span time.Duration) error
	Has(key string) bool
	Sweep()
	IsInterfaceNil() bool
}
