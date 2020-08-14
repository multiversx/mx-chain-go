package fullHistory

// HistoryRepositoryFactory can create new instances of HistoryRepository
type HistoryRepositoryFactory interface {
	Create() (HistoryRepository, error)
	IsInterfaceNil() bool
}

// HistoryRepository provides methods needed for the history data processing
type HistoryRepository interface {
	PutTransactionsData(htd *HistoryTransactionsData) error
	GetTransactionsGroupMetadata(hash []byte) (*TransactionsGroupMetadataWithEpoch, error)
	GetEpochByHash(hash []byte) (uint32, error)
	IsEnabled() bool
	IsInterfaceNil() bool
}
