package factory

import "github.com/multiversx/mx-chain-go/dataRetriever"

// AdditionalStorageServiceFactoryStub -
type AdditionalStorageServiceFactoryStub struct {
	CreateAdditionalStorageUnitsCalled func(func(store dataRetriever.StorageService, shardID string) error, dataRetriever.StorageService, string) error
}

// NewAdditionalStorageServiceFactoryStub -
func NewAdditionalStorageServiceFactoryStub() *AdditionalStorageServiceFactoryStub {
	return &AdditionalStorageServiceFactoryStub{}
}

// CreateAdditionalStorageUnits -
func (a *AdditionalStorageServiceFactoryStub) CreateAdditionalStorageUnits(func(store dataRetriever.StorageService, shardID string) error, dataRetriever.StorageService, string) error {
	return nil
}

// IsInterfaceNil -
func (a *AdditionalStorageServiceFactoryStub) IsInterfaceNil() bool {
	return false
}
