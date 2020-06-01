package immunitycache

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewImmunityCache(t *testing.T) {
	config := CacheConfig{
		Name:                        "test",
		NumChunks:                   16,
		MaxNumItems:                 math.MaxUint32,
		MaxNumBytes:                 maxNumBytesUpperBound,
		NumItemsToPreemptivelyEvict: 100,
	}

	cache, err := NewImmunityCache(config)
	require.Nil(t, err)
	require.NotNil(t, cache)
	require.Equal(t, math.MaxUint32, cache.MaxSize())

	invalidConfig := config
	invalidConfig.Name = ""
	requireErrorOnNewCache(t, invalidConfig, storage.ErrInvalidConfig, "config.Name")

	invalidConfig = config
	invalidConfig.NumChunks = 0
	requireErrorOnNewCache(t, invalidConfig, storage.ErrInvalidConfig, "config.NumChunks")

	invalidConfig = config
	invalidConfig.MaxNumItems = 0
	requireErrorOnNewCache(t, invalidConfig, storage.ErrInvalidConfig, "config.MaxNumItems")

	invalidConfig = config
	invalidConfig.MaxNumBytes = 0
	requireErrorOnNewCache(t, invalidConfig, storage.ErrInvalidConfig, "config.MaxNumBytes")

	invalidConfig = config
	invalidConfig.NumItemsToPreemptivelyEvict = 0
	requireErrorOnNewCache(t, invalidConfig, storage.ErrInvalidConfig, "config.NumItemsToPreemptivelyEvict")
}

func requireErrorOnNewCache(t *testing.T, config CacheConfig, errExpected error, errPartialMessage string) {
	cache, errReceived := NewImmunityCache(config)
	require.Nil(t, cache)
	require.True(t, errors.Is(errReceived, errExpected))
	require.Contains(t, errReceived.Error(), errPartialMessage)
}

func TestImmunityCache_ImmunizeAgainstEviction(t *testing.T) {
	cache := newCacheToTest(1, 8, maxNumBytesUpperBound)

	cache.addTestItems("a", "b", "c", "d")
	numNow, numFuture := cache.ImmunizeKeys(keysAsBytes([]string{"a", "b", "e", "f"}))
	require.Equal(t, 2, numNow)
	require.Equal(t, 2, numFuture)
	require.Equal(t, 4, cache.Len())
	require.Equal(t, 4, cache.CountImmune())

	cache.addTestItems("e", "f", "g", "h")
	require.ElementsMatch(t, []string{"a", "b", "c", "d", "e", "f", "g", "h"}, keysAsStrings(cache.Keys()))

	cache.addTestItems("i", "j", "k", "l")
	require.ElementsMatch(t, []string{"a", "b", "e", "f", "i", "j", "k", "l"}, keysAsStrings(cache.Keys()))

	require.Equal(t, 4, cache.CountImmune())
	cache.Remove([]byte("e"))
	cache.Remove([]byte("f"))
	require.Equal(t, 2, cache.CountImmune())
}

func TestImmunityCache_AddThenRemove(t *testing.T) {
	cache := newCacheToTest(1, 8, maxNumBytesUpperBound)

	_, _ = cache.Add(newCacheItem("a"))
	_, _ = cache.Add(newCacheItem("b"))
	_, _ = cache.HasOrAdd(nil, newCacheItem("c"), 0) // Same as Add()
	_ = cache.Put(nil, newCacheItem("d"), 0)         // Same as Add()
	require.Equal(t, 4, cache.Len())
	require.True(t, cache.Has([]byte("a")))
	require.True(t, cache.Has([]byte("c")))

	// Duplicates are not added
	ok, added := cache.Add(newCacheItem("a"))
	require.True(t, ok)
	require.False(t, added)

	// Won't remove if not exists
	ok = cache.RemoveWithResult([]byte("x"))
	require.False(t, ok)

	cache.Remove([]byte("a"))
	cache.Remove([]byte("c"))
	require.Equal(t, 2, cache.Len())
	require.False(t, cache.Has([]byte("a")))
	require.False(t, cache.Has([]byte("c")))

	cache.addTestItems("e", "f", "g", "h", "i", "j")
	require.Equal(t, 8, cache.Len())

	// Now eviction takes place
	cache.addTestItems("k", "l", "m", "n")
	require.Equal(t, 8, cache.Len())
	require.ElementsMatch(t, []string{"g", "h", "i", "j", "k", "l", "m", "n"}, keysAsStrings(cache.Keys()))

	cache.Clear()
	require.Equal(t, 0, cache.Len())
}

