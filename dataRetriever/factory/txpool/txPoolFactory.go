package txpool

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/txpool"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// CreateTxPool creates a new tx pool
func CreateTxPool(args txpool.ArgShardedTxPool) (dataRetriever.ShardedDataCacherNotifier, error) {
	switch args.Config.Type {
	case storageUnit.FIFOShardedCache:
		return shardedData.NewShardedData(args.Config)
	case storageUnit.LRUCache:
		return shardedData.NewShardedData(args.Config)
	default:
		return txpool.NewShardedTxPool(args)
	}
}
