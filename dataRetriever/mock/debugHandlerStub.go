package mock

// DebugHandler -
type DebugHandler struct {
	LogRequestedDataCalled          func(topic string, hash [][]byte, numReqIntra int, numReqCross int)
	LogFailedToResolveDataCalled    func(topic string, hash []byte, err error)
	LogSucceededToResolveDataCalled func(topic string, hash []byte)
	EnabledCalled                   func() bool
}

// LogSucceededToResolveData -
func (dh *DebugHandler) LogSucceededToResolveData(topic string, hash []byte) {
	if dh.LogSucceededToResolveDataCalled != nil {
		dh.LogSucceededToResolveDataCalled(topic, hash)
	}
}

// LogRequestedData -
func (dh *DebugHandler) LogRequestedData(topic string, hashes [][]byte, numReqIntra int, numReqCross int) {
	if dh.LogRequestedDataCalled != nil {
		dh.LogRequestedDataCalled(topic, hashes, numReqIntra, numReqCross)
	}
}

// LogFailedToResolveData -
func (dh *DebugHandler) LogFailedToResolveData(topic string, hash []byte, err error) {
	if dh.LogFailedToResolveDataCalled != nil {
		dh.LogFailedToResolveDataCalled(topic, hash, err)
	}
}

// IsInterfaceNil -
func (dh *DebugHandler) IsInterfaceNil() bool {
	return dh == nil
}
