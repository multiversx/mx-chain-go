package mock

type BlackListHandlerStub struct {
	AddCalled   func(key string) error
	HasCalled   func(key string) bool
	SweepCalled func()
}

func (blhs *BlackListHandlerStub) Add(key string) error {
	if blhs.AddCalled == nil {
		return nil
	}

	return blhs.AddCalled(key)
}

func (blhs *BlackListHandlerStub) Has(key string) bool {
	if blhs.HasCalled == nil {
		return false
	}

	return blhs.HasCalled(key)
}

func (blhs *BlackListHandlerStub) Sweep() {
	if blhs.SweepCalled == nil {
		return
	}

	blhs.SweepCalled()
}

func (blhs *BlackListHandlerStub) IsInterfaceNil() bool {
	return blhs == nil
}
