package factory

import "github.com/multiversx/mx-chain-go/dataRetriever"

// AdditionalStorageServiceFactoryStub is a mock for AdditionalStorageServiceFactory
type AdditionalStorageServiceFactoryStub struct {
	WorkAsSovereign bool
}

// CreateAdditionalStorageUnits -
func (asf *AdditionalStorageServiceFactoryStub) CreateAdditionalStorageUnits(f func(store dataRetriever.StorageService, shardID string) error, store dataRetriever.StorageService, shardID string) error {
	if asf.WorkAsSovereign {
		return f(store, shardID)
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (asf *AdditionalStorageServiceFactoryStub) IsInterfaceNil() bool {
	return false
}
