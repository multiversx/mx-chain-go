package fullHistory

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
	RecordBlock(blockHeaderHash []byte, blockHeader data.HeaderHandler, blockBody data.BodyHandler) error
	GetTransactionsGroupMetadata(hash []byte) (*TransactionsGroupMetadataWithEpoch, error)
	GetEpochByHash(hash []byte) (uint32, error)
	IsEnabled() bool
	IsInterfaceNil() bool
}
