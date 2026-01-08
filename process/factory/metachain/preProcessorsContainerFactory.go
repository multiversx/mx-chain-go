package metachain

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

// ArgsPreProcessorsContainerFactory holds the arguments needed for creating a new preProcessorsContainerFactory instance
type ArgsPreProcessorsContainerFactory struct {
	ShardCoordinator             sharding.Coordinator
	Store                        dataRetriever.StorageService
	Marshalizer                  marshal.Marshalizer
	Hasher                       hashing.Hasher
	DataPool                     dataRetriever.PoolsHolder
	Accounts                     state.AccountsAdapter
	AccountsProposal             state.AccountsAdapter
	RequestHandler               process.RequestHandler
	TxProcessor                  process.TransactionProcessor
	ScResultProcessor            process.SmartContractResultProcessor
	EconomicsFee                 process.FeeHandler
	GasHandler                   process.GasHandler
	BlockTracker                 preprocess.BlockTracker
	PubkeyConverter              core.PubkeyConverter
	BlockSizeComputation         preprocess.BlockSizeComputationHandler
	BalanceComputation           preprocess.BalanceComputationHandler
	EnableEpochsHandler          common.EnableEpochsHandler
	EpochNotifier                process.EpochNotifier
	EnableRoundsHandler          common.EnableRoundsHandler
	RoundNotifier                process.RoundNotifier
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
	txProcessor                  process.TransactionProcessor
	scResultProcessor            process.SmartContractResultProcessor
	accounts                     state.AccountsAdapter
	accountsProposal             state.AccountsAdapter
	requestHandler               process.RequestHandler
	economicsFee                 process.FeeHandler
	gasHandler                   process.GasHandler
	blockTracker                 preprocess.BlockTracker
	pubkeyConverter              core.PubkeyConverter
	blockSizeComputation         preprocess.BlockSizeComputationHandler
	balanceComputation           preprocess.BalanceComputationHandler
	enableEpochsHandler          common.EnableEpochsHandler
	epochNotifier                process.EpochNotifier
	enableRoundsHandler          common.EnableRoundsHandler
	roundNotifier                process.RoundNotifier
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
	if check.IfNil(args.TxProcessor) {
		return nil, process.ErrNilTxProcessor
	}
	if check.IfNil(args.Accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(args.AccountsProposal) {
		return nil, fmt.Errorf("%w for proposal", process.ErrNilAccountsAdapter)
	}
	if check.IfNil(args.RequestHandler) {
		return nil, process.ErrNilRequestHandler
	}
	if check.IfNil(args.EconomicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(args.ScResultProcessor) {
		return nil, process.ErrNilSmartContractResultProcessor
	}
	if check.IfNil(args.GasHandler) {
		return nil, process.ErrNilGasHandler
	}
	if check.IfNil(args.BlockTracker) {
		return nil, process.ErrNilBlockTracker
	}
	if check.IfNil(args.PubkeyConverter) {
		return nil, process.ErrNilPubkeyConverter
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
	if check.IfNil(args.EnableRoundsHandler) {
		return nil, process.ErrNilEnableRoundsHandler
	}

	return &preProcessorsContainerFactory{
		shardCoordinator:             args.ShardCoordinator,
		store:                        args.Store,
		marshalizer:                  args.Marshalizer,
		hasher:                       args.Hasher,
		dataPool:                     args.DataPool,
		txProcessor:                  args.TxProcessor,
		accounts:                     args.Accounts,
		accountsProposal:             args.AccountsProposal,
		requestHandler:               args.RequestHandler,
		economicsFee:                 args.EconomicsFee,
		scResultProcessor:            args.ScResultProcessor,
		gasHandler:                   args.GasHandler,
		blockTracker:                 args.BlockTracker,
		pubkeyConverter:              args.PubkeyConverter,
		blockSizeComputation:         args.BlockSizeComputation,
		balanceComputation:           args.BalanceComputation,
		enableEpochsHandler:          args.EnableEpochsHandler,
		epochNotifier:                args.EpochNotifier,
		enableRoundsHandler:          args.EnableRoundsHandler,
		roundNotifier:                args.RoundNotifier,
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
			EpochNotifier:              ppcm.epochNotifier,
			EnableRoundsHandler:        ppcm.enableRoundsHandler,
			RoundNotifier:              ppcm.roundNotifier,
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
			EpochNotifier:              ppcm.epochNotifier,
			EnableRoundsHandler:        ppcm.enableRoundsHandler,
			RoundNotifier:              ppcm.roundNotifier,
		},
		ScrProcessor: ppcm.scResultProcessor,
	}

	return preprocess.NewSmartContractResultPreprocessor(args)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppcm *preProcessorsContainerFactory) IsInterfaceNil() bool {
	return ppcm == nil
}
