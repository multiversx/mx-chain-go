package queue

import "sync"

type sliceQueue struct {
	queue [][]byte
	size  uint
	*sync.RWMutex
}

// NewSliceQueue creates a new sliceQueue
func NewSliceQueue(numHashes uint) *sliceQueue {
	return &sliceQueue{
		queue:   make([][]byte, 0),
		size:    numHashes,
		RWMutex: &sync.RWMutex{},
	}
}

// Add adds a new element to the queue, and evicts the first element if the queue is full
func (hq *sliceQueue) Add(data []byte) []byte {
	if hq.size == 0 {
		return data
	}

	hq.Lock()
	defer hq.Unlock()

	if uint(len(hq.queue)) == hq.size {
		dataToEvict := hq.queue[0]
		hq.queue = hq.queue[1:]
		hq.queue = append(hq.queue, data)

		return dataToEvict
	}

	hq.queue = append(hq.queue, data)

	return nil
}
