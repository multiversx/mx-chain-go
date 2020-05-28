package immunitycache

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImmunityCache_ImmunizeAgainstEviction(t *testing.T) {
	cache := newCacheToTest(1, 8, math.MaxUint32)

	cache.addTestItems("a", "b", "c", "d")
	numNow, numFuture := cache.ImmunizeItemsAgainstEviction(keysAsBytes([]string{"a", "b", "e", "f"}))
	require.Equal(t, 2, numNow)
	require.Equal(t, 2, numFuture)
	require.Equal(t, 4, cache.Len())

	cache.addTestItems("e", "f", "g", "h")
	require.ElementsMatch(t, []string{"a", "b", "c", "d", "e", "f", "g", "h"}, keysAsStrings(cache.Keys()))

	cache.addTestItems("i", "j", "k", "l")
	require.ElementsMatch(t, []string{"a", "b", "e", "f", "i", "j", "k", "l"}, keysAsStrings(cache.Keys()))
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

func newUnconstrainedCacheToTest(numChunks uint32) *ImmunityCache {
	cache, err := NewImmunityCache(CacheConfig{
		Name:                        "test",
		NumChunks:                   numChunks,
		MaxNumItems:                 math.MaxUint32,
		MaxNumBytes:                 math.MaxUint32,
		NumItemsToPreemptivelyEvict: math.MaxUint32,
	})
	if err != nil {
		panic(fmt.Sprintf("newUnconstrainedCacheToTest(): %s", err))
	}

	return cache
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

func (cache *ImmunityCache) addTestItems(keys ...string) {
	for _, key := range keys {
		_, _ = cache.Add(newCacheItem(key))
	}
}
