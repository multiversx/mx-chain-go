package track

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type blockNotarizerHandler interface {
	AddNotarizedHeader(shardID uint32, notarizedHeader data.HeaderHandler, notarizedHeaderHash []byte)
	CleanupNotarizedHeadersBehindNonce(shardID uint32, nonce uint64)
	DisplayNotarizedHeaders(shardID uint32, message string)
	GetLastNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error)
	GetFirstNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error)
	GetLastNotarizedHeaderNonce(shardID uint32) uint64
	GetNotarizedHeader(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error)
	InitNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error
	RemoveLastNotarizedHeader()
	RestoreNotarizedHeadersToGenesis()
	IsInterfaceNil() bool
}

type blockNotifierHandler interface {
	CallHandlers(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)
	RegisterHandler(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	GetNumRegisteredHandlers() int
	IsInterfaceNil() bool
}

type blockProcessorHandler interface {
	ComputeLongestChain(shardID uint32, header data.HeaderHandler) ([]data.HeaderHandler, [][]byte)
	ProcessReceivedHeader(header data.HeaderHandler)
	IsInterfaceNil() bool
}

type blockTrackerHandler interface {
	GetSelfHeaders(headerHandler data.HeaderHandler) []*HeaderInfo
	ComputeCrossInfo(headers []data.HeaderHandler)
	ComputeLongestSelfChain() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte)
	SortHeadersFromNonce(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte)
	IsInterfaceNil() bool
}

type blockBalancerHandler interface {
	GetNumPendingMiniBlocks(shardID uint32) uint32
	SetNumPendingMiniBlocks(shardID uint32, numPendingMiniBlocks uint32)
	GetLastShardProcessedMetaNonce(shardID uint32) uint64
	SetLastShardProcessedMetaNonce(shardID uint32, nonce uint64)
	IsInterfaceNil() bool
}
