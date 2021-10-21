package detector

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestRoundProposerDataCache_Add_OneRound_TwoProposers_FourInterceptedData(t *testing.T) {
	t.Parallel()
	dataCache := NewRoundValidatorDataCache(1)

	err := dataCache.Add(1, []byte("proposer1"), slash.HeaderInfo{Hash: []byte("hash1")})
	require.Nil(t, err)

	err = dataCache.Add(1, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash1")
		},
	})
	require.Equal(t, process.ErrHeadersNotDifferentHashes, err)

	err = dataCache.Add(1, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash2")
		},
	})
	require.Nil(t, err)

	err = dataCache.Add(1, []byte("proposer2"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash3")
		},
	})
	require.Nil(t, err)

	// One round
	require.Len(t, dataCache.cache, 1)
	// Two proposers in same round
	require.Len(t, dataCache.cache[1], 2)

	// First proposer: proposed two headers
	require.Len(t, dataCache.cache[1]["proposer1"], 2)
	require.Equal(t, dataCache.cache[1]["proposer1"][0].Hash(), []byte("hash1"))
	require.Equal(t, dataCache.cache[1]["proposer1"][1].Hash(), []byte("hash2"))

	// Second proposer: proposed one header
	require.Len(t, dataCache.cache[1]["proposer2"], 1)
	require.Equal(t, dataCache.cache[1]["proposer2"][0].Hash(), []byte("hash3"))
}

func TestRoundProposerDataCache_Add_CacheSizeTwo_FourEntriesInCache_ExpectOldestRoundInCacheRemoved(t *testing.T) {
	t.Parallel()
	dataCache := NewRoundValidatorDataCache(2)

	err := dataCache.Add(1, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash1")
		},
	})
	require.Nil(t, err)

	err = dataCache.Add(2, []byte("proposer2"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash2")
		},
	})
	require.Nil(t, err)

	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[1], 1)
	require.Len(t, dataCache.cache[2], 1)
	require.Len(t, dataCache.cache[1]["proposer1"], 1)
	require.Len(t, dataCache.cache[2]["proposer2"], 1)

	require.Equal(t, dataCache.cache[1]["proposer1"][0].Hash(), []byte("hash1"))
	require.Equal(t, dataCache.cache[2]["proposer2"][0].Hash(), []byte("hash2"))

	err = dataCache.Add(0, []byte("proposer3"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash3")
		},
	})
	require.Equal(t, process.ErrHeaderRoundNotRelevant, err)

	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[1], 1)
	require.Len(t, dataCache.cache[2], 1)
	require.Len(t, dataCache.cache[1]["proposer1"], 1)
	require.Len(t, dataCache.cache[2]["proposer2"], 1)

	require.Equal(t, dataCache.cache[1]["proposer1"][0].Hash(), []byte("hash1"))
	require.Equal(t, dataCache.cache[2]["proposer2"][0].Hash(), []byte("hash2"))

	err = dataCache.Add(3, []byte("proposer3"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash3")
		},
	})
	require.Nil(t, err)

	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[2], 1)
	require.Len(t, dataCache.cache[3], 1)
	require.Len(t, dataCache.cache[2]["proposer2"], 1)
	require.Len(t, dataCache.cache[3]["proposer3"], 1)

	require.Equal(t, dataCache.cache[2]["proposer2"][0].Hash(), []byte("hash2"))
	require.Equal(t, dataCache.cache[3]["proposer3"][0].Hash(), []byte("hash3"))

	err = dataCache.Add(4, []byte("proposer4"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash4")
		},
	})
	require.Nil(t, err)

	require.Len(t, dataCache.cache, 2)
	require.Len(t, dataCache.cache[3], 1)
	require.Len(t, dataCache.cache[4], 1)
	require.Len(t, dataCache.cache[3]["proposer3"], 1)
	require.Len(t, dataCache.cache[4]["proposer4"], 1)

	require.Equal(t, dataCache.cache[3]["proposer3"][0].Hash(), []byte("hash3"))
	require.Equal(t, dataCache.cache[4]["proposer4"][0].Hash(), []byte("hash4"))
}

func TestRoundProposerDataCache_GetData(t *testing.T) {
	t.Parallel()
	dataCache := NewRoundValidatorDataCache(3)

	err := dataCache.Add(1, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash1")
		},
	})
	require.Nil(t, err)

	err = dataCache.Add(1, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash2")
		},
	})
	require.Nil(t, err)

	err = dataCache.Add(2, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash2")
		},
	})
	require.Nil(t, err)

	err = dataCache.Add(2, []byte("proposer2"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash3")
		},
	})
	require.Nil(t, err)

	require.Len(t, dataCache.cache, 2)

	data1 := dataCache.GetData(1, []byte("proposer1"))
	require.Len(t, data1, 2)
	require.Equal(t, data1[0].Hash(), []byte("hash1"))
	require.Equal(t, data1[1].Hash(), []byte("hash2"))

	data1 = dataCache.GetData(2, []byte("proposer1"))
	require.Len(t, data1, 1)
	require.Equal(t, data1[0].Hash(), []byte("hash2"))

	data2 := dataCache.GetData(2, []byte("proposer2"))
	require.Len(t, data2, 1)
	require.Equal(t, data2[0].Hash(), []byte("hash3"))

	data3 := dataCache.GetData(444, []byte("this proposer is not cached"))
	require.Nil(t, data3)
}

func TestRoundProposerDataCache_GetValidators(t *testing.T) {
	t.Parallel()
	dataCache := NewRoundValidatorDataCache(2)

	err := dataCache.Add(1, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash1")
		},
	})
	require.Nil(t, err)

	err = dataCache.Add(1, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash2")
		},
	})
	require.Nil(t, err)

	err = dataCache.Add(1, []byte("proposer2"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash2")
		},
	})
	require.Nil(t, err)

	err = dataCache.Add(2, []byte("proposer2"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash2")
		},
	})
	require.Nil(t, err)

	require.Len(t, dataCache.cache, 2)

	validatorsRound1 := dataCache.GetPubKeys(1)
	require.Len(t, validatorsRound1, 2)
	require.Contains(t, validatorsRound1, []byte("proposer1"))
	require.Contains(t, validatorsRound1, []byte("proposer2"))

	validatorsRound2 := dataCache.GetPubKeys(2)
	require.Len(t, validatorsRound2, 1)
	require.Equal(t, []byte("proposer2"), validatorsRound2[0])
}

func TestRoundProposerDataCache_Contains(t *testing.T) {
	t.Parallel()
	dataCache := NewRoundValidatorDataCache(2)

	err := dataCache.Add(1, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash1")
		},
	})
	require.Nil(t, err)

	err = dataCache.Add(1, []byte("proposer1"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash2")
		},
	})
	require.Nil(t, err)

	err = dataCache.Add(2, []byte("proposer2"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash3")
		},
	})
	require.Nil(t, err)

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
	require.False(t, dataCache.contains(3, []byte("proposer1"), expectedData1))
}

func TestRoundHeadersCache_IsInterfaceNil(t *testing.T) {
	cache := NewRoundHeadersCache(1)
	require.False(t, cache.IsInterfaceNil())
	cache = nil
	require.True(t, cache.IsInterfaceNil())
}
