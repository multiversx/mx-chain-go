package trie

import (
	"sync"
)

type snapshotsBuffer struct {
	mutOp  sync.Mutex
	buffer map[string]struct{}
}

func newSnapshotsBuffer() *snapshotsBuffer {
	return &snapshotsBuffer{
		buffer: make(map[string]struct{}),
	}
}

func (sb *snapshotsBuffer) add(rootHash []byte) {
	sb.mutOp.Lock()
	defer sb.mutOp.Unlock()

	sb.buffer[string(rootHash)] = struct{}{}
}

func (sb *snapshotsBuffer) contains(rootHash []byte) bool {
	sb.mutOp.Lock()
	defer sb.mutOp.Unlock()

	_, ok := sb.buffer[string(rootHash)]
	return ok
}

func (sb *snapshotsBuffer) remove(rootHash []byte) {
	sb.mutOp.Lock()
	defer sb.mutOp.Unlock()

	delete(sb.buffer, string(rootHash))
}

func (sb *snapshotsBuffer) len() int {
	sb.mutOp.Lock()
	defer sb.mutOp.Unlock()

	return len(sb.buffer)
}
