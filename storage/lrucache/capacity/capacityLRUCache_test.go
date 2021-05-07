package capacity

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
)

func createDefaultCache() *capacityLRU {
	cache, _ := NewCapacityLRU(100, 100)
	return cache
}

//------- NewCapacityLRU

func TestNewCapacityLRU_WithInvalidSize(t *testing.T) {
	t.Parallel()

	size := 0
	capacity := int64(1)
	cache, err := NewCapacityLRU(size, capacity)
	assert.Nil(t, cache)
	assert.Equal(t, storage.ErrCacheSizeInvalid, err)
}

func TestNewCapacityLRU_WithInvalidCapacity(t *testing.T) {
	t.Parallel()

	size := 1
	capacity := int64(0)
	cache, err := NewCapacityLRU(size, capacity)
	assert.Nil(t, cache)
	assert.Equal(t, storage.ErrCacheCapacityInvalid, err)
}

func TestNewCapacityLRU(t *testing.T) {
	t.Parallel()

	size := 1
	capacity := int64(5)

	cache, err := NewCapacityLRU(size, capacity)
	assert.NotNil(t, cache)
	assert.Nil(t, err)
	assert.Equal(t, size, cache.size)
	assert.Equal(t, capacity, cache.maxCapacityInBytes)
	assert.Equal(t, int64(0), cache.currentCapacityInBytes)
	assert.NotNil(t, cache.evictList)
	assert.NotNil(t, cache.items)
}

//------- AddSized

func TestCapacityLRUCache_AddSizedNegativeSizeInBytesShouldReturn(t *testing.T) {
	t.Parallel()

	c := createDefaultCache()
	data := []byte("test")
	key := "key"
	c.AddSized(key, data, -1)

	assert.Equal(t, 0, c.Len())
}

func TestCapacityLRUCache_AddSizedSimpleTestShouldWork(t *testing.T) {
	t.Parallel()

	c := createDefaultCache()
	data := []byte("test")
	key := "key"
	capacity := int64(5)
	c.AddSized(key, data, capacity)

	v, ok := c.Get(key)
	assert.True(t, ok)
	assert.NotNil(t, v)
	assert.Equal(t, data, v)

	keys := c.Keys()
	assert.Equal(t, 1, len(keys))
	assert.Equal(t, key, keys[0])
}

func TestCapacityLRUCache_AddSizedEvictionByCacheSizeShouldWork(t *testing.T) {
	t.Parallel()

	c, _ := NewCapacityLRU(3, 100000)

	keys := []string{"key1", "key2", "key3", "key4", "key5"}

	c.AddSized(keys[0], struct{}{}, 0)
	assert.Equal(t, 1, c.Len())

	c.AddSized(keys[1], struct{}{}, 0)
	assert.Equal(t, 2, c.Len())

	c.AddSized(keys[2], struct{}{}, 0)
	assert.Equal(t, 3, c.Len())

	c.AddSized(keys[3], struct{}{}, 0)
	assert.Equal(t, 3, c.Len())
	assert.False(t, c.Contains(keys[0]))
	assert.True(t, c.Contains(keys[3]))

	c.AddSized(keys[4], struct{}{}, 0)
	assert.Equal(t, 3, c.Len())
	assert.False(t, c.Contains(keys[1]))
	assert.True(t, c.Contains(keys[4]))
}

func TestCapacityLRUCache_AddSizedEvictionBySizeInBytesShouldWork(t *testing.T) {
	t.Parallel()

	c, _ := NewCapacityLRU(100000, 1000)

	keys := []string{"key1", "key2", "key3", "key4"}

	c.AddSized(keys[0], struct{}{}, 500)
	assert.Equal(t, 1, c.Len())

	c.AddSized(keys[1], struct{}{}, 500)
	assert.Equal(t, 2, c.Len())

	c.AddSized(keys[2], struct{}{}, 500)
	assert.Equal(t, 2, c.Len())
	assert.False(t, c.Contains(keys[0]))
	assert.True(t, c.Contains(keys[2]))

	c.AddSized(keys[3], struct{}{}, 500)
	assert.Equal(t, 2, c.Len())
	assert.False(t, c.Contains(keys[1]))
	assert.True(t, c.Contains(keys[3]))
}

