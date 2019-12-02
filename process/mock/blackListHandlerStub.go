package mock

type BlackListHandlerStub struct {
	AddCalled   func(key string, withSweep bool) error
	HasCalled   func(key string, withSweep bool) bool
	SweepCalled func()
}

func (blhs *BlackListHandlerStub) Add(key string, withSweep bool) error {
	return blhs.AddCalled(key, withSweep)
}

func (blhs *BlackListHandlerStub) Has(key string, withSweep bool) bool {
	return blhs.HasCalled(key, withSweep)
}

func (blhs *BlackListHandlerStub) Sweep() {
	blhs.SweepCalled()
}

func (blhs *BlackListHandlerStub) IsInterfaceNil() bool {
	return blhs == nil
}
