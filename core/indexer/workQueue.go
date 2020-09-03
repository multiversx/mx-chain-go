package indexer

import (
	"sync"
	"time"
)

const (
	backOffTime   = time.Second * 10
	maxBackOff    = time.Minute * 5
	workCycleTime = time.Second * 2
)

type workQueue struct {
	backOff         time.Duration
	workMut         sync.RWMutex
	pendingRequests []*workItem
}

// NewWorkQueue creates and initializes a new work queue
func NewWorkQueue() (*workQueue, error) {
	return &workQueue{
		backOff:         0,
		pendingRequests: make([]*workItem, 0),
	}, nil
}

// GetCycleTime returns the wait duration a worker should use between requests
func (wq *workQueue) GetCycleTime() time.Duration {
	return workCycleTime + wq.backOff
}

// Length returns the number of work items in the work queue
func (wq *workQueue) Length() int {
	wq.workMut.RLock()
	length := len(wq.pendingRequests)
	wq.workMut.RUnlock()

	return length
}

// Next returns the next item to be processed
func (wq *workQueue) Next() *workItem {
	if wq.Length() == 0 {
		return nil
	}

	wq.workMut.RLock()
	nextItem := wq.pendingRequests[0]
	wq.workMut.RUnlock()

	return nextItem
}

// Add inserts a new work item at the end of the queue
func (wq *workQueue) Add(item *workItem) {
	wq.workMut.Lock()
	wq.pendingRequests = append(wq.pendingRequests, item)
	wq.workMut.Unlock()
}

// Done signals that a work item was processed successfully, thus resets the backoff if any, and removes the item
//  from the top of the list
func (wq *workQueue) Done() {
	wq.backOff = 0
	if wq.Length() == 0 {
		return
	}

	wq.workMut.Lock()
	wq.pendingRequests = wq.pendingRequests[1:]
	wq.workMut.Unlock()
}

// GotBackOff signals that a backoff was received, and the backoff value should increase
func (wq *workQueue) GotBackOff() {
	if wq.backOff == 0 {
		wq.backOff = backOffTime
		return
	}

	if wq.backOff >= maxBackOff {
		return
	}

	wq.backOff += wq.backOff / 5
}

// GetBackOffTime will return back off time
func (wq *workQueue) GetBackOffTime() int64 {
	return int64(wq.backOff)
}
