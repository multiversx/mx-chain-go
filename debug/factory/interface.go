package factory

// InterceptorResolverDebugHandler hold information about requested and received information
type InterceptorResolverDebugHandler interface {
	LogRequestedData(topic string, hashes [][]byte, numReqIntra int, numReqCross int)
	LogReceivedHashes(topic string, hashes [][]byte)
	LogProcessedHashes(topic string, hashes [][]byte, err error)
	LogFailedToResolveData(topic string, hash []byte, err error)
	LogSucceededToResolveData(topic string, hash []byte)
	Query(topic string) []string
	IsInterfaceNil() bool
}
