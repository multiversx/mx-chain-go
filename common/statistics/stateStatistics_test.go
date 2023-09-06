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

func TestStateStatistics_Operations(t *testing.T) {
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

		assert.Equal(t, uint64(0), ss.PersisterOp())

		ss.IncrPersisterOp()
		ss.IncrPersisterOp()
		assert.Equal(t, uint64(2), ss.PersisterOp())

		ss.IncrPersisterOp()
		assert.Equal(t, uint64(3), ss.PersisterOp())

		ss.Reset()
		assert.Equal(t, uint64(0), ss.PersisterOp())
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

func TestStateStatistics_ConcurrenyOperations(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	numIterations := 10000

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
				ss.IncrPersisterOp()
			case 3:
				ss.IncrTrieOp()
			case 7:
				_ = ss.CacheOp()
			case 8:
				_ = ss.PersisterOp()
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
