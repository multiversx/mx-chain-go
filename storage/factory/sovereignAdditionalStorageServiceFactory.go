package factory

import "github.com/multiversx/mx-chain-go/dataRetriever"

type SovereignAdditionalStorageServiceFactory struct {
}

func (s *SovereignAdditionalStorageServiceFactory) CreateAdditionalStorageService(f func(store dataRetriever.StorageService, shardID string) error, store dataRetriever.StorageService, shardID string) error {
	return f(store, shardID)
}

func (s *SovereignAdditionalStorageServiceFactory) IsInterfaceNil() bool {
	return s == nil
}
