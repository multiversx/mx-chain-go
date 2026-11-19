package triesHolder

import (
	"errors"
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

func TestReadOnlyDataTriesHolder_SkipsOversizedEntries(t *testing.T) {
	t.Parallel()

	holder, err := NewReadOnlyDataTriesHolder(maxTrieSizeMinValue)
	require.NoError(t, err)
	oversized := &trieMock.TrieStub{SizeInMemoryCalled: func() int { return 2 * maxTrieSizeMinValue }}
	holder.Put([]byte("large"), oversized)
	require.Nil(t, holder.Get([]byte("large")))
	require.Zero(t, holder.cacher.SizeInBytesContained())

	small := &trieMock.TrieStub{SizeInMemoryCalled: func() int { return 1 }}
	holder.Put([]byte("small"), small)
	holder.Put([]byte("large"), oversized)
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

type growthReportingTrieStub struct {
	trieMock.TrieStub
	read func(key []byte) ([]byte, uint32, uint32, int, error)
}

func (tr *growthReportingTrieStub) GetWithVersionAndSizeOnGrowth(key []byte) ([]byte, uint32, uint32, int, error) {
	return tr.read(key)
}

func TestReadOnlyDataTriesHolder_ReportsGrowthOnMissAndError(t *testing.T) {
	t.Parallel()

	holder, err := NewReadOnlyDataTriesHolder(maxTrieSizeMinValue)
	require.NoError(t, err)
	key := []byte("account")
	sizeSamples := 0
	readErr := errors.New("descendant unreadable")
	tr := &growthReportingTrieStub{
		TrieStub: trieMock.TrieStub{
			SizeInMemoryCalled: func() int {
				sizeSamples++
				return 100
			},
		},
		read: func(key []byte) ([]byte, uint32, uint32, int, error) {
			switch string(key) {
			case "miss":
				return nil, 2, 0, 200, nil
			case "failure":
				return nil, 3, 0, 300, readErr
			default:
				return []byte("warm"), 3, 1, 0, nil
			}
		},
	}
	cached := holder.PutAndTrack(key, tr)
	require.Same(t, cached, holder.Get(key))
	require.EqualValues(t, 100, holder.cacher.SizeInBytesContained())

	value, depth, err := cached.Get([]byte("miss"))
	require.NoError(t, err)
	require.Nil(t, value)
	require.EqualValues(t, 2, depth)
	require.EqualValues(t, 200, holder.cacher.SizeInBytesContained())

	value, depth, version, err := cached.GetWithVersion([]byte("failure"))
	require.ErrorIs(t, err, readErr)
	require.Nil(t, value)
	require.EqualValues(t, 3, depth)
	require.Zero(t, version)
	require.EqualValues(t, 300, holder.cacher.SizeInBytesContained())

	value, depth, version, err = cached.GetWithVersion([]byte("success"))
	require.NoError(t, err)
	require.Equal(t, []byte("warm"), value)
	require.EqualValues(t, 3, depth)
	require.EqualValues(t, 1, version)
	require.EqualValues(t, 300, holder.cacher.SizeInBytesContained())
	require.Equal(t, 1, sizeSamples, "only admission should sample SizeInMemory")
}

func TestReadOnlyDataTriesHolder_DiscardsLateGrowthFromDetachedTries(t *testing.T) {
	t.Parallel()

	for _, action := range []string{"remove", "reset", "replace", "replace-oversized", "evict"} {
		t.Run(action, func(t *testing.T) {
			holder, err := NewReadOnlyDataTriesHolder(maxTrieSizeMinValue)
			require.NoError(t, err)
			key := []byte("account")
			reportedSize := 768 << 10
			if action == "replace-oversized" {
				reportedSize = maxTrieSizeMinValue + 1
			}
			readStarted, finishRead, readFinished := make(chan struct{}), make(chan struct{}), make(chan struct{})
			tr := &growthReportingTrieStub{
				TrieStub: trieMock.TrieStub{SizeInMemoryCalled: func() int { return 512 << 10 }},
				read: func(_ []byte) ([]byte, uint32, uint32, int, error) {
					close(readStarted)
					<-finishRead
					return []byte("original value"), 1, 0, reportedSize, nil
				},
			}
			original := holder.PutAndTrack(key, tr)
			var readValue []byte
			var readErr error
			go func() {
				defer close(readFinished)
				readValue, _, _, readErr = original.GetWithVersion([]byte("key"))
			}()
			<-readStarted

			replacement := &trieMock.TrieStub{SizeInMemoryCalled: func() int { return 600 << 10 }}
			switch action {
			case "remove":
				holder.Remove(key)
			case "reset":
				holder.Reset()
			case "replace", "replace-oversized":
				holder.Put(key, replacement)
			case "evict":
				holder.Put([]byte("other account"), replacement)
			}
			close(finishRead)
			<-readFinished
			require.NoError(t, readErr)
			require.Equal(t, []byte("original value"), readValue)

			replaced := action == "replace" || action == "replace-oversized"
			if replaced {
				require.Same(t, replacement, holder.Get(key))
			} else {
				require.Nil(t, holder.Get(key))
			}
			if action == "evict" {
				require.Same(t, replacement, holder.Get([]byte("other account")))
			}
			if replaced || action == "evict" {
				require.EqualValues(t, 600<<10, holder.cacher.SizeInBytesContained())
			} else {
				require.Zero(t, holder.cacher.SizeInBytesContained())
			}
		})
	}
}

func TestReadOnlyDataTriesHolder_OutOfOrderGrowthCannotReduceWeight(t *testing.T) {
	t.Parallel()

	holder, err := NewReadOnlyDataTriesHolder(maxTrieSizeMinValue)
	require.NoError(t, err)
	key := []byte("account")
	tr := &growthReportingTrieStub{
		TrieStub: trieMock.TrieStub{SizeInMemoryCalled: func() int { return 64 }},
	}
	cacheEntry := holder.PutAndTrack(key, tr).(*trackedReadOnlyTrie)
	for _, size := range []int64{100, 300, 200} {
		holder.updateSize(cacheEntry, size)
	}
	require.Same(t, cacheEntry, holder.Get(key))
	require.EqualValues(t, 300, holder.cacher.SizeInBytesContained())
}

func TestReadOnlyDataTriesHolder_GrowthRetainsWholeEntriesOnly(t *testing.T) {
	t.Parallel()

	holder, err := NewReadOnlyDataTriesHolder(maxTrieSizeMinValue)
	require.NoError(t, err)
	var entries []*trackedReadOnlyTrie
	for i := 0; i < 3; i++ {
		key := []byte(fmt.Sprint(i))
		tr := &growthReportingTrieStub{
			TrieStub: trieMock.TrieStub{SizeInMemoryCalled: func() int { return 256 << 10 }},
		}
		entries = append(entries, holder.PutAndTrack(key, tr).(*trackedReadOnlyTrie))
	}
	holder.updateSize(entries[2], 700<<10)
	require.Nil(t, holder.Get([]byte("0")))
	require.Same(t, entries[1], holder.Get([]byte("1")))
	require.Same(t, entries[2], holder.Get([]byte("2")))
	require.EqualValues(t, (256+700)<<10, holder.cacher.SizeInBytesContained())

	holder.updateSize(entries[2], maxTrieSizeMinValue+1)
	require.Nil(t, holder.Get([]byte("2")))
	require.Same(t, entries[1], holder.Get([]byte("1")))
	require.EqualValues(t, 256<<10, holder.cacher.SizeInBytesContained())
	require.Empty(t, holder.GetAll())
}
