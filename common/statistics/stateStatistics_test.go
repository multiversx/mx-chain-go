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

		assert.Equal(t, uint64(0), ss.TrieOp())

		ss.IncrTrieOp()
		ss.IncrTrieOp()
		assert.Equal(t, uint64(2), ss.TrieOp())

		ss.IncrTrieOp()
		assert.Equal(t, uint64(3), ss.TrieOp())

		ss.Reset()
		assert.Equal(t, uint64(0), ss.TrieOp())
	})

	t.Run("persister operations", func(t *testing.T) {
		t.Parallel()

		ss := NewStateStatistics()

		epoch := uint32(1)

		assert.Equal(t, uint64(0), ss.PersisterOp(epoch))

		ss.IncrPersisterOp(epoch)
		ss.IncrPersisterOp(epoch)
		assert.Equal(t, uint64(2), ss.PersisterOp(epoch))

		ss.IncrPersisterOp(epoch)
		assert.Equal(t, uint64(3), ss.PersisterOp(epoch))

		ss.Reset()
		assert.Equal(t, uint64(0), ss.PersisterOp(epoch))
	})

	t.Run("cache operations", func(t *testing.T) {
		t.Parallel()

		ss := NewStateStatistics()

		assert.Equal(t, uint64(0), ss.CacheOp())

		ss.IncrCacheOp()
		ss.IncrCacheOp()
		assert.Equal(t, uint64(2), ss.CacheOp())

		ss.IncrCacheOp()
		assert.Equal(t, uint64(3), ss.CacheOp())

		ss.Reset()
		assert.Equal(t, uint64(0), ss.CacheOp())
	})
}

func TestStateStatistics_Snapshot(t *testing.T) {
	t.Parallel()

	t.Run("persister operations", func(t *testing.T) {
		t.Parallel()

		ss := NewStateStatistics()

		epoch := uint32(1)

		assert.Equal(t, uint64(0), ss.SnapshotPersisterOp(epoch))

		ss.IncrSnapshotPersisterOp(epoch)
		ss.IncrSnapshotPersisterOp(epoch)
		assert.Equal(t, uint64(2), ss.SnapshotPersisterOp(epoch))

		ss.IncrSnapshotPersisterOp(epoch)
		assert.Equal(t, uint64(3), ss.SnapshotPersisterOp(epoch))

		ss.ResetSnapshot()
		assert.Equal(t, uint64(0), ss.SnapshotPersisterOp(epoch))
	})

	t.Run("cache operations", func(t *testing.T) {
		t.Parallel()

		ss := NewStateStatistics()

		assert.Equal(t, uint64(0), ss.CacheOp())

		ss.IncrSnapshotCacheOp()
		ss.IncrSnapshotCacheOp()
		assert.Equal(t, uint64(2), ss.SnapshotCacheOp())

		ss.IncrSnapshotCacheOp()
		assert.Equal(t, uint64(3), ss.SnapshotCacheOp())

		ss.ResetSnapshot()
		assert.Equal(t, uint64(0), ss.SnapshotCacheOp())
	})
}

func TestStateStatistics_Sync(t *testing.T) {
	t.Parallel()

	ss := NewStateStatistics()

	epoch := uint32(1)

	assert.Equal(t, uint64(0), ss.SyncCacheOp())
	assert.Equal(t, uint64(0), ss.SyncPersisterOp(epoch))

	ss.IncrPersisterOp(epoch)
	ss.IncrPersisterOp(epoch)

	ss.IncrCacheOp()
	ss.IncrCacheOp()

	ss.ResetSync()

	assert.Equal(t, uint64(2), ss.SyncPersisterOp(epoch))
	assert.Equal(t, uint64(2), ss.SyncCacheOp())

	ss.IncrPersisterOp(epoch)
	ss.IncrPersisterOp(epoch)

	ss.IncrCacheOp()
	ss.IncrCacheOp()

	cacheSync, persisterSync := ss.SyncStats()

	assert.Equal(t, uint64(2), cacheSync)
	assert.Equal(t, uint64(2), persisterSync[epoch])
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
				ss.IncrCacheOp()
			case 2:
				ss.IncrPersisterOp(epoch)
			case 3:
				ss.IncrTrieOp()
			case 7:
				_ = ss.CacheOp()
			case 8:
				_ = ss.PersisterOp(epoch)
			case 9:
				_ = ss.TrieOp()
			case 10:
				_ = ss.ToString()
			}

			wg.Done()
		}(i)
	}

	wg.Wait()
}
