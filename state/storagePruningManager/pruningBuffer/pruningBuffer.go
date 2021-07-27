package pruningBuffer

import (
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
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
func (pb *pruningBuffer) Add(rootHash []byte) {
	pb.mutOp.Lock()
	defer pb.mutOp.Unlock()

	if uint32(len(pb.buffer)) == pb.size {
		log.Warn("pruning buffer is full", "rootHash", rootHash)
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
	pb.buffer = pb.buffer[:0]

	return buffer
}

// Len returns the number of elements from the buffer
func (pb *pruningBuffer) Len() int {
	pb.mutOp.RLock()
	defer pb.mutOp.RUnlock()

	return len(pb.buffer)
}

// MaximumSize returns the maximum, provided, number of element the pruning buffer can hold
func (pb *pruningBuffer) MaximumSize() int {
	return int(pb.size)
}

// IsInterfaceNil returns true if there is no value under the interface
func (pb *pruningBuffer) IsInterfaceNil() bool {
	return pb == nil
}
