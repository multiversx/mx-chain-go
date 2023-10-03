package common

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPidQueue_PushPopShouldWork(t *testing.T) {
	t.Parallel()

	pq := NewPidQueue()
	pid0 := core.PeerID("pid 0")
	pid1 := core.PeerID("pid 1")

	evicted := pq.Pop()
	assert.Equal(t, core.PeerID(""), evicted)

	pq.Push(pid0)
	pq.Push(pid1)

	require.Equal(t, 2, pq.Len())
	assert.Equal(t, pid0, pq.Get(0))
	assert.Equal(t, pid1, pq.Get(1))

	evicted = pq.Pop()
	assert.Equal(t, pid0, evicted)

	evicted = pq.Pop()
	assert.Equal(t, pid1, evicted)

	require.Equal(t, 0, pq.Len())
}

func TestPidQueue_IndexOfShouldWork(t *testing.T) {
	t.Parallel()

	pq := NewPidQueue()
	pid0 := core.PeerID("pid 0")
	pid1 := core.PeerID("pid 1")
	pid2 := core.PeerID("pid 2")

	pq.Push(pid0)
	pq.Push(pid1)

	assert.Equal(t, 0, pq.IndexOf(pid0))
	assert.Equal(t, 1, pq.IndexOf(pid1))
	assert.Equal(t, indexNotFound, pq.IndexOf(pid2))
}

func TestPidQueue_PromoteNoElementsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not have panicked")
		}
	}()

	pq := NewPidQueue()

	pq.Promote(0)
	pq.Promote(1)
	pq.Promote(10)
}

func TestPidQueue_PromoteOneElementShouldWork(t *testing.T) {
	t.Parallel()

	pq := NewPidQueue()
	pid0 := core.PeerID("pid 0")
	pq.Push(pid0)

	pq.Promote(0)

	assert.Equal(t, 0, pq.IndexOf(pid0))
}

func TestPidQueue_PromoteTwoElementsShouldWork(t *testing.T) {
	t.Parallel()

	pq := NewPidQueue()
	pid0 := core.PeerID("pid 0")
	pid1 := core.PeerID("pid 1")
	pq.Push(pid0)
	pq.Push(pid1)

	pq.Promote(-1) // early return
	pq.Promote(2)  // early return

	pq.Promote(0)

	assert.Equal(t, 0, pq.IndexOf(pid1))
	assert.Equal(t, 1, pq.IndexOf(pid0))
}

func TestPidQueue_PromoteThreeElementsShouldWork(t *testing.T) {
	t.Parallel()

	pq := NewPidQueue()
	pid0 := core.PeerID("pid 0")
	pid1 := core.PeerID("pid 1")
	pid2 := core.PeerID("pid 2")
	pq.Push(pid0)
	pq.Push(pid1)
	pq.Push(pid2)

	pq.Promote(1)

	assert.Equal(t, 0, pq.IndexOf(pid0))
	assert.Equal(t, 1, pq.IndexOf(pid2))
	assert.Equal(t, 2, pq.IndexOf(pid1))

	pq.Promote(0)

	assert.Equal(t, 0, pq.IndexOf(pid2))
	assert.Equal(t, 1, pq.IndexOf(pid1))
	assert.Equal(t, 2, pq.IndexOf(pid0))

	pq.Promote(2)

	assert.Equal(t, 0, pq.IndexOf(pid2))
	assert.Equal(t, 1, pq.IndexOf(pid1))
	assert.Equal(t, 2, pq.IndexOf(pid0))
}

func TestPidQueue_RemoveShouldWork(t *testing.T) {
	t.Parallel()

	pq := NewPidQueue()
	pid0 := core.PeerID("pid 0")
	pid1 := core.PeerID("pid 1")
	pid2 := core.PeerID("pid 2")
	pq.Push(pid0)
	pq.Push(pid1)
	pq.Push(pid2)
	pq.Push(pid0)

	pq.Remove(pid0)

	require.Equal(t, 2, pq.Len())
	assert.Equal(t, 0, pq.IndexOf(pid1))
	assert.Equal(t, 1, pq.IndexOf(pid2))
}

func TestPidQueue_DataSizeInBytes(t *testing.T) {
	t.Parallel()

	pq := NewPidQueue()
	assert.Equal(t, 0, pq.DataSizeInBytes())

	pq.Push("pid 0")
	assert.Equal(t, 5, pq.DataSizeInBytes())

	pq.Push("pid 1")
	assert.Equal(t, 10, pq.DataSizeInBytes())

	pq.Push("0")
	assert.Equal(t, 11, pq.DataSizeInBytes())
}

func TestPidQueue_GetShouldWork(t *testing.T) {
	t.Parallel()

	pq := NewPidQueue()
	assert.Equal(t, core.PeerID(""), pq.Get(-1))
	assert.Equal(t, core.PeerID(""), pq.Get(0))
	assert.Equal(t, core.PeerID(""), pq.Get(1))

	pid0 := core.PeerID("pid 0")
	pid1 := core.PeerID("pid 1")
	pq.Push(pid0)
	pq.Push(pid1)
	assert.Equal(t, pid0, pq.Get(0))
	assert.Equal(t, pid1, pq.Get(1))
}

func TestPidQueue_LenShouldWork(t *testing.T) {
	t.Parallel()

	pq := NewPidQueue()
	assert.Equal(t, 0, pq.Len())

	numOfPids := 50
	for i := 0; i < numOfPids; i++ {
		pq.Push(core.PeerID(fmt.Sprintf("pid%d", i)))
	}

	assert.Equal(t, numOfPids, pq.Len())
}

func TestPidQueue_TestConcurrency(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("test should not have paniced: %v", r))
		}
	}()

	pq := NewPidQueue()

	numOperations := 1000
	var wg sync.WaitGroup
	wg.Add(numOperations)

	for idx := 0; idx < numOperations; idx++ {
		go func(index int) {
			time.Sleep(time.Duration(index) * 1 * time.Millisecond)
			pid := core.PeerID(fmt.Sprintf("pid%d", index))
			switch index % 9 {
			case 1:
				pq.Push(pid)
			case 2:
				pq.Get(index)
			case 3:
				pq.Pop()
			case 4:
				pq.IndexOf(pid)
			case 5:
				pq.Promote(index)
			case 6:
				pq.Remove(pid)
			case 7:
				pq.DataSizeInBytes()
			case 8:
				pq.Len()
			}
			wg.Done()
		}(idx)
	}

	wg.Wait()
}

func TestPidQueue_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var pq *pidQueue
	assert.True(t, pq.IsInterfaceNil())

	pq = NewPidQueue()
	assert.False(t, pq.IsInterfaceNil())
}
