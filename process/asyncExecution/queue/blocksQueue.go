package queue

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("process/asyncExecution/queue")

// blocksQueue represents a thread-safe queue for storing and processing blockchain headers.
// It provides methods to add headers at the end or beginning of the queue and retrieve them
// for processing in a FIFO (First In, First Out) manner
type blocksQueue struct {
	mutex               *sync.Mutex
	headerBodyPairs     []HeaderBodyPair
	lastAddedNonce      uint64
	closed              bool
	notifyCh            chan struct{} // used only for blocking
	mutEvictionHandlers sync.RWMutex
	evictionHandlers    []BlocksQueueEvictionSubscriber
}

// NewBlocksQueue creates and returns a new instance of blocksQueue
func NewBlocksQueue() *blocksQueue {
	mutex := &sync.Mutex{}

	return &blocksQueue{
		mutex:            mutex,
		headerBodyPairs:  make([]HeaderBodyPair, 0),
		notifyCh:         make(chan struct{}, 1), // buffered so send won't block if not read yet
		evictionHandlers: make([]BlocksQueueEvictionSubscriber, 0),
	}
}

// AddOrReplace appends a HeaderBodyPair to the end of the queue,
// or replaces the element with the same nonce, removing all higher nonces.
func (bq *blocksQueue) AddOrReplace(pair HeaderBodyPair) error {
	if check.IfNil(pair.Header) {
		return common.ErrNilHeaderHandler
	}
	if check.IfNil(pair.Body) {
		return data.ErrNilBlockBody
	}

	bq.mutex.Lock()
	defer bq.mutex.Unlock()
	if bq.closed {
		log.Warn("blocksQueue.AddOrReplace - block queue is closed")
		return nil
	}

	nonce := pair.Header.GetNonce()
	switch {
	case nonce == bq.lastAddedNonce+1 || len(bq.headerBodyPairs) == 0:
		// append next
		bq.lastAddedNonce = nonce
		bq.headerBodyPairs = append(bq.headerBodyPairs, pair)
	case nonce <= bq.lastAddedNonce:
		// replace at provided nonce and remove all higher nonces
		err := bq.replaceAndRemoveHigherNonces(pair, nonce)
		if err != nil {
			return err
		}
	default:
		// mismatch
		return fmt.Errorf("%w: last header nonce: %d, current header nonce %d",
			ErrHeaderNonceMismatch, bq.lastAddedNonce, nonce)
	}

	log.Debug("blocksQueue.AddOrReplace - block queue has been added", "queue size", len(bq.headerBodyPairs))

	if len(bq.headerBodyPairs) > 1 {
		return nil
	}

	select {
	case bq.notifyCh <- struct{}{}:
	default:
	}

	return nil
}

func (bq *blocksQueue) getIndexForNonce(nonce uint64) int {
	for i, bp := range bq.headerBodyPairs {
		if bp.Header.GetNonce() == nonce {
			return i
		}
	}

	return -1
}

func (bq *blocksQueue) getPairsWithHigherNonces(nonce uint64) ([]HeaderBodyPair, int) {
	pairsToBeRemoved := make([]HeaderBodyPair, 0)
	firstIndex := len(bq.headerBodyPairs)
	for i, bp := range bq.headerBodyPairs {
		if bp.Header.GetNonce() <= nonce {
			continue
		}

		if i < firstIndex {
			firstIndex = i
		}

		pairsToBeRemoved = append(pairsToBeRemoved, bp)
	}

	return pairsToBeRemoved, firstIndex
}

func (bq *blocksQueue) removeAfterNonce(nonce uint64) {
	pairsToBeRemoved, firstIndex := bq.getPairsWithHigherNonces(nonce)
	if len(pairsToBeRemoved) == 0 {
		return
	}

	bq.headerBodyPairs = bq.headerBodyPairs[:firstIndex]
	bq.notifyEvictedPairs(pairsToBeRemoved)
}

func (bq *blocksQueue) replaceAndRemoveHigherNonces(pair HeaderBodyPair, nonce uint64) error {
	bq.removeAfterNonce(nonce)

	indexToReplace := bq.getIndexForNonce(nonce)
	if indexToReplace == -1 {
		// the provided nonce was not found, append it
		bq.headerBodyPairs = append(bq.headerBodyPairs, pair)
		bq.lastAddedNonce = nonce
		return nil
	}

	bq.headerBodyPairs[indexToReplace] = pair
	bq.lastAddedNonce = nonce

	return nil
}

