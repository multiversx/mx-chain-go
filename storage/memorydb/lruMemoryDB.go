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

func NewlruDB(size uint32) (*lruDB, error) {
	cacher, err := lrucache.NewCache(int(size))
	if err != nil {
		return nil, err
	}

	return &lruDB{cacher: cacher}, nil
}

func (l *lruDB) Put(key, val []byte) error {
	_ = l.cacher.Put(key, val)
	return nil
}

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

func (l *lruDB) Has(key []byte) error {
	has := l.cacher.Has(key)
	if has {
		return nil
	}
	return storage.ErrKeyNotFound
}

func (l *lruDB) Init() error {
	l.cacher.Clear()
	return nil
}

func (l *lruDB) Close() error {
	l.cacher.Clear()
	return nil
}

func (l *lruDB) Remove(key []byte) error {
	l.cacher.Remove(key)
	return nil
}

func (l *lruDB) Destroy() error {
	l.cacher.Clear()
	return nil
}
