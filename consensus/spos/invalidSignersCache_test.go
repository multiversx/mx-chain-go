package spos

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInvalidSignersCache_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var cache *invalidSignersCache
	require.True(t, cache.IsInterfaceNil())

	cache = NewInvalidSignersCache()
	require.False(t, cache.IsInterfaceNil())
}

func TestInvalidSignersCache(t *testing.T) {
	t.Parallel()

	t.Run("all ops should work", func(t *testing.T) {
		t.Parallel()

		cache := NewInvalidSignersCache()
		require.NotNil(t, cache)

		cache.AddInvalidSigners("") // early return, for coverage only

		hash := "hash"
		require.False(t, cache.HasInvalidSigners(hash))

		cache.AddInvalidSigners(hash)
		require.True(t, cache.HasInvalidSigners(hash))

		cache.Reset()
		require.False(t, cache.HasInvalidSigners(hash))
	})
	t.Run("concurrent ops should work", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				require.Fail(t, "should have not panicked")
			}
		}()

		cache := NewInvalidSignersCache()
		require.NotNil(t, cache)

		numCalls := 1000
		wg := sync.WaitGroup{}
		wg.Add(numCalls)

		for i := 0; i < numCalls; i++ {
			go func(idx int) {
				switch idx % 3 {
				case 0:
					cache.AddInvalidSigners(fmt.Sprintf("hash_%d", idx))
				case 1:
					cache.HasInvalidSigners(fmt.Sprintf("hash_%d", idx))
				case 2:
					cache.Reset()
				default:
					require.Fail(t, "should not happen")
				}

				wg.Done()
			}(i)
		}

		wg.Wait()
	})
}
