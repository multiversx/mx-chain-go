package mock

type FloodPreventerStub struct {
	AccumulateGlobalCalled func(identifier string, size uint64) bool
	AccumulateCalled       func(identifier string, size uint64) bool
	ResetCalled            func()
}

func (fps *FloodPreventerStub) AccumulateGlobal(identifier string, size uint64) bool {
	return fps.AccumulateGlobalCalled(identifier, size)
}

func (fps *FloodPreventerStub) Accumulate(identifier string, size uint64) bool {
	return fps.AccumulateCalled(identifier, size)
}

func (fps *FloodPreventerStub) Reset() {
	fps.ResetCalled()
}

func (fps *FloodPreventerStub) IsInterfaceNil() bool {
	return fps == nil
}
