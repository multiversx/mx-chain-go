package mock

type StorageBootstrapperMock struct {
	LoadFromStorageCalled func() error
}

func (sbm *StorageBootstrapperMock) LoadFromStorage() error {
	if sbm.LoadFromStorageCalled == nil {
		return nil
	}
	return sbm.LoadFromStorageCalled()
}

func (sbm *StorageBootstrapperMock) IsInterfaceNil() bool {
	return false
}
