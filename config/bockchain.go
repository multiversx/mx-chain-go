package config

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"math/big"
	"path/filepath"
)

var (
	// testnet holds the configuration of testnet blockchain
	testnet = &blockchain.Config{
		BlockChainID: big.NewInt(0),

		BlockStorage: &storage.UnitConfig{
			CacheConf: &storage.CacheConfig{
				Size: 100,
				Type: storage.LRUCache,
			},
			DBConf: &storage.DBConfig{
				FileName: filepath.Join(DefaultPath(), "Blocks"),
				Type:     storage.LvlDB,
			},
		},

		BlockHeaderStorage: &storage.UnitConfig{
			CacheConf: &storage.CacheConfig{
				Size: 100,
				Type: storage.LRUCache,
			},
			DBConf: &storage.DBConfig{
				FileName: filepath.Join(DefaultPath(), "BlockHeaders"),
				Type:     storage.LvlDB,
			},
		},

		TxStorage: &storage.UnitConfig{
			CacheConf: &storage.CacheConfig{
				Size: 100000,
				Type: storage.LRUCache,
			},
			DBConf: &storage.DBConfig{
				FileName: filepath.Join(DefaultPath(), "Transactions"),
				Type:     storage.LvlDB,
			},
		},

		BlockCache: &storage.CacheConfig{
			Size: 100,
			Type: storage.LRUCache,
		},

		TxPoolStorage: &storage.CacheConfig{
				Size: 1000,
				Type: storage.LRUCache,
		},
	}

	// TODO: Add specifications for mainnet
	mainnet = (*blockchain.Config)(nil)

	// Blockchain holds the configuration of the blockchain
	Blockchain = testnet
)