package latestData

import "github.com/multiversx/mx-chain-go/storage"

type sovereignLatestDataProviderFactory struct {
}

// NewSovereignLatestDataProviderFactory creates a latest data provider factory for sovereign run type chain
func NewSovereignLatestDataProviderFactory() *sovereignLatestDataProviderFactory {
	return &sovereignLatestDataProviderFactory{}
}

// CreateLatestDataProvider creates a latest data provider for sovereign run type chain
func (f *sovereignLatestDataProviderFactory) CreateLatestDataProvider(args ArgsLatestDataProvider) (storage.LatestStorageDataProviderHandler, error) {
	return NewSovereignLatestDataProvider(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *sovereignLatestDataProviderFactory) IsInterfaceNil() bool {
	return f == nil
}
