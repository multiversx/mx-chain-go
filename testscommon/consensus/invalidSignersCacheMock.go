package consensus

// InvalidSignersCacheMock -
type InvalidSignersCacheMock struct {
	AddInvalidSignersCalled        func(headerHash []byte, invalidSigners []byte, invalidPublicKeys []string)
	CheckKnownInvalidSignersCalled func(headerHash []byte, invalidSigners []byte) bool
	ResetCalled                    func()
}

// AddInvalidSigners -
func (mock *InvalidSignersCacheMock) AddInvalidSigners(headerHash []byte, invalidSigners []byte, invalidPublicKeys []string) {
	if mock.AddInvalidSignersCalled != nil {
		mock.AddInvalidSignersCalled(headerHash, invalidSigners, invalidPublicKeys)
	}
}

// CheckKnownInvalidSigners -
func (mock *InvalidSignersCacheMock) CheckKnownInvalidSigners(headerHash []byte, invalidSigners []byte) bool {
	if mock.CheckKnownInvalidSignersCalled != nil {
		return mock.CheckKnownInvalidSignersCalled(headerHash, invalidSigners)
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