func TestCapacityLRUCache_AddSizedEvictionBySizeInBytesOneLargeElementShouldWork(t *testing.T) {
	t.Parallel()

	c, _ := NewCapacityLRU(100000, 1000)

	keys := []string{"key1", "key2", "key3", "key4"}

	c.AddSized(keys[0], struct{}{}, 500)
	assert.Equal(t, 1, c.Len())

	c.AddSized(keys[1], struct{}{}, 500)
	assert.Equal(t, 2, c.Len())

	c.AddSized(keys[2], struct{}{}, 500)
	assert.Equal(t, 2, c.Len())
	assert.False(t, c.Contains(keys[0]))
	assert.True(t, c.Contains(keys[2]))

	c.AddSized(keys[3], struct{}{}, 500000)
	assert.Equal(t, 1, c.Len())
	assert.False(t, c.Contains(keys[0]))
	assert.False(t, c.Contains(keys[1]))
	assert.False(t, c.Contains(keys[2]))
	assert.True(t, c.Contains(keys[3]))
}

func TestCapacityLRUCache_AddSizedEvictionBySizeInBytesOneLargeElementEvictedBySmallElementsShouldWork(t *testing.T) {
	t.Parallel()

	c, _ := NewCapacityLRU(100000, 1000)

	keys := []string{"key1", "key2", "key3"}

	c.AddSized(keys[0], struct{}{}, 500000)
	assert.Equal(t, 1, c.Len())

	c.AddSized(keys[1], struct{}{}, 500)
	assert.Equal(t, 1, c.Len())

	c.AddSized(keys[2], struct{}{}, 500)
	assert.Equal(t, 2, c.Len())
	assert.False(t, c.Contains(keys[0]))
	assert.True(t, c.Contains(keys[1]))
	assert.True(t, c.Contains(keys[2]))
}

func TestCapacityLRUCache_AddSizedEvictionBySizeInBytesExistingOneLargeElementShouldWork(t *testing.T) {
	t.Parallel()

	c, _ := NewCapacityLRU(100000, 1000)

	keys := []string{"key1", "key2"}

	c.AddSized(keys[0], struct{}{}, 500)
	assert.Equal(t, 1, c.Len())

	c.AddSized(keys[1], struct{}{}, 500)
	assert.Equal(t, 2, c.Len())

	c.AddSized(keys[0], struct{}{}, 500000)
	assert.Equal(t, 1, c.Len())
	assert.True(t, c.Contains(keys[0]))
	assert.False(t, c.Contains(keys[1]))
}

//------- AddSizedIfMissing

func TestCapacityLRUCache_AddSizedIfMissing(t *testing.T) {
	t.Parallel()

	c := createDefaultCache()
	data := []byte("data1")
	key := "key"

	found, evicted := c.AddSizedIfMissing(key, data, 1)
	assert.False(t, found)
	assert.False(t, evicted)

	v, ok := c.Get(key)
	assert.True(t, ok)
	assert.NotNil(t, v)
	assert.Equal(t, data, v)

	data2 := []byte("data2")
	found, evicted = c.AddSizedIfMissing(key, data2, 1)
	assert.True(t, found)
	assert.False(t, evicted)

	v, ok = c.Get(key)
	assert.True(t, ok)
	assert.NotNil(t, v)
	assert.Equal(t, data, v)
}

func TestCapacityLRUCache_AddSizedIfMissingNegativeSizeInBytesShouldReturnFalse(t *testing.T) {
	t.Parallel()

	c := createDefaultCache()
	data := []byte("data1")
	key := "key"

	has, evicted := c.AddSizedIfMissing(key, data, -1)
	assert.False(t, has)
	assert.False(t, evicted)
	assert.Equal(t, 0, c.Len())
}

//------- Get

func TestCapacityLRUCache_GetShouldWork(t *testing.T) {
	t.Parallel()

	key := "key"
	value := &struct{ A int }{A: 10}

	c := createDefaultCache()
	c.AddSized(key, value, 0)

	recovered, exists := c.Get(key)
	assert.True(t, value == recovered) //pointer testing
	assert.True(t, exists)

	recovered, exists = c.Get("key not found")
	assert.Nil(t, recovered)
	assert.False(t, exists)
}

//------- Purge

func TestCapacityLRUCache_PurgeShouldWork(t *testing.T) {
	t.Parallel()

	c, _ := NewCapacityLRU(100000, 1000)

	keys := []string{"key1", "key2"}
	c.AddSized(keys[0], struct{}{}, 500)
	c.AddSized(keys[1], struct{}{}, 500)

	c.Purge()

	assert.Equal(t, 0, c.Len())
	assert.Equal(t, int64(0), c.currentCapacityInBytes)
}

