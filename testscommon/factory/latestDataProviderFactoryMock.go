package factory

import (
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/latestData"
	storageMock "github.com/multiversx/mx-chain-go/storage/mock"
)

// LatestDataProviderFactoryMock -
type LatestDataProviderFactoryMock struct {
	CreateLatestDataProviderCalled func(args latestData.ArgsLatestDataProvider) (storage.LatestStorageDataProviderHandler, error)
}

// CreateLatestDataProvider -
func (mock *LatestDataProviderFactoryMock) CreateLatestDataProvider(args latestData.ArgsLatestDataProvider) (storage.LatestStorageDataProviderHandler, error) {
	if mock.CreateLatestDataProviderCalled != nil {
		return mock.CreateLatestDataProviderCalled(args)
	}

	return &storageMock.LatestStorageDataProviderStub{}, nil
}

// IsInterfaceNil -
func (mock *LatestDataProviderFactoryMock) IsInterfaceNil() bool {
	return mock == nil
}
