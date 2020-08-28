package dblookupext

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// HistoryRepositoryFactory can create new instances of HistoryRepository
type HistoryRepositoryFactory interface {
	Create() (HistoryRepository, error)
	IsInterfaceNil() bool
}

// HistoryRepository provides methods needed for the history data processing
type HistoryRepository interface {
	RegisterToBlockTracker(blockTracker BlockTracker)
	RecordBlock(blockHeaderHash []byte, blockHeader data.HeaderHandler, blockBody data.BodyHandler) error
	GetMiniblockMetadataByTxHash(hash []byte) (*MiniblockMetadata, error)
	GetEpochByHash(hash []byte) (uint32, error)
	IsEnabled() bool
	IsInterfaceNil() bool
}

// BlockTracker defines the interface of the block tracker
type BlockTracker interface {
	RegisterCrossNotarizedHeadersHandler(func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RegisterSelfNotarizedHeadersHandler(func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	IsInterfaceNil() bool
}
