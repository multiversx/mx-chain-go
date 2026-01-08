package queue

import (
	"fmt"
	"time"
)

// SetLastAddedNonce will set the last added nonce
func (bq *blocksQueue) SetLastAddedNonce(lastAddedNonce uint64) {
	bq.mutex.Lock()
	defer bq.mutex.Unlock()

	bq.lastAddedNonce = lastAddedNonce
}

// TestingPop is a duplicate of the original Pop method but has a delay in between <-bq.notifyCh and the next bq.mutex.Lock()
func (bq *blocksQueue) TestingPop(delay time.Duration) (HeaderBodyPair, bool) {
	bq.mutex.Lock()
	if len(bq.headerBodyPairs) > 1 {
		item := bq.headerBodyPairs[0]
		bq.headerBodyPairs = bq.headerBodyPairs[1:]
		bq.mutex.Unlock()
		return item, true
	}
	if bq.closed {
		bq.mutex.Unlock()
		return HeaderBodyPair{}, false
	}
	bq.mutex.Unlock()

	// Wait until notified or closed
	_, ok := <-bq.notifyCh
	if !ok {
		return HeaderBodyPair{}, false
	}

	// the only difference from the original Pop method
	time.Sleep(delay)

	// After being notified, check again
	bq.mutex.Lock()
	defer bq.mutex.Unlock()
	if len(bq.headerBodyPairs) > 0 {
		item := bq.headerBodyPairs[0]
		bq.headerBodyPairs = bq.headerBodyPairs[1:]
		println(fmt.Sprintf("popped %d", item.Header.GetNonce()))
		return item, true
	}

	if bq.closed {
		return HeaderBodyPair{}, false
	}

	log.Warn("blocksQueue.Pop - blocks queue is empty")

	return HeaderBodyPair{}, true
}
