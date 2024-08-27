package sovereign

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/factory/containers"
	"github.com/multiversx/mx-chain-go/process/factory/shard"
	"github.com/multiversx/mx-chain-go/process/factory/shard/data"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
)

var _ process.PreProcessorsContainerFactory = (*sovereignPreProcessorsContainerFactory)(nil)

type sovereignPreProcessorsContainerFactory struct {
	shardCoordinator             sharding.Coordinator
	store                        dataRetriever.StorageService
	marshaller                   marshal.Marshalizer
	hasher                       hashing.Hasher
	dataPool                     dataRetriever.PoolsHolder
	pubkeyConverter              core.PubkeyConverter
	txProcessor                  process.TransactionProcessor
	scProcessor                  process.SmartContractProcessor
	scResultProcessor            process.SmartContractResultProcessor
	rewardsTxProcessor           process.RewardTransactionProcessor
	accounts                     state.AccountsAdapter
	requestHandler               process.RequestHandler
	economicsFee                 process.FeeHandler
	gasHandler                   process.GasHandler
	blockTracker                 preprocess.BlockTracker
	blockSizeComputation         preprocess.BlockSizeComputationHandler
	balanceComputation           preprocess.BalanceComputationHandler
	enableEpochsHandler          common.EnableEpochsHandler
	txTypeHandler                process.TxTypeHandler
	scheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	processedMiniBlocksTracker   process.ProcessedMiniBlocksTracker
	txExecutionOrderHandler      common.TxExecutionOrderHandler
	runTypeComponents            data.RunTypeComponents
}

// NewSovereignPreProcessorsContainerFactory is responsible for creating a new sovereign pre-processors factory object
func NewSovereignPreProcessorsContainerFactory(args data.ArgPreProcessorsContainerFactory) (*sovereignPreProcessorsContainerFactory, error) {
	err := shard.CheckPreProcessorContainerFactoryNilParameters(args)
	if err != nil {
		return nil, err
	}

	return &sovereignPreProcessorsContainerFactory{
		shardCoordinator:             args.ShardCoordinator,
		store:                        args.Store,
		marshaller:                   args.Marshaller,
		hasher:                       args.Hasher,
		dataPool:                     args.DataPool,
		pubkeyConverter:              args.PubkeyConverter,
		txProcessor:                  args.TxProcessor,
		accounts:                     args.Accounts,
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
		txTypeHandler:                args.TxTypeHandler,
		scheduledTxsExecutionHandler: args.ScheduledTxsExecutionHandler,
		processedMiniBlocksTracker:   args.ProcessedMiniBlocksTracker,
		txExecutionOrderHandler:      args.TxExecutionOrderHandler,
		runTypeComponents:            args.RunTypeComponents,
	}, nil
}

// Create returns a preprocessor container that will hold all preprocessors in the system
func (ppcf *sovereignPreProcessorsContainerFactory) Create() (process.PreProcessorsContainer, error) {
	container := containers.NewPreProcessorsContainer()

	preproc, err := ppcf.createTxPreProcessor()
	if err != nil {
		return nil, err
	}

	err = container.Add(block.TxBlock, preproc)
	if err != nil {
		return nil, err
	}

	preproc, err = ppcf.createSmartContractResultPreProcessor()
	if err != nil {
		return nil, err
	}

	err = container.Add(block.SmartContractResultBlock, preproc)
	if err != nil {
		return nil, err
	}

	preproc, err = ppcf.createValidatorInfoPreProcessor()
	if err != nil {
		return nil, err
	}

	err = container.Add(block.PeerBlock, preproc)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (ppcf *sovereignPreProcessorsContainerFactory) createTxPreProcessor() (process.PreProcessor, error) {
	args := preprocess.ArgsTransactionPreProcessor{
		TxDataPool:                   ppcf.dataPool.Transactions(),
		Store:                        ppcf.store,
		Hasher:                       ppcf.hasher,
		Marshalizer:                  ppcf.marshaller,
		TxProcessor:                  ppcf.txProcessor,
		ShardCoordinator:             ppcf.shardCoordinator,
		Accounts:                     ppcf.accounts,
		OnRequestTransaction:         ppcf.requestHandler.RequestTransaction,
		EconomicsFee:                 ppcf.economicsFee,
		GasHandler:                   ppcf.gasHandler,
		BlockTracker:                 ppcf.blockTracker,
		BlockType:                    block.TxBlock,
		PubkeyConverter:              ppcf.pubkeyConverter,
		BlockSizeComputation:         ppcf.blockSizeComputation,
		BalanceComputation:           ppcf.balanceComputation,
		EnableEpochsHandler:          ppcf.enableEpochsHandler,
		TxTypeHandler:                ppcf.txTypeHandler,
		ScheduledTxsExecutionHandler: ppcf.scheduledTxsExecutionHandler,
		ProcessedMiniBlocksTracker:   ppcf.processedMiniBlocksTracker,
		TxExecutionOrderHandler:      ppcf.txExecutionOrderHandler,
	}

	return ppcf.runTypeComponents.TxPreProcessorCreator().CreateTxPreProcessor(args)
}

func (ppcf *sovereignPreProcessorsContainerFactory) createSmartContractResultPreProcessor() (process.PreProcessor, error) {
	arg := preprocess.SmartContractResultPreProcessorCreatorArgs{
		ScrDataPool:                  ppcf.dataPool.UnsignedTransactions(),
		Store:                        ppcf.store,
		Hasher:                       ppcf.hasher,
		Marshalizer:                  ppcf.marshaller,
		ScrProcessor:                 ppcf.scResultProcessor,
		ShardCoordinator:             ppcf.shardCoordinator,
		Accounts:                     ppcf.accounts,
		OnRequestSmartContractResult: ppcf.requestHandler.RequestUnsignedTransactions,
		GasHandler:                   ppcf.gasHandler,
		EconomicsFee:                 ppcf.economicsFee,
		PubkeyConverter:              ppcf.pubkeyConverter,
		BlockSizeComputation:         ppcf.blockSizeComputation,
		BalanceComputation:           ppcf.balanceComputation,
		EnableEpochsHandler:          ppcf.enableEpochsHandler,
		ProcessedMiniBlocksTracker:   ppcf.processedMiniBlocksTracker,
		TxExecutionOrderHandler:      ppcf.txExecutionOrderHandler,
	}

	return ppcf.runTypeComponents.SCResultsPreProcessorCreator().CreateSmartContractResultPreProcessor(arg)
}

func (ppcf *sovereignPreProcessorsContainerFactory) createValidatorInfoPreProcessor() (process.PreProcessor, error) {
	validatorInfoPreprocessor, err := preprocess.NewValidatorInfoPreprocessor(
		ppcf.hasher,
		ppcf.marshaller,
		ppcf.blockSizeComputation,
		ppcf.dataPool.ValidatorsInfo(),
		ppcf.store,
		ppcf.enableEpochsHandler,
	)

	return validatorInfoPreprocessor, err
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppcf *sovereignPreProcessorsContainerFactory) IsInterfaceNil() bool {
	return ppcf == nil
}
