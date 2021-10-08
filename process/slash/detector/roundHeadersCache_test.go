package detector

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestRoundDataCache_Add_OneRound_FourHeaders(t *testing.T) {
	t.Parallel()

	dataCache := NewRoundHeadersCache(1)

	dataCache.Add(1, []byte("hash1"), &testscommon.HeaderHandlerStub{
		TimestampField: 1,
	})
	dataCache.Add(1, []byte("hash1"), &testscommon.HeaderHandlerStub{
		TimestampField: 2,
	})
	dataCache.Add(1, []byte("hash2"), &testscommon.HeaderHandlerStub{
		TimestampField: 3,
	})
	dataCache.Add(1, []byte("hash3"), &testscommon.HeaderHandlerStub{
		TimestampField: 4,
	})

	require.Len(t, dataCache.cache, 1)
	require.Len(t, dataCache.cache[1], 4)

	require.Equal(t, "hash1", dataCache.cache[1][0].hash)
	require.Equal(t, uint64(1), dataCache.cache[1][0].header.GetTimeStamp())

	require.Equal(t, "hash1", dataCache.cache[1][1].hash)
	require.Equal(t, uint64(2), dataCache.cache[1][1].header.GetTimeStamp())

	require.Equal(t, "hash2", dataCache.cache[1][2].hash)
	require.Equal(t, uint64(3), dataCache.cache[1][2].header.GetTimeStamp())

	require.Equal(t, "hash3", dataCache.cache[1][3].hash)
	require.Equal(t, uint64(4), dataCache.cache[1][3].header.GetTimeStamp())

}

func TestRoundDataCache_Add_CacheSizeTwo_FourEntriesInCache_ExpectOldestRoundInCacheRemoved(t *testing.T) {
	t.Parallel()

	dataCache := NewRoundHeadersCache(2)

	dataCache.Add(1, []byte("hash1"), &testscommon.HeaderHandlerStub{
		TimestampField: 1,
	})
	dataCache.Add(2, []byte("hash2"), &testscommon.HeaderHandlerStub{
		TimestampField: 2,
	})

	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[1], 1)
	require.Len(t, dataCache.cache[2], 1)

	require.Equal(t, "hash1", dataCache.cache[1][0].hash)
	require.Equal(t, uint64(1), dataCache.cache[1][0].header.GetTimeStamp())

	require.Equal(t, "hash2", dataCache.cache[2][0].hash)
	require.Equal(t, uint64(2), dataCache.cache[2][0].header.GetTimeStamp())

	dataCache.Add(0, []byte("hash0"), &testscommon.HeaderHandlerStub{})

	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[1], 1)
	require.Len(t, dataCache.cache[2], 1)

	require.Equal(t, "hash1", dataCache.cache[1][0].hash)
	require.Equal(t, uint64(1), dataCache.cache[1][0].header.GetTimeStamp())

	require.Equal(t, "hash2", dataCache.cache[2][0].hash)
	require.Equal(t, uint64(2), dataCache.cache[2][0].header.GetTimeStamp())

	dataCache.Add(3, []byte("hash3"), &testscommon.HeaderHandlerStub{
		TimestampField: 3,
	})

	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[2], 1)
	require.Len(t, dataCache.cache[3], 1)

	require.Equal(t, "hash2", dataCache.cache[2][0].hash)
	require.Equal(t, uint64(2), dataCache.cache[2][0].header.GetTimeStamp())

	require.Equal(t, "hash3", dataCache.cache[3][0].hash)
	require.Equal(t, uint64(3), dataCache.cache[3][0].header.GetTimeStamp())
}

func TestRoundDataCache_Contains(t *testing.T) {
	t.Parallel()

	dataCache := NewRoundHeadersCache(2)

	go dataCache.Add(1, []byte("hash1"), &testscommon.HeaderHandlerStub{
		TimestampField: 1,
	})
	go dataCache.Add(1, []byte("hash2"), &testscommon.HeaderHandlerStub{
		TimestampField: 2,
	})
	go dataCache.Add(2, []byte("hash3"), &testscommon.HeaderHandlerStub{
		TimestampField: 3,
	})
	time.Sleep(time.Millisecond)

	require.True(t, dataCache.Contains(1, []byte("hash1")))
	require.True(t, dataCache.Contains(1, []byte("hash2")))
	require.True(t, dataCache.Contains(2, []byte("hash3")))

	require.False(t, dataCache.Contains(1, []byte("hash3")))
	require.False(t, dataCache.Contains(3, []byte("hash1")))
}

func TestRoundValidatorsDataCache_IsInterfaceNil(t *testing.T) {
	cache := NewRoundValidatorDataCache(1)
	require.False(t, cache.IsInterfaceNil())
	cache = nil
	require.True(t, cache.IsInterfaceNil())
}
