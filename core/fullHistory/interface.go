package fullHistory

// HistoryProcessorFactory can create new instances of HistoryHandler
type HistoryProcessorFactory interface {
	Create() (HistoryHandler, error)
	IsInterfaceNil() bool
}

// HistoryHandler provides methods needed for the history data processing
type HistoryHandler interface {
	PutTransactionsData(htd *HistoryTransactionsData) error
	GetTransaction(hash []byte) (*HistoryTransactionWithEpoch, error)
	IsEnabled() bool
	IsInterfaceNil() bool
}

type hashEpochHandler interface {
	GetEpoch(hash []byte) (uint32, error)
	SaveEpoch(hash []byte, epoch uint32) error
}
