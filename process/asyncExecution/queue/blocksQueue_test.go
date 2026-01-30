package queue

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
)

func TestNewHeadersQueue(t *testing.T) {
	t.Parallel()

	hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
	require.NotNil(t, hq)
	require.NotNil(t, hq.headerBodyPairs)
	require.Equal(t, 0, len(hq.headerBodyPairs))
	require.False(t, hq.IsInterfaceNil())
}

func TestHeadersQueue_Add(t *testing.T) {
	t.Parallel()

	t.Run("nil header should return error", func(t *testing.T) {
		t.Parallel()
		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		err := hq.AddOrReplace(HeaderBodyPair{})
		assert.Equal(t, common.ErrNilHeaderHandler, err)
	})

	t.Run("nil body should return error", func(t *testing.T) {
		t.Parallel()
		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		err := hq.AddOrReplace(HeaderBodyPair{Header: &block.Header{Nonce: 1}})
		assert.Equal(t, data.ErrNilBlockBody, err)
	})

	t.Run("valid header should be added", func(t *testing.T) {
		t.Parallel()
		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		header := &block.Header{Nonce: 1}
		err := hq.AddOrReplace(HeaderBodyPair{Header: header, Body: &block.Body{}})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(hq.headerBodyPairs))
	})

	t.Run("add headers with same nonce", func(t *testing.T) {
		t.Parallel()
		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		header := &block.Header{Nonce: 1, Round: 1}
		err := hq.AddOrReplace(HeaderBodyPair{Header: header, Body: &block.Body{}})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(hq.headerBodyPairs))
		assert.Equal(t, uint64(1), hq.headerBodyPairs[0].Header.GetRound())

		header = &block.Header{Nonce: 1, Round: 2}
		err = hq.AddOrReplace(HeaderBodyPair{Header: header, Body: &block.Body{}})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(hq.headerBodyPairs))
		assert.Equal(t, uint64(2), hq.headerBodyPairs[0].Header.GetRound())
	})
}

func TestHeadersQueue_Pop(t *testing.T) {
	t.Parallel()

	t.Run("pop should be blocking", func(t *testing.T) {
		t.Parallel()
		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})

		go func() {
			time.Sleep(1 * time.Second)
			hq.Close()
		}()

		_, shouldContinue := hq.Pop()
		assert.False(t, shouldContinue)

	})

	t.Run("should return first header and remove it from queue", func(t *testing.T) {
		t.Parallel()
		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		pair1 := HeaderBodyPair{Header: &block.Header{Nonce: 1}, Body: &block.Body{}}
		pair2 := HeaderBodyPair{Header: &block.Header{Nonce: 2}, Body: &block.Body{}}
		_ = hq.AddOrReplace(pair1)
		_ = hq.AddOrReplace(pair2)

		firstPair, shouldContinue := hq.Pop()
		assert.True(t, shouldContinue)
		assert.Equal(t, uint64(1), firstPair.Header.GetNonce())
		assert.Equal(t, 1, len(hq.headerBodyPairs))
		assert.Equal(t, uint64(2), hq.headerBodyPairs[0].Header.GetNonce())
	})
}

func TestHeadersQueue_Concurrency(t *testing.T) {
	t.Parallel()

	hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
	const numGoroutines = 10
	const headersPerGoroutine = 10

	done := make(chan struct{})

	var nonceCounter uint64
	writeMutex := &sync.Mutex{}
	for i := 0; i < numGoroutines; i++ {
		go func(gid int) {
			for j := 0; j < headersPerGoroutine; j++ {
				writeMutex.Lock()
				h := &block.Header{Nonce: nonceCounter}
				pair := HeaderBodyPair{Header: h, Body: &block.Body{}}

				err := hq.AddOrReplace(pair)
				require.Nil(t, err)
				nonceCounter++
				writeMutex.Unlock()
			}
			done <- struct{}{}
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	totalHeaders := numGoroutines * headersPerGoroutine

	results := make(chan uint64, totalHeaders)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for {
				pair, shouldContinue := hq.Pop()
				valuesOk := !check.IfNil(pair.Header) && !check.IfNil(pair.Body)
				if !shouldContinue || !valuesOk {
					return
				}
				hdr := pair.Header.(*block.Header)
				results <- hdr.Nonce
			}
		}()
	}

	nonces := make(map[uint64]struct{})
	for i := 0; i < totalHeaders; i++ {
		n := <-results
		nonces[n] = struct{}{}
	}
	assert.Equal(t, totalHeaders, len(nonces))

	go func() {
		time.Sleep(time.Second)
		hq.Close()
	}()

	// pop will return false and empty a pair after close
	res, shouldContinue := hq.Pop()
	require.Nil(t, res.Header)
	require.Nil(t, res.Body)
	require.False(t, shouldContinue)

	pair := HeaderBodyPair{Header: &block.Header{}, Body: &block.Body{}}
	err := hq.AddOrReplace(pair)
	require.Nil(t, err)

	res, shouldContinue = hq.Pop()
	require.Nil(t, res.Header)
	require.Nil(t, res.Body)
	require.False(t, shouldContinue)

}

