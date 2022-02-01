package dataRetriever

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ StorageService = (*ChainStorer)(nil)

// ChainStorer is a StorageService implementation that can hold multiple storages
//  grouped by storage unit type
type ChainStorer struct {
	lock  sync.RWMutex
	chain map[UnitType]storage.Storer
}

// NewChainStorer returns a new initialised ChainStorer
func NewChainStorer() *ChainStorer {
	return &ChainStorer{
		chain: make(map[UnitType]storage.Storer),
	}
}

// AddStorer will add a new storer to the chain map
func (bc *ChainStorer) AddStorer(key UnitType, s storage.Storer) {
	bc.lock.Lock()
	bc.chain[key] = s
	bc.lock.Unlock()
}

// GetStorer returns the storer from the chain map or nil if the storer was not found
func (bc *ChainStorer) GetStorer(unitType UnitType) storage.Storer {
	bc.lock.RLock()
	storer := bc.chain[unitType]
	bc.lock.RUnlock()

	return storer
}

// Has returns true if the key is found in the selected Unit or false otherwise
// It can return an error if the provided unit type is not supported or if the
// underlying implementation of the storage unit reports an error.
func (bc *ChainStorer) Has(unitType UnitType, key []byte) error {
	bc.lock.RLock()
	storer := bc.chain[unitType]
	bc.lock.RUnlock()

	if storer == nil {
		return ErrNoSuchStorageUnit
	}

	return storer.Has(key)
}

// Get returns the value for the given key if found in the selected storage unit,
// nil otherwise. It can return an error if the provided unit type is not supported
// or if the storage unit underlying implementation reports an error
func (bc *ChainStorer) Get(unitType UnitType, key []byte) ([]byte, error) {
	bc.lock.RLock()
	storer := bc.chain[unitType]
	bc.lock.RUnlock()

	if storer == nil {
		return nil, ErrNoSuchStorageUnit
	}

	return storer.Get(key)
}

// Put stores the key, value pair in the selected storage unit
// It can return an error if the provided unit type is not supported
// or if the storage unit underlying implementation reports an error
func (bc *ChainStorer) Put(unitType UnitType, key []byte, value []byte) error {
	bc.lock.RLock()
	storer := bc.chain[unitType]
	bc.lock.RUnlock()

	if storer == nil {
		return fmt.Errorf("%w for unit type %s", ErrNoSuchStorageUnit, unitType.String())
	}

	return storer.Put(key, value)
}

// SetEpochForPutOperation will set the epoch to be used in all persisters for the put operation
func (bc *ChainStorer) SetEpochForPutOperation(epoch uint32) {
	bc.lock.Lock()
	for _, storer := range bc.chain {
		storerWithPutInEpoch, ok := storer.(storage.StorerWithPutInEpoch)
		if !ok {
			continue
		}

		storerWithPutInEpoch.SetEpochForPutOperation(epoch)
	}
	bc.lock.Unlock()
}

// GetAll gets all the elements with keys in the keys array, from the selected storage unit
// It can report an error if the provided unit type is not supported, if there is a missing
// key in the unit, or if the underlying implementation of the storage unit reports an error.
func (bc *ChainStorer) GetAll(unitType UnitType, keys [][]byte) (map[string][]byte, error) {
	bc.lock.RLock()
	storer := bc.chain[unitType]
	bc.lock.RUnlock()

	if storer == nil {
		return nil, ErrNoSuchStorageUnit
	}

	m := map[string][]byte{}

	for _, key := range keys {
		val, err := storer.Get(key)

		if err != nil {
			return nil, err
		}

		m[string(key)] = val
	}

	return m, nil
}

// GetAllStorers returns all the available storers
func (bc *ChainStorer) GetAllStorers() map[UnitType]storage.Storer {
	bc.lock.RLock()
	chainMapCopy := make(map[UnitType]storage.Storer, len(bc.chain))
	for key, value := range bc.chain {
		chainMapCopy[key] = value
	}
	bc.lock.RUnlock()

	return chainMapCopy
}

// Destroy removes the underlying files/resources used by the storage service
func (bc *ChainStorer) Destroy() error {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	var err error

	for _, v := range bc.chain {
		err = v.DestroyUnit()
		if err != nil {
			return err
		}
	}
	bc.chain = nil
	return nil
}

// CloseAll will close all the active units
func (bc *ChainStorer) CloseAll() error {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	closedSuccessfully := true
	for _, unit := range bc.chain {
		err := unit.Close()
		if err != nil {
			closedSuccessfully = false
		}
	}

	if closedSuccessfully {
		return nil
	}

	return storage.ErrClosingPersisters
}

// IsInterfaceNil returns true if there is no value under the interface
func (bc *ChainStorer) IsInterfaceNil() bool {
	return bc == nil
}
