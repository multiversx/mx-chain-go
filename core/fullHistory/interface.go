package fullHistory

// HistoryRepositoryFactory can create new instances of HistoryRepository
type HistoryRepositoryFactory interface {
	Create() (HistoryRepository, error)
	IsInterfaceNil() bool
}

// HistoryRepository provides methods needed for the history data processing
type HistoryRepository interface {
	PutTransactionsData(htd *HistoryTransactionsData) error
	GetTransaction(hash []byte) (*HistoryTransactionWithEpoch, error)
	GetEpochForHash(hash []byte) (uint32, error)
	IsEnabled() bool
	IsInterfaceNil() bool
}

type hashEpochRepository interface {
	GetEpoch(hash []byte) (uint32, error)
	SaveEpoch(hash []byte, epoch uint32) error
}
