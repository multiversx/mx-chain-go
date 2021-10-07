package detector

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestRoundProposerDataCache_Add_OneRound_TwoProposers_FourInterceptedData(t *testing.T) {
	t.Parallel()
	dataCache := newRoundProposerDataCache(1)

	dataCache.add(1, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash1")
		},
	})
	dataCache.add(1, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash1")
		},
	})
	dataCache.add(1, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash2")
		},
	})
	dataCache.add(1, []byte("proposer2"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash3")
		},
	})

	// One round
	require.Len(t, dataCache.cache, 1)
	// Two proposers in same round
	require.Len(t, dataCache.cache[1], 2)

	// First proposer: proposed three headers
	require.Len(t, dataCache.cache[1]["proposer1"], 3)
	require.Equal(t, dataCache.cache[1]["proposer1"][0].Hash(), []byte("hash1"))
	require.Equal(t, dataCache.cache[1]["proposer1"][1].Hash(), []byte("hash1"))
	require.Equal(t, dataCache.cache[1]["proposer1"][2].Hash(), []byte("hash2"))

	// Second proposer: proposed one header
	require.Len(t, dataCache.cache[1]["proposer2"], 1)
	require.Equal(t, dataCache.cache[1]["proposer2"][0].Hash(), []byte("hash3"))
}

func TestRoundProposerDataCache_Add_CacheSizeTwo_FourEntriesInCache_ExpectOldestRoundInCacheRemoved(t *testing.T) {
	t.Parallel()
	dataCache := newRoundProposerDataCache(2)

	dataCache.add(1, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash1")
		},
	})
	dataCache.add(2, []byte("proposer2"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash2")
		},
	})

	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[1], 1)
	require.Len(t, dataCache.cache[2], 1)
	require.Len(t, dataCache.cache[1]["proposer1"], 1)
	require.Len(t, dataCache.cache[2]["proposer2"], 1)

	require.Equal(t, dataCache.cache[1]["proposer1"][0].Hash(), []byte("hash1"))
	require.Equal(t, dataCache.cache[2]["proposer2"][0].Hash(), []byte("hash2"))

	dataCache.add(0, []byte("proposer3"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash3")
		},
	})

	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[1], 1)
	require.Len(t, dataCache.cache[2], 1)
	require.Len(t, dataCache.cache[1]["proposer1"], 1)
	require.Len(t, dataCache.cache[2]["proposer2"], 1)

	require.Equal(t, dataCache.cache[1]["proposer1"][0].Hash(), []byte("hash1"))
	require.Equal(t, dataCache.cache[2]["proposer2"][0].Hash(), []byte("hash2"))

	dataCache.add(3, []byte("proposer3"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash3")
		},
	})

	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[2], 1)
	require.Len(t, dataCache.cache[3], 1)
	require.Len(t, dataCache.cache[2]["proposer2"], 1)
	require.Len(t, dataCache.cache[3]["proposer3"], 1)

	require.Equal(t, dataCache.cache[2]["proposer2"][0].Hash(), []byte("hash2"))
	require.Equal(t, dataCache.cache[3]["proposer3"][0].Hash(), []byte("hash3"))
}

func TestRoundProposerDataCache_GetData(t *testing.T) {
	t.Parallel()
	dataCache := newRoundProposerDataCache(3)

	dataCache.add(1, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash1")
		},
	})
	dataCache.add(1, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash2")
		},
	})
	dataCache.add(2, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash2")
		},
	})
	dataCache.add(2, []byte("proposer2"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash3")
		},
	})

	require.Len(t, dataCache.cache, 2)

	data1 := dataCache.data(1, []byte("proposer1"))
	require.Len(t, data1, 2)
	require.Equal(t, data1[0].Hash(), []byte("hash1"))
	require.Equal(t, data1[1].Hash(), []byte("hash2"))

	data1 = dataCache.data(2, []byte("proposer1"))
	require.Len(t, data1, 1)
	require.Equal(t, data1[0].Hash(), []byte("hash2"))

	data2 := dataCache.data(2, []byte("proposer2"))
	require.Len(t, data2, 1)
	require.Equal(t, data2[0].Hash(), []byte("hash3"))

	data3 := dataCache.data(444, []byte("this proposer is not cached"))
	require.Nil(t, data3)
}

func TestRoundProposerDataCache_GetValidators(t *testing.T) {
	t.Parallel()
	dataCache := newRoundProposerDataCache(2)

	dataCache.add(1, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash1")
		},
	})
	dataCache.add(1, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash2")
		},
	})
	dataCache.add(1, []byte("proposer2"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash2")
		},
	})
	dataCache.add(2, []byte("proposer2"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash2")
		},
	})

	require.Len(t, dataCache.cache, 2)

	validatorsRound1 := dataCache.validators(1)
	require.Len(t, validatorsRound1, 2)
	require.Equal(t, []byte("proposer1"), validatorsRound1[0])
	require.Equal(t, []byte("proposer2"), validatorsRound1[1])

	validatorsRound2 := dataCache.validators(2)
	require.Len(t, validatorsRound2, 1)
	require.Equal(t, []byte("proposer2"), validatorsRound2[0])
}

func TestRoundProposerDataCache_Contains(t *testing.T) {
	t.Parallel()
	dataCache := newRoundProposerDataCache(2)

	go dataCache.add(1, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash1")
		},
	})
	go dataCache.add(1, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash2")
		},
	})
	go dataCache.add(2, []byte("proposer2"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash3")
		},
	})
	time.Sleep(time.Millisecond)

	expectedData1 := &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash1")
		},
	}
	expectedData2 := &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash2")
		},
	}
	expectedData3 := &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash3")
		},
	}
	require.True(t, dataCache.contains(1, []byte("proposer1"), expectedData1))
	require.True(t, dataCache.contains(1, []byte("proposer1"), expectedData2))
	require.True(t, dataCache.contains(2, []byte("proposer2"), expectedData3))

	require.False(t, dataCache.contains(1, []byte("proposer1"), expectedData3))
	require.False(t, dataCache.contains(2, []byte("proposer1"), expectedData1))
	require.False(t, dataCache.contains(1, []byte("proposer2"), expectedData3))
	require.False(t, dataCache.contains(1, []byte("proposer2"), expectedData2))
}
