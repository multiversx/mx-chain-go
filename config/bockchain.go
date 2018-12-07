package config

import (
	"math/big"
	"path/filepath"
)

// CacheType represents the type of the supported caches
type CacheType uint8

// DBType represents the type of the supported databases
type DBType uint8

// LRUCache is currently the only supported Cache type
const (
	LRUCache CacheType = 0
)

// LvlDB currently the only supported DBs
// More to be added
const (
	LvlDB DBType = 0
)

var (
	// TestnetBlockchainConfig holds the configuration of testnet blockchain
	TestnetBlockchainConfig = &BlockChainConfig{
		BlockChainID: big.NewInt(0),

		BlockStorage: &StorageUnitConfig{
			CacheConf: &CacheConfig{
				Size: 100,
				Type: LRUCache,
			},
			DBConf: &DBConfig{
				FileName: filepath.Join(DefaultPath(), "Blocks"),
				Type:     LvlDB,
			},
		},

		BlockHeaderStorage: &StorageUnitConfig{
			CacheConf: &CacheConfig{
				Size: 100,
				Type: LRUCache,
			},
			DBConf: &DBConfig{
				FileName: filepath.Join(DefaultPath(), "BlockHeaders"),
				Type:     LvlDB,
			},
		},

		TxStorage: &StorageUnitConfig{
			CacheConf: &CacheConfig{
				Size: 100000,
				Type: LRUCache,
			},
			DBConf: &DBConfig{
				FileName: filepath.Join(DefaultPath(), "Transactions"),
				Type:     LvlDB,
			},
		},

		BBlockCache: &CacheConfig{
			Size: 100,
			Type: LRUCache,
		},

		TxPoolStorage: &CacheConfig{
			Size: 1000,
			Type: LRUCache,
		},
	}

	// MainnetBlockchainConfig is currently not specified
	MainnetBlockchainConfig = (*BlockChainConfig)(nil)
)

// BlockChainConfig holds the configurable elements of the blockchain
type BlockChainConfig struct {
	BlockChainID       *big.Int
	BlockStorage       *StorageUnitConfig
	BlockHeaderStorage *StorageUnitConfig
	TxStorage          *StorageUnitConfig
	TxPoolStorage      *CacheConfig
	BBlockCache        *CacheConfig
}

// StorageUnitConfig holds the configurable elements of the storage unit
type StorageUnitConfig struct {
	CacheConf *CacheConfig
	DBConf    *DBConfig
}

// CacheConfig holds the configurable elements of a cache
type CacheConfig struct {
	Size uint32
	Type CacheType
}

// DBConfig holds the configurable elements of a database
type DBConfig struct {
	FileName string
	Type     DBType
}
