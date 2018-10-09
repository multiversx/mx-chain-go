package storage

import (
	"ElrondNetwork/elrond-go-sandbox/config"
	"ElrondNetwork/elrond-go-sandbox/storage/leveldb"
	"ElrondNetwork/elrond-go-sandbox/storage/lrucache"
	"sync"

	"github.com/pkg/errors"
)

type Persister interface {
	// Add the value to the (key, val) persistance medium
	Put(key, val []byte) error

	// gets the value associated to the key
	Get(key []byte) ([]byte, error)

	// returns true if the given key is present in the persistance medium
	Has(key []byte) (bool, error)

	// initialized the persistance medium and prepares it for usage
	Init() error

	// Closes the files/resources associated to the persistance medium
	Close() error

	// Removes the data associated to the given key
	Remove(key []byte) error

	// Removes the persistance medium stored data
	Destroy() error
}

type Cacher interface {
	// Clear is used to completely clear the cache.
	Clear()

	// Add adds a value to the cache.  Returns true if an eviction occurred.
	Put(key, value []byte) (evicted bool)

	// Get looks up a key's value from the cache.
	Get(key []byte) (value []byte, ok bool)

	// Has checks if a key is in the cache, without updating the
	// recent-ness or deleting it for being stale.
	Has(key []byte) bool

	// Peek returns the key value (or undefined if not found) without updating
	// the "recently used"-ness of the key.
	Peek(key []byte) (value []byte, ok bool)

	// HasOrAdd checks if a key is in the cache  without updating the
	// recent-ness or deleting it for being stale,  and if not, adds the value.
	// Returns whether found and whether an eviction occurred.
	HasOrAdd(key, value []byte) (ok, evicted bool)

	// Remove removes the provided key from the cache.
	Remove(key []byte)

	// RemoveOldest removes the oldest item from the cache.
	RemoveOldest()

	// Keys returns a slice of the keys in the cache, from oldest to newest.
	Keys() [][]byte

	// Len returns the number of items in the cache.
	Len() int
}

type Storer interface {
	Put(key, data []byte) error
	Get(key []byte) ([]byte, error)
	Has(key []byte) (bool, error)
	HasOrAdd(key []byte, value []byte) (bool, error)
	Remove(key []byte) error
	ClearCache()
	DestroyUnit() error
}

type StorageUnit struct {
	lock      sync.RWMutex
	persister Persister
	cacher    Cacher
}

// add data to both cache and persistance medium
func (s *StorageUnit) Put(key, data []byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.cacher.Put(key, data)
	err := s.persister.Put(key, data)

	return err
}

// search the data associated to the key in
func (s *StorageUnit) Get(key []byte) ([]byte, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	v, ok := s.cacher.Get(key)
	var err error

	if ok == false {
		// not found in cache
		// search it in second persistance medium
		v, err = s.persister.Get(key)

		if err != nil {
			return nil, err
		}

		// if found in persistance unit, add it in cache
		s.cacher.Put(key, v)
	}

	return v, nil
}

// check if the key is in the storageUnit.
// it first checks the cache and if not present it checks the db
func (s *StorageUnit) Has(key []byte) (bool, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	has := s.cacher.Has(key)

	if has {
		return has, nil
	}

	return s.persister.Has(key)
}

// checks if the key is present in the storage and if not adds it.
// it updates the cache either way
// it returns if the value was originally found
func (s *StorageUnit) HasOrAdd(key []byte, value []byte) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	has := s.cacher.Has(key)

	if has {
		return has, nil
	}

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

// deletes the data associated to the given key from both cache and persistance medium
func (s *StorageUnit) Remove(key []byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.cacher.Remove(key)
	err := s.persister.Remove(key)

	return err
}

// clean up the entire cache
func (s *StorageUnit) ClearCache() {
	s.cacher.Clear()
}

//clean up both the cache and db
func (s *StorageUnit) DestroyUnit() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.cacher.Clear()
	s.persister.Close()
	return s.persister.Destroy()
}

func NewStorageUnit(c Cacher, p Persister) (*StorageUnit, error) {
	if p == nil {
		return nil, errors.New("expected not nil persister")
	}

	if c == nil {
		return nil, errors.New("expected not nil cacher")
	}

	sUnit := &StorageUnit{
		persister: p,
		cacher:    c,
	}

	sUnit.persister.Init()

	return sUnit, nil
}

func NewStorageUnitFromConf(conf *config.StorageUnitConfig) (*StorageUnit, error) {
	var cache Cacher
	var db Persister
	var err error

	cache, err = CreateCacheFromConf(conf.CacheConf)

	if err != nil {
		return nil, err
	}

	db, err = CreateDBFromConf(conf.DBConf)

	if err != nil {
		return nil, err
	}

	st, err := NewStorageUnit(cache, db)

	if err != nil {
		return nil, err
	}

	return st, err
}

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
