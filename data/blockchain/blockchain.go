package blockchain

import (
	"ElrondNetwork/elrond-go-sandbox/data/block"
	"ElrondNetwork/elrond-go-sandbox/storage"
	"math/big"

	"ElrondNetwork/elrond-go-sandbox/config"

	"sync"

	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("prefix", "blockchain")

const (
	TransactionUnit UnitType = 0
	BlockUnit       UnitType = 1
	BlockHeaderUnit UnitType = 2
)

type UnitType uint8

type BlockChainService interface {
	Has(unitType UnitType, key []byte) bool
	Get(unitType UnitType, key []byte) []byte
	Put(unitType UnitType, key []byte, value []byte)
	GetAll(unitType UnitType, keys [][]byte) map[string][]byte
}

type BlockChain struct {
	lock          sync.RWMutex
	genesisBlock  *block.Block
	currentBlock  *block.Block
	localHeight   *big.Int
	networkHeight *big.Int
	badBlocks     storage.Cacher
	chain         map[UnitType]*storage.StorageUnit
}

// NewData returns an initialized blockchain
func NewData() (*BlockChain, error) {
	// TODO: choose the blockchain config based on some parameter/env variable/config variable
	// currently testnet blockchain config is the only valid one

	txStorage, err := storage.NewStorageUnitFromConf(config.TestnetBlockchainConfig.TxStorage)

	if err != nil {
		log.Fatalf("transaction storage could not be created: %s", err)
	}

	blStorage, err := storage.NewStorageUnitFromConf(config.TestnetBlockchainConfig.BlockStorage)
	if err != nil {
		log.Fatalf("block storage could not be created: %s", err)
	}

	blHeadStorage, err := storage.NewStorageUnitFromConf(config.TestnetBlockchainConfig.BlockHeaderStorage)
	if err != nil {
		log.Fatalf("block header storage could not be created: %s", err)
	}

	badBlocksCache, err := storage.CreateCacheFromConf(config.TestnetBlockchainConfig.BBlockCache)

	if err != nil {
		log.Fatalf("bad block cache could not be created: %s", err)
	}

	data := &BlockChain{
		genesisBlock:  nil,
		currentBlock:  nil,
		localHeight:   big.NewInt(-1),
		networkHeight: big.NewInt(-1),
		badBlocks:     badBlocksCache,
		chain: map[UnitType]*storage.StorageUnit{
			TransactionUnit: txStorage,
			BlockUnit:       blStorage,
			BlockHeaderUnit: blHeadStorage,
		},
	}

	return data, nil
}

func (bc *BlockChain) Has(unitType UnitType, key []byte) bool {
	bc.lock.RLock()
	storer := bc.chain[unitType]
	bc.lock.RUnlock()

	if storer == nil {
		log.Debug("no such unit type %d", unitType)
		return false
	}

	has, err := storer.Has(key)

	if err != nil {
		log.Debug("not found %s", key)
	}

	return has
}

func (bc *BlockChain) Get(unitType UnitType, key []byte) []byte {
	bc.lock.RLock()
	storer := bc.chain[unitType]
	bc.lock.RUnlock()

	if storer == nil {
		log.Debug("no such unit type %d", unitType)
		return nil
	}

	elem, err := storer.Get(key)

	if err != nil {
		log.Debug("not found %s", key)
	}

	return elem
}

func (bc *BlockChain) Put(unitType UnitType, key []byte, value []byte) {
	bc.lock.RLock()
	storer := bc.chain[unitType]
	bc.lock.RUnlock()

	if storer == nil {
		log.Debug("no such unit type %d", unitType)
		return
	}

	err := storer.Put(key, value)

	if err != nil {
		log.Error("not able to put key %s into unit %d", key, unitType)
	}
}

func (bc *BlockChain) GetAll(unitType UnitType, keys [][]byte) map[string][]byte {
	bc.lock.RLock()
	storer := bc.chain[unitType]
	bc.lock.RUnlock()

	if storer == nil {
		log.Debug("no such unit type %d", unitType)
		return nil
	}

	m := map[string][]byte{}

	for _, key := range keys {
		val, err := storer.Get(key)

		if err != nil {
			log.Debug("could not find value for key %s", key)
			continue
		}

		m[string(key)] = val
	}

	return m
}
