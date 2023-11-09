package pruningBuffer

import (
	"sync"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("state/pruningBuffer")

type pruningBuffer struct {
	mutOp  sync.RWMutex
	buffer [][]byte
	size   uint32
}

// NewPruningBuffer creates a new instance of pruning buffer
func NewPruningBuffer(pruningBufferLen uint32) *pruningBuffer {
	return &pruningBuffer{
		buffer: make([][]byte, 0),
		size:   pruningBufferLen,
	}
}

// Add appends a new byteArray to the buffer if there is any space left
// Otherwise, it will remove the oldest record
func (pb *pruningBuffer) Add(rootHash []byte) {
	pb.mutOp.Lock()
	defer pb.mutOp.Unlock()

	if pb.size == 0 {
		return
	}

	if uint32(len(pb.buffer)) == pb.size {
		removedHash := pb.buffer[0]
		pb.buffer = pb.buffer[1:]
		pb.buffer = append(pb.buffer, rootHash)
		log.Debug("pruning buffer is full, removing oldest", "added hash", rootHash, "removed hash", removedHash)

		return
	}

	pb.buffer = append(pb.buffer, rootHash)
	log.Trace("pruning buffer add", "rootHash", rootHash)
}

// RemoveAll empties the buffer and returns all the contained byteArrays
func (pb *pruningBuffer) RemoveAll() [][]byte {
	pb.mutOp.Lock()
	defer pb.mutOp.Unlock()

	log.Trace("pruning buffer", "len", len(pb.buffer))

	buffer := make([][]byte, len(pb.buffer))
	copy(buffer, pb.buffer)
	pb.buffer = make([][]byte, 0)

	return buffer
}

// Len returns the number of elements from the buffer
func (pb *pruningBuffer) Len() int {
	pb.mutOp.RLock()
	defer pb.mutOp.RUnlock()

	return len(pb.buffer)
}