func TestBlocksQueue_ConcurrentAddPopValidate(t *testing.T) {
	t.Parallel()

	hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})

	const numGoroutines = 3
	const operationsPerGoroutine = 10

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	wg.Add(numGoroutines * 3) // 3 types of operations

	// Goroutines for AddOrReplace operations
	for i := 0; i < numGoroutines; i++ {
		go func(gid int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					nonce := uint64(gid*operationsPerGoroutine + j)
					pair := HeaderBodyPair{
						Header: &block.Header{Nonce: nonce},
						Body:   &block.Body{},
					}
					_ = hq.AddOrReplace(pair) // Ignore errors for concurrency test
					time.Sleep(time.Millisecond)
				}
			}
		}(i)
	}

	// Goroutines for Pop operations with timeout protection
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine/2; j++ { // Fewer pops than adds
				select {
				case <-ctx.Done():
					return
				default:
					// Use Peek to check if queue has items before attempting Pop
					if _, ok := hq.Peek(); ok {
						_, _ = hq.Pop()
					}
					time.Sleep(time.Millisecond)
				}
			}
		}()
	}

	// Goroutines for ValidateQueueIntegrity operations
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					_ = hq.ValidateQueueIntegrity() // Ignore validation errors for concurrency test
					time.Sleep(time.Millisecond)
				}
			}
		}()
	}

	// Wait for all goroutines to complete
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait with timeout to prevent hanging
	select {
	case <-done:
		// Test passed - no deadlocks or panics
		assert.True(t, true, "Concurrent operations completed without deadlock")
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out - possible deadlock in concurrent operations")
	}

	// Clean shutdown
	hq.Close()
}

