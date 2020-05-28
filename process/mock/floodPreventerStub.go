package mock

// FloodPreventerStub -
type FloodPreventerStub struct {
	IncreaseLoadCalled func(identifier string, size uint64) error
	ResetCalled        func()
}

// IncreaseLoad -
func (fps *FloodPreventerStub) IncreaseLoad(identifier string, size uint64) error {
	return fps.IncreaseLoadCalled(identifier, size)
}

// Reset -
func (fps *FloodPreventerStub) Reset() {
	fps.ResetCalled()
}

// IsInterfaceNil -
func (fps *FloodPreventerStub) IsInterfaceNil() bool {
	return fps == nil
}
