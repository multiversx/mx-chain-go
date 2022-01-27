package slashMocks

// RoundHashCacheStub -
type RoundHashCacheStub struct {
	AddCalled func(round uint64, hash []byte) error
}

// Add -
func (rhc *RoundHashCacheStub) Add(round uint64, hash []byte) error {
	if rhc.AddCalled != nil {
		return rhc.AddCalled(round, hash)
	}
	return nil
}

// Remove -
func (rhc *RoundHashCacheStub) Remove(_ uint64, _ []byte) {
}

// IsInterfaceNil -
func (rhc *RoundHashCacheStub) IsInterfaceNil() bool {
	return rhc == nil
}
