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

func (bq *blocksQueue) replaceAndRemoveHigherNonces(pair HeaderBodyPair, nonce uint64) error {
	indexToReplace := -1
	for i, bp := range bq.headerBodyPairs {
		if bp.Header.GetNonce() == nonce {
			indexToReplace = i
			break
		}
	}
	if indexToReplace == -1 {
		return fmt.Errorf("%w for nonce %d", ErrMissingHeaderNonce, nonce)
	}

	initialLen := len(bq.headerBodyPairs)
	if indexToReplace == initialLen-1 {
		// last element, replace it
		bq.headerBodyPairs[indexToReplace] = pair
		bq.lastAddedNonce = nonce
		return nil
	}

	// replace at the found index and truncate everything after it
	bq.headerBodyPairs[indexToReplace] = pair
	bq.headerBodyPairs = bq.headerBodyPairs[:indexToReplace+1]
	bq.lastAddedNonce = nonce
	log.Debug("blocksQueue.replaceAndRemoveHigherNonces",
		"nonce", nonce,
		"removed count", initialLen-indexToReplace-1)

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
