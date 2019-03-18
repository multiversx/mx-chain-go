package blockchain

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

const (
	// TransactionUnit is the transactions storage unit identifier
	TransactionUnit UnitType = 0
	// MiniBlockUnit is the transaction block body storage unit identifier
	MiniBlockUnit UnitType = 1
	// PeerChangesUnit is the peer change block body storage unit identifier
	PeerChangesUnit UnitType = 2
	// BlockHeaderUnit is the Block Headers Storage unit identifier
	BlockHeaderUnit UnitType = 3
	// MetaBlockUnit is the metachain blocks storage unit identifier
	MetaBlockUnit UnitType = 4
	// MetaShardDataUnit is the metachain shard data unit identifier
	MetaShardDataUnit UnitType = 5
	// MetaPeerDataUnit is the metachain peer data unit identifier
	MetaPeerDataUnit UnitType = 6
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
