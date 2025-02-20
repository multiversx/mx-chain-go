package latestData

import "github.com/multiversx/mx-chain-go/storage"

type latestDataProviderFactory struct {
}

// NewLatestDataProviderFactory creates a latest data provider factory for normal run type chain
func NewLatestDataProviderFactory() *latestDataProviderFactory {
	return &latestDataProviderFactory{}
}

// CreateLatestDataProvider creates a latest data provider for normal run type chain
func (f *latestDataProviderFactory) CreateLatestDataProvider(args ArgsLatestDataProvider) (storage.LatestStorageDataProviderHandler, error) {
	return NewLatestDataProvider(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *latestDataProviderFactory) IsInterfaceNil() bool {
	return f == nil
}
