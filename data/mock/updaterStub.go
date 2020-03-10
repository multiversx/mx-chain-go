package mock

// UpdaterStub -
type UpdaterStub struct {
	UpdateCalled func(key, value []byte) error
}

// Update -
func (updater *UpdaterStub) Update(key, value []byte) error {
	return updater.UpdateCalled(key, value)
}

// IsInterfaceNil returns true if there is no value under the interface
func (updater *UpdaterStub) IsInterfaceNil() bool {
	return updater == nil
}
