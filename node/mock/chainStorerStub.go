package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/pkg/errors"
)

// ChainStorerStub is a mock implementation of the ChainStorer interface
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
func (bc *ChainStorerStub) CloseAll() error {
	if bc.CloseAllCalled != nil {
		return bc.CloseAllCalled()
	}
	return nil
}

// AddStorer -
func (bc *ChainStorerStub) AddStorer(key dataRetriever.UnitType, s storage.Storer) {
	if bc.AddStorerCalled != nil {
		bc.AddStorerCalled(key, s)
	}
}

// GetStorer -
func (bc *ChainStorerStub) GetStorer(unitType dataRetriever.UnitType) storage.Storer {
	if bc.GetStorerCalled != nil {
		return bc.GetStorerCalled(unitType)
	}
	return nil
}

// Has -
func (bc *ChainStorerStub) Has(unitType dataRetriever.UnitType, key []byte) error {
	if bc.HasCalled != nil {
		return bc.HasCalled(unitType, key)
	}
	return errors.New("Key not found")
}

// Get -
func (bc *ChainStorerStub) Get(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
	if bc.GetCalled != nil {
		return bc.GetCalled(unitType, key)
	}
	return nil, nil
}

// Put -
func (bc *ChainStorerStub) Put(unitType dataRetriever.UnitType, key []byte, value []byte) error {
	if bc.PutCalled != nil {
		return bc.PutCalled(unitType, key, value)
	}
	return nil
}

// GetAll -
func (bc *ChainStorerStub) GetAll(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
	if bc.GetAllCalled != nil {
		return bc.GetAllCalled(unitType, keys)
	}
	return nil, nil
}

// SetEpochForPutOperation -
func (bc *ChainStorerStub) SetEpochForPutOperation(_ uint32) {
}

// Destroy -
func (bc *ChainStorerStub) Destroy() error {
	if bc.DestroyCalled != nil {
		return bc.DestroyCalled()
	}
	return nil
}

// IsInterfaceNil -
func (bc *ChainStorerStub) IsInterfaceNil() bool {
	return bc == nil
}
