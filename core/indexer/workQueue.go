package indexer

import (
	"sync"
	"time"
)

const backOffTime = time.Second*10
const maxBackOff = time.Minute*5
const workCycleTime = time.Second*2

type workQueue struct {
	backOff time.Duration
	workMut sync.RWMutex
	pendingRequests []*workItem
}

// NewWorkQueue -
func NewWorkQueue() (*workQueue, error) {
	return &workQueue{
		backOff:         0,
		pendingRequests: make([]*workItem, 0),
	}, nil
}

// GetCycleTime -
func (wq *workQueue) GetCycleTime() time.Duration {
	return workCycleTime + wq.backOff
}

// Length -
func (wq *workQueue) Length() int {
	wq.workMut.RLock()
	length := len(wq.pendingRequests)
	wq.workMut.RUnlock()

	return length
}

// Next -
func (wq *workQueue) Next() *workItem {
	if wq.Length() == 0 {
		return nil
	}

	wq.workMut.RLock()
	nextItem := wq.pendingRequests[0]
	wq.workMut.RUnlock()

	return nextItem
}

// Add -
func (wq *workQueue) Add(item *workItem) {
	wq.workMut.Lock()
	wq.pendingRequests = append(wq.pendingRequests, item)
	wq.workMut.Unlock()
}

// Done -
func (wq *workQueue) Done() {
	wq.backOff = 0
	if wq.Length() == 0 {
		return
	}

	wq.workMut.Lock()
	wq.pendingRequests = wq.pendingRequests[1:]
	wq.workMut.Unlock()
}

// GotBackOff -
func (wq *workQueue) GotBackOff() {
	if wq.backOff == 0 {
		wq.backOff = backOffTime
		return
	}

	if wq.backOff >= maxBackOff {
		return
	}

	wq.backOff += wq.backOff/5
}
