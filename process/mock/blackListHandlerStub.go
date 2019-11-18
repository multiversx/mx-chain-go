package mock

type BlackListHandlerStub struct {
	AddCalled func(key string) error
	HasCalled func(key string) bool
}

func (blhs *BlackListHandlerStub) Add(key string) error {
	return blhs.AddCalled(key)
}

func (blhs *BlackListHandlerStub) Has(key string) bool {
	return blhs.HasCalled(key)
}

func (blhs *BlackListHandlerStub) IsInterfaceNil() bool {
	return blhs == nil
}
