package consensus

// InvalidSignersCacheMock -
type InvalidSignersCacheMock struct {
	AddInvalidSignersCalled func(hash string)
	HasInvalidSignersCalled func(hash string) bool
	ResetCalled             func()
}

// AddInvalidSigners -
func (mock *InvalidSignersCacheMock) AddInvalidSigners(hash string) {
	if mock.AddInvalidSignersCalled != nil {
		mock.AddInvalidSignersCalled(hash)
	}
}

// HasInvalidSigners -
func (mock *InvalidSignersCacheMock) HasInvalidSigners(hash string) bool {
	if mock.HasInvalidSignersCalled != nil {
		return mock.HasInvalidSignersCalled(hash)
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
