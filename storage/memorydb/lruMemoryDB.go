package memorydb

import (
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
)

// lruDB represents the memory database storage. It holds a LRU of key value pairs
// and a mutex to handle concurrent accesses to the map
type lruDB struct {
	cacher storage.Cacher
}

// NewlruDB creates a lruDB according to size
func NewlruDB(size uint32) (*lruDB, error) {
	cacher, err := lrucache.NewCache(int(size))
	if err != nil {
		return nil, err
	}

	return &lruDB{cacher: cacher}, nil
}

// Put adds the value to the (key, val) storage medium
func (l *lruDB) Put(key, val []byte) error {
	_ = l.cacher.Put(key, val)
	return nil
}

// Get gets the value associated to the key, or reports an error
func (l *lruDB) Get(key []byte) ([]byte, error) {
	val, ok := l.cacher.Get(key)
	if !ok {
		return nil, storage.ErrKeyNotFound
	}

	mrsVal, ok := val.([]byte)
	if !ok {
		return nil, storage.ErrKeyNotFound
	}
	return mrsVal, nil
}

// Has returns true if the given key is present in the persistence medium, false otherwise
func (l *lruDB) Has(key []byte) error {
	has := l.cacher.Has(key)
	if has {
		return nil
	}
	return storage.ErrKeyNotFound
}

// Init initializes the storage medium and prepares it for usage
func (l *lruDB) Init() error {
	l.cacher.Clear()
	return nil
}

// Close closes the files/resources associated to the storage medium
func (l *lruDB) Close() error {
	l.cacher.Clear()
	return nil
}

// Remove removes the data associated to the given key
func (l *lruDB) Remove(key []byte) error {
	l.cacher.Remove(key)
	return nil
}

// Destroy removes the storage medium stored data
func (l *lruDB) Destroy() error {
	l.cacher.Clear()
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (l *lruDB) IsInterfaceNil() bool {
	if l == nil {
		return true
	}
	return false
}
