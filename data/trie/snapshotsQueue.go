package trie

import (
	"sync"
)

type snapshotsQueue struct {
	queue []*snapshotsQueueEntry
	mut   sync.RWMutex
}

type snapshotsQueueEntry struct {
	rootHash []byte
	newDb    bool
}

func newSnapshotsQueue() *snapshotsQueue {
	return &snapshotsQueue{
		queue: make([]*snapshotsQueueEntry, 0),
	}
}

func (sq *snapshotsQueue) add(rootHash []byte, newDb bool) {
	sq.mut.Lock()
	newSnapshot := &snapshotsQueueEntry{
		rootHash: rootHash,
		newDb:    newDb,
	}
	sq.queue = append(sq.queue, newSnapshot)
	sq.mut.Unlock()
}

func (sq *snapshotsQueue) len() int {
	sq.mut.Lock()
	defer sq.mut.Unlock()

	return len(sq.queue)
}

func (sq *snapshotsQueue) clone() snapshotsBuffer {
	sq.mut.Lock()
	defer sq.mut.Unlock()

	newQueue := make([]*snapshotsQueueEntry, len(sq.queue))
	for i := range newQueue {
		newQueue[i] = &snapshotsQueueEntry{
			rootHash: sq.queue[i].rootHash,
			newDb:    sq.queue[i].newDb,
		}
	}

	return &snapshotsQueue{queue: newQueue}
}

func (sq *snapshotsQueue) getFirst() *snapshotsQueueEntry {
	sq.mut.Lock()
	defer sq.mut.Unlock()

	return sq.queue[0]
}

func (sq *snapshotsQueue) removeFirst() bool {
	sq.mut.Lock()
	defer sq.mut.Unlock()

	sq.queue = sq.queue[1:]
	return len(sq.queue) == 0
}
