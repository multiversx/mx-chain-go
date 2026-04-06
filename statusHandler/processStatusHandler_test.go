package statusHandler

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewProcessStatusHandler(t *testing.T) {
	t.Parallel()

	psh := NewProcessStatusHandler()
	assert.False(t, check.IfNil(psh))
}

func TestProcessStatusHandler_AllMethods(t *testing.T) {
	t.Parallel()

	psh := NewProcessStatusHandler()
	assert.True(t, psh.IsIdle())

	psh.SetBusy("reason 1")
	assert.False(t, psh.IsIdle())

	psh.SetIdle()
	assert.True(t, psh.IsIdle())

	psh.SetBusy("reason 2")
	assert.False(t, psh.IsIdle())

	psh.SetIdle()
	assert.True(t, psh.IsIdle())
}

func TestProcessStatusHandler_TrySetBusy(t *testing.T) {
	t.Parallel()

	t.Run("should succeed when idle", func(t *testing.T) {
		t.Parallel()

		psh := NewProcessStatusHandler()
		assert.True(t, psh.IsIdle())

		result := psh.TrySetBusy("reason")
		assert.True(t, result)
		assert.False(t, psh.IsIdle())
	})

	t.Run("should fail when already busy", func(t *testing.T) {
		t.Parallel()

		psh := NewProcessStatusHandler()
		psh.SetBusy("first reason")

		result := psh.TrySetBusy("second reason")
		assert.False(t, result)
		assert.False(t, psh.IsIdle())
	})

	t.Run("should succeed after SetIdle", func(t *testing.T) {
		t.Parallel()

		psh := NewProcessStatusHandler()
		psh.SetBusy("first reason")
		psh.SetIdle()

		result := psh.TrySetBusy("second reason")
		assert.True(t, result)
		assert.False(t, psh.IsIdle())
	})

	t.Run("second TrySetBusy should fail when first succeeded", func(t *testing.T) {
		t.Parallel()

		psh := NewProcessStatusHandler()

		result1 := psh.TrySetBusy("first")
		assert.True(t, result1)

		result2 := psh.TrySetBusy("second")
		assert.False(t, result2)
		assert.False(t, psh.IsIdle())
	})

	t.Run("TrySetBusy should succeed again after SetIdle following a successful TrySetBusy", func(t *testing.T) {
		t.Parallel()

		psh := NewProcessStatusHandler()

		assert.True(t, psh.TrySetBusy("first"))
		psh.SetIdle()
		assert.True(t, psh.TrySetBusy("second"))
		assert.False(t, psh.IsIdle())
	})
}

func TestNewProcessStatusHandler_ParallelCalls(t *testing.T) {
	t.Parallel()

	psh := NewProcessStatusHandler()
	numCalls := 1000
	wg := sync.WaitGroup{}
	wg.Add(numCalls)

	for i := 0; i < numCalls; i++ {
		operationIndex := i % 3

		go func() {
			switch operationIndex {
			case 0:
				psh.SetBusy("reason")
			case 1:
				psh.SetIdle()
			case 2:
				psh.IsIdle()
			}

			wg.Done()
		}()
	}

	wg.Wait()
}

func TestNewProcessStatusHandler_TrySetBusyConcurrency(t *testing.T) {
	t.Parallel()

	psh := NewProcessStatusHandler()
	numGoroutines := 100
	successCount := int32(0)
	wg := sync.WaitGroup{}
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			if psh.TrySetBusy("concurrent reason") {
				atomic.AddInt32(&successCount, 1)
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, int32(1), atomic.LoadInt32(&successCount))
	assert.False(t, psh.IsIdle())
}
