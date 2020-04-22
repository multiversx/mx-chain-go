package mock

// InterceptedDebugHandlerStub -
type InterceptedDebugHandlerStub struct {
	LogReceivedHashCalled  func(topic string, hash []byte)
	LogProcessedHashCalled func(topic string, hash []byte, err error)
	EnabledCalled          func() bool
}

// LogReceivedHash -
func (idhs *InterceptedDebugHandlerStub) LogReceivedHash(topic string, hash []byte) {
	if idhs.LogReceivedHashCalled != nil {
		idhs.LogReceivedHashCalled(topic, hash)
	}
}

// LogProcessedHash -
func (idhs *InterceptedDebugHandlerStub) LogProcessedHash(topic string, hash []byte, err error) {
	if idhs.LogProcessedHashCalled != nil {
		idhs.LogProcessedHashCalled(topic, hash, err)
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
