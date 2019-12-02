package mock

type RequestedItemsHandlerStub struct {
	AddCalled   func(key string, withSweep bool) error
	HasCalled   func(key string, withSweep bool) bool
	SweepCalled func()
}

func (rihs *RequestedItemsHandlerStub) Add(key string, withSweep bool) error {
	return rihs.AddCalled(key, withSweep)
}

func (rihs *RequestedItemsHandlerStub) Has(key string, withSweep bool) bool {
	return rihs.HasCalled(key, withSweep)
}

func (rihs *RequestedItemsHandlerStub) Sweep() {
	rihs.SweepCalled()
}

func (rihs *RequestedItemsHandlerStub) IsInterfaceNil() bool {
	return rihs == nil
}
