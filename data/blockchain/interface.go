package blockchain

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

const (
	// TransactionUnit is the transactions storage unit identifier
	TransactionUnit UnitType = 0
	// TxBlockBodyUnit is the transaction block body storage unit identifier
	TxBlockBodyUnit UnitType = 1
	// StateBlockBodyUnit is the state block body storage unit identifier
	StateBlockBodyUnit UnitType = 2
	// PeerBlockBodyUnit is the peer change block body storage unit identifier
	PeerBlockBodyUnit UnitType = 3
	// BlockHeaderUnit is the Block Headers Storage unit identifier
	BlockHeaderUnit UnitType = 4
	// MetaBlockUnit is the metachain blocks storage unit identifier
	MetaBlockUnit UnitType = 5
)

// UnitType is the type for Storage unit identifiers
type UnitType uint8

// StorageService is the interface for blockChain storage unit provided services
type StorageService interface {
	// GetStorer returns the storer from the chain map
	GetStorer(unitType UnitType) storage.Storer
	// AddStorer will add a new storer to the chain map
	AddStorer(key UnitType, s storage.Storer)
	// Has returns true if the key is found in the selected Unit or false otherwise
	Has(unitType UnitType, key []byte) (bool, error)
	// Get returns the value for the given key if found in the selected storage unit, nil otherwise
	Get(unitType UnitType, key []byte) ([]byte, error)
	// Put stores the key, value pair in the selected storage unit
	Put(unitType UnitType, key []byte, value []byte) error
	// GetAll gets all the elements with keys in the keys array, from the selected storage unit
	// If there is a missing key in the unit, it returns an error
	GetAll(unitType UnitType, keys [][]byte) (map[string][]byte, error)
	// Destroy removes the underlying files/resources used by the storage service
	Destroy() error
}
