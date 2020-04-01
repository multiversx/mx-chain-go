package mock

// InterceptedDebugHandlerStub -
type InterceptedDebugHandlerStub struct {
	ReceivedHashCalled  func(topic string, hash []byte)
	ProcessedHashCalled func(topic string, hash []byte, err error)
	EnabledCalled       func() bool
}

// ReceivedHash -
func (idhs *InterceptedDebugHandlerStub) ReceivedHash(topic string, hash []byte) {
	if idhs.ReceivedHashCalled != nil {
		idhs.ReceivedHashCalled(topic, hash)
	}
}

// ProcessedHash -
func (idhs *InterceptedDebugHandlerStub) ProcessedHash(topic string, hash []byte, err error) {
	if idhs.ProcessedHashCalled != nil {
		idhs.ProcessedHashCalled(topic, hash, err)
	}
}

// Enabled -
func (idhs *InterceptedDebugHandlerStub) Enabled() bool {
	if idhs.EnabledCalled != nil {
		return idhs.EnabledCalled()
	}

	return false
}

// IsInterfaceNil -
func (idhs *InterceptedDebugHandlerStub) IsInterfaceNil() bool {
	return idhs == nil
}
