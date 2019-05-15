package mock

type HasherStub struct {
	ComputeCalled   func(s string) []byte
	EmptyHashCalled func() []byte
	SizeCalled      func() int
}

// Compute will output the SHA's equivalent of the input string
func (hs *HasherStub) Compute(s string) []byte {
	return hs.ComputeCalled(s)
}

// EmptyHash will return the equivalent of empty string SHA's
func (hs *HasherStub) EmptyHash() []byte {
	return hs.EmptyHashCalled()
}

// Size returns the required size in bytes
func (hs *HasherStub) Size() int {
	return hs.SizeCalled()
}
