package trie

import (
	"sync"
)

type pruningBuffer struct {
	mutOp  sync.RWMutex
	buffer map[string]struct{}
	size   uint32
}

func newPruningBuffer(pruningBufferLen uint32) *pruningBuffer {
	return &pruningBuffer{
		buffer: make(map[string]struct{}),
		size:   pruningBufferLen,
	}
}

func (pb *pruningBuffer) add(rootHash []byte) {
	pb.mutOp.Lock()
	defer pb.mutOp.Unlock()

	if uint32(len(pb.buffer)) == pb.size {
		log.Trace("pruning buffer is full", "rootHash", rootHash)
		return
	}

	pb.buffer[string(rootHash)] = struct{}{}
	log.Trace("pruning buffer add", "rootHash", rootHash)
}

func (pb *pruningBuffer) remove(rootHash []byte) {
	pb.mutOp.Lock()
	defer pb.mutOp.Unlock()

	delete(pb.buffer, string(rootHash))
}

func (pb *pruningBuffer) removeAll() map[string]struct{} {
	pb.mutOp.Lock()
	defer pb.mutOp.Unlock()

	log.Trace("pruning buffer", "len", len(pb.buffer))

	buffer := make(map[string]struct{})
	for key := range pb.buffer {
		buffer[key] = struct{}{}
	}

	pb.buffer = make(map[string]struct{})

	return buffer
}
