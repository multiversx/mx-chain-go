package queue

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHeadersQueue(t *testing.T) {
	t.Parallel()

	hq, err := NewBlocksQueue()
	require.Nil(t, err)
	require.NotNil(t, hq)
	require.NotNil(t, hq.headerBodyPairs)
	require.Equal(t, 0, len(hq.headerBodyPairs))
	require.False(t, hq.IsInterfaceNil())
}

func TestHeadersQueue_Add(t *testing.T) {
	t.Parallel()

	t.Run("nil header should return error", func(t *testing.T) {
		t.Parallel()
		hq, _ := NewBlocksQueue()
		err := hq.AddOrReplace(HeaderBodyPair{})
		assert.Equal(t, common.ErrNilHeaderHandler, err)
	})

	t.Run("nil body should return error", func(t *testing.T) {
		t.Parallel()
		hq, _ := NewBlocksQueue()
		err := hq.AddOrReplace(HeaderBodyPair{Header: &block.Header{Nonce: 1}})
		assert.Equal(t, data.ErrNilBlockBody, err)
	})

	t.Run("valid header should be added", func(t *testing.T) {
		t.Parallel()
		hq, _ := NewBlocksQueue()
		header := &block.Header{Nonce: 1}
		err := hq.AddOrReplace(HeaderBodyPair{Header: header, Body: &block.Body{}})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(hq.headerBodyPairs))
	})
}

func TestHeadersQueue_AddFirstMultiple(t *testing.T) {
	t.Parallel()

	t.Run("empty slice should not modify queue", func(t *testing.T) {
		t.Parallel()
		hq, _ := NewBlocksQueue()
		err := hq.AddFirstMultiple(nil)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(hq.headerBodyPairs))
	})

	t.Run("multiple headers should be added at beginning", func(t *testing.T) {
		t.Parallel()
		hq, _ := NewBlocksQueue()
		firstPair := HeaderBodyPair{Header: &block.Header{Nonce: 1}, Body: &block.Body{}}
		_ = hq.AddOrReplace(firstPair)

		newHeaders := []HeaderBodyPair{
			{
				Header: &block.Header{Nonce: 2},
				Body:   &block.Body{},
			},
			{
				Header: &block.Header{Nonce: 3},
				Body:   &block.Body{},
			},
		}
		err := hq.AddFirstMultiple(newHeaders)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(hq.headerBodyPairs))
		assert.Equal(t, uint64(2), hq.headerBodyPairs[0].Header.GetNonce())
		assert.Equal(t, uint64(3), hq.headerBodyPairs[1].Header.GetNonce())
		assert.Equal(t, uint64(1), hq.headerBodyPairs[2].Header.GetNonce())
	})
}

func TestHeadersQueue_Pop(t *testing.T) {
	t.Parallel()

	t.Run("pop should be blocking", func(t *testing.T) {
		t.Parallel()
		hq, _ := NewBlocksQueue()

		go func() {
			time.Sleep(1 * time.Second)
			hq.Close()
		}()

		_, ok := hq.Pop()
		assert.False(t, ok)

	})

	t.Run("should return first header and remove it from queue", func(t *testing.T) {
		t.Parallel()
		hq, _ := NewBlocksQueue()
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

	hq, _ := NewBlocksQueue()
	const numGoroutines = 10
	const headersPerGoroutine = 10

	done := make(chan struct{})

	for i := 0; i < numGoroutines; i++ {
		go func(gid int) {
			for j := 0; j < headersPerGoroutine; j++ {
				h := &block.Header{Nonce: uint64(gid*headersPerGoroutine + j)}
				pair := HeaderBodyPair{Header: h, Body: &block.Body{}}
				_ = hq.AddOrReplace(pair)
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

	// pop wil return false and empty a pair after close
	res, ok := hq.Pop()
	require.Nil(t, res.Header)
	require.Nil(t, res.Body)
	require.False(t, ok)

	pair := HeaderBodyPair{Header: &block.Header{}, Body: &block.Body{}}
	_ = hq.AddOrReplace(pair)
	_, ok = hq.Pop()
	require.Nil(t, res.Header)
	require.Nil(t, res.Body)
	require.False(t, ok)

}
