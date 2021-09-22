package detector

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestRoundProposerDataCache_Add_OneRound_TwoProposers_ThreeInterceptedData(t *testing.T) {
	dataCache := newRoundProposerDataCache(1)

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

	// First proposer: proposed two headers
	require.Len(t, dataCache.cache[1]["proposer1"], 2)
	require.Equal(t, dataCache.cache[1]["proposer1"][0].Hash(), []byte("hash1"))
	require.Equal(t, dataCache.cache[1]["proposer1"][1].Hash(), []byte("hash2"))

	// Second proposer: proposed one header
	require.Len(t, dataCache.cache[1]["proposer2"], 1)
	require.Equal(t, dataCache.cache[1]["proposer2"][0].Hash(), []byte("hash3"))
}

func TestRoundProposerDataCache_Add_CacheSizeTwo_ThreeRounds_ExpectOldestEntryInCacheRemoved(t *testing.T) {
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

	dataCache.add(1, []byte("proposer2"), &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash3")
		},
	})

}

func TestRoundProposerDataCache_Add(t *testing.T) {
	dataCache := newRoundProposerDataCache(3)

	dataCache.add(1, []byte("proposer1"), &interceptedBlocks.InterceptedHeader{})
	dataCache.add(1, []byte("proposer1"), &interceptedBlocks.InterceptedHeader{})
	dataCache.add(1, []byte("proposer1"), &interceptedBlocks.InterceptedHeader{})
	dataCache.add(1, []byte("proposer2"), &interceptedBlocks.InterceptedHeader{})
	dataCache.add(2, []byte("proposer3"), &interceptedBlocks.InterceptedHeader{})
	dataCache.add(3, []byte("proposer4"), &interceptedBlocks.InterceptedHeader{})

	require.Len(t, dataCache.cache, 3)
	require.Len(t, dataCache.cache[1], 2)
	require.Len(t, dataCache.cache[1]["proposer1"], 3)
	require.Len(t, dataCache.cache[1]["proposer2"], 1)

	require.Len(t, dataCache.cache[2], 1)
	require.Len(t, dataCache.cache[2]["proposer3"], 1)

	require.Len(t, dataCache.cache[3], 1)
	require.Len(t, dataCache.cache[3]["proposer4"], 1)

	dataCache.add(4, []byte("proposer5"), &interceptedBlocks.InterceptedHeader{})
	require.Len(t, dataCache.cache, 3)
	require.Len(t, dataCache.cache[2], 1)
	require.Len(t, dataCache.cache[2]["proposer3"], 1)

	require.Len(t, dataCache.cache[3], 1)
	require.Len(t, dataCache.cache[3]["proposer4"], 1)

	require.Len(t, dataCache.cache[4], 1)
	require.Len(t, dataCache.cache[4]["proposer5"], 1)
}
