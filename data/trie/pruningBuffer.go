package trie

import (
	"sync"
)

type pruningBuffer struct {
	mutOp  sync.RWMutex
	buffer [][]byte
	size   uint32
}

func newPruningBuffer(pruningBufferLen uint32) *pruningBuffer {
	return &pruningBuffer{
		buffer: make([][]byte, 0),
		size:   pruningBufferLen,
	}
}

func (pb *pruningBuffer) add(rootHash []byte) {
	pb.mutOp.Lock()
	defer pb.mutOp.Unlock()

	if uint32(len(pb.buffer)) == pb.size {
		log.Warn("pruning buffer is full", "rootHash", rootHash)
		return
	}

	pb.buffer = append(pb.buffer, rootHash)
	log.Trace("pruning buffer add", "rootHash", rootHash)
}

func (pb *pruningBuffer) removeAll() [][]byte {
	pb.mutOp.Lock()
	defer pb.mutOp.Unlock()

	log.Trace("pruning buffer", "len", len(pb.buffer))

	buffer := make([][]byte, len(pb.buffer))
	copy(buffer, pb.buffer)
	pb.buffer = pb.buffer[:0]

	return buffer
}

func (pb *pruningBuffer) len() int {
	pb.mutOp.RLock()
	defer pb.mutOp.RUnlock()

	return len(pb.buffer)
}
