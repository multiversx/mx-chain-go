package preprocess

import "github.com/multiversx/mx-chain-go/process"

type rewardsTxPreProcFactory struct {
}

// NewRewardsTxPreProcFactory creates a rewards tx pre-processor factory
func NewRewardsTxPreProcFactory() *rewardsTxPreProcFactory {
	return &rewardsTxPreProcFactory{}
}

// CreateRewardsTxPreProcessor creates a rewards tx pre-processor for normal chain run type
func (f *rewardsTxPreProcFactory) CreateRewardsTxPreProcessor(args ArgsRewardTxPreProcessor) (process.PreProcessor, error) {
	return NewRewardTxPreprocessor(
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
}

// IsInterfaceNil checks if the underyling pointer is nil
func (f *rewardsTxPreProcFactory) IsInterfaceNil() bool {
	return f == nil
}
