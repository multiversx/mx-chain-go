package detector

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/stretchr/testify/require"
)

func TestRoundDataCache_Add_OneRound_FourHeaders(t *testing.T) {
	t.Parallel()

	dataCache := NewRoundHeadersCache(1)

	h1 := &slash.HeaderInfo{Header: &block.HeaderV2{Header: &block.Header{TimeStamp: 1}}, Hash: []byte("hash1")}
	h2 := &slash.HeaderInfo{Header: &block.HeaderV2{Header: &block.Header{TimeStamp: 2}}, Hash: []byte("hash1")}
	h3 := &slash.HeaderInfo{Header: &block.HeaderV2{Header: &block.Header{TimeStamp: 3}}, Hash: []byte("hash2")}
	h4 := &slash.HeaderInfo{Header: &block.HeaderV2{Header: &block.Header{TimeStamp: 4}}, Hash: []byte("hash3")}

	err := dataCache.Add(1, h1)
	require.Nil(t, err)

	err = dataCache.Add(1, h2)
	require.Equal(t, process.ErrHeadersNotDifferentHashes, err)

	err = dataCache.Add(1, h3)
	require.Nil(t, err)

	err = dataCache.Add(1, h4)
	require.Nil(t, err)

	require.Len(t, dataCache.cache, 1)
	require.Len(t, dataCache.cache[1], 3)

	require.Equal(t, []byte("hash1"), dataCache.cache[1][0].GetHash())
	require.Equal(t, uint64(1), dataCache.cache[1][0].GetHeaderHandler().GetTimeStamp())

	require.Equal(t, []byte("hash2"), dataCache.cache[1][1].GetHash())
	require.Equal(t, uint64(3), dataCache.cache[1][1].GetHeaderHandler().GetTimeStamp())

	require.Equal(t, []byte("hash3"), dataCache.cache[1][2].GetHash())
	require.Equal(t, uint64(4), dataCache.cache[1][2].GetHeaderHandler().GetTimeStamp())

}

func TestRoundDataCache_Add_CacheSizeTwo_FourEntriesInCache_ExpectOldestRoundInCacheRemoved(t *testing.T) {
	t.Parallel()

	dataCache := NewRoundHeadersCache(2)

	h0 := &slash.HeaderInfo{Header: &block.HeaderV2{Header: &block.Header{}}, Hash: []byte("hash0")}
	h1 := &slash.HeaderInfo{Header: &block.HeaderV2{Header: &block.Header{TimeStamp: 1}}, Hash: []byte("hash1")}
	h2 := &slash.HeaderInfo{Header: &block.HeaderV2{Header: &block.Header{TimeStamp: 2}}, Hash: []byte("hash2")}
	h3 := &slash.HeaderInfo{Header: &block.HeaderV2{Header: &block.Header{TimeStamp: 3}}, Hash: []byte("hash3")}
	h4 := &slash.HeaderInfo{Header: &block.HeaderV2{Header: &block.Header{TimeStamp: 4}}, Hash: []byte("hash4")}

	err := dataCache.Add(1, h1)
	require.Nil(t, err)

	err = dataCache.Add(2, h2)
	require.Nil(t, err)

	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[1], 1)
	require.Len(t, dataCache.cache[2], 1)

	require.Equal(t, []byte("hash1"), dataCache.cache[1][0].GetHash())
	require.Equal(t, uint64(1), dataCache.cache[1][0].GetHeaderHandler().GetTimeStamp())

	require.Equal(t, []byte("hash2"), dataCache.cache[2][0].GetHash())
	require.Equal(t, uint64(2), dataCache.cache[2][0].GetHeaderHandler().GetTimeStamp())

	err = dataCache.Add(0, h0)
	require.Equal(t, process.ErrHeaderRoundNotRelevant, err)

	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[1], 1)
	require.Len(t, dataCache.cache[2], 1)

	require.Equal(t, []byte("hash1"), dataCache.cache[1][0].GetHash())
	require.Equal(t, uint64(1), dataCache.cache[1][0].GetHeaderHandler().GetTimeStamp())

	require.Equal(t, []byte("hash2"), dataCache.cache[2][0].GetHash())
	require.Equal(t, uint64(2), dataCache.cache[2][0].GetHeaderHandler().GetTimeStamp())

	err = dataCache.Add(3, h3)
	require.Nil(t, err)

	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[2], 1)
	require.Len(t, dataCache.cache[3], 1)

	require.Equal(t, []byte("hash2"), dataCache.cache[2][0].GetHash())
	require.Equal(t, uint64(2), dataCache.cache[2][0].GetHeaderHandler().GetTimeStamp())

	require.Equal(t, []byte("hash3"), dataCache.cache[3][0].GetHash())
	require.Equal(t, uint64(3), dataCache.cache[3][0].GetHeaderHandler().GetTimeStamp())

	err = dataCache.Add(4, h4)
	require.Nil(t, err)

	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[3], 1)
	require.Len(t, dataCache.cache[4], 1)

	require.Equal(t, []byte("hash3"), dataCache.cache[3][0].GetHash())
	require.Equal(t, uint64(3), dataCache.cache[3][0].GetHeaderHandler().GetTimeStamp())

	require.Equal(t, []byte("hash4"), dataCache.cache[4][0].GetHash())
	require.Equal(t, uint64(4), dataCache.cache[4][0].GetHeaderHandler().GetTimeStamp())
}

func TestRoundDataCache_Contains(t *testing.T) {
	t.Parallel()

	dataCache := NewRoundHeadersCache(2)

	h1 := &slash.HeaderInfo{Header: &block.HeaderV2{Header: &block.Header{TimeStamp: 1}}, Hash: []byte("hash1")}
	h2 := &slash.HeaderInfo{Header: &block.HeaderV2{Header: &block.Header{TimeStamp: 2}}, Hash: []byte("hash2")}
	h3 := &slash.HeaderInfo{Header: &block.HeaderV2{Header: &block.Header{TimeStamp: 3}}, Hash: []byte("hash3")}

	err := dataCache.Add(1, h1)
	require.Nil(t, err)

	err = dataCache.Add(1, h2)
	require.Nil(t, err)

	err = dataCache.Add(2, h3)
	require.Nil(t, err)

	require.True(t, dataCache.contains(1, []byte("hash1")))
	require.True(t, dataCache.contains(1, []byte("hash2")))
	require.True(t, dataCache.contains(2, []byte("hash3")))

	require.False(t, dataCache.contains(1, []byte("hash3")))
	require.False(t, dataCache.contains(3, []byte("hash1")))
}

func TestRoundValidatorsDataCache_IsInterfaceNil(t *testing.T) {
	cache := NewRoundValidatorHeaderCache(1)
	require.False(t, cache.IsInterfaceNil())
	cache = nil
	require.True(t, cache.IsInterfaceNil())
}
