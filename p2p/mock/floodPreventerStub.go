package mock

type FloodPreventerStub struct {
	IncrementCalled func(identifier string, size uint64) bool
	ResetCalled     func()
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
