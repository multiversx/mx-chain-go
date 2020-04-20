package factory

// InterceptorResolverDebugHandler hold information about requested and received information
type InterceptorResolverDebugHandler interface {
	RequestedData(topic string, hash []byte, numReqIntra int, numReqCross int)
	ReceivedHash(topic string, hash []byte)
	ProcessedHash(topic string, hash []byte, err error)
	FailedToResolveData(topic string, hash []byte, err error)
	Enabled() bool
	Query(topic string) []string
	IsInterfaceNil() bool
}
