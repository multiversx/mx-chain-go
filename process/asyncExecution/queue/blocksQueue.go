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
	cond            *sync.Cond
	headerBodyPairs []HeaderBodyPair
	closed          bool
}

// NewBlocksQueue creates and returns a new instance of blocksQueue
func NewBlocksQueue() (*blocksQueue, error) {
	mutex := &sync.Mutex{}

	return &blocksQueue{
		mutex:           mutex,
		headerBodyPairs: make([]HeaderBodyPair, 0),
		cond:            sync.NewCond(mutex),
	}, nil
}

// AddOrReplace appends a HeaderBodyPair to the end of the queue,
// or replaces the last element if it has the same nonce.
func (hq *blocksQueue) AddOrReplace(pair HeaderBodyPair) error {
	if check.IfNil(pair.Header) {
		return common.ErrNilHeaderHandler
	}
	if check.IfNil(pair.Body) {
		return data.ErrNilBlockBody
	}

	hq.mutex.Lock()
	defer hq.mutex.Unlock()
	if hq.closed {
		log.Warn("blocksQueue.AddOrReplace - block queue is closed")
		return nil
	}

	lastIndex := len(hq.headerBodyPairs) - 1
	if lastIndex >= 0 && hq.headerBodyPairs[lastIndex].Header.GetNonce() == pair.Header.GetNonce() {
		// if the last header from queue has the same nonce, we should replace it with the current pair
		hq.headerBodyPairs[lastIndex] = pair
	} else {
		hq.headerBodyPairs = append(hq.headerBodyPairs, pair)
	}

	hq.cond.Signal()

	return nil
}

// Pop removes and returns the first HeaderBodyPair from the queue.
// If the queue is empty, the method blocks until a new item is available.
func (hq *blocksQueue) Pop() (HeaderBodyPair, bool) {
	hq.mutex.Lock()
	defer hq.mutex.Unlock()

	for len(hq.headerBodyPairs) == 0 && !hq.closed {
		hq.cond.Wait()
	}

	if hq.closed {
		return HeaderBodyPair{}, false
	}

	headerForProcessing := hq.headerBodyPairs[0]
	hq.headerBodyPairs = hq.headerBodyPairs[1:]

	return headerForProcessing, true
}

// Close will close the queue
func (hq *blocksQueue) Close() {
	hq.mutex.Lock()
	defer hq.mutex.Unlock()

	hq.closed = true
	hq.cond.Broadcast()
}

// IsInterfaceNil returns true if there is no value under the interface
func (hq *blocksQueue) IsInterfaceNil() bool {
	return hq == nil
}
