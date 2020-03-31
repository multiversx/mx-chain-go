package trie

import (
	"sync"
)

type pruningBuffer struct {
	mutOp  sync.RWMutex
	buffer map[string]struct{}
}

func newPruningBuffer() *pruningBuffer {
	return &pruningBuffer{
		buffer: make(map[string]struct{}),
	}
}

func (sb *pruningBuffer) add(rootHash []byte) {
	sb.mutOp.Lock()
	defer sb.mutOp.Unlock()

	if len(sb.buffer) == pruningBufferLen {
		log.Trace("pruning buffer is full", "rootHash", rootHash)
		return
	}

	sb.buffer[string(rootHash)] = struct{}{}
	log.Trace("pruning buffer add", "rootHash", rootHash)
}

func (sb *pruningBuffer) remove(rootHash []byte) {
	sb.mutOp.Lock()
	defer sb.mutOp.Unlock()

	delete(sb.buffer, string(rootHash))
}

func (sb *pruningBuffer) removeAll() map[string]struct{} {
	sb.mutOp.Lock()
	defer sb.mutOp.Unlock()

	log.Trace("pruning buffer", "len", len(sb.buffer))

	buffer := sb.buffer
	sb.buffer = make(map[string]struct{})

	return buffer
}
