package shard

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/factory/containers"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
)

var _ process.PreProcessorsContainerFactory = (*preProcessorsContainerFactory)(nil)

// ArgsPreProcessorsContainerFactory holds the arguments needed for creating a preProcessorsContainerFactory
type ArgsPreProcessorsContainerFactory struct {
	ShardCoordinator             sharding.Coordinator
	Store                        dataRetriever.StorageService
	Marshalizer                  marshal.Marshalizer
	Hasher                       hashing.Hasher
	DataPool                     dataRetriever.PoolsHolder
	PubkeyConverter              core.PubkeyConverter
	Accounts                     state.AccountsAdapter
	AccountsProposal             state.AccountsAdapter
	RequestHandler               process.RequestHandler
	TxProcessor                  process.TransactionProcessor
	ScProcessor                  process.SmartContractProcessor
	ScResultProcessor            process.SmartContractResultProcessor
	RewardsTxProcessor           process.RewardTransactionProcessor
	EconomicsFee                 process.FeeHandler
	GasHandler                   process.GasHandler
	BlockTracker                 preprocess.BlockTracker
	BlockSizeComputation         preprocess.BlockSizeComputationHandler
	BalanceComputation           preprocess.BalanceComputationHandler
	EnableEpochsHandler          common.EnableEpochsHandler
	EnableRoundsHandler          common.EnableRoundsHandler
	TxTypeHandler                process.TxTypeHandler
	ScheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	ProcessedMiniBlocksTracker   process.ProcessedMiniBlocksTracker
	TxExecutionOrderHandler      common.TxExecutionOrderHandler
	TxCacheSelectionConfig       config.TxCacheSelectionConfig
}

type preProcessorsContainerFactory struct {
	shardCoordinator             sharding.Coordinator
	store                        dataRetriever.StorageService
	marshalizer                  marshal.Marshalizer
	hasher                       hashing.Hasher
	dataPool                     dataRetriever.PoolsHolder
	pubkeyConverter              core.PubkeyConverter
	txProcessor                  process.TransactionProcessor
	scProcessor                  process.SmartContractProcessor
	scResultProcessor            process.SmartContractResultProcessor
	rewardsTxProcessor           process.RewardTransactionProcessor
	accounts                     state.AccountsAdapter
	accountsProposal             state.AccountsAdapter
	requestHandler               process.RequestHandler
	economicsFee                 process.FeeHandler
	gasHandler                   process.GasHandler
	blockTracker                 preprocess.BlockTracker
	blockSizeComputation         preprocess.BlockSizeComputationHandler
	balanceComputation           preprocess.BalanceComputationHandler
	enableEpochsHandler          common.EnableEpochsHandler
	enableRoundsHandler          common.EnableRoundsHandler
	txTypeHandler                process.TxTypeHandler
	scheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	processedMiniBlocksTracker   process.ProcessedMiniBlocksTracker
	txExecutionOrderHandler      common.TxExecutionOrderHandler
	txCacheSelectionConfig       config.TxCacheSelectionConfig
}

