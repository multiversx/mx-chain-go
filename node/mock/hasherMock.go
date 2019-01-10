package mock

type HasherMock struct {
}

func (hash HasherMock) Compute(s string) []byte {
	return nil
}
func (hash HasherMock) EmptyHash() []byte {
	return nil
}
func (HasherMock) Size() int {
	return 0
}
