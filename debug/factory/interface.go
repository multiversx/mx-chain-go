package factory

// InterceptorDebugHandler hold information about requested and received information
type InterceptorDebugHandler interface {
	LogRequestedData(topic string, hashes [][]byte, numReqIntra int, numReqCross int)
	LogReceivedHashes(topic string, hashes [][]byte)
	LogProcessedHashes(topic string, hashes [][]byte, err error)
	LogFailedToResolveData(topic string, hash []byte, err error)
	LogSucceededToResolveData(topic string, hash []byte)
	Query(topic string) []string
	Close() error
	IsInterfaceNil() bool
}

// ProcessDebugger defines what a process debugger implementation should do
type ProcessDebugger interface {
	SetLastCommittedBlockRound(round uint64)
	Close() error
	IsInterfaceNil() bool
}
