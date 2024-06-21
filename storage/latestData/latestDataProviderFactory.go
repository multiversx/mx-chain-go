package latestData

import "github.com/multiversx/mx-chain-go/storage"

type latestDataProviderFactory struct {
}

func NewLatestDataProviderFactory() *latestDataProviderFactory {
	return &latestDataProviderFactory{}
}

func (f *latestDataProviderFactory) CreateLatestDataProvider(args ArgsLatestDataProvider) (storage.LatestStorageDataProviderHandler, error) {
	return NewLatestDataProvider(args)
}

func (f *latestDataProviderFactory) IsInterfaceNil() bool {
	return f == nil
}
