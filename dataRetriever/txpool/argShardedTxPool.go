package txpool

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/storage/txcache"
)

// ArgShardedTxPool is the argument for ShardedTxPool's constructor
type ArgShardedTxPool struct {
	Config               storageunit.CacheConfig
	EpochNotifier        dataRetriever.EpochNotifier
	TxGasHandler         txcache.TxGasHandler
	AccountNonceProvider dataRetriever.AccountNonceProvider
	NumberOfShards       uint32
	SelfShardID          uint32
}

// TODO: Upon further analysis and brainstorming, add some sensible minimum accepted values for the appropriate fields.
func (args *ArgShardedTxPool) verify() error {
	config := args.Config

	if config.SizeInBytes == 0 {
		return fmt.Errorf("%w: config.SizeInBytes is not valid", dataRetriever.ErrCacheConfigInvalidSizeInBytes)
	}
	if config.SizeInBytesPerSender == 0 {
		return fmt.Errorf("%w: config.SizeInBytesPerSender is not valid", dataRetriever.ErrCacheConfigInvalidSizeInBytes)
	}
	if config.Capacity == 0 {
		return fmt.Errorf("%w: config.Capacity is not valid", dataRetriever.ErrCacheConfigInvalidSize)
	}
	if config.SizePerSender == 0 {
		return fmt.Errorf("%w: config.SizePerSender is not valid", dataRetriever.ErrCacheConfigInvalidSize)
	}
	if config.Shards == 0 {
		return fmt.Errorf("%w: config.Shards (map chunks) is not valid", dataRetriever.ErrCacheConfigInvalidShards)
	}
	if check.IfNil(args.EpochNotifier) {
		return fmt.Errorf("%w: EpochNotifier is not valid", dataRetriever.ErrNilEpochNotifier)
	}
	if check.IfNil(args.TxGasHandler) {
		return fmt.Errorf("%w: TxGasHandler is not valid", dataRetriever.ErrNilTxGasHandler)
	}
	if args.TxGasHandler.MinGasPrice() == 0 {
		return fmt.Errorf("%w: MinGasPrice is not valid", dataRetriever.ErrCacheConfigInvalidEconomics)
	}
	if check.IfNil(args.AccountNonceProvider) {
		return fmt.Errorf("%w: AccountNonceProvider is not valid", dataRetriever.ErrNilAccountNonceProvider)
	}
	if args.NumberOfShards == 0 {
		return fmt.Errorf("%w: NumberOfShards is not valid", dataRetriever.ErrCacheConfigInvalidSharding)
	}

	return nil
}
