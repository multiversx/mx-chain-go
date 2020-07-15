package mock

// HardforkStorerStub -
type HardforkStorerStub struct {
	WriteCalled              func(identifier string, key []byte, value []byte) error
	FinishedIdentifierCalled func(identifier string) error
	RangeKeysCalled          func(handler func(identifier string, keys [][]byte) bool)
	GetCalled                func(identifier string, key []byte) ([]byte, error)
	CloseCalled              func() error
}

// Write -
func (hss *HardforkStorerStub) Write(identifier string, key []byte, value []byte) error {
	if hss.WriteCalled != nil {
		return hss.WriteCalled(identifier, key, value)
	}

	return nil
}

// FinishedIdentifier -
func (hss *HardforkStorerStub) FinishedIdentifier(identifier string) error {
	if hss.FinishedIdentifierCalled != nil {
		return hss.FinishedIdentifierCalled(identifier)
	}

	return nil
}

// RangeKeys -
func (hss *HardforkStorerStub) RangeKeys(handler func(identifier string, keys [][]byte) bool) {
	if hss.RangeKeysCalled != nil {
		hss.RangeKeysCalled(handler)
	}
}

// Get -
func (hss *HardforkStorerStub) Get(identifier string, key []byte) ([]byte, error) {
	if hss.GetCalled != nil {
		return hss.GetCalled(identifier, key)
	}

	return make([]byte, 0), nil
}

// Close -
func (hss *HardforkStorerStub) Close() error {
	if hss.CloseCalled != nil {
		return hss.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (hss *HardforkStorerStub) IsInterfaceNil() bool {
	return hss == nil
}
