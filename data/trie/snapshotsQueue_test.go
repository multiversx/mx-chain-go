package trie

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSnapshotsQueue(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, newSnapshotsQueue())
}

func TestSnapshotsQueue_Add(t *testing.T) {
	t.Parallel()

	sq := newSnapshotsQueue()
	sq.add([]byte("root hash"), true)

	assert.Equal(t, 1, len(sq.queue))
	assert.Equal(t, []byte("root hash"), sq.queue[0].rootHash)
	assert.True(t, sq.queue[0].newDb)
}

func TestSnapshotsQueue_AddConcurrently(t *testing.T) {
	t.Parallel()

	sq := newSnapshotsQueue()
	numSnapshots := 100

	wg := sync.WaitGroup{}
	wg.Add(numSnapshots)

	for i := 0; i < numSnapshots; i++ {
		go func(index int) {
			sq.add([]byte(strconv.Itoa(i)), true)
			wg.Done()
		}(i)
	}

	wg.Wait()

	assert.Equal(t, numSnapshots, len(sq.queue))
}

func TestSnapshotsQueue_Len(t *testing.T) {
	t.Parallel()

	sq := newSnapshotsQueue()
	numSnapshots := 100

	for i := 0; i < numSnapshots; i++ {
		sq.add([]byte(strconv.Itoa(i)), true)
	}

	assert.Equal(t, numSnapshots, sq.len())
}

func TestSnapshotsQueue_Clone(t *testing.T) {
	t.Parallel()

	sq := newSnapshotsQueue()
	sq.add([]byte("root hash"), true)

	newSq := sq.clone()
	assert.Equal(t, sq.len(), newSq.len())

	sq.queue[0].newDb = false
	assert.True(t, newSq.getFirst().newDb)

	sq.add([]byte("root hash1"), true)
	assert.NotEqual(t, sq.len(), newSq.len())
}

func TestSnapshotsQueue_GetFirst(t *testing.T) {
	t.Parallel()

	sq := newSnapshotsQueue()
	numSnapshots := 10

	for i := 0; i < numSnapshots; i++ {
		sq.add([]byte(strconv.Itoa(i)), true)
	}

	firstEntry := sq.getFirst()
	assert.Equal(t, []byte(strconv.Itoa(0)), firstEntry.rootHash)
	assert.True(t, firstEntry.newDb)
}

func TestSnapshotsQueue_RemoveFirst(t *testing.T) {
	t.Parallel()

	sq := newSnapshotsQueue()
	numSnapshots := 2

	for i := 0; i < numSnapshots; i++ {
		sq.add([]byte(strconv.Itoa(i)), true)
	}

	assert.False(t, sq.removeFirst())
	assert.True(t, sq.removeFirst())
}
