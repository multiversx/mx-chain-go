package networksharding

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

const indexNotFound = -1

type pidQueue struct {
	data []core.PeerID
}

func newPidQueue() *pidQueue {
	return &pidQueue{
		data: make([]core.PeerID, 0),
	}
}

func (pq *pidQueue) push(pid core.PeerID) {
	pq.data = append(pq.data, pid)
}

func (pq *pidQueue) pop() core.PeerID {
	evicted := pq.data[0]
	pq.data = pq.data[1:]

	return evicted
}

func (pq *pidQueue) indexOf(pid core.PeerID) int {
	for idx, p := range pq.data {
		if p == pid {
			return idx
		}
	}

	return indexNotFound
}

func (pq *pidQueue) promote(idx int) {
	if len(pq.data) < 2 {
		return
	}

	promoted := pq.data[idx]
	pq.data = append(pq.data[:idx], pq.data[idx+1:]...)
	pq.data = append(pq.data, promoted)
}

func (pq *pidQueue) remove(pid core.PeerID) {
	newData := make([]core.PeerID, 0, len(pq.data))

	for _, p := range pq.data {
		if p == pid {
			continue
		}

		newData = append(newData, p)
	}

	pq.data = newData
}

func (pq *pidQueue) size() int {
	sum := 0
	for _, pid := range pq.data {
		sum += len(pid)
	}

	return sum
}
