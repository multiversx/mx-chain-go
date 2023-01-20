package mock

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/storage"
)

// BootStrapDataProviderStub -
type BootStrapDataProviderStub struct {
	LoadForPathCalled func(persisterFactory storage.PersisterFactory, path string) (*bootstrapStorage.BootstrapData, storage.Storer, error)
	GetStorerCalled   func(storer storage.Storer) (process.BootStorer, error)
}

// LoadForPath -
func (b *BootStrapDataProviderStub) LoadForPath(persisterFactory storage.PersisterFactory, path string) (*bootstrapStorage.BootstrapData, storage.Storer, error) {
	if b.LoadForPathCalled != nil {
		return b.LoadForPathCalled(persisterFactory, path)
	}

	return &bootstrapStorage.BootstrapData{}, nil, nil
}

// GetStorer -
func (b *BootStrapDataProviderStub) GetStorer(storer storage.Storer) (process.BootStorer, error) {
	if b.GetStorerCalled != nil {
		return b.GetStorerCalled(storer)
	}

	return nil, nil
}

// IsInterfaceNil -
func (b *BootStrapDataProviderStub) IsInterfaceNil() bool {
	return b == nil
}
