package txpool

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
)

// ArgShardedTxPool is the argument for ShardedTxPool's constructor
type ArgShardedTxPool struct {
	Config                         storageunit.CacheConfig
	TxGasHandler                   txGasHandler
	Marshalizer                    marshal.Marshalizer
	NumberOfShards                 uint32
	SelfShardID                    uint32
	MaxNumBytesPerSenderUpperBound uint32
	SortedTransactionsConfig       config.SortedTransactionsConfig
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
	if check.IfNil(args.TxGasHandler) {
		return fmt.Errorf("%w: TxGasHandler is not valid", dataRetriever.ErrNilTxGasHandler)
	}
	if check.IfNil(args.Marshalizer) {
		return fmt.Errorf("%w: Marshalizer is not valid", dataRetriever.ErrNilMarshalizer)
	}
	if args.NumberOfShards == 0 {
		return fmt.Errorf("%w: NumberOfShards is not valid", dataRetriever.ErrCacheConfigInvalidSharding)
	}

	if args.SortedTransactionsConfig.SelectionLoopDurationCheckInterval == 0 {
		return fmt.Errorf("%w: SelectionLoopDurationCheckInterval is not valid", dataRetriever.ErrBadSelectionLoopDurationCheckInterval)
	}

	if args.SortedTransactionsConfig.TxCacheSelectionMaxNumTxs == 0 {
		return fmt.Errorf("%w: TxCacheSelectionMaxNumTxs is not valid", dataRetriever.ErrBadTxCacheSelectionMaxNumTxs)
	}

	if args.SortedTransactionsConfig.TxCacheSelectionGasRequested == 0 {
		return fmt.Errorf("%w: TxCacheSelectionMaxNumTxs is not valid", dataRetriever.ErrBadTxCacheSelectionGasRequested)
	}

	if args.SortedTransactionsConfig.TxCacheSelectionLoopMaximumDuration == 0 {
		return fmt.Errorf("%w: TxCacheSelectionMaxNumTxs is not valid", dataRetriever.ErrBadTxCacheSelectionLoopMaximumDuration)
	}

	return nil
}
