package detector

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/stretchr/testify/require"
)

func TestRoundProposerDataCache_AddProposerData(t *testing.T) {
	cache := newRoundProposerDataCache(3)

	cache.addProposerData(1, []byte("proposer1"), &interceptedBlocks.InterceptedHeader{})
	cache.addProposerData(1, []byte("proposer1"), &interceptedBlocks.InterceptedHeader{})
	cache.addProposerData(1, []byte("proposer1"), &interceptedBlocks.InterceptedHeader{})
	cache.addProposerData(1, []byte("proposer2"), &interceptedBlocks.InterceptedHeader{})
	cache.addProposerData(2, []byte("proposer3"), &interceptedBlocks.InterceptedHeader{})
	cache.addProposerData(3, []byte("proposer4"), &interceptedBlocks.InterceptedHeader{})

	require.Len(t, cache.cache, 3)
	require.Len(t, cache.cache[1], 2)
	require.Len(t, cache.cache[1]["proposer1"], 3)
	require.Len(t, cache.cache[1]["proposer2"], 1)

	require.Len(t, cache.cache[2], 1)
	require.Len(t, cache.cache[2]["proposer3"], 1)

	require.Len(t, cache.cache[3], 1)
	require.Len(t, cache.cache[3]["proposer4"], 1)

	cache.addProposerData(4, []byte("proposer5"), &interceptedBlocks.InterceptedHeader{})
	require.Len(t, cache.cache, 3)
	require.Len(t, cache.cache[2], 1)
	require.Len(t, cache.cache[2]["proposer3"], 1)

	require.Len(t, cache.cache[3], 1)
	require.Len(t, cache.cache[3]["proposer4"], 1)

	require.Len(t, cache.cache[4], 1)
	require.Len(t, cache.cache[4]["proposer5"], 1)
}
