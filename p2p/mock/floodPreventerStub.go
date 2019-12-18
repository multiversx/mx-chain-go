package mock

type FloodPreventerStub struct {
	IncrementAddingToSumCalled func(identifier string, size uint64) bool
	IncrementCalled            func(identifier string, size uint64) bool
	ResetCalled                func()
}

func (fps *FloodPreventerStub) IncrementAddingToSum(identifier string, size uint64) bool {
	return fps.IncrementAddingToSumCalled(identifier, size)
}

func (fps *FloodPreventerStub) Increment(identifier string, size uint64) bool {
	return fps.IncrementCalled(identifier, size)
}

func (fps *FloodPreventerStub) Reset() {
	fps.ResetCalled()
}

func (fps *FloodPreventerStub) IsInterfaceNil() bool {
	return fps == nil
}
