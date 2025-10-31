package queue

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHeadersQueue(t *testing.T) {
	t.Parallel()

	hq := NewBlocksQueue()
	require.NotNil(t, hq)
	require.NotNil(t, hq.headerBodyPairs)
	require.Equal(t, 0, len(hq.headerBodyPairs))
	require.False(t, hq.IsInterfaceNil())
}

func TestHeadersQueue_Add(t *testing.T) {
	t.Parallel()

	t.Run("nil header should return error", func(t *testing.T) {
		t.Parallel()
		hq := NewBlocksQueue()
		err := hq.AddOrReplace(HeaderBodyPair{})
		assert.Equal(t, common.ErrNilHeaderHandler, err)
	})

	t.Run("nil body should return error", func(t *testing.T) {
		t.Parallel()
		hq := NewBlocksQueue()
		err := hq.AddOrReplace(HeaderBodyPair{Header: &block.Header{Nonce: 1}})
		assert.Equal(t, data.ErrNilBlockBody, err)
	})

	t.Run("valid header should be added", func(t *testing.T) {
		t.Parallel()
		hq := NewBlocksQueue()
		header := &block.Header{Nonce: 1}
		err := hq.AddOrReplace(HeaderBodyPair{Header: header, Body: &block.Body{}})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(hq.headerBodyPairs))
	})

	t.Run("add headers with same nonce", func(t *testing.T) {
		t.Parallel()
		hq := NewBlocksQueue()
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
		hq := NewBlocksQueue()

		go func() {
			time.Sleep(1 * time.Second)
			hq.Close()
		}()

		_, ok := hq.Pop()
		assert.False(t, ok)

	})

	t.Run("should return first header and remove it from queue", func(t *testing.T) {
		t.Parallel()
		hq := NewBlocksQueue()
		pair1 := HeaderBodyPair{Header: &block.Header{Nonce: 1}, Body: &block.Body{}}
		pair2 := HeaderBodyPair{Header: &block.Header{Nonce: 2}, Body: &block.Body{}}
		_ = hq.AddOrReplace(pair1)
		_ = hq.AddOrReplace(pair2)

		firstPair, ok := hq.Pop()
		assert.True(t, ok)
		assert.Equal(t, uint64(1), firstPair.Header.GetNonce())
		assert.Equal(t, 1, len(hq.headerBodyPairs))
		assert.Equal(t, uint64(2), hq.headerBodyPairs[0].Header.GetNonce())
	})
}

func TestHeadersQueue_Concurrency(t *testing.T) {
	t.Parallel()

	hq := NewBlocksQueue()
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
				pair, ok := hq.Pop()
				if !ok {
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
	res, ok := hq.Pop()
	require.Nil(t, res.Header)
	require.Nil(t, res.Body)
	require.False(t, ok)

	pair := HeaderBodyPair{Header: &block.Header{}, Body: &block.Body{}}
	err := hq.AddOrReplace(pair)
	require.Nil(t, err)

	res, ok = hq.Pop()
	require.Nil(t, res.Header)
	require.Nil(t, res.Body)
	require.False(t, ok)

}

func TestMultipleAddOrReplaceShouldNotBlock(t *testing.T) {
	t.Parallel()

	hq := NewBlocksQueue()

	for i := 0; i < 10; i++ {
		pair := HeaderBodyPair{Header: &block.Header{Nonce: uint64(i)}, Body: &block.Body{}}
		err := hq.AddOrReplace(pair)
		require.Nil(t, err)
	}

	res, ok := hq.Pop()
	require.True(t, ok)
	require.Equal(t, uint64(0), res.Header.GetNonce())
}

func TestAddWrongNonce(t *testing.T) {
	t.Parallel()

	hq := NewBlocksQueue()
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

		hq := NewBlocksQueue()
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

		hq := NewBlocksQueue()

		_, ok := hq.Peek()
		require.False(t, ok)
	})

}

func TestBlocksQueue_AddOrReplaceWithLowerNonce(t *testing.T) {
	t.Parallel()

	t.Run("replace at middle nonce and remove all higher nonces", func(t *testing.T) {
		t.Parallel()

		hq := NewBlocksQueue()
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

		hq := NewBlocksQueue()
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

		hq := NewBlocksQueue()
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

	hq := NewBlocksQueue()
	hq.Close()
	hq.Close() // for coverage, should already be closed
}

func TestBlocksQueue_Clean(t *testing.T) {
	t.Parallel()

	hq := NewBlocksQueue()
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

		hq := NewBlocksQueue()
		err := hq.RemoveAtNonceAndHigher(5)
		require.Nil(t, err)
		require.Equal(t, 0, len(hq.headerBodyPairs))
	})

	t.Run("remove from middle nonce removes that nonce and all higher", func(t *testing.T) {
		t.Parallel()

		hq := NewBlocksQueue()

		// register a mocked subscriber to check if the callback is called for each nonce
		evictedNonces := make(map[uint64]struct{})
		subscriber := &queue.BlocksQueueEvictionSubscriberMock{
			OnHeaderEvictedCalled: func(headerNonce uint64) {
				evictedNonces[headerNonce] = struct{}{}
			},
		}
		hq.RegisterEvictionSubscriber(subscriber)
		hq.RegisterEvictionSubscriber(nil) // coverage only

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
		require.Zero(t, len(evictedNonces))

		// Remove from nonce 3 onwards
		err := hq.RemoveAtNonceAndHigher(3)
		require.Nil(t, err)

		// Should have only 2 elements now (nonces 1, 2)
		require.Equal(t, 2, len(hq.headerBodyPairs))
		require.Equal(t, uint64(2), hq.lastAddedNonce)

		// Verify remaining elements
		require.Equal(t, uint64(1), hq.headerBodyPairs[0].Header.GetNonce())
		require.Equal(t, uint64(2), hq.headerBodyPairs[1].Header.GetNonce())

		// Check evicted nonces
		require.Len(t, evictedNonces, 3)
		require.Contains(t, evictedNonces, uint64(3))
		require.Contains(t, evictedNonces, uint64(4))
		require.Contains(t, evictedNonces, uint64(5))
	})

	t.Run("remove from first nonce clears entire queue", func(t *testing.T) {
		t.Parallel()

		hq := NewBlocksQueue()
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
		err := hq.RemoveAtNonceAndHigher(1)
		require.Nil(t, err)

		// Queue should be empty
		require.Equal(t, 0, len(hq.headerBodyPairs))
		require.Equal(t, uint64(0), hq.lastAddedNonce)
	})

	t.Run("remove from last nonce removes only last element", func(t *testing.T) {
		t.Parallel()

		hq := NewBlocksQueue()
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
		err := hq.RemoveAtNonceAndHigher(4)
		require.Nil(t, err)

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

		hq := NewBlocksQueue()
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
		err := hq.RemoveAtNonceAndHigher(5)
		require.NoError(t, err)

		// Queue should remain unchanged
		require.Equal(t, 0, len(hq.headerBodyPairs))
		require.Equal(t, uint64(4), hq.lastAddedNonce)
	})

	t.Run("remove from first nonce with nonce 0", func(t *testing.T) {
		t.Parallel()

		hq := NewBlocksQueue()
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
		err = hq.RemoveAtNonceAndHigher(0)
		require.Nil(t, err)

		// Queue should be empty, lastAddedNonce should be 0
		require.Equal(t, 0, len(hq.headerBodyPairs))
		require.Equal(t, uint64(0), hq.lastAddedNonce)

		hq.Close()
		err = hq.RemoveAtNonceAndHigher(10) // coverage only, should early exit
		require.Nil(t, err)
	})
}
