package mock

type HasherMock struct {
	ComputeCalled   func(s string) []byte
	EmptyHashCalled func() []byte
}

func (hash HasherMock) Compute(s string) []byte {
	if hash.ComputeCalled != nil {
		return hash.ComputeCalled(s)
	}
	return nil
}
func (hash HasherMock) EmptyHash() []byte {
	if hash.EmptyHashCalled != nil {
		hash.EmptyHashCalled()
	}
	return nil
}
func (HasherMock) Size() int {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (hash *HasherMock) IsInterfaceNil() bool {
	if hash == nil {
		return true
	}
	return false
}
