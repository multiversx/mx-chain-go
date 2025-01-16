package statistics

import (
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewStateStatistics_ShouldWork(t *testing.T) {
	t.Parallel()

	ss := NewStateStatistics()

	assert.False(t, check.IfNil(ss))
}

func TestStateStatistics_Processing(t *testing.T) {
	t.Parallel()

	t.Run("trie operations", func(t *testing.T) {
		t.Parallel()

		ss := NewStateStatistics()

		assert.Equal(t, uint64(0), ss.Trie())

		ss.IncrementTrie()
		ss.IncrementTrie()
		assert.Equal(t, uint64(2), ss.Trie())

		ss.IncrementTrie()
		assert.Equal(t, uint64(3), ss.Trie())

		ss.Reset()
		assert.Equal(t, uint64(0), ss.Trie())
	})

	t.Run("persister operations", func(t *testing.T) {
		t.Parallel()

		ss := NewStateStatistics()

		epoch := uint32(1)

		assert.Equal(t, uint64(0), ss.Persister(epoch))

		ss.IncrementPersister(epoch)
		ss.IncrementPersister(epoch)
		assert.Equal(t, uint64(2), ss.Persister(epoch))

		ss.IncrementPersister(epoch)
		assert.Equal(t, uint64(3), ss.Persister(epoch))

		ss.Reset()
		assert.Equal(t, uint64(0), ss.Persister(epoch))
	})

	t.Run("cache operations", func(t *testing.T) {
		t.Parallel()

		ss := NewStateStatistics()

		assert.Equal(t, uint64(0), ss.Cache())

		ss.IncrementCache()
		ss.IncrementCache()
		assert.Equal(t, uint64(2), ss.Cache())

		ss.IncrementCache()
		assert.Equal(t, uint64(3), ss.Cache())

		ss.Reset()
		assert.Equal(t, uint64(0), ss.Cache())
	})
}

func TestStateStatistics_Snapshot(t *testing.T) {
	t.Parallel()

	t.Run("persister operations", func(t *testing.T) {
		t.Parallel()

		ss := NewStateStatistics()

		epoch := uint32(1)

		assert.Equal(t, uint64(0), ss.SnapshotPersister(epoch))

		ss.IncrementSnapshotPersister(epoch)
		ss.IncrementSnapshotPersister(epoch)
		assert.Equal(t, uint64(2), ss.SnapshotPersister(epoch))

		ss.IncrementSnapshotPersister(epoch)
		assert.Equal(t, uint64(3), ss.SnapshotPersister(epoch))

		ss.ResetSnapshot()
		assert.Equal(t, uint64(0), ss.SnapshotPersister(epoch))
	})

	t.Run("cache operations", func(t *testing.T) {
		t.Parallel()

		ss := NewStateStatistics()

		assert.Equal(t, uint64(0), ss.Cache())

		ss.IncrementSnapshotCache()
		ss.IncrementSnapshotCache()
		assert.Equal(t, uint64(2), ss.SnapshotCache())

		ss.IncrementSnapshotCache()
		assert.Equal(t, uint64(3), ss.SnapshotCache())

		ss.ResetSnapshot()
		assert.Equal(t, uint64(0), ss.SnapshotCache())
	})
}

func TestStateStatistics_ConcurrenyOperations(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	numIterations := 10000

	epoch := uint32(1)

	ss := NewStateStatistics()

	wg := sync.WaitGroup{}
	wg.Add(numIterations)

	for i := 0; i < numIterations; i++ {
		go func(idx int) {
			switch idx % 11 {
			case 0:
				ss.Reset()
			case 1:
				ss.IncrementCache()
			case 2:
				ss.IncrementPersister(epoch)
			case 3:
				ss.IncrementTrie()
			case 7:
				_ = ss.Cache()
			case 8:
				_ = ss.Persister(epoch)
			case 9:
				_ = ss.Trie()
			case 10:
				_ = ss.ProcessingStats()
			}

			wg.Done()
		}(i)
	}

	wg.Wait()
}
