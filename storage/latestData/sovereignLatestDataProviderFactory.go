package latestData

import "github.com/multiversx/mx-chain-go/storage"

type sovereignLatestDataProviderFactory struct {
}

func NewSovereignLatestDataProviderFactory() *sovereignLatestDataProviderFactory {
	return &sovereignLatestDataProviderFactory{}
}

func (f *sovereignLatestDataProviderFactory) CreateLatestDataProvider(args ArgsLatestDataProvider) (storage.LatestStorageDataProviderHandler, error) {
	return NewSovereignLatestDataProvider(args)
}

func (f *sovereignLatestDataProviderFactory) IsInterfaceNil() bool {
	return f == nil
}
