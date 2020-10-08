package core

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEmptyChannelShouldWorkOnBufferedChannel(t *testing.T) {
	ch := make(chan bool, 10)

	assert.Equal(t, 0, len(ch))
	readsCnt := EmptyChannel(ch)
	assert.Equal(t, 0, len(ch))
	assert.Equal(t, 0, readsCnt)

	ch <- true
	ch <- true
	ch <- true

	assert.Equal(t, 3, len(ch))
	readsCnt = EmptyChannel(ch)
	assert.Equal(t, 0, len(ch))
	assert.Equal(t, 3, readsCnt)
}

func TestEmptyChannelShouldWorkOnNotBufferedChannel(t *testing.T) {
	ch := make(chan bool)

	assert.Equal(t, 0, len(ch))
	readsCnt := int32(EmptyChannel(ch))
	assert.Equal(t, 0, len(ch))
	assert.Equal(t, int32(0), readsCnt)

	wg := sync.WaitGroup{}
	wgChanWasWritten := sync.WaitGroup{}
	numConcurrentWrites := 50
	wg.Add(numConcurrentWrites)
	wgChanWasWritten.Add(numConcurrentWrites)
	for i := 0; i < numConcurrentWrites; i++ {
		go func() {
			wg.Done()
			time.Sleep(time.Millisecond)
			ch <- true
			wgChanWasWritten.Done()
		}()
	}

	// wait for go routines to start
	wg.Wait()

	go func() {
		for readsCnt < int32(numConcurrentWrites) {
			atomic.AddInt32(&readsCnt, int32(EmptyChannel(ch)))
		}
	}()

	// wait for go routines to finish
	wgChanWasWritten.Wait()

	assert.Equal(t, 0, len(ch))
	assert.Equal(t, int32(numConcurrentWrites), atomic.LoadInt32(&readsCnt))
}

func TestGetPBFTThreshold_ShouldWork(t *testing.T) {
	assert.Equal(t, 2, GetPBFTThreshold(2))
	assert.Equal(t, 3, GetPBFTThreshold(3))
	assert.Equal(t, 3, GetPBFTThreshold(4))
	assert.Equal(t, 4, GetPBFTThreshold(5))
	assert.Equal(t, 5, GetPBFTThreshold(6))
	assert.Equal(t, 5, GetPBFTThreshold(7))
}

func TestGetPBFTFallbackThreshold_ShouldWork(t *testing.T) {
	assert.Equal(t, 2, GetPBFTFallbackThreshold(2))
	assert.Equal(t, 2, GetPBFTFallbackThreshold(3))
	assert.Equal(t, 3, GetPBFTFallbackThreshold(4))
	assert.Equal(t, 3, GetPBFTFallbackThreshold(5))
	assert.Equal(t, 4, GetPBFTFallbackThreshold(6))
	assert.Equal(t, 4, GetPBFTFallbackThreshold(7))
}
