package track

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// blockTracker is the interface needed by base block track to deal with shards and meta nodes while they track blocks
type blockTracker interface {
	getSelfHeaders(headerHandler data.HeaderHandler) []*headerInfo
	computeLongestSelfChain() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte)
	sortHeadersForShardFromNonce(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte)
}

type blockNotarizerHandler interface {
	addNotarizedHeader(shardID uint32, notarizedHeader data.HeaderHandler, notarizedHeaderHash []byte)
	cleanupNotarizedHeadersBehindNonce(shardID uint32, nonce uint64)
	displayNotarizedHeaders(shardID uint32)
	getLastNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error)
	getLastNotarizedHeaderNonce(shardID uint32) uint64
	getNotarizedHeader(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error)
	initNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error
	removeLastNotarizedHeader()
	restoreNotarizedHeadersToGenesis()
}

type blockNotifierHandler interface {
	callHandlers(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)
	registerHandler(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
}

type blockProcessorHandler interface {
	computeLongestChain(shardID uint32, header data.HeaderHandler) ([]data.HeaderHandler, [][]byte)
	processReceivedHeader(header data.HeaderHandler)
}