//------- Peek

func TestCapacityLRUCache_PeekNotFoundShouldWork(t *testing.T) {
	t.Parallel()

	c, _ := NewCapacityLRU(100000, 1000)
	val, found := c.Peek("key not found")

	assert.Nil(t, val)
	assert.False(t, found)
}

func TestCapacityLRUCache_PeekShouldWork(t *testing.T) {
	t.Parallel()

	c, _ := NewCapacityLRU(100000, 1000)
	key1 := "key1"
	key2 := "key2"
	val1 := &struct{}{}

	c.AddSized(key1, val1, 0)
	c.AddSized(key2, struct{}{}, 0)

	//at this point key2 is more "recent" than key1
	assert.True(t, c.evictList.Front().Value.(*entry).key == key2)

	val, found := c.Peek(key1)
	assert.True(t, val == val1) //pointer testing
	assert.True(t, found)

	//recentness should not have been altered
	assert.True(t, c.evictList.Front().Value.(*entry).key == key2)
}

//------- Remove

func TestCapacityLRUCache_RemoveNotFoundShouldWork(t *testing.T) {
	t.Parallel()

	c, _ := NewCapacityLRU(100000, 1000)
	removed := c.Remove("key not found")

	assert.False(t, removed)
}

func TestCapacityLRUCache_RemovedShouldWork(t *testing.T) {
	t.Parallel()

	c, _ := NewCapacityLRU(100000, 1000)
	key1 := "key1"
	key2 := "key2"

	c.AddSized(key1, struct{}{}, 0)
	c.AddSized(key2, struct{}{}, 0)

	assert.Equal(t, 2, c.Len())

	c.Remove(key1)

	assert.Equal(t, 1, c.Len())
	assert.True(t, c.Contains(key2))
}

// ---------- AddSizedAndReturnEvicted

func TestCapacityLRUCache_AddSizedAndReturnEvictedNegativeSizeInBytesShouldReturn(t *testing.T) {
	t.Parallel()

	c := createDefaultCache()
	data := []byte("test")
	key := "key"
	c.AddSizedAndReturnEvicted(key, data, -1)

	assert.Equal(t, 0, c.Len())
}

func TestCapacityLRUCache_AddSizedAndReturnEvictedSimpleTestShouldWork(t *testing.T) {
	t.Parallel()

	c := createDefaultCache()
	data := []byte("test")
	key := "key"
	capacity := int64(5)
	c.AddSizedAndReturnEvicted(key, data, capacity)

	v, ok := c.Get(key)
	assert.True(t, ok)
	assert.NotNil(t, v)
	assert.Equal(t, data, v)

	keys := c.Keys()
	assert.Equal(t, 1, len(keys))
	assert.Equal(t, key, keys[0])
}

func TestCapacityLRUCache_AddSizedAndReturnEvictedEvictionByCacheSizeShouldWork(t *testing.T) {
	t.Parallel()

	c, _ := NewCapacityLRU(3, 100000)

	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	values := []string{"val1", "val2", "val3", "val4", "val5"}

	evicted := c.AddSizedAndReturnEvicted(keys[0], values[0], int64(len(values[0])))
	assert.Equal(t, 0, len(evicted))
	assert.Equal(t, 1, c.Len())

	evicted = c.AddSizedAndReturnEvicted(keys[1], values[1], int64(len(values[1])))
	assert.Equal(t, 0, len(evicted))
	assert.Equal(t, 2, c.Len())

	evicted = c.AddSizedAndReturnEvicted(keys[2], values[2], int64(len(values[2])))
	assert.Equal(t, 0, len(evicted))
	assert.Equal(t, 3, c.Len())

	evicted = c.AddSizedAndReturnEvicted(keys[3], values[3], int64(len(values[3])))
	assert.Equal(t, 3, c.Len())
	assert.False(t, c.Contains(keys[0]))
	assert.True(t, c.Contains(keys[3]))
	assert.Equal(t, 1, len(evicted))
	assert.Equal(t, values[0], evicted[keys[0]])

	evicted = c.AddSizedAndReturnEvicted(keys[4], values[4], int64(len(values[4])))
	assert.Equal(t, 3, c.Len())
	assert.False(t, c.Contains(keys[1]))
	assert.True(t, c.Contains(keys[4]))
	assert.Equal(t, 1, len(evicted))
	assert.Equal(t, values[1], evicted[keys[1]])
}

