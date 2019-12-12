package track

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// RegisterSelfNotarizedHeadersHandler registers a new handler to be called when self notarized header is changed
func (bbt *baseBlockTrack) RegisterSelfNotarizedHeadersHandler(handler func(headers []data.HeaderHandler, headersHashes [][]byte)) {
	if handler == nil {
		log.Debug("attempt to register a nil handler to a tracker object")
		return
	}

	bbt.mutSelfNotarizedHeadersHandlers.Lock()
	bbt.selfNotarizedHeadersHandlers = append(bbt.selfNotarizedHeadersHandlers, handler)
	bbt.mutSelfNotarizedHeadersHandlers.Unlock()
}

func (bbt *baseBlockTrack) callSelfNotarizedHeadersHandlers(headers []data.HeaderHandler, headersHashes [][]byte) {
	bbt.mutSelfNotarizedHeadersHandlers.RLock()
	for _, handler := range bbt.selfNotarizedHeadersHandlers {
		go handler(headers, headersHashes)
	}
	bbt.mutSelfNotarizedHeadersHandlers.RUnlock()
}
