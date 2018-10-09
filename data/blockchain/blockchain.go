package blockchain

import (
	"ElrondNetwork/elrond-go-sandbox/data/block"
	"ElrondNetwork/elrond-go-sandbox/storage"
	"math/big"

	"ElrondNetwork/elrond-go-sandbox/config"

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
	Get(unitType UnitType, key []byte) []byte
	Put(unitType UnitType, key []byte)
	GetLocal(unitType UnitType, key []byte) []byte
	Contains(unitType UnitType, key []byte) bool
	GetAllLocal(unitType UnitType, keys [][]byte) map[string][]byte
}

type BlockChain struct {
	genesisBlock  *block.Block
	currentBlock  *block.Block
	localHeight   *big.Int
	networkHeight *big.Int
	badBlocks     storage.Cacher
	chain         map[UnitType]storage.Storer
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
		chain: map[UnitType]storage.Storer{
			TransactionUnit: txStorage,
			BlockUnit:       blStorage,
			BlockHeaderUnit: blHeadStorage,
		},
	}

	return data, nil
}
