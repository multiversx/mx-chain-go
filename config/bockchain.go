package config

import "math/big"

type CacheType uint8
type DBType uint8

// supported Caches
const (
	LRUCache CacheType = 0
)

// supported DBs
const (
	LvlDB DBType = 0
)

var (
	TestnetBlockchainConfig = &BlockChainConfig{
		BlockChainID: big.NewInt(0),

		BlockStorage: &StorageUnitConfig{
			CacheConf: &CacheConfig{
				Size: 100,
				Type: LRUCache,
			},
			DBConf: &DBConfig{
				FileName: "Blocks",
				Type:     LvlDB,
			},
		},

		BlockHeaderStorage: &StorageUnitConfig{
			CacheConf: &CacheConfig{
				Size: 100,
				Type: LRUCache,
			},
			DBConf: &DBConfig{
				FileName: "Blocks",
				Type:     LvlDB,
			},
		},

		TxStorage: &StorageUnitConfig{
			CacheConf: &CacheConfig{
				Size: 100000,
				Type: LRUCache,
			},
			DBConf: &DBConfig{
				FileName: "Transactions",
				Type:     LvlDB,
			},
		},

		BBlockCache: &CacheConfig{
			Size: 100,
			Type: LRUCache,
		},
	}

	// Mainnet config to follow
	MainnetBlockchainConfig = (*BlockChainConfig)(nil)
)

type BlockChainConfig struct {
	BlockChainID       *big.Int
	BlockStorage       *StorageUnitConfig
	BlockHeaderStorage *StorageUnitConfig
	TxStorage          *StorageUnitConfig
	BBlockCache        *CacheConfig
}

type StorageUnitConfig struct {
	CacheConf *CacheConfig
	DBConf    *DBConfig
}

type CacheConfig struct {
	Size uint32
	Type CacheType
}

type DBConfig struct {
	FileName string
	Type     DBType
}
