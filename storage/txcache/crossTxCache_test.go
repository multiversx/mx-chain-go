package txcache

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCrossTxCache_DoImmunizeTxsAgainstEviction(t *testing.T) {
	cache := newCrossTxCacheToTest(1, 8, math.MaxUint32)

	cache.addTestTxs("a", "b", "c", "d")
	numNow, numFuture := cache.doImmunizeTxsAgainstEviction(hashesAsBytes([]string{"a", "b", "e", "f"}))
	require.Equal(t, 2, numNow)
	require.Equal(t, 2, numFuture)
	require.Equal(t, 4, cache.Len())

	cache.addTestTxs("e", "f", "g", "h")
	require.ElementsMatch(t, []string{"a", "b", "c", "d", "e", "f", "g", "h"}, hashesAsStrings(cache.Keys()))

	cache.addTestTxs("i", "j", "k", "l")
	require.ElementsMatch(t, []string{"a", "b", "e", "f", "i", "j", "k", "l"}, hashesAsStrings(cache.Keys()))
}

// This information about (hash to chunk) distribution is useful to write tests
func TestCrossTxCache_Fnv32Hash(t *testing.T) {
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

func newUnconstrainedCrossTxCacheToTest(numChunks uint32) *crossTxCache {
	cache, err := NewCrossTxCache(ConfigDestinationMe{
		NumChunks:                   numChunks,
		MaxNumItems:                 math.MaxUint32,
		MaxNumBytes:                 math.MaxUint32,
		NumItemsToPreemptivelyEvict: math.MaxUint32,
	})
	if err != nil {
		panic(fmt.Sprintf("newUnconstrainedCrossTxCacheToTest(): %s", err))
	}

	return cache
}

func newCrossTxCacheToTest(numChunks uint32, maxNumItems uint32, numMaxBytes uint32) *crossTxCache {
	cache, err := NewCrossTxCache(ConfigDestinationMe{
		NumChunks:                   numChunks,
		MaxNumItems:                 maxNumItems,
		MaxNumBytes:                 numMaxBytes,
		NumItemsToPreemptivelyEvict: numChunks * 1,
	})
	if err != nil {
		panic(fmt.Sprintf("newCrossTxCacheToTest(): %s", err))
	}

	return cache
}

func (cache *crossTxCache) addTestTxs(hashes ...string) {
	for _, hash := range hashes {
		_, _ = cache.addTestTx(hash)
	}
}

func (cache *crossTxCache) addTestTx(hash string) (ok, added bool) {
	return cache.AddTx(createTx([]byte(hash), ".", uint64(42)))
}
