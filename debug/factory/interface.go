package factory

// InterceptorResolverDebugHandler hold information about requested and received information
type InterceptorResolverDebugHandler interface {
	LogRequestedData(topic string, hash []byte, numReqIntra int, numReqCross int)
	LogReceivedHash(topic string, hash []byte)
	LogProcessedHash(topic string, hash []byte, err error)
	LogFailedToResolveData(topic string, hash []byte, err error)
	Enabled() bool
	Query(topic string) []string
	IsInterfaceNil() bool
}
