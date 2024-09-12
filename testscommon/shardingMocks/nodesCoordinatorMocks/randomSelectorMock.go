package nodesCoordinatorMocks

// RandomSelectorMock is a mock for the RandomSelector interface
type RandomSelectorMock struct {
	SelectCalled func(randSeed []byte, sampleSize uint32) ([]uint32, error)
}

// Select calls the mocked method
func (rsm *RandomSelectorMock) Select(randSeed []byte, sampleSize uint32) ([]uint32, error) {
	if rsm.SelectCalled != nil {
		return rsm.SelectCalled(randSeed, sampleSize)
	}
	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rsm *RandomSelectorMock) IsInterfaceNil() bool {
	return rsm == nil
}