func TestBlocksQueue_ValidateQueueIntegrity(t *testing.T) {
	t.Parallel()

	t.Run("empty queue should be valid", func(t *testing.T) {
		t.Parallel()
		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		defer hq.Close()

		err := hq.ValidateQueueIntegrity()
		require.NoError(t, err)
	})

	t.Run("single item queue should be valid", func(t *testing.T) {
		t.Parallel()
		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		defer hq.Close()

		pair := HeaderBodyPair{
			Header: &block.Header{Nonce: 5},
			Body:   &block.Body{},
		}
		err := hq.AddOrReplace(pair)
		require.NoError(t, err)

		err = hq.ValidateQueueIntegrity()
		require.NoError(t, err)
	})

	t.Run("sequential nonces should be valid", func(t *testing.T) {
		t.Parallel()
		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		defer hq.Close()

		// Add sequential nonces: 1, 2, 3, 4, 5
		for i := 1; i <= 5; i++ {
			pair := HeaderBodyPair{
				Header: &block.Header{Nonce: uint64(i)},
				Body:   &block.Body{},
			}
			err := hq.AddOrReplace(pair)
			require.NoError(t, err)
		}

		err := hq.ValidateQueueIntegrity()
		require.NoError(t, err)
	})

	t.Run("non-sequential nonces should return error", func(t *testing.T) {
		t.Parallel()
		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		defer hq.Close()

		// Add nonce 1
		pair1 := HeaderBodyPair{
			Header: &block.Header{Nonce: 1},
			Body:   &block.Body{},
		}
		err := hq.AddOrReplace(pair1)
		require.NoError(t, err)

		// Manually insert nonce 3 (skipping 2) to create invalid state
		hq.mutex.Lock()
		pair3 := HeaderBodyPair{
			Header: &block.Header{Nonce: 3},
			Body:   &block.Body{},
		}
		hq.headerBodyPairs = append(hq.headerBodyPairs, pair3)
		hq.mutex.Unlock()

		err = hq.ValidateQueueIntegrity()
		require.Error(t, err)
		require.ErrorIs(t, err, ErrQueueIntegrityViolation)
	})

	t.Run("lastAddedNonce mismatch should return error", func(t *testing.T) {
		t.Parallel()
		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		defer hq.Close()

		// Add some items normally
		for i := 1; i <= 3; i++ {
			pair := HeaderBodyPair{
				Header: &block.Header{Nonce: uint64(i)},
				Body:   &block.Body{},
			}
			err := hq.AddOrReplace(pair)
			require.NoError(t, err)
		}

		// Manually corrupt lastAddedNonce
		hq.mutex.Lock()
		hq.lastAddedNonce = 5 // Should be 3
		hq.mutex.Unlock()

		err := hq.ValidateQueueIntegrity()
		require.Error(t, err)
		require.ErrorIs(t, err, ErrQueueIntegrityViolation)
	})

	t.Run("validation after pop operations", func(t *testing.T) {
		t.Parallel()
		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		defer hq.Close()

		// Add sequential nonces: 1, 2, 3, 4
		for i := 1; i <= 4; i++ {
			pair := HeaderBodyPair{
				Header: &block.Header{Nonce: uint64(i)},
				Body:   &block.Body{},
			}
			err := hq.AddOrReplace(pair)
			require.NoError(t, err)
		}

		// Pop first two items
		_, ok := hq.Pop()
		require.True(t, ok)
		_, ok = hq.Pop()
		require.True(t, ok)

		// Remaining queue should still be valid (nonces 3, 4)
		err := hq.ValidateQueueIntegrity()
		require.NoError(t, err)
	})

	t.Run("validation after replace operations", func(t *testing.T) {
		t.Parallel()
		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		defer hq.Close()

		// Add sequential nonces: 1, 2, 3, 4, 5
		for i := 1; i <= 5; i++ {
			pair := HeaderBodyPair{
				Header: &block.Header{Nonce: uint64(i)},
				Body:   &block.Body{},
			}
			err := hq.AddOrReplace(pair)
			require.NoError(t, err)
		}

		// Replace at nonce 3 (should remove 3, 4, 5 and add new 3)
		pair3New := HeaderBodyPair{
			Header: &block.Header{Nonce: 3, Round: 999}, // Different round to distinguish
			Body:   &block.Body{},
		}
		err := hq.AddOrReplace(pair3New)
		require.NoError(t, err)

		// Queue should be valid with nonces 1, 2, 3
		err = hq.ValidateQueueIntegrity()
		require.NoError(t, err)

		// Verify the queue actually contains the right items
		hq.mutex.Lock()
		assert.Equal(t, 3, len(hq.headerBodyPairs))
		assert.Equal(t, uint64(1), hq.headerBodyPairs[0].Header.GetNonce())
		assert.Equal(t, uint64(2), hq.headerBodyPairs[1].Header.GetNonce())
		assert.Equal(t, uint64(3), hq.headerBodyPairs[2].Header.GetNonce())
		assert.Equal(t, uint64(999), hq.headerBodyPairs[2].Header.GetRound()) // Verify it's the new one
		hq.mutex.Unlock()
	})

	t.Run("validation on closed queue", func(t *testing.T) {
		t.Parallel()
		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})

		// Add some items
		for i := 1; i <= 3; i++ {
			pair := HeaderBodyPair{
				Header: &block.Header{Nonce: uint64(i)},
				Body:   &block.Body{},
			}
			err := hq.AddOrReplace(pair)
			require.NoError(t, err)
		}

		// Close the queue
		hq.Close()

		// Validation should still work on closed queue
		err := hq.ValidateQueueIntegrity()
		require.NoError(t, err)
	})

	t.Run("validation with RemoveAtNonceAndHigher", func(t *testing.T) {
		t.Parallel()
		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		defer hq.Close()

		// Add sequential nonces: 1, 2, 3, 4, 5
		for i := 1; i <= 5; i++ {
			pair := HeaderBodyPair{
				Header: &block.Header{Nonce: uint64(i)},
				Body:   &block.Body{},
			}
			err := hq.AddOrReplace(pair)
			require.NoError(t, err)
		}

		// Remove from nonce 3 and higher
		removedNonces := hq.RemoveAtNonceAndHigher(3)
		expectedRemoved := []uint64{3, 4, 5}
		assert.Equal(t, expectedRemoved, removedNonces)

		// Queue should be valid with remaining nonces 1, 2
		err := hq.ValidateQueueIntegrity()
		require.NoError(t, err)
	})
}

