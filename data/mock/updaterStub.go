package mock

type UpdaterStub struct {
	UpdateCalled func(key, value []byte) error
}

func (updater *UpdaterStub) Update(key, value []byte) error {
	return updater.UpdateCalled(key, value)
}

// IsInterfaceNil returns true if there is no value under the interface
func (updater *UpdaterStub) IsInterfaceNil() bool {
	if updater == nil {
		return true
	}
	return false
}
