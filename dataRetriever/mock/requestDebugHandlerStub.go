package mock

// RequestDebugHandlerStub -
type RequestDebugHandlerStub struct {
	RequestedDataCalled func(hash []byte, topic string, numReqIntra int, numReqCross int)
	EnabledCalled       func() bool
}

// RequestedData -
func (rdhs *RequestDebugHandlerStub) RequestedData(topic string, hash []byte, numReqIntra int, numReqCross int) {
	if rdhs.RequestedDataCalled != nil {
		rdhs.RequestedDataCalled(hash, topic, numReqIntra, numReqCross)
	}
}

// Enabled -
func (rdhs *RequestDebugHandlerStub) Enabled() bool {
	if rdhs.EnabledCalled != nil {
		return rdhs.EnabledCalled()
	}

	return false
}

// IsInterfaceNil -
func (rdhs *RequestDebugHandlerStub) IsInterfaceNil() bool {
	return rdhs == nil
}
