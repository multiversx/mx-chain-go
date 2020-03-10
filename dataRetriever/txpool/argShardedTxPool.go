package txpool

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// ArgShardedTxPool is the argument for ShardedTxPool's constructor
type ArgShardedTxPool struct {
	Config         storageUnit.CacheConfig
	MinGasPrice    uint64
	NumberOfShards uint32
	SelfShardID    uint32
}

func (args *ArgShardedTxPool) verify() error {
	config := args.Config

	if config.SizeInBytes < dataRetriever.TxPoolMinSizeInBytes {
		return fmt.Errorf("%w: config.SizeInBytes is less than [dataRetriever.TxPoolMinSizeInBytes]", dataRetriever.ErrCacheConfigInvalidSizeInBytes)
	}
	if config.Size < 1 {
		return fmt.Errorf("%w: config.Size is less than 1", dataRetriever.ErrCacheConfigInvalidSize)
	}
	if config.Shards < 1 {
		return fmt.Errorf("%w: config.Shards (map chunks) is less than 1", dataRetriever.ErrCacheConfigInvalidShards)
	}
	if args.MinGasPrice < 1 {
		return fmt.Errorf("%w: MinGasPrice is less than 1", dataRetriever.ErrCacheConfigInvalidEconomics)
	}
	if args.NumberOfShards < 1 {
		return fmt.Errorf("%w: NumberOfShards is less than 1", dataRetriever.ErrCacheConfigInvalidSharding)
	}

	return nil
}
