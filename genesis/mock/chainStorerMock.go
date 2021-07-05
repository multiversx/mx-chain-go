package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/pkg/errors"
)

// ChainStorerStub -
type ChainStorerStub struct {
	AddStorerCalled               func(key dataRetriever.UnitType, s storage.Storer)
	GetStorerCalled               func(unitType dataRetriever.UnitType) storage.Storer
	HasCalled                     func(unitType dataRetriever.UnitType, key []byte) error
	GetCalled                     func(unitType dataRetriever.UnitType, key []byte) ([]byte, error)
	PutCalled                     func(unitType dataRetriever.UnitType, key []byte, value []byte) error
	GetAllCalled                  func(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error)
	DestroyCalled                 func() error
	CloseAllCalled                func() error
	SetEpochForPutOperationCalled func(epoch uint32)
	GetAllStorersCalled           func() map[dataRetriever.UnitType]storage.Storer
}

// SetEpochForPutOperation -
func (css *ChainStorerStub) SetEpochForPutOperation(epoch uint32) {
	if css.SetEpochForPutOperationCalled != nil {
		css.SetEpochForPutOperationCalled(epoch)
	}
}

// CloseAll -
func (css *ChainStorerStub) CloseAll() error {
	if css.CloseAllCalled != nil {
		return css.CloseAllCalled()
	}

	return nil
}

// AddStorer will add a new storer to the chain map
func (css *ChainStorerStub) AddStorer(key dataRetriever.UnitType, s storage.Storer) {
	if css.AddStorerCalled != nil {
		css.AddStorerCalled(key, s)
	}
}

// GetStorer returns the storer from the chain map or nil if the storer was not found
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
	return errors.New("Key not found")
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

// GetAllStorers -
func (css *ChainStorerStub) GetAllStorers() map[dataRetriever.UnitType]storage.Storer {
	if css.GetAllStorersCalled != nil {
		return css.GetAllStorersCalled()
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
