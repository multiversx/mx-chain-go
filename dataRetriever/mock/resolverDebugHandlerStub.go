package mock

// ResolverDebugHandler -
type ResolverDebugHandler struct {
	RequestedDataCalled       func(topic string, hash []byte, numReqIntra int, numReqCross int)
	FailedToResolveDataCalled func(topic string, hash []byte, err error)
	EnabledCalled             func() bool
}

// RequestedData -
func (rdh *ResolverDebugHandler) RequestedData(topic string, hash []byte, numReqIntra int, numReqCross int) {
	if rdh.RequestedDataCalled != nil {
		rdh.RequestedDataCalled(topic, hash, numReqIntra, numReqCross)
	}
}

// FailedToResolveData -
func (rdh *ResolverDebugHandler) FailedToResolveData(topic string, hash []byte, err error) {
	if rdh.FailedToResolveDataCalled != nil {
		rdh.FailedToResolveDataCalled(topic, hash, err)
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