func TestMultipleAddOrReplaceShouldNotBlock(t *testing.T) {
	t.Parallel()

	hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})

	for i := 0; i < 10; i++ {
		pair := HeaderBodyPair{Header: &block.Header{Nonce: uint64(i)}, Body: &block.Body{}}
		err := hq.AddOrReplace(pair)
		require.Nil(t, err)
	}

	res, shouldContinue := hq.Pop()
	require.True(t, shouldContinue)
	require.Equal(t, uint64(0), res.Header.GetNonce())
}

func TestAddWrongNonce(t *testing.T) {
	t.Parallel()

	hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
	pair := HeaderBodyPair{Header: &block.Header{Nonce: uint64(1)}, Body: &block.Body{}}
	err := hq.AddOrReplace(pair)
	require.Nil(t, err)

	pair = HeaderBodyPair{Header: &block.Header{Nonce: uint64(3)}, Body: &block.Body{}}
	err = hq.AddOrReplace(pair)
	require.True(t, errors.Is(err, ErrHeaderNonceMismatch))
}

func TestBlocksQueue_Peak(t *testing.T) {
	t.Parallel()

	t.Run("peak should return first element", func(t *testing.T) {
		t.Parallel()

		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		pair1 := HeaderBodyPair{Header: &block.Header{Nonce: uint64(1)}, Body: &block.Body{}}
		err := hq.AddOrReplace(pair1)
		require.Nil(t, err)

		pair2 := HeaderBodyPair{Header: &block.Header{Nonce: uint64(2)}, Body: &block.Body{}}
		err = hq.AddOrReplace(pair2)
		require.Nil(t, err)

		res, ok := hq.Peek()
		require.True(t, ok)
		require.Equal(t, pair1, res)
		require.Equal(t, 2, len(hq.headerBodyPairs))
	})

	t.Run("peak emtpy queue", func(t *testing.T) {
		t.Parallel()

		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})

		_, ok := hq.Peek()
		require.False(t, ok)
	})

}

