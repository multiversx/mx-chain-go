package mock

// InterceptedDebugHandlerStub -
type InterceptedDebugHandlerStub struct {
	LogReceivedHashesCalled  func(topic string, hashes [][]byte)
	LogProcessedHashesCalled func(topic string, hashes [][]byte, err error)
}

// LogReceivedHashes -
func (idhs *InterceptedDebugHandlerStub) LogReceivedHashes(topic string, hashes [][]byte) {
	if idhs.LogReceivedHashesCalled != nil {
		idhs.LogReceivedHashesCalled(topic, hashes)
	}
}

// LogProcessedHashes -
func (idhs *InterceptedDebugHandlerStub) LogProcessedHashes(topic string, hashes [][]byte, err error) {
	if idhs.LogProcessedHashesCalled != nil {
		idhs.LogProcessedHashesCalled(topic, hashes, err)
	}
}

// IsInterfaceNil -
func (idhs *InterceptedDebugHandlerStub) IsInterfaceNil() bool {
	return idhs == nil
}