// Pop removes and returns the first HeaderBodyPair from the queue.
// If the queue is empty, the method blocks until a new item is available.
func (bq *blocksQueue) Pop() (HeaderBodyPair, bool) {
	bq.mutex.Lock()
	if len(bq.headerBodyPairs) > 0 {
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

	// After being notified, check again
	bq.mutex.Lock()
	defer bq.mutex.Unlock()
	if len(bq.headerBodyPairs) > 0 {
		item := bq.headerBodyPairs[0]
		bq.headerBodyPairs = bq.headerBodyPairs[1:]
		return item, true
	}
	return HeaderBodyPair{}, false
}

// Peek returns the first element from queue
func (bq *blocksQueue) Peek() (HeaderBodyPair, bool) {
	bq.mutex.Lock()
	defer bq.mutex.Unlock()

	if bq.closed || len(bq.headerBodyPairs) == 0 {
		return HeaderBodyPair{}, false
	}

	return bq.headerBodyPairs[0], true

}

// RemoveAtNonceAndHigher removes the header-body pair at the specified nonce
// and all pairs with higher nonces from the queue
func (bq *blocksQueue) RemoveAtNonceAndHigher(nonce uint64) error {
	bq.mutex.Lock()
	defer bq.mutex.Unlock()

	if bq.closed {
		return nil
	}

	bq.removeAfterNonce(nonce)

	// notify subscribers the nonce should be deleted no matter it was already popped or not
	bq.notifyHeaderEvicted(nonce)

	indexToRemove := bq.getIndexForNonce(nonce)
	if indexToRemove == -1 {
		// nonce already popped
		return nil
	}

	if indexToRemove == 0 {
		// removing from the beginning, clear the entire queue
		bq.headerBodyPairs = make([]HeaderBodyPair, 0)
		if nonce > 0 {
			bq.lastAddedNonce = nonce - 1
			return nil
		}

		bq.lastAddedNonce = 0
		return nil
	}

	bq.headerBodyPairs = bq.headerBodyPairs[:indexToRemove]
	bq.lastAddedNonce = bq.headerBodyPairs[len(bq.headerBodyPairs)-1].Header.GetNonce()

	return nil
}

// RegisterEvictionSubscriber registers a new eviction subscriber
func (bq *blocksQueue) RegisterEvictionSubscriber(subscriber BlocksQueueEvictionSubscriber) {
	if bq.closed || check.IfNil(subscriber) {
		return
	}

	bq.mutEvictionHandlers.Lock()
	defer bq.mutEvictionHandlers.Unlock()

	bq.evictionHandlers = append(bq.evictionHandlers, subscriber)
}

func (bq *blocksQueue) notifyEvictedPairs(evicted []HeaderBodyPair) {
	for _, evictedPair := range evicted {
		bq.notifyHeaderEvicted(evictedPair.Header.GetNonce())
	}
}

func (bq *blocksQueue) notifyHeaderEvicted(headerNonce uint64) {
	bq.mutEvictionHandlers.RLock()
	defer bq.mutEvictionHandlers.RUnlock()

	for _, subscriber := range bq.evictionHandlers {
		subscriber.OnHeaderEvicted(headerNonce)
	}
}

// Clean cleanup the queue and set the provided last added nonce
func (bq *blocksQueue) Clean(lastAddedNonce uint64) {
	bq.mutex.Lock()
	defer bq.mutex.Unlock()

	bq.headerBodyPairs = make([]HeaderBodyPair, 0)
	bq.lastAddedNonce = lastAddedNonce
}

// SetLastAddedNonce will set the last added nonce
func (bq *blocksQueue) SetLastAddedNonce(lastAddedNonce uint64) {
	bq.mutex.Lock()
	defer bq.mutex.Unlock()

	bq.lastAddedNonce = lastAddedNonce
}

// Close will close the queue
func (bq *blocksQueue) Close() {
	bq.mutex.Lock()
	defer bq.mutex.Unlock()

	if bq.closed {
		return
	}

	bq.closed = true
	close(bq.notifyCh)
}

// IsInterfaceNil returns true if there is no value under the interface
func (bq *blocksQueue) IsInterfaceNil() bool {
	return bq == nil
}
