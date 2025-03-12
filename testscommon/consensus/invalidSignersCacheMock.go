package consensus

// InvalidSignersCacheMock -
type InvalidSignersCacheMock struct {
	AddInvalidSignersCalled func(headerHash []byte, invalidSigners []byte, invalidPublicKeys []string)
	HasInvalidSignersCalled func(headerHash []byte, invalidSigners []byte) bool
	ResetCalled             func()
}

// AddInvalidSigners -
func (mock *InvalidSignersCacheMock) AddInvalidSigners(headerHash []byte, invalidSigners []byte, invalidPublicKeys []string) {
	if mock.AddInvalidSignersCalled != nil {
		mock.AddInvalidSignersCalled(headerHash, invalidSigners, invalidPublicKeys)
	}
}

// HasInvalidSigners -
func (mock *InvalidSignersCacheMock) HasInvalidSigners(headerHash []byte, invalidSigners []byte) bool {
	if mock.HasInvalidSignersCalled != nil {
		return mock.HasInvalidSignersCalled(headerHash, invalidSigners)
	}
	return false
}

// Reset -
func (mock *InvalidSignersCacheMock) Reset() {
	if mock.ResetCalled != nil {
		mock.ResetCalled()
	}
}

// IsInterfaceNil -
func (mock *InvalidSignersCacheMock) IsInterfaceNil() bool {
	return mock == nil
}
