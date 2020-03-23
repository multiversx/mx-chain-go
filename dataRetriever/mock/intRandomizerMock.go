package mock

// IntRandomizerMock -
type IntRandomizerMock struct {
	IntnCalled func(n int) int
}

// Intn -
func (irm *IntRandomizerMock) Intn(n int) int {
	if irm.IntnCalled != nil {
		return irm.IntnCalled(n)
	}

	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (irm *IntRandomizerMock) IsInterfaceNil() bool {
	return irm == nil
}
