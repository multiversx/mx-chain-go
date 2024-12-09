package testscommon

import "github.com/multiversx/mx-chain-go/dataRetriever"

// AdditionalStorageServiceFactoryMock is a mock for AdditionalStorageServiceFactory
type AdditionalStorageServiceFactoryMock struct {
	WorkAsSovereign bool
}

// CreateAdditionalStorageUnits -
func (asf *AdditionalStorageServiceFactoryMock) CreateAdditionalStorageUnits(f func(store dataRetriever.StorageService, shardID string) error, store dataRetriever.StorageService, shardID string) error {
	if asf.WorkAsSovereign {
		return f(store, shardID)
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (asf *AdditionalStorageServiceFactoryMock) IsInterfaceNil() bool {
	return false
}
