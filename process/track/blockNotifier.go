package track

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
)

type blockNotifier struct {
	mutNotarizedHeadersHandlers sync.RWMutex
	notarizedHeadersHandlers    []func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)
}

// NewBlockNotifier creates a block notifier object which implements blockNotifierHandler interface
func NewBlockNotifier() (*blockNotifier, error) {
	bn := blockNotifier{}
	bn.notarizedHeadersHandlers = make([]func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte), 0)
	return &bn, nil
}

// CallHandlers calls all the registered handlers to notify them about the new received headers
func (bn *blockNotifier) CallHandlers(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
	if len(headers) == 0 {
		return
	}

	bn.mutNotarizedHeadersHandlers.RLock()
	for _, handler := range bn.notarizedHeadersHandlers {
		go handler(shardID, headers, headersHashes)
	}
	bn.mutNotarizedHeadersHandlers.RUnlock()
}

// RegisterHandler registers a handler which wants to be notified when new headers are received
func (bn *blockNotifier) RegisterHandler(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)) {
	if handler == nil {
		log.Warn("attempt to register a nil handler to a tracker object")
		return
	}

	bn.mutNotarizedHeadersHandlers.Lock()
	bn.notarizedHeadersHandlers = append(bn.notarizedHeadersHandlers, handler)
	bn.mutNotarizedHeadersHandlers.Unlock()
}

// GetNumRegisteredHandlers gets the total number of registered handlers
func (bn *blockNotifier) GetNumRegisteredHandlers() int {
	bn.mutNotarizedHeadersHandlers.RLock()
	defer bn.mutNotarizedHeadersHandlers.RUnlock()

	return len(bn.notarizedHeadersHandlers)
}

// IsInterfaceNil returns true if there is no value under the interface
func (bn *blockNotifier) IsInterfaceNil() bool {
	return bn == nil
}
