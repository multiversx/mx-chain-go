package factory

import "github.com/multiversx/mx-chain-go/dataRetriever"

type shardAdditionalStorageServiceFactory struct {
}

// NewShardAdditionalStorageServiceFactory creates a new instance of shardAdditionalStorageServiceFactory
func NewShardAdditionalStorageServiceFactory() (*shardAdditionalStorageServiceFactory, error) {
	return &shardAdditionalStorageServiceFactory{}, nil
}

// CreateAdditionalStorageUnits does nothing
func (s *shardAdditionalStorageServiceFactory) CreateAdditionalStorageUnits(_ func(store dataRetriever.StorageService, shardID string) error, _ dataRetriever.StorageService, _ string) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *shardAdditionalStorageServiceFactory) IsInterfaceNil() bool {
	return s == nil
}
