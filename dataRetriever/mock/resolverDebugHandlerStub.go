package mock

// ResolverDebugHandler -
type ResolverDebugHandler struct {
	LogRequestedDataCalled          func(topic string, hash [][]byte, numReqIntra int, numReqCross int)
	LogFailedToResolveDataCalled    func(topic string, hash []byte, err error)
	LogSucceededToResolveDataCalled func(topic string, hash []byte)
	EnabledCalled                   func() bool
}

// LogSucceededToResolveData -
func (rdh *ResolverDebugHandler) LogSucceededToResolveData(topic string, hash []byte) {
	if rdh.LogSucceededToResolveDataCalled != nil {
		rdh.LogSucceededToResolveDataCalled(topic, hash)
	}
}

// LogRequestedData -
func (rdh *ResolverDebugHandler) LogRequestedData(topic string, hashes [][]byte, numReqIntra int, numReqCross int) {
	if rdh.LogRequestedDataCalled != nil {
		rdh.LogRequestedDataCalled(topic, hashes, numReqIntra, numReqCross)
	}
}

// LogFailedToResolveData -
func (rdh *ResolverDebugHandler) LogFailedToResolveData(topic string, hash []byte, err error) {
	if rdh.LogFailedToResolveDataCalled != nil {
		rdh.LogFailedToResolveDataCalled(topic, hash, err)
	}
}

// IsInterfaceNil -
func (rdh *ResolverDebugHandler) IsInterfaceNil() bool {
	return rdh == nil
}
