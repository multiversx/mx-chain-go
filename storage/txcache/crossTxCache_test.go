package txcache

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCrossTxCache_DoImmunizeTxsAgainstEviction(t *testing.T) {
	cache := newCrossTxCacheToTest(1, 8, math.MaxUint16)

	cache.addTestTxs("a", "b", "c", "d")
	numNow, numFuture := cache.ImmunizeKeys(hashesAsBytes([]string{"a", "b", "e", "f"}))
	require.Equal(t, 2, numNow)
	require.Equal(t, 2, numFuture)
	require.Equal(t, 4, cache.Len())

	cache.addTestTxs("e", "f", "g", "h")
	require.ElementsMatch(t, []string{"a", "b", "c", "d", "e", "f", "g", "h"}, hashesAsStrings(cache.Keys()))

	cache.addTestTxs("i", "j", "k", "l")
	require.ElementsMatch(t, []string{"a", "b", "e", "f", "i", "j", "k", "l"}, hashesAsStrings(cache.Keys()))
}

func newUnconstrainedCrossTxCacheToTest(numChunks uint32) *CrossTxCache {
	cache, err := NewCrossTxCache(ConfigDestinationMe{
		Name:                        "test",
		NumChunks:                   numChunks,
		MaxNumItems:                 math.MaxUint32,
		MaxNumBytes:                 math.MaxUint16,
		NumItemsToPreemptivelyEvict: math.MaxUint32,
	})
	if err != nil {
		panic(fmt.Sprintf("newUnconstrainedCrossTxCacheToTest(): %s", err))
	}

	return cache
}

func newCrossTxCacheToTest(numChunks uint32, maxNumItems uint32, numMaxBytes uint32) *CrossTxCache {
	cache, err := NewCrossTxCache(ConfigDestinationMe{
		Name:                        "test",
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

func (cache *CrossTxCache) addTestTxs(hashes ...string) {
	for _, hash := range hashes {
		_, _ = cache.addTestTx(hash)
	}
}

func (cache *CrossTxCache) addTestTx(hash string) (ok, added bool) {
	return cache.AddTx(createTx([]byte(hash), ".", uint64(42)))
}
