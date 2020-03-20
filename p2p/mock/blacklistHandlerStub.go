package mock

type BlacklistHandlerStub struct {
	HasCalled func(key string) bool
}

func (bhs *BlacklistHandlerStub) Has(key string) bool {
	return bhs.HasCalled(key)
}

func (bhs *BlacklistHandlerStub) IsInterfaceNil() bool {
	return bhs == nil
}