func TestBlocksQueue_AddOrReplaceWithLowerNonce(t *testing.T) {
	t.Parallel()

	t.Run("replace at middle nonce and remove all higher nonces", func(t *testing.T) {
		t.Parallel()

		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		// Add blocks with nonces 1, 2, 3, 4, 5
		for i := uint64(1); i <= 5; i++ {
			pair := HeaderBodyPair{
				Header: &block.Header{Nonce: i},
				Body:   &block.Body{},
			}
			err := hq.AddOrReplace(pair)
			require.Nil(t, err)
		}

		require.Equal(t, 5, len(hq.headerBodyPairs))
		require.Equal(t, uint64(5), hq.lastAddedNonce)

		// Replace at nonce 3 with a different round
		pairAtNonce3 := HeaderBodyPair{
			Header: &block.Header{Nonce: 3, Round: 100},
			Body:   &block.Body{},
		}
		err := hq.AddOrReplace(pairAtNonce3)
		require.Nil(t, err)

		// Should have only 3 elements now (nonces 1, 2, 3)
		require.Equal(t, 3, len(hq.headerBodyPairs))
		require.Equal(t, uint64(3), hq.lastAddedNonce)

		// Verify the replacement happened
		require.Equal(t, uint64(100), hq.headerBodyPairs[2].Header.GetRound())

		// Verify the order of remaining elements
		require.Equal(t, uint64(1), hq.headerBodyPairs[0].Header.GetNonce())
		require.Equal(t, uint64(2), hq.headerBodyPairs[1].Header.GetNonce())
		require.Equal(t, uint64(3), hq.headerBodyPairs[2].Header.GetNonce())
	})

	t.Run("replace at first nonce and remove all higher nonces", func(t *testing.T) {
		t.Parallel()

		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		// Add blocks with nonces 1, 2, 3, 4
		for i := uint64(1); i <= 4; i++ {
			pair := HeaderBodyPair{
				Header: &block.Header{Nonce: i, Round: i},
				Body:   &block.Body{},
			}
			err := hq.AddOrReplace(pair)
			require.Nil(t, err)
		}

		require.Equal(t, 4, len(hq.headerBodyPairs))
		require.Equal(t, uint64(4), hq.lastAddedNonce)

		// Replace at nonce 1
		pairAtNonce1 := HeaderBodyPair{
			Header: &block.Header{Nonce: 1, Round: 200},
			Body:   &block.Body{},
		}
		err := hq.AddOrReplace(pairAtNonce1)
		require.Nil(t, err)

		// Should have only 1 element now
		require.Equal(t, 1, len(hq.headerBodyPairs))
		require.Equal(t, uint64(1), hq.lastAddedNonce)
		require.Equal(t, uint64(200), hq.headerBodyPairs[0].Header.GetRound())
	})

	t.Run("replace with nonce lower than first element should remove all higher nonces", func(t *testing.T) {
		t.Parallel()

		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		hq.SetLastAddedNonce(10)

		// Add blocks with nonces 11, 12, 13
		for i := uint64(11); i <= 13; i++ {
			pair := HeaderBodyPair{
				Header: &block.Header{Nonce: i, Round: i},
				Body:   &block.Body{},
			}
			err := hq.AddOrReplace(pair)
			require.Nil(t, err)
		}

		require.Equal(t, 3, len(hq.headerBodyPairs))
		require.Equal(t, uint64(13), hq.lastAddedNonce)

		// Try to replace at nonce 5 (which is lower than first element nonce 11)
		pairAtNonce5 := HeaderBodyPair{
			Header: &block.Header{Nonce: 5, Round: 500},
			Body:   &block.Body{},
		}
		err := hq.AddOrReplace(pairAtNonce5)
		require.NoError(t, err)

		// Queue should remain unchanged
		require.Equal(t, 1, len(hq.headerBodyPairs))
		require.Equal(t, uint64(5), hq.lastAddedNonce)
	})
}

func TestBlocksQueue_Close(t *testing.T) {
	t.Parallel()

	hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
	hq.Close()
	hq.Close() // for coverage, should already be closed
}

func TestBlocksQueue_Clean(t *testing.T) {
	t.Parallel()

	hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
	// Add blocks with nonces 2, 3, 4, 5
	for i := uint64(2); i <= 5; i++ {
		pair := HeaderBodyPair{
			Header: &block.Header{Nonce: i, Round: i},
			Body:   &block.Body{},
		}
		err := hq.AddOrReplace(pair)
		require.Nil(t, err)
	}

	require.Equal(t, 4, len(hq.headerBodyPairs))
	require.Equal(t, uint64(5), hq.lastAddedNonce)

	hq.Clean(1)
	require.Equal(t, 0, len(hq.headerBodyPairs))
	require.Equal(t, uint64(1), hq.lastAddedNonce)
}

