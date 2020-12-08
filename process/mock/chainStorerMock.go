package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/pkg/errors"
)

// ChainStorerMock is a mock implementation of the ChainStorer interface
type ChainStorerMock struct {
	AddStorerCalled func(key dataRetriever.UnitType, s storage.Storer)
	GetStorerCalled func(unitType dataRetriever.UnitType) storage.Storer
	HasCalled       func(unitType dataRetriever.UnitType, key []byte) error
	GetCalled       func(unitType dataRetriever.UnitType, key []byte) ([]byte, error)
	PutCalled       func(unitType dataRetriever.UnitType, key []byte, value []byte) error
	GetAllCalled    func(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error)
	DestroyCalled   func() error
	CloseAllCalled  func() error
}

// CloseAll -
func (csm *ChainStorerMock) CloseAll() error {
	if csm.CloseAllCalled != nil {
		return csm.CloseAllCalled()
	}

	return nil
}

// AddStorer will add a new storer to the chain map
func (csm *ChainStorerMock) AddStorer(key dataRetriever.UnitType, s storage.Storer) {
	if csm.AddStorerCalled != nil {
		csm.AddStorerCalled(key, s)
	}
}

// GetStorer returns the storer from the chain map or nil if the storer was not found
func (csm *ChainStorerMock) GetStorer(unitType dataRetriever.UnitType) storage.Storer {
	if csm.GetStorerCalled != nil {
		return csm.GetStorerCalled(unitType)
	}
	return &StorerMock{}
}

// Has returns true if the key is found in the selected Unit or false otherwise
// It can return an error if the provided unit type is not supported or if the
// underlying implementation of the storage unit reports an error.
func (csm *ChainStorerMock) Has(unitType dataRetriever.UnitType, key []byte) error {
	if csm.HasCalled != nil {
		return csm.HasCalled(unitType, key)
	}
	return errors.New("Key not found")
}

// Get returns the value for the given key if found in the selected storage unit,
// nil otherwise. It can return an error if the provided unit type is not supported
// or if the storage unit underlying implementation reports an error
func (csm *ChainStorerMock) Get(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
	if csm.GetCalled != nil {
		return csm.GetCalled(unitType, key)
	}
	return nil, nil
}

// Put stores the key, value pair in the selected storage unit
// It can return an error if the provided unit type is not supported
// or if the storage unit underlying implementation reports an error
func (csm *ChainStorerMock) Put(unitType dataRetriever.UnitType, key []byte, value []byte) error {
	if csm.PutCalled != nil {
		return csm.PutCalled(unitType, key, value)
	}
	return nil
}

// GetAll gets all the elements with keys in the keys array, from the selected storage unit
// It can report an error if the provided unit type is not supported, if there is a missing
// key in the unit, or if the underlying implementation of the storage unit reports an error.
func (csm *ChainStorerMock) GetAll(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
	if csm.GetAllCalled != nil {
		return csm.GetAllCalled(unitType, keys)
	}
	return nil, nil
}

// SetEpochForPutOperation won't do anything
func (csm *ChainStorerMock) SetEpochForPutOperation(_ uint32) {
}

// Destroy removes the underlying files/resources used by the storage service
func (csm *ChainStorerMock) Destroy() error {
	if csm.DestroyCalled != nil {
		return csm.DestroyCalled()
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (csm *ChainStorerMock) IsInterfaceNil() bool {
	return csm == nil
}
