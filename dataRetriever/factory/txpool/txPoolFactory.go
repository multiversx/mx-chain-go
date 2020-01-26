package txpool

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/txpool"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// CreateTxPool creates a new tx pool, according to the configuration
func CreateTxPool(config storageUnit.CacheConfig, economics txpool.Economics) (dataRetriever.ShardedDataCacherNotifier, error) {
	switch config.Type {
	case storageUnit.FIFOShardedCache:
		return shardedData.NewShardedData(config)
	case storageUnit.LRUCache:
		return shardedData.NewShardedData(config)
	default:
		return txpool.NewShardedTxPool(config, economics)
	}
}
