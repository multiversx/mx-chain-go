package blockchain

import (
	"ElrondNetwork/elrond-go-sandbox/data/block"
	"ElrondNetwork/elrond-go-sandbox/storage"
	"math/big"

	"ElrondNetwork/elrond-go-sandbox/config"

	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("prefix", "blockchain")

const (
	TransactionUnit UnitType = 0
	BlockUnit       UnitType = 1
	BlockHeaderUnit UnitType = 2
)

type UnitType uint8

type BlockChainStorageService interface {
	Has(unitType UnitType, key []byte) (bool, error)
	Get(unitType UnitType, key []byte) ([]byte, error)
	Put(unitType UnitType, key []byte, value []byte) error
	GetAll(unitType UnitType, keys [][]byte) (map[string][]byte, error)
	Destroy() error
}

type BlockChain struct {
	lock          sync.RWMutex
	GenesisBlock  *block.Block
	CurrentBlock  *block.Block
	LocalHeight   *big.Int
	NetworkHeight *big.Int
	badBlocks     storage.Cacher
	chain         map[UnitType]*storage.StorageUnit
}

// NewData returns an initialized blockchain
func NewData() (*BlockChain, error) {
	// TODO: choose the blockchain config based on some parameter/env variable/config variable
	// currently testnet blockchain config is the only valid one

	txStorage, err := storage.NewStorageUnitFromConf(config.TestnetBlockchainConfig.TxStorage)

	if err != nil {
		panic(err)
	}

	blStorage, err := storage.NewStorageUnitFromConf(config.TestnetBlockchainConfig.BlockStorage)
	if err != nil {
		txStorage.DestroyUnit()
		panic(err)
	}

	blHeadStorage, err := storage.NewStorageUnitFromConf(config.TestnetBlockchainConfig.BlockHeaderStorage)
	if err != nil {
		txStorage.DestroyUnit()
		blStorage.DestroyUnit()
		panic(err)
	}

	badBlocksCache, err := storage.CreateCacheFromConf(config.TestnetBlockchainConfig.BBlockCache)

	if err != nil {
		txStorage.DestroyUnit()
		blStorage.DestroyUnit()
		blHeadStorage.DestroyUnit()
		panic(err)
	}

	data := &BlockChain{
		GenesisBlock:  nil,
		CurrentBlock:  nil,
		LocalHeight:   big.NewInt(-1),
		NetworkHeight: big.NewInt(-1),
		badBlocks:     badBlocksCache,
		chain: map[UnitType]*storage.StorageUnit{
			TransactionUnit: txStorage,
			BlockUnit:       blStorage,
			BlockHeaderUnit: blHeadStorage,
		},
	}

	return data, nil
}

func (bc *BlockChain) Has(unitType UnitType, key []byte) (bool, error) {
	bc.lock.RLock()
	storer := bc.chain[unitType]
	bc.lock.RUnlock()

	if storer == nil {
		log.Error("no such unit type ", unitType)
		return false, errors.New("no such unit type")
	}

	return storer.Has(key)
}

func (bc *BlockChain) Get(unitType UnitType, key []byte) ([]byte, error) {
	bc.lock.RLock()
	storer := bc.chain[unitType]
	bc.lock.RUnlock()

	if storer == nil {
		log.Error("no such unit type ", unitType)
		return nil, errors.New("no such unit type")
	}

	return storer.Get(key)
}

func (bc *BlockChain) Put(unitType UnitType, key []byte, value []byte) error {
	bc.lock.RLock()
	storer := bc.chain[unitType]
	bc.lock.RUnlock()

	if storer == nil {
		log.Error("no such unit type ", unitType)
		return errors.New("no such unit type")
	}

	return storer.Put(key, value)
}

func (bc *BlockChain) GetAll(unitType UnitType, keys [][]byte) (map[string][]byte, error) {
	bc.lock.RLock()
	storer := bc.chain[unitType]
	bc.lock.RUnlock()

	if storer == nil {
		log.Error("no such unit type ", unitType)
		return nil, errors.New("no such unit type")
	}

	m := map[string][]byte{}

	for _, key := range keys {
		val, err := storer.Get(key)

		if err != nil {
			log.Debug("could not find value for key ", key)
			return nil, err
		}

		m[string(key)] = val
	}

	return m, nil
}

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

func (bc *BlockChain) IsBadBlock(blockHash []byte) bool {
	return bc.badBlocks.Has(blockHash)
}

func (bc *BlockChain) PutBadBlock(key []byte, value []byte) {
	bc.badBlocks.Put(key, value)
}
