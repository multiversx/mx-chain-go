package mock

type UpdaterStub struct {
	UpdateCalled func(key, value []byte) error
}

func (updater *UpdaterStub) Update(key, value []byte) error {
	return updater.UpdateCalled(key, value)
}
