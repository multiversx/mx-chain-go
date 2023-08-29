package factory

import "github.com/multiversx/mx-chain-go/dataRetriever"

// ShardAdditionalStorageServiceFactory is the factory for creating additional storage service for shards
type ShardAdditionalStorageServiceFactory struct {
}

// CreateAdditionalStorageService creates an additional storage service for a shard
func (s *ShardAdditionalStorageServiceFactory) CreateAdditionalStorageService(_ func(store dataRetriever.StorageService, shardID string) error, _ dataRetriever.StorageService, _ string) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *ShardAdditionalStorageServiceFactory) IsInterfaceNil() bool {
	return s == nil
}
