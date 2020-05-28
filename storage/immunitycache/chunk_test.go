package immunitycache

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImmunityChunk_ImmunizeKeys(t *testing.T) {
	chunk := newUnconstrainedChunkToTest()

	chunk.addTestItems("x", "y", "z")
	require.Equal(t, 3, chunk.CountItems())

	// No immune items, all removed
	numRemoved := chunk.RemoveOldest(42)
	require.Equal(t, 3, numRemoved)
	require.Equal(t, 0, chunk.CountItems())

	chunk.addTestItems("x", "y", "z")
	require.Equal(t, 3, chunk.CountItems())

	// Immunize some items
	numNow, numFuture := chunk.ImmunizeKeys(keysAsBytes([]string{"x", "z"}))
	require.Equal(t, 2, numNow)
	require.Equal(t, 0, numFuture)

	numRemoved = chunk.RemoveOldest(42)
	require.Equal(t, 1, numRemoved)
	require.Equal(t, 2, chunk.CountItems())
	require.Equal(t, []string{"x", "z"}, keysAsStrings(chunk.KeysInOrder()))
}

func TestImmunityChunk_AddItemIgnoresDuplicates(t *testing.T) {
	chunk := newUnconstrainedChunkToTest()
	chunk.addTestItems("x", "y", "z")
	require.Equal(t, 3, chunk.CountItems())

	ok, added := chunk.addItem(newCacheItem("a"))
	require.True(t, ok)
	require.True(t, added)
	require.Equal(t, 4, chunk.CountItems())

	ok, added = chunk.addItem(newCacheItem("x"))
	require.True(t, ok)
	require.False(t, added)
	require.Equal(t, 4, chunk.CountItems())
}

func TestImmunityChunk_AddItemEvictsWhenTooMany(t *testing.T) {
	chunk := newChunkToTest(3, math.MaxUint32)
	chunk.addTestItems("x", "y", "z")
	require.Equal(t, 3, chunk.CountItems())

	chunk.addTestItems("a", "b")
	require.Equal(t, []string{"z", "a", "b"}, keysAsStrings(chunk.KeysInOrder()))
}

func TestImmunityChunk_AddItemDoesNotEvictImmuneItems(t *testing.T) {
	chunk := newChunkToTest(3, math.MaxUint32)
	chunk.addTestItems("x", "y", "z")
	require.Equal(t, 3, chunk.CountItems())

	_, _ = chunk.ImmunizeKeys(keysAsBytes([]string{"x", "y"}))

	chunk.addTestItems("a")
	require.Equal(t, []string{"x", "y", "a"}, keysAsStrings(chunk.KeysInOrder()))
	chunk.addTestItems("b")
	require.Equal(t, []string{"x", "y", "b"}, keysAsStrings(chunk.KeysInOrder()))

	_, _ = chunk.ImmunizeKeys(keysAsBytes([]string{"b"}))
	ok, added := chunk.addItem(newCacheItem("c"))
	require.False(t, ok)
	require.False(t, added)
	require.Equal(t, []string{"x", "y", "b"}, keysAsStrings(chunk.KeysInOrder()))
}

func newUnconstrainedChunkToTest() *immunityChunk {
	chunk := newImmunityChunk(immunityChunkConfig{
		maxNumItems:                 math.MaxUint32,
		maxNumBytes:                 math.MaxUint32,
		numItemsToPreemptivelyEvict: math.MaxUint32,
	})

	return chunk
}

func newChunkToTest(maxNumItems uint32, numMaxBytes uint32) *immunityChunk {
	chunk := newImmunityChunk(immunityChunkConfig{
		maxNumItems:                 maxNumItems,
		maxNumBytes:                 numMaxBytes,
		numItemsToPreemptivelyEvict: 1,
	})

	return chunk
}

func (chunk *immunityChunk) addTestItems(keys ...string) {
	for _, key := range keys {
		_, _ = chunk.addItem(newCacheItem(key))
	}
}
