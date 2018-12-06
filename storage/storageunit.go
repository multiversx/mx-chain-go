package storage

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/config"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/bloom"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/leveldb"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/lrucache"

	"github.com/pkg/errors"
)

// Persister provides storage of data services in a database like construct
type Persister interface {
	// Put add the value to the (key, val) persistance medium
	Put(key, val []byte) error

	// Get gets the value associated to the key
	Get(key []byte) ([]byte, error)

	// Has returns true if the given key is present in the persistance medium
	Has(key []byte) (bool, error)

	// Init initializes the persistance medium and prepares it for usage
	Init() error

	// Close closes the files/resources associated to the persistance medium
	Close() error

	// Remove removes the data associated to the given key
	Remove(key []byte) error

	// Destroy removes the persistance medium stored data
	Destroy() error
}

// Cacher provides caching services
type Cacher interface {
	// Clear is used to completely clear the cache.
	Clear()

	// Put adds a value to the cache.  Returns true if an eviction occurred.
	Put(key []byte, value interface{}) (evicted bool)

	// Get looks up a key's value from the cache.
	Get(key []byte) (value interface{}, ok bool)

	// Has checks if a key is in the cache, without updating the
	// recent-ness or deleting it for being stale.
	Has(key []byte) bool

	// Peek returns the key value (or undefined if not found) without updating
	// the "recently used"-ness of the key.
	Peek(key []byte) (value interface{}, ok bool)

	// HasOrAdd checks if a key is in the cache  without updating the
	// recent-ness or deleting it for being stale,  and if not, adds the value.
	// Returns whether found and whether an eviction occurred.
	HasOrAdd(key []byte, value interface{}) (ok, evicted bool)

	// Remove removes the provided key from the cache.
	Remove(key []byte)

	// RemoveOldest removes the oldest item from the cache.
	RemoveOldest()

	// Keys returns a slice of the keys in the cache, from oldest to newest.
	Keys() [][]byte

	// Len returns the number of items in the cache.
	Len() int
}

// BloomFilter provides services for filtering database requests
type BloomFilter interface {

	//Add adds the value to the bloom filter
	Add([]byte)

	//Test checks if the value is in in the set. If it returns 'false',
	//the item is definitely not in the DB
	MayContain([]byte) bool

	//Clear sets all the bits from the filter to 0
	Clear()
}

// Storer provides storage services in a two layered storage construct, where the first layer is
// represented by a cache and second layer by a persitent storage (DB-like)
type Storer interface {
	Put(key, data []byte) error
	Get(key []byte) ([]byte, error)
	Has(key []byte) (bool, error)
	HasOrAdd(key []byte, value []byte) (bool, error)
	Remove(key []byte) error
	ClearCache()
	DestroyUnit() error
}

// StorageUnit represents a storer's data bank
// holding the cache, persistance unit and bloom filter
type StorageUnit struct {
	lock        sync.RWMutex
	persister   Persister
	cacher      Cacher
	bloomFilter BloomFilter
}

// Put adds data to both cache and persistance medium, and updates the bloom filter
func (s *StorageUnit) Put(key, data []byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// no need to add if already present in cache
	has := s.cacher.Has(key)

	if has {
		return nil
	}

	s.cacher.Put(key, data)

	err := s.persister.Put(key, data)

	if err != nil {
		s.cacher.Remove(key)
		return err
	}

	if s.bloomFilter != nil {
		s.bloomFilter.Add(key)
	}

	return err

}

// Get searches the key in the cache. In case it is not found, it searches
// for the key in bloom filter first and if found
// it further searches it in the associated database.
// In case it is found in the database, the cache is updated with the value as well.
func (s *StorageUnit) Get(key []byte) ([]byte, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	v, ok := s.cacher.Get(key)
	var err error

	if !ok {
		// not found in cache
		// search it in second persistence medium
		if s.bloomFilter == nil || s.bloomFilter.MayContain(key) == true {
			v, err = s.persister.Get(key)

			if err != nil {
				return nil, err
			}

			// if found in persistance unit, add it in cache
			s.cacher.Put(key, v)
		} else {
			return nil, errors.New(fmt.Sprintf("key: %s not found", string(key)))
		}
	}

	return v.([]byte), nil
}

