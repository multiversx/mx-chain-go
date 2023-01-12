package common

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
)

const indexNotFound = -1

type pidQueue struct {
	data    []core.PeerID
	mutData sync.RWMutex
}

// NewPidQueue returns a new instance of pid queue
func NewPidQueue() *pidQueue {
	return &pidQueue{
		data: make([]core.PeerID, 0),
	}
}

// Push adds a new pid at the end of the queue
func (pq *pidQueue) Push(pid core.PeerID) {
	pq.mutData.Lock()
	defer pq.mutData.Unlock()

	pq.data = append(pq.data, pid)
}

// Pop removes first pid from queue
func (pq *pidQueue) Pop() core.PeerID {
	pq.mutData.Lock()
	defer pq.mutData.Unlock()

	if len(pq.data) == 0 {
		return ""
	}

	evicted := pq.data[0]
	pq.data = pq.data[1:]

	return evicted
}

// IndexOf returns the index of the pid
func (pq *pidQueue) IndexOf(pid core.PeerID) int {
	pq.mutData.RLock()
	defer pq.mutData.RUnlock()

	for idx, p := range pq.data {
		if p == pid {
			return idx
		}
	}

	return indexNotFound
}

// Promote moves the pid at the specified index at the end of the queue
func (pq *pidQueue) Promote(idx int) {
	pq.mutData.Lock()
	defer pq.mutData.Unlock()

	if len(pq.data) < 2 {
		return
	}

	if idx < 0 || idx >= len(pq.data) {
		return
	}

	promoted := pq.data[idx]
	pq.data = append(pq.data[:idx], pq.data[idx+1:]...)
	pq.data = append(pq.data, promoted)
}

// Remove removes the pid from the queue
func (pq *pidQueue) Remove(pid core.PeerID) {
	pq.mutData.Lock()
	defer pq.mutData.Unlock()

	newData := make([]core.PeerID, 0, len(pq.data))

	for _, p := range pq.data {
		if p == pid {
			continue
		}

		newData = append(newData, p)
	}

	pq.data = newData
}

// DataSizeInBytes returns the data size in bytes
func (pq *pidQueue) DataSizeInBytes() int {
	pq.mutData.RLock()
	defer pq.mutData.RUnlock()

	sum := 0
	for _, pid := range pq.data {
		sum += len(pid)
	}

	return sum
}

// Get returns the pid from the specified index
func (pq *pidQueue) Get(idx int) core.PeerID {
	pq.mutData.RLock()
	defer pq.mutData.RUnlock()
	if idx < 0 || idx >= len(pq.data) {
		return ""
	}

	return pq.data[idx]
}

// Len returns the number of pids in queue
func (pq *pidQueue) Len() int {
	pq.mutData.RLock()
	defer pq.mutData.RUnlock()

	return len(pq.data)
}

// IsInterfaceNil returns true if there is no value under the interface
func (pq *pidQueue) IsInterfaceNil() bool {
	return pq == nil
}
