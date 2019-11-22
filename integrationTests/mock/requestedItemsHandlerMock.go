package mock

type RequestedItemsHandlerMock struct {
	AddCalled func(key string) error
	HasCalled func(key string) bool
}

func (rihm *RequestedItemsHandlerMock) Add(key string) error {
	return rihm.AddCalled(key)
}

func (rihm *RequestedItemsHandlerMock) Has(key string) bool {
	return rihm.HasCalled(key)
}

func (rihm *RequestedItemsHandlerMock) IsInterfaceNil() bool {
	return rihm == nil
}
