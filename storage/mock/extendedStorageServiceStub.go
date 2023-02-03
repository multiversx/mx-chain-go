package mock

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage"
)

// StorageListProviderStub -
type StorageListProviderStub struct {
	GetAllStorersCalled func() map[dataRetriever.UnitType]storage.Storer
}

// GetAllStorers -
func (slps *StorageListProviderStub) GetAllStorers() map[dataRetriever.UnitType]storage.Storer {
	if slps.GetAllStorersCalled != nil {
		return slps.GetAllStorersCalled()
	}

	return nil
}

// IsInterfaceNil -
func (slps *StorageListProviderStub) IsInterfaceNil() bool {
	return slps == nil
}
