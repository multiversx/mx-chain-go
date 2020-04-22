package mock

import "github.com/ElrondNetwork/elrond-go/storage"

// LatestStorageDataProviderStub -
type LatestStorageDataProviderStub struct {
	GetParentDirAndLastEpochCalled func() (string, uint32, error)
	GetCalled                      func() (storage.LatestDataFromStorage, error)
	GetShardsFromDirectoryCalled   func(path string) ([]string, error)
}

// GetParentDirAndLastEpoch -
func (lsdps *LatestStorageDataProviderStub) GetParentDirAndLastEpoch() (string, uint32, error) {
	if lsdps.GetParentDirAndLastEpochCalled != nil {
		return lsdps.GetParentDirAndLastEpochCalled()
	}

	return "", 0, nil
}

// Get -
func (lsdps *LatestStorageDataProviderStub) Get() (storage.LatestDataFromStorage, error) {
	if lsdps.GetCalled != nil {
		return lsdps.GetCalled()
	}

	return storage.LatestDataFromStorage{}, nil
}

// GetShardsFromDirectory -
func (lsdps *LatestStorageDataProviderStub) GetShardsFromDirectory(path string) ([]string, error) {
	if lsdps.GetShardsFromDirectoryCalled != nil {
		return lsdps.GetShardsFromDirectoryCalled(path)
	}

	return nil, nil
}

// IsInterfaceNil --
func (lsdps *LatestStorageDataProviderStub) IsInterfaceNil() bool {
	return lsdps == nil
}
