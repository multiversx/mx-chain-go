package trie

import (
	"sync"
)

type snapshotsBuffer struct {
	mutOp  sync.RWMutex
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
	log.Trace("snapshots buffer add", "rootHash", rootHash)
}

func (sb *snapshotsBuffer) contains(rootHash []byte) bool {
	sb.mutOp.RLock()
	defer sb.mutOp.RUnlock()

	_, ok := sb.buffer[string(rootHash)]
	return ok
}

func (sb *snapshotsBuffer) remove(rootHash []byte) {
	sb.mutOp.Lock()
	defer sb.mutOp.Unlock()

	delete(sb.buffer, string(rootHash))
	log.Trace("snapshots buffer remove", "rootHash", rootHash)
}

func (sb *snapshotsBuffer) len() int {
	sb.mutOp.RLock()
	defer sb.mutOp.RUnlock()

	log.Trace("snapshots buffer", "len", len(sb.buffer))
	return len(sb.buffer)
}
