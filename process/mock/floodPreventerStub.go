package mock

// FloodPreventerStub -
type FloodPreventerStub struct {
	IncreaseLoadCalled       func(identifier string, size uint64) error
	ApplyConsensusSizeCalled func(size int)
	ResetCalled              func()
}

// IncreaseLoad -
func (fps *FloodPreventerStub) IncreaseLoad(identifier string, size uint64) error {
	return fps.IncreaseLoadCalled(identifier, size)
}

// ApplyConsensusSize -
func (fps *FloodPreventerStub) ApplyConsensusSize(size int) {
	if fps.ApplyConsensusSizeCalled != nil {
		fps.ApplyConsensusSizeCalled(size)
	}
}

// Reset -
func (fps *FloodPreventerStub) Reset() {
	fps.ResetCalled()
}

// IsInterfaceNil -
func (fps *FloodPreventerStub) IsInterfaceNil() bool {
	return fps == nil
}
