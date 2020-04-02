package mock

// ResolverDebugHandlerStub -
type ResolverDebugHandlerStub struct {
	RequestedDataCalled       func(hash []byte, topic string, numReqIntra int, numReqCross int)
	FailedToResolveDataCalled func(topic string, hash []byte, err error)
	EnabledCalled             func() bool
}

// RequestedData -
func (rdhs *ResolverDebugHandlerStub) RequestedData(topic string, hash []byte, numReqIntra int, numReqCross int) {
	if rdhs.RequestedDataCalled != nil {
		rdhs.RequestedDataCalled(hash, topic, numReqIntra, numReqCross)
	}
}

// Enabled -
func (rdhs *ResolverDebugHandlerStub) Enabled() bool {
	if rdhs.EnabledCalled != nil {
		return rdhs.EnabledCalled()
	}

	return false
}

// FailedToResolveData -
func (rdhs *ResolverDebugHandlerStub) FailedToResolveData(topic string, hash []byte, err error) {
	if rdhs.FailedToResolveDataCalled != nil {
		rdhs.FailedToResolveDataCalled(topic, hash, err)
	}
}

// IsInterfaceNil -
func (rdhs *ResolverDebugHandlerStub) IsInterfaceNil() bool {
	return rdhs == nil
}
