package mock

// ResolverDebugHandler -
type ResolverDebugHandler struct {
	LogRequestedDataCalled       func(topic string, hash []byte, numReqIntra int, numReqCross int)
	LogFailedToResolveDataCalled func(topic string, hash []byte, err error)
	EnabledCalled                func() bool
}

// LogRequestedData -
func (rdh *ResolverDebugHandler) LogRequestedData(topic string, hash []byte, numReqIntra int, numReqCross int) {
	if rdh.LogRequestedDataCalled != nil {
		rdh.LogRequestedDataCalled(topic, hash, numReqIntra, numReqCross)
	}
}

// LogFailedToResolveData -
func (rdh *ResolverDebugHandler) LogFailedToResolveData(topic string, hash []byte, err error) {
	if rdh.LogFailedToResolveDataCalled != nil {
		rdh.LogFailedToResolveDataCalled(topic, hash, err)
	}
}

// Enabled -
func (rdh *ResolverDebugHandler) Enabled() bool {
	if rdh.EnabledCalled != nil {
		return rdh.EnabledCalled()
	}

	return false
}

// IsInterfaceNil -
func (rdh *ResolverDebugHandler) IsInterfaceNil() bool {
	return rdh == nil
}
