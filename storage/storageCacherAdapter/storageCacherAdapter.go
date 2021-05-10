package storageCacherAdapter

import (
	"math"
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("storageCacherAdapter")

type storageCacherAdapter struct {
	cacher storage.AdaptedSizedLRUCache
	db     storage.Persister
	lock   sync.RWMutex

	storedDataFactory storage.StoredDataFactory
	marshalizer       marshal.Marshalizer
}

// NewStorageCacherAdapter creates a new storageCacherAdapter
func NewStorageCacherAdapter(
	cacher storage.AdaptedSizedLRUCache,
	db storage.Persister,
	storedDataFactory storage.StoredDataFactory,
	marshalizer marshal.Marshalizer,
) (*storageCacherAdapter, error) {
	if check.IfNil(cacher) {
		return nil, storage.ErrNilCacher
	}
	if check.IfNil(db) {
		return nil, storage.ErrNilPersister
	}
	if check.IfNil(marshalizer) {
		return nil, storage.ErrNilMarshalizer
	}
	if check.IfNil(storedDataFactory) {
		return nil, storage.ErrNilStoredDataFactory
	}

	return &storageCacherAdapter{
		cacher:            cacher,
		db:                db,
		lock:              sync.RWMutex{},
		storedDataFactory: storedDataFactory,
		marshalizer:       marshalizer,
	}, nil
}

// Clear clears the underlying cacher and db
func (c *storageCacherAdapter) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()

	// TODO also clear the underlying db
	c.cacher.Purge()
}

// Put adds the given value in the cacher. If the cacher is full, the evicted values will be persisted to the db
func (c *storageCacherAdapter) Put(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	evictedValues := c.cacher.AddSizedAndReturnEvicted(string(key), value, int64(sizeInBytes))
	for evictedKey, evictedVal := range evictedValues {
		evictedKeyStr, ok := evictedKey.(string)
		if !ok {
			log.Warn("invalid key type")
			continue
		}

		evictedValBytes, err := c.marshalizer.Marshal(evictedVal)
		if err != nil {
			log.Error("could not marshall value", "error", err)
			continue
		}

		err = c.db.Put([]byte(evictedKeyStr), evictedValBytes)
		if err != nil {
			log.Error("could not save to db", "error", err)
			continue
		}
	}

	return len(evictedValues) != 0
}

// Get returns the value at the given key
func (c *storageCacherAdapter) Get(key []byte) (value interface{}, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	val, ok := c.cacher.Get(string(key))
	if ok {
		return val, true
	}

	valBytes, err := c.db.Get(key)
	if err != nil {
		return nil, false
	}

	storedData := c.storedDataFactory.CreateEmpty()
	err = c.marshalizer.Unmarshal(storedData, valBytes)
	if err != nil {
		log.Error("could not unmarshall", "error", err)
		return nil, false
	}

	return storedData, true
}

// Has checks if the given key is present in the storageUnit
func (c *storageCacherAdapter) Has(key []byte) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	isPresent := c.cacher.Contains(string(key))
	if isPresent {
		return true
	}

	err := c.db.Has(key)
	if err == nil {
		return true
	}

	return false
}

// Peek returns the value at the given key
func (c *storageCacherAdapter) Peek(key []byte) (value interface{}, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.cacher.Peek(string(key))
}

// HasOrAdd adds the given value in the storageUnit
func (c *storageCacherAdapter) HasOrAdd(key []byte, value interface{}, sizeInBytes int) (has, added bool) {
	_ = c.Put(key, value, sizeInBytes)

	return false, true
}

// Remove deletes the given key from the storageUnit
func (c *storageCacherAdapter) Remove(key []byte) {
	c.lock.Lock()
	defer c.lock.Unlock()

	removed := c.cacher.Remove(string(key))
	if removed {
		return
	}

	_ = c.db.Remove(key)
}

// Keys returns all the keys present in the storageUnit
func (c *storageCacherAdapter) Keys() [][]byte {
	c.lock.RLock()
	defer c.lock.RUnlock()

	cacherKeys := c.cacher.Keys()
	storedKeys := make([][]byte, 0)
	for i := range cacherKeys {
		key, ok := cacherKeys[i].(string)
		if !ok {
			continue
		}

		storedKeys = append(storedKeys, []byte(key))
	}

	getKeys := func(key []byte, _ []byte) bool {
		storedKeys = append(storedKeys, key)
		return true
	}

	c.db.RangeKeys(getKeys)
	return storedKeys
}

// Len returns the number of elements from the storageUnit
func (c *storageCacherAdapter) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()

	cacheLen := c.cacher.Len()
	countValues := func(_ []byte, _ []byte) bool {
		cacheLen++
		return true
	}

	c.db.RangeKeys(countValues)
	return cacheLen
}

// SizeInBytesContained returns the number of bytes stored in the storageUnit
func (c *storageCacherAdapter) SizeInBytesContained() uint64 {
	c.lock.RLock()
	defer c.lock.RUnlock()

	size := c.cacher.SizeInBytesContained()
	countSize := func(_ []byte, val []byte) bool {
		size += uint64(len(val))
		return true
	}

	c.db.RangeKeys(countSize)
	return size
}

// MaxSize returns MaxInt64
func (c *storageCacherAdapter) MaxSize() int {
	return math.MaxInt64
}

// RegisterHandler does nothing
func (c *storageCacherAdapter) RegisterHandler(_ func(key []byte, _ interface{}), _ string) {
	return
}

// UnRegisterHandler does nothing
func (c *storageCacherAdapter) UnRegisterHandler(_ string) {
	return
}

// Close closes the underlying db
func (c *storageCacherAdapter) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.db.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *storageCacherAdapter) IsInterfaceNil() bool {
	return c == nil
}
