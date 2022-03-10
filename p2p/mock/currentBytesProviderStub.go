package mock

// CurrentBytesProviderStub -
type CurrentBytesProviderStub struct {
	BytesToSendToNewPeersCalled func() ([]byte, bool)
}

// BytesToSendToNewPeers -
func (stub *CurrentBytesProviderStub) BytesToSendToNewPeers() ([]byte, bool) {
	if stub.BytesToSendToNewPeersCalled != nil {
		return stub.BytesToSendToNewPeersCalled()
	}

	return make([]byte, 0), false
}

// IsInterfaceNil -
func (stub *CurrentBytesProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
