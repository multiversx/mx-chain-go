package blockchain

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
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
)

// UnitType is the type for Storage unit identifiers
type UnitType uint8

// StorageService is the interface for blockChain storage unit provided services
type StorageService interface {
	// GetStorer returns the storer from the chain map
	GetStorer(unitType UnitType) storage.Storer
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

// BlockChain holds the block information for the current shard.
//
// The BlockChain through it's Storage units map manages the storage,
// retrieval search of blocks (body), transactions, block headers,
// bad blocks.
//
// The BlockChain also holds pointers to the Genesis block, the current block
// the height of the local chain and the perceived height of the chain in the network.
type BlockChain struct {
	lock               sync.RWMutex
	GenesisBlock       *block.Header               // Genesys Block pointer
	CurrentBlockHeader *block.Header               // Current Block pointer
	LocalHeight        int64                       // Height of the local chain
	NetworkHeight      int64                       // Percieved height of the network chain
	badBlocks          storage.Cacher              // Bad blocks cache
	chain              map[UnitType]storage.Storer // chains for each unit type. Together they form the blockchain
}

// NewBlockChain returns an initialized blockchain
// It uses a config file to setup it's supported storage units map
func NewBlockChain(
	badBlocksCache storage.Cacher,
	txUnit storage.Storer,
	txBlockUnit storage.Storer,
	stateBlockUnit storage.Storer,
	peerBlockUnit storage.Storer,
	headerUnit storage.Storer) (*BlockChain, error) {

	if badBlocksCache == nil {
		return nil, ErrBadBlocksCacheNil
	}

	if txUnit == nil {
		return nil, ErrTxUnitNil
	}

	if txBlockUnit == nil {
		return nil, ErrTxBlockUnitNil
	}

	if stateBlockUnit == nil {
		return nil, ErrStateBlockUnitNil
	}

	if peerBlockUnit == nil {
		return nil, ErrPeerBlockUnitNil
	}

	if headerUnit == nil {
		return nil, ErrHeaderUnitNil
	}

	data := &BlockChain{
		GenesisBlock:       nil,
		CurrentBlockHeader: nil,
		LocalHeight:        -1,
		NetworkHeight:      -1,
		badBlocks:          badBlocksCache,
		chain: map[UnitType]storage.Storer{
			TransactionUnit:    txUnit,
			TxBlockBodyUnit:    txBlockUnit,
			StateBlockBodyUnit: stateBlockUnit,
			PeerBlockBodyUnit:  peerBlockUnit,
			BlockHeaderUnit:    headerUnit,
		},
	}

	return data, nil
}

// GetStorer returns the storer from the chain map or nil if the storer was not found
func (bc *BlockChain) GetStorer(unitType UnitType) storage.Storer {
	bc.lock.RLock()
	storer := bc.chain[unitType]
	bc.lock.RUnlock()

	return storer
}

// Has returns true if the key is found in the selected Unit or false otherwise
// It can return an error if the provided unit type is not supported or if the
// underlying implementation of the storage unit reports an error.
func (bc *BlockChain) Has(unitType UnitType, key []byte) (bool, error) {
	bc.lock.RLock()
	storer := bc.chain[unitType]
	bc.lock.RUnlock()

	if storer == nil {
		return false, ErrNoSuchStorageUnit
	}

	return storer.Has(key)
}

// Get returns the value for the given key if found in the selected storage unit,
// nil otherwise. It can return an error if the provided unit type is not supported
// or if the storage unit underlying implementation reports an error
func (bc *BlockChain) Get(unitType UnitType, key []byte) ([]byte, error) {
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
func (bc *BlockChain) Put(unitType UnitType, key []byte, value []byte) error {
	bc.lock.RLock()
	storer := bc.chain[unitType]
	bc.lock.RUnlock()

	if storer == nil {
		return ErrNoSuchStorageUnit
	}

	return storer.Put(key, value)
}

// GetAll gets all the elements with keys in the keys array, from the selected storage unit
// It can report an error if the provided unit type is not supported, if there is a missing
// key in the unit, or if the underlying implementation of the storage unit reports an error.
func (bc *BlockChain) GetAll(unitType UnitType, keys [][]byte) (map[string][]byte, error) {
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

// Destroy removes the underlying files/resources used by the storage service
func (bc *BlockChain) Destroy() error {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	var err error

	for _, v := range bc.chain {
		err = v.DestroyUnit()
		if err != nil {
			return err
		}
	}

	return nil
}

// IsBadBlock returns true if the provided hash is blacklisted as a bad block, or false otherwise
func (bc *BlockChain) IsBadBlock(blockHash []byte) bool {
	return bc.badBlocks.Has(blockHash)
}

// PutBadBlock adds the given serialized block to the bad block cache, blacklisting it
func (bc *BlockChain) PutBadBlock(blockHash []byte) {
	bc.badBlocks.Put(blockHash, struct{}{})
}