func TestCapacityLRUCache_AddSizedAndReturnEvictedEvictionBySizeInBytesShouldWork(t *testing.T) {
	t.Parallel()

	c, _ := NewCapacityLRU(100000, 1000)

	keys := []string{"key1", "key2", "key3", "key4"}
	values := []string{"val1", "val2", "val3", "val4"}

	evicted := c.AddSizedAndReturnEvicted(keys[0], values[0], 500)
	assert.Equal(t, 0, len(evicted))
	assert.Equal(t, 1, c.Len())

	evicted = c.AddSizedAndReturnEvicted(keys[1], values[1], 500)
	assert.Equal(t, 0, len(evicted))
	assert.Equal(t, 2, c.Len())

	evicted = c.AddSizedAndReturnEvicted(keys[2], values[2], 500)
	assert.Equal(t, 2, c.Len())
	assert.False(t, c.Contains(keys[0]))
	assert.True(t, c.Contains(keys[2]))
	assert.Equal(t, 1, len(evicted))
	assert.Equal(t, values[0], evicted[keys[0]])

	evicted = c.AddSizedAndReturnEvicted(keys[3], values[3], 500)
	assert.Equal(t, 2, c.Len())
	assert.False(t, c.Contains(keys[1]))
	assert.True(t, c.Contains(keys[3]))
	assert.Equal(t, 1, len(evicted))
	assert.Equal(t, values[1], evicted[keys[1]])
}

func TestCapacityLRUCache_AddSizedAndReturnEvictedEvictionBySizeInBytesOneLargeElementShouldWork(t *testing.T) {
	t.Parallel()

	c, _ := NewCapacityLRU(100000, 1000)

	keys := []string{"key1", "key2", "key3", "key4"}
	values := []string{"val1", "val2", "val3", "val4"}

	evicted := c.AddSizedAndReturnEvicted(keys[0], values[0], 500)
	assert.Equal(t, 0, len(evicted))
	assert.Equal(t, 1, c.Len())

	evicted = c.AddSizedAndReturnEvicted(keys[1], values[1], 500)
	assert.Equal(t, 0, len(evicted))
	assert.Equal(t, 2, c.Len())

	evicted = c.AddSizedAndReturnEvicted(keys[2], values[2], 500)
	assert.Equal(t, 2, c.Len())
	assert.False(t, c.Contains(keys[0]))
	assert.True(t, c.Contains(keys[2]))
	assert.Equal(t, 1, len(evicted))
	assert.Equal(t, values[0], evicted[keys[0]])

	evicted = c.AddSizedAndReturnEvicted(keys[3], values[3], 500000)
	assert.Equal(t, 1, c.Len())
	assert.False(t, c.Contains(keys[0]))
	assert.False(t, c.Contains(keys[1]))
	assert.False(t, c.Contains(keys[2]))
	assert.True(t, c.Contains(keys[3]))
	assert.Equal(t, 2, len(evicted))
	assert.Equal(t, values[1], evicted[keys[1]])
	assert.Equal(t, values[2], evicted[keys[2]])
}

func TestCapacityLRUCache_AddSizedAndReturnEvictedEvictionBySizeInBytesOneLargeElementEvictedBySmallElementsShouldWork(t *testing.T) {
	t.Parallel()

	c, _ := NewCapacityLRU(100000, 1000)

	keys := []string{"key1", "key2", "key3"}
	values := []string{"val1", "val2", "val3"}

	evicted := c.AddSizedAndReturnEvicted(keys[0], values[0], 500000)
	assert.Equal(t, 0, len(evicted))
	assert.Equal(t, 1, c.Len())

	evicted = c.AddSizedAndReturnEvicted(keys[1], values[1], 500)
	assert.Equal(t, 1, c.Len())
	assert.Equal(t, 1, len(evicted))
	assert.Equal(t, values[0], evicted[keys[0]])

	evicted = c.AddSizedAndReturnEvicted(keys[2], values[2], 500)
	assert.Equal(t, 0, len(evicted))
	assert.Equal(t, 2, c.Len())
	assert.False(t, c.Contains(keys[0]))
	assert.True(t, c.Contains(keys[1]))
	assert.True(t, c.Contains(keys[2]))
}
