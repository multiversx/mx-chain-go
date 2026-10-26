package triesHolder

import (
	"strconv"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/require"
)

func TestDataTriesHolder_GetAllRetainsPendingTries(t *testing.T) {
	dth, err := NewDataTriesHolder(dthSize)
	require.NoError(t, err)
	var expected []common.Trie
	for i := 0; i < 3; i++ {
		tr := &trieMock.TrieStub{SizeInMemoryCalled: func() int { return dthSize / 2 }}
		dth.Put([]byte(strconv.Itoa(i)), tr)
		expected = append(expected, tr)
	}
	require.Len(t, dth.evictedBuffer, 1)
	for attempt := 0; attempt < 3; attempt++ {
		require.ElementsMatch(t, expected, dth.GetAll())
		require.Len(t, dth.dirtyTries, 3)
		require.Len(t, dth.evictedBuffer, 1)
		require.Len(t, dth.touchedTries, 3)
	}
}

func TestDataTriesHolder_GetAllDoesNotEvictGrowingPendingTrie(t *testing.T) {
	dth, err := NewDataTriesHolder(dthSize)
	require.NoError(t, err)
	size := dthSize / 2
	first := &trieMock.TrieStub{SizeInMemoryCalled: func() int { return size }}
	second := &trieMock.TrieStub{SizeInMemoryCalled: func() int { return dthSize / 2 }}
	dth.Put([]byte("first"), first)
	dth.Put([]byte("second"), second)
	// The trie expands after admission. Reconciling cache weights before commit
	// must not evict an unpersisted trie without retaining it for retry.
	size = dthSize
	for attempt := 0; attempt < 3; attempt++ {
		require.ElementsMatch(t, []common.Trie{first, second}, dth.GetAll())
		require.Equal(t, uint64(dthSize), dth.cacher.SizeInBytesContained())
	}
}

func TestDataTriesHolder_MarkAllAsCommittedAcrossGrowingBatches(t *testing.T) {
	dth, err := NewDataTriesHolder(dthSize)
	require.NoError(t, err)
	for cycle := 0; cycle < 8; cycle++ {
		var pending []common.Trie
		var last common.Trie
		var lastKey []byte
		for i := 0; i < 12; i++ {
			lastKey = []byte(strconv.Itoa(cycle*12 + i))
			last = &trieMock.TrieStub{SizeInMemoryCalled: func() int { return 3 * dthSize / 8 }}
			dth.Put(lastKey, last)
			pending = append(pending, last)
			// New pending tries continue arriving while earlier snapshots have
			// not been acknowledged. Duplicate snapshots must lose no work.
			for attempt := 0; attempt < 2; attempt++ {
				require.ElementsMatch(t, pending, dth.GetAll())
			}
		}
		require.Len(t, dth.evictedBuffer, 10)
		dth.MarkAllAsCommitted()
		require.Empty(t, dth.GetAll())
		require.Empty(t, dth.dirtyTries)
		require.Empty(t, dth.evictedBuffer)
		require.Empty(t, dth.touchedTries)
		require.Same(t, last, dth.Get(lastKey), "acknowledgment must keep clean resident tries warm")
		dth.MarkAllAsCommitted()
		require.Empty(t, dth.GetAll(), "duplicate acknowledgment must not reintroduce pending work")
		require.LessOrEqual(t, dth.cacher.SizeInBytesContained(), uint64(dthSize))
	}
}

func TestDataTriesHolder_MarkAllAsCommittedUsesPostCommitSize(t *testing.T) {
	dth, err := NewDataTriesHolder(dthSize)
	require.NoError(t, err)
	size := dthSize / 2
	first := &trieMock.TrieStub{SizeInMemoryCalled: func() int { return size }}
	second := &trieMock.TrieStub{SizeInMemoryCalled: func() int { return dthSize / 2 }}
	dth.Put([]byte("first"), first)
	dth.Put([]byte("second"), second)
	require.Len(t, dth.GetAll(), 2)
	// Successful trie commits can collapse expanded nodes. Charge their size
	// after committing, so clean cache entries need not be evicted prematurely.
	size = oneKB
	dth.MarkAllAsCommitted()
	require.Empty(t, dth.GetAll())
	require.Equal(t, uint64(dthSize/2+oneKB), dth.cacher.SizeInBytesContained())
	require.Same(t, first, dth.Get([]byte("first")))
	require.Same(t, second, dth.Get([]byte("second")))
}

func TestDataTriesHolder_ResetDiscardsUnacknowledgedBatch(t *testing.T) {
	dth, err := NewDataTriesHolder(dthSize)
	require.NoError(t, err)
	for i := 0; i < 3; i++ {
		dth.Put([]byte(strconv.Itoa(i)), &trieMock.TrieStub{SizeInMemoryCalled: func() int { return dthSize / 2 }})
	}
	require.Len(t, dth.GetAll(), 3)
	require.Len(t, dth.evictedBuffer, 1)
	dth.Reset()
	require.Empty(t, dth.GetAll())
	require.Empty(t, dth.evictedBuffer)
	require.Empty(t, dth.touchedTries)
	require.Zero(t, dth.cacher.SizeInBytesContained())
}
