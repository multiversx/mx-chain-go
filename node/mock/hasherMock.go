package mock

// HasherMock -
type HasherMock struct {
	ComputeCalled   func(s string) []byte
	EmptyHashCalled func() []byte
}

// Compute -
func (hash HasherMock) Compute(s string) []byte {
	if hash.ComputeCalled != nil {
		return hash.ComputeCalled(s)
	}
	return nil
}

// EmptyHash -
func (hash HasherMock) EmptyHash() []byte {
	if hash.EmptyHashCalled != nil {
		hash.EmptyHashCalled()
	}
	return nil
}

// Size -
func (HasherMock) Size() int {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (hash HasherMock) IsInterfaceNil() bool {
	return false
}
