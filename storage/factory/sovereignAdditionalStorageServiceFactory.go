package factory

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/errors"
)

type sovereignAdditionalStorageServiceFactory struct {
}

// NewSovereignAdditionalStorageServiceFactory creates a new instance of sovereignAdditionalStorageServiceFactory
func NewSovereignAdditionalStorageServiceFactory() (*sovereignAdditionalStorageServiceFactory, error) {
	return &sovereignAdditionalStorageServiceFactory{}, nil
}

// CreateAdditionalStorageUnits creates additional storage units for a sovereign shard
func (s *sovereignAdditionalStorageServiceFactory) CreateAdditionalStorageUnits(f func(store dataRetriever.StorageService, shardID string) error, store dataRetriever.StorageService, shardID string) error {
	if f == nil {
		return errors.ErrNilFunction
	}
	return f(store, shardID)
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *sovereignAdditionalStorageServiceFactory) IsInterfaceNil() bool {
	return s == nil
}
