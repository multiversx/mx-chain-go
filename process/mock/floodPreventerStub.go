package mock

// FloodPreventerStub -
type FloodPreventerStub struct {
	AccumulateGlobalCalled func(identifier string, size uint64) bool
	AccumulateCalled       func(identifier string, size uint64) bool
	ResetCalled            func()
}

// AccumulateGlobal -
func (fps *FloodPreventerStub) AccumulateGlobal(identifier string, size uint64) bool {
	return fps.AccumulateGlobalCalled(identifier, size)
}

// Accumulate -
func (fps *FloodPreventerStub) Accumulate(identifier string, size uint64) bool {
	return fps.AccumulateCalled(identifier, size)
}

// Reset -
func (fps *FloodPreventerStub) Reset() {
	fps.ResetCalled()
}

// IsInterfaceNil -
func (fps *FloodPreventerStub) IsInterfaceNil() bool {
	return fps == nil
}
