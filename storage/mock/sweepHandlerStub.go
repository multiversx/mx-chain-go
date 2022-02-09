package mock

// SweepHandlerStub -
type SweepHandlerStub struct {
	OnSweepCalled func(key []byte)
}

// OnSweep -
func (sh *SweepHandlerStub) OnSweep(key []byte) {
	if sh.OnSweepCalled != nil {
		sh.OnSweepCalled(key)
	}
}
