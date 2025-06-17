package queue

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
)

// headersQueue represents a thread-safe queue for storing and processing blockchain headers
// It provides methods to add headers at the end or beginning of the queue and retrieve them
// for processing in a FIFO (First In, First Out) manner
type headersQueue struct {
	mutex   sync.RWMutex
	headers []data.HeaderHandler
}

// NewHeadersQueue creates and returns a new instance of headersQueue

func NewHeadersQueue() (*headersQueue, error) {
	return &headersQueue{
		headers: make([]data.HeaderHandler, 0),
	}, nil
}

// Add appends a single header to the end of the queue
func (hq *headersQueue) Add(header data.HeaderHandler) error {
	if check.IfNil(header) {
		return common.ErrNilHeaderHandler
	}

	hq.mutex.Lock()
	defer hq.mutex.Unlock()

	hq.headers = append(hq.headers, header)

	return nil
}

// AddFirstMultiple adds multiple headers at the beginning of the queue
func (hq *headersQueue) AddFirstMultiple(headers []data.HeaderHandler) error {
	if len(headers) == 0 {
		return nil
	}

	hq.mutex.Lock()
	defer hq.mutex.Unlock()

	hq.headers = append(headers, hq.headers...)

	return nil
}

// TakeFirstHeaderForProcessing removes and returns the first header from the queue
func (hq *headersQueue) TakeFirstHeaderForProcessing() (data.HeaderHandler, error) {
	hq.mutex.Lock()
	defer hq.mutex.Unlock()

	if len(hq.headers) == 0 {
		return nil, ErrNoHeaderForProcessing
	}

	headerForProcessing := hq.headers[0]
	hq.headers = hq.headers[1:]

	return headerForProcessing, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (hq *headersQueue) IsInterfaceNil() bool {
	return hq == nil
}
