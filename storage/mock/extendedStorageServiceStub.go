package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// StorageListProviderStub -
type StorageListProviderStub struct {
	GetAllStorersCalled func() map[dataRetriever.UnitType]storage.Storer
}

// GetAllStorers -
func (sis *StorageListProviderStub) GetAllStorers() map[dataRetriever.UnitType]storage.Storer {
	if sis.GetAllStorersCalled != nil {
		return sis.GetAllStorersCalled()
	}

	return nil
}

// IsInterfaceNil -
func (slps *StorageListProviderStub) IsInterfaceNil() bool {
	return slps == nil
}
