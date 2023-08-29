package testscommon

import "github.com/multiversx/mx-chain-go/dataRetriever"

// AdditionalStorageServiceFactoryMock is a mock for AdditionalStorageServiceFactory
type AdditionalStorageServiceFactoryMock struct {
	ExecuteFunction bool
}

// CreateAdditionalStorageService creates an additional storage service for a shard
func (psf *AdditionalStorageServiceFactoryMock) CreateAdditionalStorageService(f func(store dataRetriever.StorageService, shardID string) error, store dataRetriever.StorageService, shardID string) error {
	if psf.ExecuteFunction {
		return f(store, shardID)
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (psf *AdditionalStorageServiceFactoryMock) IsInterfaceNil() bool {
	return false
}
