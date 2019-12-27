package txpool

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/txpool"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// CreateTxPool creates a new tx pool, according to the configuration
func CreateTxPool(config storageUnit.CacheConfig) (dataRetriever.ShardedDataCacherNotifier, error) {
	isOldTxPool := config.Type == storageUnit.FIFOShardedCache || config.Type == storageUnit.LRUCache

	if isOldTxPool {
		return shardedData.NewShardedData(config)
	}

	return txpool.NewShardedTxPool(config)
}
