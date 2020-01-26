package txpool

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// ArgShardedTxPool is the argument for ShardedTxPool's constructor
type ArgShardedTxPool struct {
	Config         storageUnit.CacheConfig
	MinGasPrice    uint64
	NumberOfShards uint32
}

func (args *ArgShardedTxPool) verify() error {
	config := args.Config

	if config.SizeInBytes < process.TxPoolMinSizeInBytes {
		return dataRetriever.ErrCacheConfigInvalidSizeInBytes
	}
	if config.Size < 1 {
		return dataRetriever.ErrCacheConfigInvalidSize
	}
	if config.Shards < 1 {
		return dataRetriever.ErrCacheConfigInvalidShards
	}
	if args.MinGasPrice < 1 {
		return dataRetriever.ErrCacheConfigInvalidEconomics
	}
	if args.NumberOfShards < 1 {
		return dataRetriever.ErrCacheConfigInvalidSharding
	}

	return nil
}
