package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"math/big"
	"path/filepath"
)

// Config holds the mocked config used only in unit tests
var Config = &blockchain.Config{
	BlockChainID: big.NewInt(0),

	BlockStorage: &storage.UnitConfig{
		CacheConf: &storage.CacheConfig{
			Size: 100,
			Type: storage.LRUCache,
		},
		DBConf: &storage.DBConfig{
			FileName: filepath.Join("", "Blocks"),
			Type:     storage.LvlDB,
		},
	},

	BlockHeaderStorage: &storage.UnitConfig{
		CacheConf: &storage.CacheConfig{
			Size: 100,
			Type: storage.LRUCache,
		},
		DBConf: &storage.DBConfig{
			FileName: filepath.Join("", "BlockHeaders"),
			Type:     storage.LvlDB,
		},
	},

	TxStorage: &storage.UnitConfig{
		CacheConf: &storage.CacheConfig{
			Size: 100000,
			Type: storage.LRUCache,
		},
		DBConf: &storage.DBConfig{
			FileName: filepath.Join("", "Transactions"),
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
