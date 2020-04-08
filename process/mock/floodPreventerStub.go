package mock

// FloodPreventerStub -
type FloodPreventerStub struct {
	IncreaseLoadGlobalCalled func(identifier string, size uint64) error
	IncreaseLoadCalled       func(identifier string, size uint64) error
	ResetCalled              func()
}

// IncreaseLoadGlobal -
func (fps *FloodPreventerStub) IncreaseLoadGlobal(identifier string, size uint64) error {
	return fps.IncreaseLoadGlobalCalled(identifier, size)
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
