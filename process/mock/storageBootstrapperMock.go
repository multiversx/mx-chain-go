package mock

// StorageBootstrapperMock -
type StorageBootstrapperMock struct {
	LoadFromStorageCalled func() error
}

// LoadFromStorage -
func (sbm *StorageBootstrapperMock) LoadFromStorage() error {
	if sbm.LoadFromStorageCalled != nil {
		return sbm.LoadFromStorageCalled()
	}
	return nil
}

// GetHighestBlockNonce -
func (sbm *StorageBootstrapperMock) GetHighestBlockNonce() uint64 {
	return 0
}

// IsInterfaceNil -
func (sbm *StorageBootstrapperMock) IsInterfaceNil() bool {
	return sbm == nil
}
