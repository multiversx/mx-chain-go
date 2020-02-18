package mock

// IntRandomizerMock -
type IntRandomizerMock struct {
	IntnCalled func(n int) (int, error)
}

// Intn -
func (irm *IntRandomizerMock) Intn(n int) (int, error) {
	return irm.IntnCalled(n)
}

// IsInterfaceNil returns true if there is no value under the interface
func (irm *IntRandomizerMock) IsInterfaceNil() bool {
	return irm == nil
}
