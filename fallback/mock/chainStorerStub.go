package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ChainStorerStub -
type ChainStorerStub struct {
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
func (css *ChainStorerStub) Has(unitType dataRetriever.UnitType, key []byte) error {
	if css.HasCalled != nil {
		return css.HasCalled(unitType, key)
	}
	return nil
}

// Get -
func (css *ChainStorerStub) Get(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
	if css.GetCalled != nil {
		return css.GetCalled(unitType, key)
	}
	return nil, nil
}

// Put -
func (css *ChainStorerStub) Put(unitType dataRetriever.UnitType, key []byte, value []byte) error {
	if css.PutCalled != nil {
		return css.PutCalled(unitType, key, value)
	}
	return nil
}

// GetAll -
func (css *ChainStorerStub) GetAll(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
	if css.GetAllCalled != nil {
		return css.GetAllCalled(unitType, keys)
	}
	return nil, nil
}

// SetEpochForPutOperation -
func (css *ChainStorerStub) SetEpochForPutOperation(epoch uint32) {
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
