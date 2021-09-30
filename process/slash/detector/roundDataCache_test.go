package detector

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestRoundDataCache_Add_OneRound_FourHeaders(t *testing.T) {
	t.Parallel()

	dataCache := newRoundDataCache(1)

	dataCache.add(1, []byte("hash1"), &testscommon.HeaderHandlerStub{
		TimestampField: 1,
	})
	dataCache.add(1, []byte("hash1"), &testscommon.HeaderHandlerStub{
		TimestampField: 2,
	})
	dataCache.add(1, []byte("hash2"), &testscommon.HeaderHandlerStub{
		TimestampField: 3,
	})
	dataCache.add(1, []byte("hash3"), &testscommon.HeaderHandlerStub{
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

	dataCache := newRoundDataCache(2)

	dataCache.add(1, []byte("hash1"), &testscommon.HeaderHandlerStub{
		TimestampField: 1,
	})
	dataCache.add(2, []byte("hash2"), &testscommon.HeaderHandlerStub{
		TimestampField: 2,
	})

	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[1], 1)
	require.Len(t, dataCache.cache[2], 1)

	require.Equal(t, "hash1", dataCache.cache[1][0].hash)
	require.Equal(t, uint64(1), dataCache.cache[1][0].header.GetTimeStamp())

	require.Equal(t, "hash2", dataCache.cache[2][0].hash)
	require.Equal(t, uint64(2), dataCache.cache[2][0].header.GetTimeStamp())

	dataCache.add(0, []byte("hash0"), &testscommon.HeaderHandlerStub{})

	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[1], 1)
	require.Len(t, dataCache.cache[2], 1)

	require.Equal(t, "hash1", dataCache.cache[1][0].hash)
	require.Equal(t, uint64(1), dataCache.cache[1][0].header.GetTimeStamp())

	require.Equal(t, "hash2", dataCache.cache[2][0].hash)
	require.Equal(t, uint64(2), dataCache.cache[2][0].header.GetTimeStamp())

	dataCache.add(3, []byte("hash3"), &testscommon.HeaderHandlerStub{
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

func TestRoundDataCache_Contains_Headers(t *testing.T) {
	t.Parallel()

	dataCache := newRoundDataCache(2)

	dataCache.add(1, []byte("hash1"), &testscommon.HeaderHandlerStub{
		TimestampField: 1,
	})
	dataCache.add(1, []byte("hash2"), &testscommon.HeaderHandlerStub{
		TimestampField: 2,
	})
	dataCache.add(2, []byte("hash3"), &testscommon.HeaderHandlerStub{
		TimestampField: 3,
	})

	require.True(t, dataCache.contains(1, []byte("hash1")))
	require.True(t, dataCache.contains(1, []byte("hash2")))
	require.True(t, dataCache.contains(2, []byte("hash3")))

	require.False(t, dataCache.contains(1, []byte("hash3")))
	require.False(t, dataCache.contains(3, []byte("hash1")))

	headers1 := dataCache.headers(1)
	require.Len(t, headers1, 2)
	require.Equal(t, "hash1", headers1[0].hash)
	require.Equal(t, uint64(1), headers1[0].header.GetTimeStamp())
	require.Equal(t, "hash2", headers1[1].hash)
	require.Equal(t, uint64(2), headers1[1].header.GetTimeStamp())

	headers2 := dataCache.headers(2)
	require.Len(t, headers2, 1)
	require.Equal(t, "hash3", headers2[0].hash)
	require.Equal(t, uint64(3), headers2[0].header.GetTimeStamp())

	headers3 := dataCache.headers(3)
	require.Len(t, headers3, 0)
}
