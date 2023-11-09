package storage

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage"
)

// ChainStorerStub -
type ChainStorerStub struct {
	AddStorerCalled     func(key dataRetriever.UnitType, s storage.Storer)
	GetStorerCalled     func(unitType dataRetriever.UnitType) (storage.Storer, error)
	HasCalled           func(unitType dataRetriever.UnitType, key []byte) error
	GetCalled           func(unitType dataRetriever.UnitType, key []byte) ([]byte, error)
	PutCalled           func(unitType dataRetriever.UnitType, key []byte, value []byte) error
	GetAllCalled        func(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error)
	GetAllStorersCalled func() map[dataRetriever.UnitType]storage.Storer
	DestroyCalled       func() error
	CloseAllCalled      func() error
}

// CloseAll -
func (stub *ChainStorerStub) CloseAll() error {
	if stub.CloseAllCalled != nil {
		return stub.CloseAllCalled()
	}
	return nil
}

// AddStorer -
func (stub *ChainStorerStub) AddStorer(key dataRetriever.UnitType, s storage.Storer) {
	if stub.AddStorerCalled != nil {
		stub.AddStorerCalled(key, s)
	}
}

// GetStorer -
func (stub *ChainStorerStub) GetStorer(unitType dataRetriever.UnitType) (storage.Storer, error) {
	if stub.GetStorerCalled != nil {
		return stub.GetStorerCalled(unitType)
	}
	return nil, storage.ErrKeyNotFound
}

// Has -
func (stub *ChainStorerStub) Has(unitType dataRetriever.UnitType, key []byte) error {
	if stub.HasCalled != nil {
		return stub.HasCalled(unitType, key)
	}
	return nil
}

// Get -
func (stub *ChainStorerStub) Get(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
	if stub.GetCalled != nil {
		return stub.GetCalled(unitType, key)
	}
	return nil, nil
}

// Put -
func (stub *ChainStorerStub) Put(unitType dataRetriever.UnitType, key []byte, value []byte) error {
	if stub.PutCalled != nil {
		return stub.PutCalled(unitType, key, value)
	}
	return nil
}

// GetAll -
func (stub *ChainStorerStub) GetAll(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
	if stub.GetAllCalled != nil {
		return stub.GetAllCalled(unitType, keys)
	}
	return nil, nil
}

// GetAllStorers -
func (stub *ChainStorerStub) GetAllStorers() map[dataRetriever.UnitType]storage.Storer {
	if stub.GetAllStorersCalled != nil {
		return stub.GetAllStorersCalled()
	}
	return nil
}

// SetEpochForPutOperation -
func (stub *ChainStorerStub) SetEpochForPutOperation(_ uint32) {
}

// Destroy -
func (stub *ChainStorerStub) Destroy() error {
	if stub.DestroyCalled != nil {
		return stub.DestroyCalled()
	}
	return nil
}

// IsInterfaceNil -
func (stub *ChainStorerStub) IsInterfaceNil() bool {
	return stub == nil
}
