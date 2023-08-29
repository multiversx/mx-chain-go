package factory

import "github.com/multiversx/mx-chain-go/dataRetriever"

// AdditionalStorageServiceFactoryStub -
type AdditionalStorageServiceFactoryStub struct {
	CreateAdditionalStorageServiceCalled func(func(store dataRetriever.StorageService, shardID string) error, dataRetriever.StorageService, string) error
}

// NewAdditionalStorageServiceFactoryStub -
func NewAdditionalStorageServiceFactoryStub() *AdditionalStorageServiceFactoryStub {
	return &AdditionalStorageServiceFactoryStub{}
}

// CreateAdditionalStorageService -
func (a *AdditionalStorageServiceFactoryStub) CreateAdditionalStorageService(func(store dataRetriever.StorageService, shardID string) error, dataRetriever.StorageService, string) error {
	return nil
}

// IsInterfaceNil -
func (a *AdditionalStorageServiceFactoryStub) IsInterfaceNil() bool {
	return false
}
