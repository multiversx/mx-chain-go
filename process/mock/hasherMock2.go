package mock

type HasherMock2 struct {
	ComputeCalled   func(s string) []byte
	EmptyHashCalled func() []byte
}

func (hash HasherMock2) Compute(s string) []byte {
	if hash.ComputeCalled != nil {
		return hash.ComputeCalled(s)
	}
	return nil
}
func (hash HasherMock2) EmptyHash() []byte {
	if hash.EmptyHashCalled != nil {
		hash.EmptyHashCalled()
	}
	return nil
}
func (HasherMock2) Size() int {
	return 0
}