func TestImmunityCache_Get(t *testing.T) {
	cache := newCacheToTest(1, 8, maxNumBytesUpperBound)

	a := newCacheItem("a")
	b := newCacheItem("b")
	ok, added := cache.Add(a)
	require.True(t, ok && added)
	ok, added = cache.Add(b)
	require.True(t, ok && added)

	item, ok := cache.GetItem([]byte("a"))
	require.True(t, ok)
	require.Equal(t, a, item)

	itemAsEmptyInterface, ok := cache.Get([]byte("a"))
	require.True(t, ok)
	require.Equal(t, a, itemAsEmptyInterface)

	itemAsEmptyInterface, ok = cache.Peek([]byte("a"))
	require.True(t, ok)
	require.Equal(t, a, itemAsEmptyInterface)

	item, ok = cache.GetItem([]byte("b"))
	require.True(t, ok)
	require.Equal(t, b, item)

	itemAsEmptyInterface, ok = cache.Get([]byte("b"))
	require.True(t, ok)
	require.Equal(t, b, itemAsEmptyInterface)

	itemAsEmptyInterface, ok = cache.Peek([]byte("b"))
	require.True(t, ok)
	require.Equal(t, b, itemAsEmptyInterface)

	item, ok = cache.GetItem([]byte("c"))
	require.False(t, ok)
	require.Nil(t, item)

	itemAsEmptyInterface, ok = cache.Get([]byte("c"))
	require.False(t, ok)
	require.Nil(t, itemAsEmptyInterface)

	itemAsEmptyInterface, ok = cache.Peek([]byte("c"))
	require.False(t, ok)
	require.Nil(t, itemAsEmptyInterface)
}

func TestImmunityCache_AddThenRemove_ChangesNumBytes(t *testing.T) {
	cache := newCacheToTest(1, 8, 1000)

	_, _ = cache.Add(newCacheItemWithSize("a", 100))
	_, _ = cache.Add(newCacheItemWithSize("b", 300))
	require.Equal(t, 400, cache.NumBytes())

	_, _ = cache.Add(newCacheItemWithSize("c", 400))
	_, _ = cache.Add(newCacheItemWithSize("d", 200))
	require.Equal(t, 1000, cache.NumBytes())

	// Eviction takes place
	_, _ = cache.Add(newCacheItemWithSize("e", 500))
	// Edge case, added item overflows.
	// Should not be an issue in practice, when we preemptively evict a large number of items.
	require.Equal(t, 1400, cache.NumBytes())
	require.ElementsMatch(t, []string{"b", "c", "d", "e"}, keysAsStrings(cache.Keys()))

	// "b" and "c" (300 + 400) will be evicted
	_, _ = cache.Add(newCacheItemWithSize("f", 400))
	require.Equal(t, 1100, cache.NumBytes())
}

func TestImmunityCache_AddDoesNotWork_WhenFullWithImmune(t *testing.T) {
	cache := newCacheToTest(1, 4, 1000)

	cache.addTestItems("a", "b", "c", "d")
	numNow, numFuture := cache.ImmunizeKeys(keysAsBytes([]string{"a", "b", "c", "d"}))
	require.Equal(t, 4, numNow)
	require.Equal(t, 0, numFuture)

	ok, added := cache.Add(newCacheItem("x"))
	require.False(t, ok)
	require.False(t, added)
	require.False(t, cache.Has([]byte("x")))
}

func TestImmunityCache_ForEachItem(t *testing.T) {
	cache := newCacheToTest(1, 4, 1000)

	keys := make([]string, 0)
	cache.addTestItems("a", "b", "c", "d")
	cache.ForEachItem(func(key []byte, value storage.CacheItem) {
		keys = append(keys, string(key))
	})

	require.ElementsMatch(t, []string{"a", "b", "c", "d"}, keys)
}

// This information about (hash to chunk) distribution is useful to write tests
func TestImmunityCache_Fnv32Hash(t *testing.T) {
	// Cache with 2 chunks
	require.Equal(t, 0, int(fnv32Hash("a")%2))
	require.Equal(t, 1, int(fnv32Hash("b")%2))
	require.Equal(t, 0, int(fnv32Hash("c")%2))
	require.Equal(t, 1, int(fnv32Hash("d")%2))

	// Cache with 4 chunks
	require.Equal(t, 2, int(fnv32Hash("a")%4))
	require.Equal(t, 1, int(fnv32Hash("b")%4))
	require.Equal(t, 0, int(fnv32Hash("c")%4))
	require.Equal(t, 3, int(fnv32Hash("d")%4))
}

func TestImmunityCache_ClearConcurrentWithRangeOverChunks(t *testing.T) {
	cache := newCacheToTest(16, 4, 1000)
	require.Equal(t, 16, len(cache.chunks))

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		for i := 0; i < 10000; i++ {
			cache.Clear()
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < 10000; i++ {
			for _, chunk := range cache.getChunksWithLock() {
				assert.Equal(t, 0, chunk.Count())
			}
		}
		wg.Done()
	}()

	wg.Wait()
}

func newCacheToTest(numChunks uint32, maxNumItems uint32, numMaxBytes uint32) *ImmunityCache {
	cache, err := NewImmunityCache(CacheConfig{
		Name:                        "test",
		NumChunks:                   numChunks,
		MaxNumItems:                 maxNumItems,
		MaxNumBytes:                 numMaxBytes,
		NumItemsToPreemptivelyEvict: numChunks * 1,
	})
	if err != nil {
		panic(fmt.Sprintf("newCacheToTest(): %s", err))
	}

	return cache
}

func (ic *ImmunityCache) addTestItems(keys ...string) {
	for _, key := range keys {
		_, _ = ic.Add(newCacheItem(key))
	}
}
