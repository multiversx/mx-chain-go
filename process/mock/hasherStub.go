package mock

type HasherStub struct {
	ComputeCalled   func(s string) []byte
	EmptyHashCalled func() []byte
}

func (hash *HasherStub) Compute(s string) []byte {
	if hash.ComputeCalled != nil {
		return hash.ComputeCalled(s)
	}
	return nil
}
func (hash *HasherStub) EmptyHash() []byte {
	if hash.EmptyHashCalled != nil {
		hash.EmptyHashCalled()
	}
	return nil
}
func (hash *HasherStub) Size() int {
	return 0
}
