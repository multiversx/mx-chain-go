package factory

import "github.com/multiversx/mx-chain-go/dataRetriever"

type sovereignAdditionalStorageServiceFactory struct {
}

// NewSovereignAdditionalStorageServiceFactory creates a new instance of sovereignAdditionalStorageServiceFactory
func NewSovereignAdditionalStorageServiceFactory() (*sovereignAdditionalStorageServiceFactory, error) {
	return &sovereignAdditionalStorageServiceFactory{}, nil
}

// CreateAdditionalStorageService creates an additional storage service for a sovereign shard
func (s *sovereignAdditionalStorageServiceFactory) CreateAdditionalStorageService(f func(store dataRetriever.StorageService, shardID string) error, store dataRetriever.StorageService, shardID string) error {
	return f(store, shardID)
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *sovereignAdditionalStorageServiceFactory) IsInterfaceNil() bool {
	return s == nil
}
