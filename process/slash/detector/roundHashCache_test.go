package detector

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/require"
)

func TestRoundDataCache_Add_OneRound_FourHeaders(t *testing.T) {
	t.Parallel()

	dataCache := NewRoundHashCache(1)

	err := dataCache.Add(1, []byte("hash1"))
	require.Nil(t, err)

	err = dataCache.Add(1, []byte("hash1"))
	require.Equal(t, process.ErrHeadersNotDifferentHashes, err)

	err = dataCache.Add(1, []byte("hash2"))
	require.Nil(t, err)

	err = dataCache.Add(1, []byte("hash3"))
	require.Nil(t, err)

	require.Len(t, dataCache.cache, 1)
	require.Len(t, dataCache.cache[1], 3)
	require.Equal(t, []byte("hash1"), dataCache.cache[1][0])
	require.Equal(t, []byte("hash2"), dataCache.cache[1][1])
	require.Equal(t, []byte("hash3"), dataCache.cache[1][2])
}

func TestRoundDataCache_Add_CacheSizeTwo_FourEntriesInCache_ExpectOldestRoundInCacheRemoved(t *testing.T) {
	t.Parallel()

	dataCache := NewRoundHashCache(2)

	err := dataCache.Add(1, []byte("hash1"))
	require.Nil(t, err)

	require.Len(t, dataCache.cache, 1)
	require.Len(t, dataCache.cache[1], 1)
	require.Equal(t, []byte("hash1"), dataCache.cache[1][0])

	err = dataCache.Add(2, []byte("hash2"))
	require.Nil(t, err)

	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[1], 1)
	require.Len(t, dataCache.cache[2], 1)
	require.Equal(t, []byte("hash1"), dataCache.cache[1][0])
	require.Equal(t, []byte("hash2"), dataCache.cache[2][0])

	err = dataCache.Add(0, []byte("hash0"))
	require.Equal(t, process.ErrHeaderRoundNotRelevant, err)

	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[1], 1)
	require.Len(t, dataCache.cache[2], 1)
	require.Equal(t, []byte("hash1"), dataCache.cache[1][0])
	require.Equal(t, []byte("hash2"), dataCache.cache[2][0])

	err = dataCache.Add(3, []byte("hash3"))
	require.Nil(t, err)

	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[2], 1)
	require.Len(t, dataCache.cache[3], 1)
	require.Equal(t, []byte("hash2"), dataCache.cache[2][0])
	require.Equal(t, []byte("hash3"), dataCache.cache[3][0])

	err = dataCache.Add(4, []byte("hash4"))
	require.Nil(t, err)

	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[3], 1)
	require.Len(t, dataCache.cache[4], 1)
	require.Equal(t, []byte("hash3"), dataCache.cache[3][0])
	require.Equal(t, []byte("hash4"), dataCache.cache[4][0])
}

func TestRoundDataCache_Contains(t *testing.T) {
	t.Parallel()

	dataCache := NewRoundHashCache(2)

	err := dataCache.Add(1, []byte("hash1"))
	require.Nil(t, err)

	err = dataCache.Add(1, []byte("hash2"))
	require.Nil(t, err)

	err = dataCache.Add(2, []byte("hash3"))
	require.Nil(t, err)

	require.True(t, dataCache.contains(1, []byte("hash1")))
	require.True(t, dataCache.contains(1, []byte("hash2")))
	require.True(t, dataCache.contains(2, []byte("hash3")))

	require.False(t, dataCache.contains(1, []byte("hash3")))
	require.False(t, dataCache.contains(3, []byte("hash1")))
}

func TestRoundHashCache_Remove(t *testing.T) {
	t.Parallel()

	dataCache := NewRoundHashCache(2)

	err := dataCache.Add(1, []byte("hash1"))
	require.Nil(t, err)
	err = dataCache.Add(1, []byte("hash2"))
	require.Nil(t, err)
	err = dataCache.Add(2, []byte("hash3"))
	require.Nil(t, err)

	dataCache.Remove(1, []byte("hash2"))
	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[1], 1)
	require.Equal(t, []byte("hash1"), dataCache.cache[1][0])
	require.Len(t, dataCache.cache[2], 1)
	require.Equal(t, []byte("hash3"), dataCache.cache[2][0])

	dataCache.Remove(2, []byte("hash3"))
	require.Len(t, dataCache.cache, 1)
	require.Len(t, dataCache.cache[1], 1)
	require.Equal(t, []byte("hash1"), dataCache.cache[1][0])

	dataCache.Remove(2, []byte("hash1"))
	require.Len(t, dataCache.cache, 1)
	require.Len(t, dataCache.cache[1], 1)
	require.Equal(t, []byte("hash1"), dataCache.cache[1][0])
}

func TestRoundValidatorsDataCache_IsInterfaceNil(t *testing.T) {
	cache := NewRoundHashCache(1)
	require.False(t, cache.IsInterfaceNil())
}
