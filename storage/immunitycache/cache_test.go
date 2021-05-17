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

func TestImmunityCache_ImmunizeDoesNothingIfCapacityReached(t *testing.T) {
	cache := newCacheToTest(1, 4, maxNumBytesUpperBound)

	numNow, numFuture := cache.ImmunizeKeys(keysAsBytes([]string{"a", "b", "c", "d"}))
	require.Equal(t, 0, numNow)
	require.Equal(t, 4, numFuture)
	require.Equal(t, 4, cache.CountImmune())

	numNow, numFuture = cache.ImmunizeKeys(keysAsBytes([]string{"e", "f", "g", "h"}))
	require.Equal(t, 0, numNow)
	require.Equal(t, 0, numFuture)
	require.Equal(t, 4, cache.CountImmune())
}

func TestImmunityCache_AddThenRemove(t *testing.T) {
	cache := newCacheToTest(1, 8, maxNumBytesUpperBound)

	_, _ = cache.HasOrAdd([]byte("a"), "foo-a", 1)
	_, _ = cache.HasOrAdd([]byte("b"), "foo-b", 1)
	_, _ = cache.HasOrAdd([]byte("c"), "foo-c", 0)
	_ = cache.Put([]byte("d"), "foo-d", 0) // Same as HasOrAdd()
	require.Equal(t, 4, cache.Len())
	require.Equal(t, 4, int(cache.hospitality.Get()))
	require.True(t, cache.Has([]byte("a")))
	require.True(t, cache.Has([]byte("c")))

	// Duplicates are not added
	_, added := cache.HasOrAdd([]byte("a"), "foo-a", 1)
	require.False(t, added)

	// Won't remove if not exists
	ok := cache.RemoveWithResult([]byte("x"))
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

	a := "foo-a"
	b := "foo-b"
	_, added := cache.HasOrAdd([]byte("a"), a, 1)
	require.True(t, added)
	_, added = cache.HasOrAdd([]byte("b"), b, 1)
	require.True(t, added)

	item, ok := cache.Get([]byte("a"))
	require.True(t, ok)
	require.Equal(t, a, item)

	itemAsEmptyInterface, ok := cache.Get([]byte("a"))
	require.True(t, ok)
	require.Equal(t, a, itemAsEmptyInterface)

	itemAsEmptyInterface, ok = cache.Peek([]byte("a"))
	require.True(t, ok)
	require.Equal(t, a, itemAsEmptyInterface)

	item, ok = cache.Get([]byte("b"))
	require.True(t, ok)
	require.Equal(t, b, item)

	itemAsEmptyInterface, ok = cache.Get([]byte("b"))
	require.True(t, ok)
	require.Equal(t, b, itemAsEmptyInterface)

	itemAsEmptyInterface, ok = cache.Peek([]byte("b"))
	require.True(t, ok)
	require.Equal(t, b, itemAsEmptyInterface)

	item, ok = cache.Get([]byte("c"))
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

	_, _ = cache.HasOrAdd([]byte("a"), "foo-a", 100)
	_, _ = cache.HasOrAdd([]byte("b"), "foo-b", 300)
	require.Equal(t, 400, cache.NumBytes())

	_, _ = cache.HasOrAdd([]byte("c"), "foo-c", 400)
	_, _ = cache.HasOrAdd([]byte("d"), "foo-d", 200)
	require.Equal(t, 1000, cache.NumBytes())

	// Eviction takes place
	_, _ = cache.HasOrAdd([]byte("e"), "foo-e", 500)
	// Edge case, added item overflows.
	// Should not be an issue in practice, when we preemptively evict a large number of items.
	require.Equal(t, 1400, cache.NumBytes())
	require.ElementsMatch(t, []string{"b", "c", "d", "e"}, keysAsStrings(cache.Keys()))

	// "b" and "c" (300 + 400) will be evicted
	_, _ = cache.HasOrAdd([]byte("f"), "foo-f", 400)
	require.Equal(t, 1100, cache.NumBytes())
}

func TestImmunityCache_AddDoesNotWork_WhenFullWithImmune(t *testing.T) {
	cache := newCacheToTest(1, 4, 1000)

	cache.addTestItems("a", "b", "c", "d")
	numNow, numFuture := cache.ImmunizeKeys(keysAsBytes([]string{"a", "b", "c", "d"}))
	require.Equal(t, 4, numNow)
	require.Equal(t, 0, numFuture)
	require.Equal(t, 4, int(cache.hospitality.Get()))

	_, added := cache.HasOrAdd([]byte("x"), "foo-x", 1)
	require.False(t, added)
	require.False(t, cache.Has([]byte("x")))
	require.Equal(t, 3, int(cache.hospitality.Get()))
}

func TestImmunityCache_ForEachItem(t *testing.T) {
	cache := newCacheToTest(1, 4, 1000)

	keys := make([]string, 0)
	cache.addTestItems("a", "b", "c", "d")
	cache.ForEachItem(func(key []byte, value interface{}) {
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

func TestImmunityCache_DiagnoseAppliesLimitToHospitality(t *testing.T) {
	cache := newCacheToTest(1, hospitalityUpperLimit*42, 1000)

	for i := 0; i < hospitalityUpperLimit*2; i++ {
		cache.addTestItems(fmt.Sprintf("%d", i))
		require.Equal(t, i+1, int(cache.hospitality.Get()))
	}

	require.Equal(t, hospitalityUpperLimit*2, int(cache.hospitality.Get()))
	cache.Diagnose(false)
	require.Equal(t, hospitalityUpperLimit, int(cache.hospitality.Get()))
}

func TestImmunityCache_DiagnoseResetsHospitalityAfterWarn(t *testing.T) {
	cache := newCacheToTest(1, 4, 1000)
	cache.addTestItems("a", "b", "c", "d")
	_, _ = cache.ImmunizeKeys(keysAsBytes([]string{"a", "b", "c", "d"}))
	require.Equal(t, 4, int(cache.hospitality.Get()))

	cache.addTestItems("e", "f", "g", "h")
	require.Equal(t, 0, int(cache.hospitality.Get()))

	for i := -1; i > hospitalityWarnThreshold; i-- {
		cache.addTestItems("foo")
		require.Equal(t, i, int(cache.hospitality.Get()))
	}

	require.Equal(t, hospitalityWarnThreshold+1, int(cache.hospitality.Get()))
	cache.Diagnose(false)
	require.Equal(t, hospitalityWarnThreshold+1, int(cache.hospitality.Get()))
	cache.addTestItems("foo")
	require.Equal(t, hospitalityWarnThreshold, int(cache.hospitality.Get()))
	cache.Diagnose(false)
	require.Equal(t, 0, int(cache.hospitality.Get()))
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

func TestImmunityCache_CloseDoesNotErr(t *testing.T) {
	cache := newCacheToTest(1, 4, 1000)

	err := cache.Close()
	assert.Nil(t, err)
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
		_, _ = ic.HasOrAdd([]byte(key), fmt.Sprintf("foo-%s", key), 100)
	}
}
