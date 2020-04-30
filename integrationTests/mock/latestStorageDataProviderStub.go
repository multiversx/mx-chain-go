package mock

import (
	"github.com/ElrondNetwork/elrond-go/storage"
)

// LatestStorageDataProviderStub -
type LatestStorageDataProviderStub struct {
	GetCalled                      func() (storage.LatestDataFromStorage, error)
	GetParentDirAndLastEpochCalled func() (string, uint32, error)
	GetShardsFromDirectoryCalled   func(path string) ([]string, error)
}

// GetParentDirAndLastEpoch -
func (l *LatestStorageDataProviderStub) GetParentDirAndLastEpoch() (string, uint32, error) {
	if l.GetParentDirAndLastEpochCalled != nil {
		return l.GetParentDirAndLastEpochCalled()
	}

	return "", 0, nil
}

// Get -
func (l *LatestStorageDataProviderStub) Get() (storage.LatestDataFromStorage, error) {
	if l.GetCalled != nil {
		return l.GetCalled()
	}

	return storage.LatestDataFromStorage{}, nil
}

// GetShardsFromDirectory --
func (l *LatestStorageDataProviderStub) GetShardsFromDirectory(path string) ([]string, error) {
	if l.GetShardsFromDirectoryCalled != nil {
		return l.GetShardsFromDirectoryCalled(path)
	}

	return nil, nil
}

// IsInterfaceNil --
func (l *LatestStorageDataProviderStub) IsInterfaceNil() bool {
	return l == nil
}
