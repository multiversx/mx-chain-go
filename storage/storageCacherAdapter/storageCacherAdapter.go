package storageCacherAdapter

import (
	"math"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("storageCacherAdapter")

type storageCacherAdapter struct {
	storageUnit storage.Storer
}

// NewStorageCacherAdapter creates a new storageCacherAdapter
func NewStorageCacherAdapter(storer storage.Storer) *storageCacherAdapter {
	return &storageCacherAdapter{storer}
}

// Clear clears the underlying cacher
func (c *storageCacherAdapter) Clear() {
	c.storageUnit.ClearCache()
}

// Put adds the given value in the storageUnit
func (c *storageCacherAdapter) Put(key []byte, value interface{}, _ int) (evicted bool) {
	valueBytes, ok := value.([]byte)
	if !ok {
		log.Error("put in storageUnit error, byte array expected")
		return false
	}

	err := c.storageUnit.Put(key, valueBytes)
	if err != nil {
		log.Error("put in storageUnit error", "error", err.Error())
		return false
	}

	return false
}

// Get returns the value at the given key
func (c *storageCacherAdapter) Get(key []byte) (value interface{}, ok bool) {
	value, err := c.storageUnit.Get(key)
	if err != nil {
		return nil, false
	}

	return value, false
}

// Has checks if the given key is present in the storageUnit
func (c *storageCacherAdapter) Has(key []byte) bool {
	err := c.storageUnit.Has(key)

	return err == nil
}

// Peek returns the value at the given key
func (c *storageCacherAdapter) Peek(key []byte) (value interface{}, ok bool) {
	return c.Get(key)
}

// HasOrAdd adds the given value in the storageUnit
func (c *storageCacherAdapter) HasOrAdd(key []byte, value interface{}, sizeInBytes int) (has, added bool) {
	_ = c.Put(key, value, sizeInBytes)

	return false, true
}

// Remove deletes the given key from the storageUnit
func (c *storageCacherAdapter) Remove(key []byte) {
	err := c.storageUnit.Remove(key)
	if err != nil {
		log.Error("remove from storageUnit error", "error", err.Error())
	}
}

// Keys returns all the keys present in the storageUnit
func (c *storageCacherAdapter) Keys() [][]byte {
	keys := make([][]byte, 0)
	getKeys := func(key []byte, val []byte) bool {
		keys = append(keys, key)
		return true
	}

	c.storageUnit.RangeKeys(getKeys)
	return keys
}

// Len returns the number of elements from the storageUnit
func (c *storageCacherAdapter) Len() int {
	cacheLen := 0
	countValues := func(key []byte, val []byte) bool {
		cacheLen++
		return true
	}

	c.storageUnit.RangeKeys(countValues)
	return cacheLen
}

// SizeInBytesContained returns the number of bytes stored in the storageUnit
func (c *storageCacherAdapter) SizeInBytesContained() uint64 {
	size := 0
	countSize := func(key []byte, val []byte) bool {
		size += len(val)
		return true
	}

	c.storageUnit.RangeKeys(countSize)
	return uint64(size)
}

// MaxSize returns MaxInt32
func (c *storageCacherAdapter) MaxSize() int {
	return math.MaxInt32
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
	return c.storageUnit.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *storageCacherAdapter) IsInterfaceNil() bool {
	return c == nil
}
