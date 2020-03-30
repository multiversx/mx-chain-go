package trie

import (
	"sync"
)

type pruningBuffer struct {
	mutOp  sync.RWMutex
	buffer [][]byte
}

func newPruningBuffer() *pruningBuffer {
	return &pruningBuffer{
		buffer: make([][]byte, 0),
	}
}

func (sb *pruningBuffer) add(rootHash []byte) {
	sb.mutOp.Lock()
	defer sb.mutOp.Unlock()

	sb.buffer = append(sb.buffer, rootHash)
	log.Trace("snapshots buffer add", "rootHash", rootHash)
}

func (sb *pruningBuffer) len() int {
	sb.mutOp.RLock()
	defer sb.mutOp.RUnlock()

	log.Trace("snapshots buffer", "len", len(sb.buffer))
	return len(sb.buffer)
}

func (sb *pruningBuffer) removeAll() [][]byte {
	sb.mutOp.Lock()
	defer sb.mutOp.Unlock()

	log.Trace("snapshots buffer", "len", len(sb.buffer))

	buffer := sb.buffer
	sb.buffer = make([][]byte, 0)

	return buffer
}
