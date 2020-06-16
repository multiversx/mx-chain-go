package networksharding

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPidQueue_PushPopShouldWork(t *testing.T) {
	t.Parallel()

	pq := newPidQueue()
	pid0 := core.PeerID("pid 0")
	pid1 := core.PeerID("pid 1")

	pq.push(pid0)
	pq.push(pid1)

	require.Equal(t, 2, len(pq.data))
	assert.Equal(t, pid0, pq.data[0])
	assert.Equal(t, pid1, pq.data[1])

	evicted := pq.pop()
	assert.Equal(t, pid0, evicted)

	evicted = pq.pop()
	assert.Equal(t, pid1, evicted)
}

func TestPidQueue_IndexOfShouldWork(t *testing.T) {
	t.Parallel()

	pq := newPidQueue()
	pid0 := core.PeerID("pid 0")
	pid1 := core.PeerID("pid 1")
	pid2 := core.PeerID("pid 2")

	pq.push(pid0)
	pq.push(pid1)

	assert.Equal(t, 0, pq.indexOf(pid0))
	assert.Equal(t, 1, pq.indexOf(pid1))
	assert.Equal(t, indexNotFound, pq.indexOf(pid2))
}

func TestPidQueue_PromoteNoElementsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not have panicked")
		}
	}()

	pq := newPidQueue()

	pq.promote(0)
	pq.promote(1)
	pq.promote(10)
}

func TestPidQueue_PromoteOneElementShouldWork(t *testing.T) {
	t.Parallel()

	pq := newPidQueue()
	pid0 := core.PeerID("pid 0")
	pq.push(pid0)

	pq.promote(0)

	assert.Equal(t, 0, pq.indexOf(pid0))
}

func TestPidQueue_PromoteTwoElementsShouldWork(t *testing.T) {
	t.Parallel()

	pq := newPidQueue()
	pid0 := core.PeerID("pid 0")
	pid1 := core.PeerID("pid 1")
	pq.push(pid0)
	pq.push(pid1)

	pq.promote(0)

	assert.Equal(t, 0, pq.indexOf(pid1))
	assert.Equal(t, 1, pq.indexOf(pid0))
}

func TestPidQueue_PromoteThreeElementsShouldWork(t *testing.T) {
	t.Parallel()

	pq := newPidQueue()
	pid0 := core.PeerID("pid 0")
	pid1 := core.PeerID("pid 1")
	pid2 := core.PeerID("pid 2")
	pq.push(pid0)
	pq.push(pid1)
	pq.push(pid2)

	pq.promote(1)

	assert.Equal(t, 0, pq.indexOf(pid0))
	assert.Equal(t, 1, pq.indexOf(pid2))
	assert.Equal(t, 2, pq.indexOf(pid1))

	pq.promote(0)

	assert.Equal(t, 0, pq.indexOf(pid2))
	assert.Equal(t, 1, pq.indexOf(pid1))
	assert.Equal(t, 2, pq.indexOf(pid0))

	pq.promote(2)

	assert.Equal(t, 0, pq.indexOf(pid2))
	assert.Equal(t, 1, pq.indexOf(pid1))
	assert.Equal(t, 2, pq.indexOf(pid0))
}

func TestPidQueue_RemoveShouldWork(t *testing.T) {
	t.Parallel()

	pq := newPidQueue()
	pid0 := core.PeerID("pid 0")
	pid1 := core.PeerID("pid 1")
	pid2 := core.PeerID("pid 2")
	pq.push(pid0)
	pq.push(pid1)
	pq.push(pid2)
	pq.push(pid0)

	pq.remove(pid0)

	require.Equal(t, 2, len(pq.data))
	assert.Equal(t, 0, pq.indexOf(pid1))
	assert.Equal(t, 1, pq.indexOf(pid2))
}

func TestPidQueue_Size(t *testing.T) {
	t.Parallel()

	pq := newPidQueue()
	assert.Equal(t, 0, pq.size())

	pq.push("pid 0")
	assert.Equal(t, 5, pq.size())

	pq.push("pid 1")
	assert.Equal(t, 10, pq.size())

	pq.push("0")
	assert.Equal(t, 11, pq.size())
}
