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

func (bn *blockNotifier) callHandlers(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
	bn.mutNotarizedHeadersHandlers.RLock()
	for _, handler := range bn.notarizedHeadersHandlers {
		go handler(shardID, headers, headersHashes)
	}
	bn.mutNotarizedHeadersHandlers.RUnlock()
}

func (bn *blockNotifier) registerHandler(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)) {
	if handler == nil {
		log.Debug("attempt to register a nil handler to a tracker object")
		return
	}

	bn.mutNotarizedHeadersHandlers.Lock()
	bn.notarizedHeadersHandlers = append(bn.notarizedHeadersHandlers, handler)
	bn.mutNotarizedHeadersHandlers.Unlock()
}
