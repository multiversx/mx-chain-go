package queue

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHeadersQueue(t *testing.T) {
	t.Parallel()

	hq, err := NewHeadersQueue()
	require.Nil(t, err)
	require.NotNil(t, hq)
	require.NotNil(t, hq.headers)
	require.Equal(t, 0, len(hq.headers))
	require.False(t, hq.IsInterfaceNil())
}

func TestHeadersQueue_Add(t *testing.T) {
	t.Parallel()

	t.Run("nil header should return error", func(t *testing.T) {
		t.Parallel()
		hq, _ := NewHeadersQueue()
		err := hq.Add(nil)
		assert.Equal(t, common.ErrNilHeaderHandler, err)
	})

	t.Run("valid header should be added", func(t *testing.T) {
		t.Parallel()
		hq, _ := NewHeadersQueue()
		header := &block.Header{Nonce: 1}
		err := hq.Add(header)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(hq.headers))
	})
}

func TestHeadersQueue_AddFirstMultiple(t *testing.T) {
	t.Parallel()

	t.Run("empty slice should not modify queue", func(t *testing.T) {
		t.Parallel()
		hq, _ := NewHeadersQueue()
		err := hq.AddFirstMultiple(nil)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(hq.headers))
	})

	t.Run("multiple headers should be added at beginning", func(t *testing.T) {
		t.Parallel()
		hq, _ := NewHeadersQueue()
		existingHeader := &block.Header{Nonce: 1}
		_ = hq.Add(existingHeader)

		newHeaders := []data.HeaderHandler{
			&block.Header{Nonce: 2},
			&block.Header{Nonce: 3},
		}
		err := hq.AddFirstMultiple(newHeaders)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(hq.headers))
		assert.Equal(t, uint64(2), hq.headers[0].GetNonce())
		assert.Equal(t, uint64(3), hq.headers[1].GetNonce())
		assert.Equal(t, uint64(1), hq.headers[2].GetNonce())
	})
}

func TestHeadersQueue_TakeFirstHeaderForProcessing(t *testing.T) {
	t.Parallel()

	t.Run("empty queue should return error", func(t *testing.T) {
		t.Parallel()
		hq, _ := NewHeadersQueue()
		header, err := hq.Pop()
		assert.Nil(t, header)
		assert.Equal(t, ErrNoHeaderForProcessing, err)
	})

	t.Run("should return first header and remove it from queue", func(t *testing.T) {
		t.Parallel()
		hq, _ := NewHeadersQueue()
		header1 := &block.Header{Nonce: 1}
		header2 := &block.Header{Nonce: 2}
		_ = hq.Add(header1)
		_ = hq.Add(header2)

		firstHeader, err := hq.Pop()
		assert.Nil(t, err)
		assert.Equal(t, uint64(1), firstHeader.GetNonce())
		assert.Equal(t, 1, len(hq.headers))
		assert.Equal(t, uint64(2), hq.headers[0].GetNonce())
	})
}

func TestHeadersQueue_Concurrency(t *testing.T) {
	t.Parallel()

	hq, _ := NewHeadersQueue()
	const numGoroutines = 10
	const headersPerGoroutine = 10

	done := make(chan struct{})

	for i := 0; i < numGoroutines; i++ {
		go func(gid int) {
			for j := 0; j < headersPerGoroutine; j++ {
				h := &block.Header{Nonce: uint64(gid*headersPerGoroutine + j)}
				_ = hq.Add(h)
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
				h, err := hq.Pop()
				if err != nil {
					return
				}
				hdr := h.(*block.Header)
				results <- hdr.Nonce
			}
		}()
	}

	nonces := make(map[uint64]struct{})
	for i := 0; i < totalHeaders; i++ {
		n := <-results
		nonces[n] = struct{}{}
	}

	_, err := hq.Pop()
	assert.Equal(t, ErrNoHeaderForProcessing, err)
	assert.Equal(t, totalHeaders, len(nonces))
}
