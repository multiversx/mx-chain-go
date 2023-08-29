package factory

import "github.com/multiversx/mx-chain-go/dataRetriever"

type shardAdditionalStorageServiceFactory struct {
}

func NewShardAdditionalStorageServiceFactory() (*shardAdditionalStorageServiceFactory, error) {
	return &shardAdditionalStorageServiceFactory{}, nil
}

// CreateAdditionalStorageService creates an additional storage service for a normal shard
func (s *shardAdditionalStorageServiceFactory) CreateAdditionalStorageService(_ func(store dataRetriever.StorageService, shardID string) error, _ dataRetriever.StorageService, _ string) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *shardAdditionalStorageServiceFactory) IsInterfaceNil() bool {
	return s == nil
}
