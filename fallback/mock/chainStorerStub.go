package mock

import (
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ChainStorerStub -
type ChainStorerStub struct {
	AddStorerCalled     func(key dataRetriever.UnitType, s storage.Storer)
	GetStorerCalled     func(unitType dataRetriever.UnitType) storage.Storer
	HasCalled           func(unitType dataRetriever.UnitType, key []byte, priority common.StorageAccessType) error
	GetCalled           func(unitType dataRetriever.UnitType, key []byte, priority common.StorageAccessType) ([]byte, error)
	PutCalled           func(unitType dataRetriever.UnitType, key []byte, value []byte, priority common.StorageAccessType) error
	GetAllCalled        func(unitType dataRetriever.UnitType, keys [][]byte, priority common.StorageAccessType) (map[string][]byte, error)
	GetAllStorersCalled func() map[dataRetriever.UnitType]storage.Storer
	DestroyCalled       func() error
	CloseAllCalled      func() error
}

// CloseAll -
func (css *ChainStorerStub) CloseAll() error {
	if css.CloseAllCalled != nil {
		return css.CloseAllCalled()
	}
	return nil
}

// AddStorer -
func (css *ChainStorerStub) AddStorer(key dataRetriever.UnitType, s storage.Storer) {
	if css.AddStorerCalled != nil {
		css.AddStorerCalled(key, s)
	}
}

// GetStorer -
func (css *ChainStorerStub) GetStorer(unitType dataRetriever.UnitType) storage.Storer {
	if css.GetStorerCalled != nil {
		return css.GetStorerCalled(unitType)
	}
	return nil
}

// Has -
func (css *ChainStorerStub) Has(unitType dataRetriever.UnitType, key []byte, priority common.StorageAccessType) error {
	if css.HasCalled != nil {
		return css.HasCalled(unitType, key, priority)
	}
	return nil
}

// Get -
func (css *ChainStorerStub) Get(unitType dataRetriever.UnitType, key []byte, priority common.StorageAccessType) ([]byte, error) {
	if css.GetCalled != nil {
		return css.GetCalled(unitType, key, priority)
	}
	return nil, nil
}

// Put -
func (css *ChainStorerStub) Put(unitType dataRetriever.UnitType, key []byte, value []byte, priority common.StorageAccessType) error {
	if css.PutCalled != nil {
		return css.PutCalled(unitType, key, value, priority)
	}
	return nil
}

// GetAll -
func (css *ChainStorerStub) GetAll(unitType dataRetriever.UnitType, keys [][]byte, priority common.StorageAccessType) (map[string][]byte, error) {
	if css.GetAllCalled != nil {
		return css.GetAllCalled(unitType, keys, priority)
	}
	return nil, nil
}

// GetAllStorers -
func (css *ChainStorerStub) GetAllStorers() map[dataRetriever.UnitType]storage.Storer {
	if css.GetAllStorersCalled != nil {
		return css.GetAllStorersCalled()
	}

	return nil
}

// SetEpochForPutOperation -
func (css *ChainStorerStub) SetEpochForPutOperation(_ uint32) {
}

// Destroy -
func (css *ChainStorerStub) Destroy() error {
	if css.DestroyCalled != nil {
		return css.DestroyCalled()
	}
	return nil
}

// IsInterfaceNil -
func (css *ChainStorerStub) IsInterfaceNil() bool {
	return css == nil
}