func TestBlocksQueue_RemoveAtNonceAndHigher(t *testing.T) {
	t.Parallel()

	t.Run("remove from empty queue should return nil", func(t *testing.T) {
		t.Parallel()

		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		removedNonces := hq.RemoveAtNonceAndHigher(5)
		require.Zero(t, len(removedNonces))
		require.Equal(t, 0, len(hq.headerBodyPairs))
	})

	t.Run("remove from middle nonce removes that nonce and all higher", func(t *testing.T) {
		t.Parallel()

		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})

		// Add blocks with nonces 1, 2, 3, 4, 5
		for i := uint64(1); i <= 5; i++ {
			pair := HeaderBodyPair{
				Header: &block.Header{Nonce: i, Round: i},
				Body:   &block.Body{},
			}
			err := hq.AddOrReplace(pair)
			require.Nil(t, err)
		}

		require.Equal(t, 5, len(hq.headerBodyPairs))
		require.Equal(t, uint64(5), hq.lastAddedNonce)

		// Remove from nonce 3 onwards
		removedNonces := hq.RemoveAtNonceAndHigher(3)
		require.Equal(t, 3, len(removedNonces))
		require.Equal(t, uint64(3), removedNonces[0])
		require.Equal(t, uint64(4), removedNonces[1])
		require.Equal(t, uint64(5), removedNonces[2])

		// Should have only 2 elements now (nonces 1, 2)
		require.Equal(t, 2, len(hq.headerBodyPairs))
		require.Equal(t, uint64(2), hq.lastAddedNonce)

		// Verify remaining elements
		require.Equal(t, uint64(1), hq.headerBodyPairs[0].Header.GetNonce())
		require.Equal(t, uint64(2), hq.headerBodyPairs[1].Header.GetNonce())
	})

	t.Run("remove from first nonce clears entire queue", func(t *testing.T) {
		t.Parallel()

		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		// Add blocks with nonces 1, 2, 3, 4
		for i := uint64(1); i <= 4; i++ {
			pair := HeaderBodyPair{
				Header: &block.Header{Nonce: i, Round: i},
				Body:   &block.Body{},
			}
			err := hq.AddOrReplace(pair)
			require.Nil(t, err)
		}

		require.Equal(t, 4, len(hq.headerBodyPairs))
		require.Equal(t, uint64(4), hq.lastAddedNonce)

		// Remove all
		removedNonces := hq.RemoveAtNonceAndHigher(1)
		require.Equal(t, 4, len(removedNonces))
		for i := uint64(1); i <= 4; i++ {
			require.Equal(t, i, removedNonces[i-1])
		}

		// Queue should be empty
		require.Equal(t, 0, len(hq.headerBodyPairs))
		require.Equal(t, uint64(0), hq.lastAddedNonce)
	})

	t.Run("remove from last nonce removes only last element", func(t *testing.T) {
		t.Parallel()

		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		// Add blocks with nonces 1, 2, 3, 4
		for i := uint64(1); i <= 4; i++ {
			pair := HeaderBodyPair{
				Header: &block.Header{Nonce: i, Round: i},
				Body:   &block.Body{},
			}
			err := hq.AddOrReplace(pair)
			require.Nil(t, err)
		}

		require.Equal(t, 4, len(hq.headerBodyPairs))
		require.Equal(t, uint64(4), hq.lastAddedNonce)

		// Remove from nonce 4 (last element)
		removedNonces := hq.RemoveAtNonceAndHigher(4)
		require.Equal(t, 1, len(removedNonces))
		require.Equal(t, uint64(4), removedNonces[0])

		// Should have 3 elements now
		require.Equal(t, 3, len(hq.headerBodyPairs))
		require.Equal(t, uint64(3), hq.lastAddedNonce)

		// Verify remaining elements
		require.Equal(t, uint64(1), hq.headerBodyPairs[0].Header.GetNonce())
		require.Equal(t, uint64(2), hq.headerBodyPairs[1].Header.GetNonce())
		require.Equal(t, uint64(3), hq.headerBodyPairs[2].Header.GetNonce())
	})

	t.Run("remove non-existent nonce removes higher ones", func(t *testing.T) {
		t.Parallel()

		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		hq.SetLastAddedNonce(10)

		// Add blocks with nonces 11, 12, 13
		for i := uint64(11); i <= 13; i++ {
			pair := HeaderBodyPair{
				Header: &block.Header{Nonce: i, Round: i},
				Body:   &block.Body{},
			}
			err := hq.AddOrReplace(pair)
			require.Nil(t, err)
		}

		require.Equal(t, 3, len(hq.headerBodyPairs))
		require.Equal(t, uint64(13), hq.lastAddedNonce)

		// Try to remove at nonce 5 (which doesn't exist)
		removedNonces := hq.RemoveAtNonceAndHigher(5)
		require.Equal(t, 3, len(removedNonces))

		// Queue should be empty
		require.Equal(t, 0, len(hq.headerBodyPairs))
		require.Equal(t, uint64(4), hq.lastAddedNonce)
	})

	t.Run("remove from first nonce with nonce 0", func(t *testing.T) {
		t.Parallel()

		hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
		// Add block with nonce 0
		pair := HeaderBodyPair{
			Header: &block.Header{Nonce: 0, Round: 1},
			Body:   &block.Body{},
		}
		err := hq.AddOrReplace(pair)
		require.Nil(t, err)

		require.Equal(t, 1, len(hq.headerBodyPairs))
		require.Equal(t, uint64(0), hq.lastAddedNonce)

		// Remove from nonce 0
		removedNonces := hq.RemoveAtNonceAndHigher(0)
		require.Equal(t, 1, len(removedNonces))

		// Queue should be empty, lastAddedNonce should be 0
		require.Equal(t, 0, len(hq.headerBodyPairs))
		require.Equal(t, uint64(0), hq.lastAddedNonce)

		hq.Close()
		hq.RemoveAtNonceAndHigher(10) // coverage only, should early exit
	})
}

