package fullHistory

// HistoryProcessorFactory can create new instances of HistoryHandler
type HistoryProcessorFactory interface {
	Create() (HistoryHandler, error)
	IsInterfaceNil() bool
}

// HistoryHandler provides methods needed for the history data processing
type HistoryHandler interface {
	PutTransactionsData(htd *HistoryTransactionsData) error
	GetTransaction(hash []byte) (*HistoryTransaction, error)
	IsEnabled() bool
	IsInterfaceNil() bool
}
