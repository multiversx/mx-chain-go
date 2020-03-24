package mock

// IntRandomizerStub -
type IntRandomizerStub struct {
	IntnCalled func(n int) int
}

// Intn -
func (irs *IntRandomizerStub) Intn(n int) int {
	if irs.IntnCalled != nil {
		return irs.IntnCalled(n)
	}

	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (irs *IntRandomizerStub) IsInterfaceNil() bool {
	return irs == nil
}