func TestBlocksQueue_AddAndPop(t *testing.T) {
	t.Parallel()

	hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})

	pair := HeaderBodyPair{
		Header: &block.Header{Nonce: 0, Round: 1},
		Body:   &block.Body{},
	}

	err := hq.AddOrReplace(pair)
	require.NoError(t, err)

	_, ok := hq.Pop()
	require.True(t, ok)

	done := make(chan struct{})

	go func() {
		defer close(done)
		_, okP := hq.Pop()
		require.True(t, okP)
	}()

	select {
	case <-done:
		t.Fatalf("expected hq.Pop() to block, but it returned")
	case <-time.After(1 * time.Second):
		t.Log("expected hq.Pop() to block, success")
	}
}

func TestBlocksQueue_RemoveAndPop(t *testing.T) {
	t.Parallel()

	hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
	defer hq.Close()

	pair := HeaderBodyPair{
		Header: &block.Header{Nonce: 0, Round: 1},
		Body:   &block.Body{},
	}

	go func() {
		// wait a bit so hq.Pop call blocks the channel
		time.Sleep(time.Millisecond * 100)

		// add a pair and remove it immediately
		err := hq.AddOrReplace(pair)
		require.NoError(t, err)

		hq.RemoveAtNonceAndHigher(pair.Header.GetNonce())
	}()

	// wait blocking, should return ok with nil header and body
	// using TestingPop to ensure that RemoveAtNonceAndHigher acquires the mutex before pop operation
	poppedPair, ok := hq.TestingPop(time.Millisecond * 20)
	require.True(t, ok)
	require.Nil(t, poppedPair.Header)
	require.Nil(t, poppedPair.Body)
}

func TestBlocksQueue_AddOrReplace_Rejection(t *testing.T) {
	t.Parallel()

	hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
	// Override maxQueueSize for testing
	hq.maxQueueSize = 5

	// Fill queue to max capacity
	for i := uint64(1); i <= 5; i++ {
		pair := HeaderBodyPair{
			Header: &block.Header{Nonce: i},
			Body:   &block.Body{},
		}
		err := hq.AddOrReplace(pair)
		require.Nil(t, err)
	}

	// Try to add one more
	pair := HeaderBodyPair{
		Header: &block.Header{Nonce: 6},
		Body:   &block.Body{},
	}
	err := hq.AddOrReplace(pair)
	require.Equal(t, ErrQueueFull, err)
	require.Equal(t, 5, len(hq.headerBodyPairs))
}

func TestBlocksQueue_AddOrReplace_ReplacementAllowed(t *testing.T) {
	t.Parallel()

	hq := NewBlocksQueue(config.HeaderBodyCacheConfig{})
	// Override maxQueueSize for testing
	hq.maxQueueSize = 5

	// Fill queue to max capacity
	for i := uint64(1); i <= 5; i++ {
		pair := HeaderBodyPair{
			Header: &block.Header{Nonce: i},
			Body:   &block.Body{},
		}
		err := hq.AddOrReplace(pair)
		require.Nil(t, err)
	}

	// Replace existing item (Nonce 3)
	pair := HeaderBodyPair{
		Header: &block.Header{Nonce: 3, Round: 100},
		Body:   &block.Body{},
	}
	err := hq.AddOrReplace(pair)
	require.Nil(t, err)
	require.Equal(t, 3, len(hq.headerBodyPairs)) // 1, 2, 3 (4 and 5 removed by replacement logic)
	require.Equal(t, uint64(100), hq.headerBodyPairs[2].Header.GetRound())
}
