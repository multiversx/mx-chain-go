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
	mutex           *sync.Mutex
	headerBodyPairs []HeaderBodyPair
	lastAddedNonce  uint64
	closed          bool
	notifyCh        chan struct{} // used only for blocking
}

// NewBlocksQueue creates and returns a new instance of blocksQueue
func NewBlocksQueue() *blocksQueue {
	mutex := &sync.Mutex{}

	return &blocksQueue{
		mutex:           mutex,
		headerBodyPairs: make([]HeaderBodyPair, 0),
		notifyCh:        make(chan struct{}, 1), // buffered so send won't block if not read yet
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
		// safe to ignore error here, as the condition inside add method is the same as this one
		_ = bq.add(pair)
	case nonce <= bq.lastAddedNonce:
		// remove all nonces starting with the new one, then add it
		err := bq.replaceAndRemoveHigherNonces(pair)
		if err != nil {
			return err
		}
	default:
		// mismatch
		return fmt.Errorf("%w: last header nonce: %d, current header nonce %d",
			ErrHeaderNonceMismatch, bq.lastAddedNonce, nonce)
	}

	log.Debug("blocksQueue.AddOrReplace - block has been added", "nonce", pair.Header.GetNonce(), "queue size", len(bq.headerBodyPairs))

	if len(bq.headerBodyPairs) > 1 {
		return nil
	}

	select {
	case bq.notifyCh <- struct{}{}:
	default:
	}

	return nil
}

func (bq *blocksQueue) getPairsFromNonce(nonce uint64) ([]HeaderBodyPair, int) {
	pairsToBeRemoved := make([]HeaderBodyPair, 0)
	firstIndex := len(bq.headerBodyPairs)
	for i, bp := range bq.headerBodyPairs {
		if bp.Header.GetNonce() < nonce {
			continue
		}

		if i < firstIndex {
			firstIndex = i
		}

		pairsToBeRemoved = append(pairsToBeRemoved, bp)
	}

	return pairsToBeRemoved, firstIndex
}

func (bq *blocksQueue) removeFromNonce(nonce uint64) []uint64 {
	removedNonces := make([]uint64, 0)
	pairsToBeRemoved, firstIndex := bq.getPairsFromNonce(nonce)
	if len(pairsToBeRemoved) == 0 {
		bq.updateLastAddedNonceBasedOnRemovingNonce(nonce)
		return removedNonces
	}

	for _, pair := range pairsToBeRemoved {
		removedNonces = append(removedNonces, pair.Header.GetNonce())
	}

	bq.headerBodyPairs = bq.headerBodyPairs[:firstIndex]
	bq.updateLastAddedNonceBasedOnRemovingNonce(nonce)

	return removedNonces
}

func (bq *blocksQueue) replaceAndRemoveHigherNonces(pair HeaderBodyPair) error {
	_ = bq.removeFromNonce(pair.Header.GetNonce())
	return bq.add(pair)
}

func (bq *blocksQueue) add(pair HeaderBodyPair) error {
	nonce := pair.Header.GetNonce()
	if len(bq.headerBodyPairs) == 0 || nonce == bq.lastAddedNonce+1 {
		bq.headerBodyPairs = append(bq.headerBodyPairs, pair)
		bq.lastAddedNonce = nonce
		return nil
	}

	return fmt.Errorf("%w for nonce %d (lastAddedNonce=%d)", ErrInvalidHeaderNonce, nonce, bq.lastAddedNonce)
}

// Pop removes and returns the first HeaderBodyPair from the queue.
// If the queue is empty, the method blocks until a new item is available.
func (bq *blocksQueue) Pop() (pair HeaderBodyPair, poppedValidPair bool, shouldContinue bool) {
	bq.mutex.Lock()
	if len(bq.headerBodyPairs) > 0 {
		item := bq.headerBodyPairs[0]
		bq.headerBodyPairs = bq.headerBodyPairs[1:]
		bq.mutex.Unlock()
		return item, true, true
	}
	if bq.closed {
		bq.mutex.Unlock()
		return HeaderBodyPair{}, false, false
	}
	bq.mutex.Unlock()

	// Wait until notified or closed
	_, ok := <-bq.notifyCh
	if !ok {
		return HeaderBodyPair{}, false, false
	}

	// After being notified, check again
	bq.mutex.Lock()
	defer bq.mutex.Unlock()
	if len(bq.headerBodyPairs) > 0 {
		item := bq.headerBodyPairs[0]
		bq.headerBodyPairs = bq.headerBodyPairs[1:]
		return item, true, true
	}

	// allow further Pops
	return HeaderBodyPair{}, false, true
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
func (bq *blocksQueue) RemoveAtNonceAndHigher(nonce uint64) []uint64 {
	bq.mutex.Lock()
	defer bq.mutex.Unlock()

	if bq.closed {
		return make([]uint64, 0)
	}

	return bq.removeFromNonce(nonce)
}

func (bq *blocksQueue) updateLastAddedNonceBasedOnRemovingNonce(removingNonce uint64) {
	if len(bq.headerBodyPairs) > 0 {
		bq.lastAddedNonce = bq.headerBodyPairs[len(bq.headerBodyPairs)-1].Header.GetNonce()
		return
	}

	if removingNonce > 0 {
		bq.lastAddedNonce = removingNonce - 1
		return
	}

	bq.lastAddedNonce = 0
}

// Clean cleanup the queue and set the provided last added nonce
func (bq *blocksQueue) Clean(lastAddedNonce uint64) {
	bq.mutex.Lock()
	defer bq.mutex.Unlock()

	bq.headerBodyPairs = make([]HeaderBodyPair, 0)
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
