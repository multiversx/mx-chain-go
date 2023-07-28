package preprocess

import "github.com/multiversx/mx-chain-go/process"

type smartContractResultPreProcessorFactory struct {
}

// NewSmartContractResultPreProcessorFactory creates a new smart contract result pre processor factory
func NewSmartContractResultPreProcessorFactory() (*smartContractResultPreProcessorFactory, error) {
	return &smartContractResultPreProcessorFactory{}, nil
}

// CreateSmartContractResultPreProcessor creates a new smart contract result pre processor
func (scppf *smartContractResultPreProcessorFactory) CreateSmartContractResultPreProcessor(args SmartContractResultPreProcessorCreatorArgs) (process.PreProcessor, error) {
	return NewSmartContractResultPreprocessor(
		args.ScrDataPool,
		args.Store,
		args.Hasher,
		args.Marshalizer,
		args.ScrProcessor,
		args.ShardCoordinator,
		args.Accounts,
		args.OnRequestSmartContractResult,
		args.GasHandler,
		args.EconomicsFee,
		args.PubkeyConverter,
		args.BlockSizeComputation,
		args.BalanceComputation,
		args.EnableEpochsHandler,
		args.ProcessedMiniBlocksTracker,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (scrppf *smartContractResultPreProcessorFactory) IsInterfaceNil() bool {
	return scrppf == nil
}
