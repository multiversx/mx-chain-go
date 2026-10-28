package triesHolder

import (
	"fmt"
	"math"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/require"
)

func TestNewReadOnlyDataTriesHolder(t *testing.T) {
	t.Parallel()

	for _, size := range []uint64{0, maxTrieSizeMinValue - 1} {
		holder, err := NewReadOnlyDataTriesHolder(size)
		require.Nil(t, holder)
		require.ErrorIs(t, err, ErrInvalidMaxTrieSizeValue)
	}
	holder, err := NewReadOnlyDataTriesHolder(math.MaxUint64)
	require.Nil(t, holder)
	require.Error(t, err, "an overflowing byte capacity must not construct a cache")
	holder, err = NewReadOnlyDataTriesHolder(maxTrieSizeMinValue)
	require.NoError(t, err)
	require.False(t, check.IfNil(holder))
	var nilHolder *readOnlyDataTriesHolder
	require.True(t, check.IfNil(nilHolder))
}

func TestReadOnlyDataTriesHolder_InvalidEntriesDoNotDisturbCache(t *testing.T) {
	t.Parallel()

	holder, err := NewReadOnlyDataTriesHolder(maxTrieSizeMinValue)
	require.NoError(t, err)
	resident := &trieMock.TrieStub{SizeInMemoryCalled: func() int { return maxTrieSizeMinValue }}
	holder.Put([]byte("account"), resident)
	var nilTrie *trieMock.TrieStub
	for _, invalid := range []common.Trie{nil, nilTrie} {
		holder.Put([]byte("account"), invalid)
	}
	holder.Put(nil, resident)
	holder.Put([]byte{}, resident)
	require.Same(t, resident, holder.Get([]byte("account")))
	require.Nil(t, holder.Get(nil))
	require.Nil(t, holder.Get([]byte("missing")))
	require.Equal(t, 1, holder.cacher.Len())
}

func TestReadOnlyDataTriesHolder_ReleasesEvictedEntriesWithoutCommit(t *testing.T) {
	t.Parallel()

	holder, err := NewReadOnlyDataTriesHolder(maxTrieSizeMinValue)
	require.NoError(t, err)
	for cycle := 0; cycle < 3; cycle++ {
		for i := 0; i < 32; i++ {
			key := []byte(fmt.Sprintf("account-%d", i))
			tr := &trieMock.TrieStub{SizeInMemoryCalled: func() int { return 600 << 10 }}
			holder.Put(key, tr)
			require.Same(t, tr, holder.Get(key))
			if i > 0 {
				require.Nil(t, holder.Get([]byte(fmt.Sprintf("account-%d", i-1))))
			}
			require.Empty(t, holder.GetAll(), "persisted reads must never become pending commits")
			require.Equal(t, 1, holder.cacher.Len())
			require.Equal(t, uint64(600<<10), holder.cacher.SizeInBytesContained())
		}
		holder.Reset()
		holder.Reset()
		require.Nil(t, holder.Get([]byte("account-31")))
		require.Zero(t, holder.cacher.SizeInBytesContained())
		require.Zero(t, holder.cacher.Len())
	}
}

func TestReadOnlyDataTriesHolder_RespectsAccessOrderAndReplacementSizes(t *testing.T) {
	t.Parallel()

	holder, err := NewReadOnlyDataTriesHolder(maxTrieSizeMinValue)
	require.NoError(t, err)
	first, second, third := &trieMock.TrieStub{}, &trieMock.TrieStub{}, &trieMock.TrieStub{}
	for _, tr := range []*trieMock.TrieStub{first, second, third} {
		tr.SizeInMemoryCalled = func() int { return maxTrieSizeMinValue / 2 }
	}
	holder.Put([]byte("first"), first)
	holder.Put([]byte("second"), second)
	require.Same(t, first, holder.Get([]byte("first")))
	holder.Put([]byte("third"), third)
	require.Nil(t, holder.Get([]byte("second")))
	require.Same(t, third, holder.Get([]byte("third")))

	replacement := &trieMock.TrieStub{SizeInMemoryCalled: func() int { return maxTrieSizeMinValue }}
	holder.Put([]byte("first"), replacement)
	require.Same(t, replacement, holder.Get([]byte("first")))
	require.Nil(t, holder.Get([]byte("third")))
	require.Equal(t, uint64(maxTrieSizeMinValue), holder.cacher.SizeInBytesContained())
	holder.MarkAsDirty([]byte("first"))
	holder.MarkAllAsCommitted()
	holder.MarkAllAsCommitted()
	require.Empty(t, holder.GetAll())
	require.Same(t, replacement, holder.Get([]byte("first")))
	holder.Remove([]byte("first"))
	holder.Remove([]byte("first"))
	require.Nil(t, holder.Get([]byte("first")))
	require.Zero(t, holder.cacher.SizeInBytesContained())
}

func TestReadOnlyDataTriesHolder_KeepsSingleOversizedEntryLikeCapacityLRU(t *testing.T) {
	t.Parallel()

	holder, err := NewReadOnlyDataTriesHolder(maxTrieSizeMinValue)
	require.NoError(t, err)
	oversized := &trieMock.TrieStub{SizeInMemoryCalled: func() int { return 2 * maxTrieSizeMinValue }}
	holder.Put([]byte("large"), oversized)
	require.Same(t, oversized, holder.Get([]byte("large")))
	small := &trieMock.TrieStub{SizeInMemoryCalled: func() int { return 1 }}
	holder.Put([]byte("small"), small)
	require.Nil(t, holder.Get([]byte("large")))
	require.Same(t, small, holder.Get([]byte("small")))
	require.Empty(t, holder.GetAll())
}

func TestReadOnlyDataTriesHolder_ConcurrentReadsEvictionsAndResets(t *testing.T) {
	t.Parallel()

	holder, err := NewReadOnlyDataTriesHolder(maxTrieSizeMinValue)
	require.NoError(t, err)
	tr := &trieMock.TrieStub{SizeInMemoryCalled: func() int { return 64 << 10 }}
	var workers sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for i := 0; i < 100; i++ {
				key := []byte(fmt.Sprintf("worker-%d/account-%d", worker, i))
				holder.Put(key, tr)
				holder.Get(key)
				holder.Remove(key)
				if i%10 == 0 {
					holder.Reset()
				}
			}
		}()
	}
	workers.Wait()
	require.Empty(t, holder.GetAll())
	require.LessOrEqual(t, holder.cacher.SizeInBytesContained(), uint64(maxTrieSizeMinValue))
}
