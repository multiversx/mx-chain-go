package mock

type RequestedItemsHandlerStub struct {
	AddCalled   func(key string) error
	HasCalled   func(key string) bool
	SweepCalled func()
}

func (rihs *RequestedItemsHandlerStub) Add(key string) error {
	return rihs.AddCalled(key)
}

func (rihs *RequestedItemsHandlerStub) Has(key string) bool {
	return rihs.HasCalled(key)
}

func (rihs *RequestedItemsHandlerStub) Sweep() {
	if rihs.SweepCalled == nil {
		return
	}

	rihs.SweepCalled()
}

func (rihs *RequestedItemsHandlerStub) IsInterfaceNil() bool {
	return rihs == nil
}
