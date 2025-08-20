package queue

import (
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
	closed          bool
	notifyCh        chan struct{} // used only for blocking
}

// NewBlocksQueue creates and returns a new instance of blocksQueue
func NewBlocksQueue() (*blocksQueue, error) {
	mutex := &sync.Mutex{}

	return &blocksQueue{
		mutex:           mutex,
		headerBodyPairs: make([]HeaderBodyPair, 0),
		notifyCh:        make(chan struct{}, 1), // buffered so send won't block if not read yet
	}, nil
}

// AddOrReplace appends a HeaderBodyPair to the end of the queue,
// or replaces the last element if it has the same nonce.
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

	lastIndex := len(bq.headerBodyPairs) - 1
	if lastIndex >= 0 && bq.headerBodyPairs[lastIndex].Header.GetNonce() == pair.Header.GetNonce() {
		// if the last header from queue has the same nonce, we should replace it with the current pair
		bq.headerBodyPairs[lastIndex] = pair
	} else {
		bq.headerBodyPairs = append(bq.headerBodyPairs, pair)
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
