package mock

// CurrentPayloadProviderStub -
type CurrentPayloadProviderStub struct {
	BytesToSendToNewPeersCalled func() ([]byte, bool)
}

// BytesToSendToNewPeers -
func (stub *CurrentPayloadProviderStub) BytesToSendToNewPeers() ([]byte, bool) {
	if stub.BytesToSendToNewPeersCalled != nil {
		return stub.BytesToSendToNewPeersCalled()
	}

	return make([]byte, 0), false
}

// IsInterfaceNil -
func (stub *CurrentPayloadProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
