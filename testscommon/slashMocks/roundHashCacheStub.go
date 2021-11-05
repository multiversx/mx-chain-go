package slashMocks

// HeadersCacheStub -
type HeadersCacheStub struct {
	AddCalled func(round uint64, hash []byte) error
}

// Add -
func (hcs *HeadersCacheStub) Add(round uint64, hash []byte) error {
	if hcs.AddCalled != nil {
		return hcs.AddCalled(round, hash)
	}
	return nil
}

// Remove -
func (hcs *HeadersCacheStub) Remove(_ uint64, _ []byte) {
}

// IsInterfaceNil -
func (hcs *HeadersCacheStub) IsInterfaceNil() bool {
	return hcs == nil
}
