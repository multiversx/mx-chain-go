package preprocess

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
)

// ArgsRewardTxPreProcessor is a structure placeholder for arguments needed to create a rewards tx pre-processor
type ArgsRewardTxPreProcessor struct {
	RewardTxDataPool           dataRetriever.ShardedDataCacherNotifier
	Store                      dataRetriever.StorageService
	Hasher                     hashing.Hasher
	Marshalizer                marshal.Marshalizer
	RewardProcessor            process.RewardTransactionProcessor
	ShardCoordinator           sharding.Coordinator
	Accounts                   state.AccountsAdapter
	OnRequestRewardTransaction func(shardID uint32, txHashes [][]byte)
	GasHandler                 process.GasHandler
	PubkeyConverter            core.PubkeyConverter
	BlockSizeComputation       BlockSizeComputationHandler
	BalanceComputation         BalanceComputationHandler
	ProcessedMiniBlocksTracker process.ProcessedMiniBlocksTracker
	TxExecutionOrderHandler    common.TxExecutionOrderHandler
}

type rewardsTxPreProcFactory struct {
}

// NewRewardsTxPreProcFactory creates a rewards tx pre-processor factory
func NewRewardsTxPreProcFactory() *rewardsTxPreProcFactory {
	return &rewardsTxPreProcFactory{}
}

// CreateRewardsTxPreProcessorAndAddToContainer creates a rewards tx pre-processor for normal chain run type and adds it to the container
func (f *rewardsTxPreProcFactory) CreateRewardsTxPreProcessorAndAddToContainer(args ArgsRewardTxPreProcessor, container process.PreProcessorsContainer) error {
	rwdTxPreProc, err := NewRewardTxPreprocessor(
		args.RewardTxDataPool,
		args.Store,
		args.Hasher,
		args.Marshalizer,
		args.RewardProcessor,
		args.ShardCoordinator,
		args.Accounts,
		args.OnRequestRewardTransaction,
		args.GasHandler,
		args.PubkeyConverter,
		args.BlockSizeComputation,
		args.BalanceComputation,
		args.ProcessedMiniBlocksTracker,
		args.TxExecutionOrderHandler,
	)
	if err != nil {
		return err
	}

	return container.Add(block.RewardsBlock, rwdTxPreProc)
}

// IsInterfaceNil checks if the underyling pointer is nil
func (f *rewardsTxPreProcFactory) IsInterfaceNil() bool {
	return f == nil
}