// NewPreProcessorsContainerFactory is responsible for creating a new preProcessors factory object
func NewPreProcessorsContainerFactory(args ArgsPreProcessorsContainerFactory) (*preProcessorsContainerFactory, error) {

	if check.IfNil(args.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.Store) {
		return nil, process.ErrNilStore
	}
	if check.IfNil(args.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(args.DataPool) {
		return nil, process.ErrNilDataPoolHolder
	}
	if check.IfNil(args.PubkeyConverter) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(args.TxProcessor) {
		return nil, process.ErrNilTxProcessor
	}
	if check.IfNil(args.Accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(args.AccountsProposal) {
		return nil, fmt.Errorf("%w for proposal", process.ErrNilAccountsAdapter)
	}
	if check.IfNil(args.ScProcessor) {
		return nil, process.ErrNilSmartContractProcessor
	}
	if check.IfNil(args.ScResultProcessor) {
		return nil, process.ErrNilSmartContractResultProcessor
	}
	if check.IfNil(args.RewardsTxProcessor) {
		return nil, process.ErrNilRewardsTxProcessor
	}
	if check.IfNil(args.RequestHandler) {
		return nil, process.ErrNilRequestHandler
	}
	if check.IfNil(args.EconomicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(args.GasHandler) {
		return nil, process.ErrNilGasHandler
	}
	if check.IfNil(args.BlockTracker) {
		return nil, process.ErrNilBlockTracker
	}
	if check.IfNil(args.BlockSizeComputation) {
		return nil, process.ErrNilBlockSizeComputationHandler
	}
	if check.IfNil(args.BalanceComputation) {
		return nil, process.ErrNilBalanceComputationHandler
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}
	if check.IfNil(args.EnableRoundsHandler) {
		return nil, process.ErrNilEnableRoundsHandler
	}
	if check.IfNil(args.TxTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(args.ScheduledTxsExecutionHandler) {
		return nil, process.ErrNilScheduledTxsExecutionHandler
	}
	if check.IfNil(args.ProcessedMiniBlocksTracker) {
		return nil, process.ErrNilProcessedMiniBlocksTracker
	}
	if check.IfNil(args.TxExecutionOrderHandler) {
		return nil, process.ErrNilTxExecutionOrderHandler
	}

	return &preProcessorsContainerFactory{
		shardCoordinator:             args.ShardCoordinator,
		store:                        args.Store,
		marshalizer:                  args.Marshalizer,
		hasher:                       args.Hasher,
		dataPool:                     args.DataPool,
		pubkeyConverter:              args.PubkeyConverter,
		txProcessor:                  args.TxProcessor,
		accounts:                     args.Accounts,
		accountsProposal:             args.AccountsProposal,
		scProcessor:                  args.ScProcessor,
		scResultProcessor:            args.ScResultProcessor,
		rewardsTxProcessor:           args.RewardsTxProcessor,
		requestHandler:               args.RequestHandler,
		economicsFee:                 args.EconomicsFee,
		gasHandler:                   args.GasHandler,
		blockTracker:                 args.BlockTracker,
		blockSizeComputation:         args.BlockSizeComputation,
		balanceComputation:           args.BalanceComputation,
		enableEpochsHandler:          args.EnableEpochsHandler,
		enableRoundsHandler:          args.EnableRoundsHandler,
		txTypeHandler:                args.TxTypeHandler,
		scheduledTxsExecutionHandler: args.ScheduledTxsExecutionHandler,
		processedMiniBlocksTracker:   args.ProcessedMiniBlocksTracker,
		txExecutionOrderHandler:      args.TxExecutionOrderHandler,
		txCacheSelectionConfig:       args.TxCacheSelectionConfig,
	}, nil
}

// Create returns a preprocessor container that will hold all preprocessors in the system
func (ppcm *preProcessorsContainerFactory) Create() (process.PreProcessorsContainer, error) {
	container := containers.NewPreProcessorsContainer()

	preproc, err := ppcm.createTxPreProcessor()
	if err != nil {
		return nil, err
	}

	err = container.Add(block.TxBlock, preproc)
	if err != nil {
		return nil, err
	}

	preproc, err = ppcm.createSmartContractResultPreProcessor()
	if err != nil {
		return nil, err
	}

	err = container.Add(block.SmartContractResultBlock, preproc)
	if err != nil {
		return nil, err
	}

	preproc, err = ppcm.createRewardsTransactionPreProcessor()
	if err != nil {
		return nil, err
	}

	err = container.Add(block.RewardsBlock, preproc)
	if err != nil {
		return nil, err
	}

	preproc, err = ppcm.createValidatorInfoPreProcessor()
	if err != nil {
		return nil, err
	}

	err = container.Add(block.PeerBlock, preproc)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (ppcm *preProcessorsContainerFactory) createTxPreProcessor() (process.PreProcessor, error) {
	args := preprocess.ArgsTransactionPreProcessor{
		BasePreProcessorArgs: preprocess.BasePreProcessorArgs{
			DataPool:                   ppcm.dataPool.Transactions(),
			Store:                      ppcm.store,
			Hasher:                     ppcm.hasher,
			Marshalizer:                ppcm.marshalizer,
			ShardCoordinator:           ppcm.shardCoordinator,
			Accounts:                   ppcm.accounts,
			AccountsProposal:           ppcm.accountsProposal,
			OnRequestTransaction:       ppcm.requestHandler.RequestTransactions,
			GasHandler:                 ppcm.gasHandler,
			PubkeyConverter:            ppcm.pubkeyConverter,
			BlockSizeComputation:       ppcm.blockSizeComputation,
			BalanceComputation:         ppcm.balanceComputation,
			ProcessedMiniBlocksTracker: ppcm.processedMiniBlocksTracker,
			TxExecutionOrderHandler:    ppcm.txExecutionOrderHandler,
			EconomicsFee:               ppcm.economicsFee,
			EnableEpochsHandler:        ppcm.enableEpochsHandler,
			EnableRoundsHandler:        ppcm.enableRoundsHandler,
		},
		TxProcessor:                  ppcm.txProcessor,
		BlockTracker:                 ppcm.blockTracker,
		BlockType:                    block.TxBlock,
		TxTypeHandler:                ppcm.txTypeHandler,
		ScheduledTxsExecutionHandler: ppcm.scheduledTxsExecutionHandler,
		TxCacheSelectionConfig:       ppcm.txCacheSelectionConfig,
	}

	return preprocess.NewTransactionPreprocessor(args)
}

func (ppcm *preProcessorsContainerFactory) createSmartContractResultPreProcessor() (process.PreProcessor, error) {
	args := preprocess.SmartContractResultsArgs{
		BasePreProcessorArgs: preprocess.BasePreProcessorArgs{
			DataPool:                   ppcm.dataPool.UnsignedTransactions(),
			Store:                      ppcm.store,
			Hasher:                     ppcm.hasher,
			Marshalizer:                ppcm.marshalizer,
			ShardCoordinator:           ppcm.shardCoordinator,
			Accounts:                   ppcm.accounts,
			AccountsProposal:           ppcm.accountsProposal,
			OnRequestTransaction:       ppcm.requestHandler.RequestUnsignedTransactions,
			GasHandler:                 ppcm.gasHandler,
			PubkeyConverter:            ppcm.pubkeyConverter,
			BlockSizeComputation:       ppcm.blockSizeComputation,
			BalanceComputation:         ppcm.balanceComputation,
			ProcessedMiniBlocksTracker: ppcm.processedMiniBlocksTracker,
			TxExecutionOrderHandler:    ppcm.txExecutionOrderHandler,
			EconomicsFee:               ppcm.economicsFee,
			EnableEpochsHandler:        ppcm.enableEpochsHandler,
			EnableRoundsHandler:        ppcm.enableRoundsHandler,
		},
		ScrProcessor: ppcm.scResultProcessor,
	}
	return preprocess.NewSmartContractResultPreprocessor(args)
}

func (ppcm *preProcessorsContainerFactory) createRewardsTransactionPreProcessor() (process.PreProcessor, error) {

	args := preprocess.RewardsPreProcessorArgs{
		BasePreProcessorArgs: preprocess.BasePreProcessorArgs{
			DataPool:                   ppcm.dataPool.RewardTransactions(),
			Store:                      ppcm.store,
			Hasher:                     ppcm.hasher,
			Marshalizer:                ppcm.marshalizer,
			ShardCoordinator:           ppcm.shardCoordinator,
			Accounts:                   ppcm.accounts,
			AccountsProposal:           ppcm.accountsProposal,
			OnRequestTransaction:       ppcm.requestHandler.RequestRewardTransactions,
			GasHandler:                 ppcm.gasHandler,
			PubkeyConverter:            ppcm.pubkeyConverter,
			BlockSizeComputation:       ppcm.blockSizeComputation,
			BalanceComputation:         ppcm.balanceComputation,
			ProcessedMiniBlocksTracker: ppcm.processedMiniBlocksTracker,
			TxExecutionOrderHandler:    ppcm.txExecutionOrderHandler,
			EconomicsFee:               ppcm.economicsFee,
			EnableEpochsHandler:        ppcm.enableEpochsHandler,
			EnableRoundsHandler:        ppcm.enableRoundsHandler,
		},
		RewardProcessor: ppcm.rewardsTxProcessor,
	}
	return preprocess.NewRewardTxPreprocessor(args)
}

func (ppcm *preProcessorsContainerFactory) createValidatorInfoPreProcessor() (process.PreProcessor, error) {
	args := preprocess.ValidatorInfoPreProcessorArgs{
		BasePreProcessorArgs: preprocess.BasePreProcessorArgs{
			DataPool:         ppcm.dataPool.ValidatorsInfo(),
			Store:            ppcm.store,
			Hasher:           ppcm.hasher,
			Marshalizer:      ppcm.marshalizer,
			ShardCoordinator: ppcm.shardCoordinator,
			Accounts:         ppcm.accounts,
			AccountsProposal: ppcm.accountsProposal,
			OnRequestTransaction: func(_ uint32, peerChangeHashes [][]byte) {
				ppcm.requestHandler.RequestValidatorsInfo(peerChangeHashes)
			},
			GasHandler:                 ppcm.gasHandler,
			PubkeyConverter:            ppcm.pubkeyConverter,
			BlockSizeComputation:       ppcm.blockSizeComputation,
			BalanceComputation:         ppcm.balanceComputation,
			ProcessedMiniBlocksTracker: ppcm.processedMiniBlocksTracker,
			TxExecutionOrderHandler:    ppcm.txExecutionOrderHandler,
			EconomicsFee:               ppcm.economicsFee,
			EnableEpochsHandler:        ppcm.enableEpochsHandler,
			EnableRoundsHandler:        ppcm.enableRoundsHandler,
		},
	}

	return preprocess.NewValidatorInfoPreprocessor(args)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppcm *preProcessorsContainerFactory) IsInterfaceNil() bool {
	return ppcm == nil
}
