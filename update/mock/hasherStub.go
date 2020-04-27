package mock

// HasherStub -
type HasherStub struct {
	ComputeCalled   func(s string) []byte
	EmptyHashCalled func() []byte
}

// Compute -
func (hash *HasherStub) Compute(s string) []byte {
	if hash.ComputeCalled != nil {
		return hash.ComputeCalled(s)
	}
	return nil
}

// EmptyHash -
func (hash *HasherStub) EmptyHash() []byte {
	if hash.EmptyHashCalled != nil {
		hash.EmptyHashCalled()
	}
	return nil
}

// Size -
func (hash *HasherStub) Size() int {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (hash *HasherStub) IsInterfaceNil() bool {
	return hash == nil
}
