package config

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"math/big"
	"path/filepath"
)

var (
	// TestnetBlockchainConfig holds the configuration of testnet blockchain
	testnet = &BlockChainConfig{
		BlockChainID: big.NewInt(0),

		BlockStorage: &storage.StorageUnitConfig{
			CacheConf: &storage.CacheConfig{
				Size: 100,
				Type: storage.LRUCache,
			},
			DBConf: &storage.DBConfig{
				FileName: filepath.Join(DefaultPath(), "Blocks"),
				Type:     storage.LvlDB,
			},
		},

		BlockHeaderStorage: &storage.StorageUnitConfig{
			CacheConf: &storage.CacheConfig{
				Size: 100,
				Type: storage.LRUCache,
			},
			DBConf: &storage.DBConfig{
				FileName: filepath.Join(DefaultPath(), "BlockHeaders"),
				Type:     storage.LvlDB,
			},
		},

		TxStorage: &storage.StorageUnitConfig{
			CacheConf: &storage.CacheConfig{
				Size: 100000,
				Type: storage.LRUCache,
			},
			DBConf: &storage.DBConfig{
				FileName: filepath.Join(DefaultPath(), "Transactions"),
				Type:     storage.LvlDB,
			},
		},

		BBlockCache: &storage.CacheConfig{
			Size: 100,
			Type: storage.LRUCache,
		},

		TxPoolStorage: &storage.CacheConfig{
				Size: 1000,
				Type: storage.LRUCache,
		},
	}

	// mainnet is currently not specified
	mainnet = (*BlockChainConfig)(nil)

	// Blockchain holds the configuration of testnet blockchain
	Blockchain = testnet
)

// BlockChainConfig holds the configurable elements of the blockchain
type BlockChainConfig struct {
	BlockChainID       *big.Int
	BlockStorage       *storage.StorageUnitConfig
	BlockHeaderStorage *storage.StorageUnitConfig
	TxStorage          *storage.StorageUnitConfig
	TxPoolStorage	   *storage.CacheConfig
	BBlockCache        *storage.CacheConfig
}