// Has checks if the key is in the storageUnit.
// It first checks the cache. If it is not found, it checks the bloom filter
// and if present it checks the db
func (s *StorageUnit) Has(key []byte) (bool, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	has := s.cacher.Has(key)

	if has {
		return has, nil
	}

	if s.bloomFilter == nil || s.bloomFilter.MayContain(key) == true {
		return s.persister.Has(key)
	}
	return false, nil
}

// HasOrAdd checks if the key is present in the storage and if not adds it.
// it updates the cache either way
// it returns if the value was originally found
func (s *StorageUnit) HasOrAdd(key []byte, value []byte) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	has := s.cacher.Has(key)

	if has {
		return has, nil
	}

	if s.bloomFilter == nil || s.bloomFilter.MayContain(key) == true {
		has, err := s.persister.Has(key)
		if err != nil {
			return has, err
		}

		//add it to the cache
		s.cacher.Put(key, value)

		if !has {
			// add it also to the persistance unit
			err = s.persister.Put(key, value)
			if err != nil {
				//revert adding to the cache
				s.cacher.Remove(key)
			}
		}
		return has, err
	}

	s.cacher.Put(key, value)
	err := s.persister.Put(key, value)

	if err != nil {
		s.cacher.Remove(key)
		return false, err
	}
	s.bloomFilter.Add(key)

	return false, err

}

// Remove removes the data associated to the given key from both cache and persistance medium
func (s *StorageUnit) Remove(key []byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.cacher.Remove(key)
	err := s.persister.Remove(key)

	return err
}

// ClearCache cleans up the entire cache
func (s *StorageUnit) ClearCache() {
	s.cacher.Clear()
}

// DestroyUnit cleans up the bloom filter, the cache, and the db
func (s *StorageUnit) DestroyUnit() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.bloomFilter != nil {
		s.bloomFilter.Clear()
	}

	s.cacher.Clear()
	s.persister.Close()
	return s.persister.Destroy()
}

// NewStorageUnit is the constructor for the storage unit, creating a new storage unit
// from the given bloom filter, cacher and persister.
func NewStorageUnit(c Cacher, p Persister, b BloomFilter) (*StorageUnit, error) {
	if p == nil {
		return nil, errors.New("expected not nil persister")
	}

	if c == nil {
		return nil, errors.New("expected not nil cacher")
	}

	sUnit := &StorageUnit{
		persister:   p,
		cacher:      c,
		bloomFilter: b,
	}

	sUnit.persister.Init()

	return sUnit, nil
}

// NewStorageUnitFromConf creates a new storage unit from a storage unit config
func NewStorageUnitFromConf(conf *config.StorageUnitConfig) (*StorageUnit, error) {
	var cache Cacher
	var db Persister
	var bf BloomFilter
	var err error

	cache, err = CreateCacheFromConf(conf.CacheConf)

	if err != nil {
		return nil, err
	}

	db, err = CreateDBFromConf(conf.DBConf)

	if err != nil {
		return nil, err
	}

	bf, err = CreateBloomFilterFromConf(conf.BloomFilterConf)

	if err != nil {
		return nil, err
	}

	st, err := NewStorageUnit(cache, db, bf)

	if err != nil {
		return nil, err
	}

	return st, err
}

//CreateCacheFromConf creates a new cache from a cache config
func CreateCacheFromConf(conf *config.CacheConfig) (Cacher, error) {
	var cacher Cacher
	var err error

	switch conf.Type {
	case config.LRUCache:
		cacher, err = lrucache.NewCache(int(conf.Size))
		// add other implementations if required
	default:
		return nil, errors.New("not supported cache type")
	}

	if err != nil {
		return nil, err
	}
	return cacher, nil
}

// CreateDBFromConf creates a new database from database config
func CreateDBFromConf(conf *config.DBConfig) (Persister, error) {
	var db Persister
	var err error

	switch conf.Type {
	case config.LvlDB:
		db, err = leveldb.NewDB(conf.FileName)
	default:
		return nil, errors.New("nit supported db type")
	}

	if err != nil {
		return nil, err
	}

	return db, nil
}

// CreateBloomFilterFromConf creates a new bloom filter from bloom filter config
func CreateBloomFilterFromConf(conf *config.BloomFilterConfig) (BloomFilter, error) {
	var bf BloomFilter
	var err error

	bf, err = bloom.NewFilter(conf.Size, conf.HashFunc)

	if err != nil {
		return nil, err
	}

	return bf, nil
}
