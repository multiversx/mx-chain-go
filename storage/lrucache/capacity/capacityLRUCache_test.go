package capacity

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewCapacityLRU_WithInvalidSize(t *testing.T) {
	size := 0
	capacity := int64(1)
	cache, err := NewCapacityLRU(size, capacity, nil)
	assert.Nil(t, cache)
	assert.Equal(t, storage.ErrCacheSizeInvalid, err)
}

func TestNewCapacityLRU_WithInvalidCapacity(t *testing.T) {
	size := 1
	capacity := int64(0)
	cache, err := NewCapacityLRU(size, capacity, nil)
	assert.Nil(t, cache)
	assert.Equal(t, storage.ErrCacheCapacityInvalid, err)
}

func TestNewCapacityLRU(t *testing.T) {
	size := 1
	capacity := int64(5)
	onEvict := EvictCallback(func(key interface{}, value interface{}) {
		return
	})
	cache, err := NewCapacityLRU(size, capacity, onEvict)
	assert.NotNil(t, cache)
	assert.Nil(t, err)
	assert.Equal(t, size, cache.size)
	assert.Equal(t, capacity, cache.maxCapacityInBytes)
	assert.Equal(t, int64(0), cache.currentCapacityInBytes)
	assert.NotNil(t, cache.onEvict)
	assert.NotNil(t, cache.evictList)
	assert.NotNil(t, cache.items)
}

func TestCapacityLRUCache_Put(t *testing.T) {
	cache := createDefaultCache()
	data := []byte("test")
	key := "key"
	capacity := int64(5)
	cache.Add(key, data, capacity)

	v, ok := cache.Get(key)
	assert.True(t, ok)
	assert.NotNil(t, v)
	assert.Equal(t, data, v)

	keys := cache.Keys()
	assert.Equal(t, 1, len(keys))
	assert.Equal(t, key, keys[0])
}

func TestCapacityLRUCache_HasOrAdd(t *testing.T) {
	cache := createDefaultCache()
	data := []byte("data1")
	key := "key"

	found, evicted := cache.HasOrAdd(key, data, 1)
	assert.False(t, found)
	assert.False(t, evicted)

	v, ok := cache.Get(key)
	assert.True(t, ok)
	assert.NotNil(t, v)
	assert.Equal(t, data, v)

	data2 := []byte("data2")
	found, evicted = cache.HasOrAdd(key, data2, 1)
	assert.True(t, found)
	assert.False(t, evicted)

	v, ok = cache.Get(key)
	assert.True(t, ok)
	assert.NotNil(t, v)
	assert.Equal(t, data, v)
}

func createDefaultCache() *CapacityLRU {
	cache, _ := NewCapacityLRU(100, 100, nil)
	return cache
}
