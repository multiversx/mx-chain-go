package metachain

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/factory/containers"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
)

var _ process.PreProcessorsContainerFactory = (*preProcessorsContainerFactory)(nil)

type preProcessorsContainerFactory struct {
	shardCoordinator             sharding.Coordinator
	store                        dataRetriever.StorageService
	marshaller                   marshal.Marshalizer
	hasher                       hashing.Hasher
	dataPool                     dataRetriever.PoolsHolder
	txProcessor                  process.TransactionProcessor
	scResultProcessor            process.SmartContractResultProcessor
	accounts                     state.AccountsAdapter
	requestHandler               process.RequestHandler
	economicsFee                 process.FeeHandler
	gasHandler                   process.GasHandler
	blockTracker                 preprocess.BlockTracker
	pubkeyConverter              core.PubkeyConverter
	blockSizeComputation         preprocess.BlockSizeComputationHandler
	balanceComputation           preprocess.BalanceComputationHandler
	enableEpochsHandler          common.EnableEpochsHandler
	txTypeHandler                process.TxTypeHandler
	scheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	processedMiniBlocksTracker   process.ProcessedMiniBlocksTracker
	txExecutionOrderHandler      common.TxExecutionOrderHandler
	txPreprocessorCreator        preprocess.TxPreProcessorCreator
}

// ArgPreProcessorsContainerFactory defines the arguments needed by the pre-processor container factory
type ArgPreProcessorsContainerFactory struct {
	ShardCoordinator             sharding.Coordinator
	Store                        dataRetriever.StorageService
	Marshaller                   marshal.Marshalizer
	Hasher                       hashing.Hasher
	DataPool                     dataRetriever.PoolsHolder
	Accounts                     state.AccountsAdapter
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
	TxTypeHandler                process.TxTypeHandler
	ScheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	ProcessedMiniBlocksTracker   process.ProcessedMiniBlocksTracker
	TxExecutionOrderHandler      common.TxExecutionOrderHandler
	TxPreProcessorCreator        preprocess.TxPreProcessorCreator
}

// NewPreProcessorsContainerFactory is responsible for creating a new preProcessors factory object
func NewPreProcessorsContainerFactory(args ArgPreProcessorsContainerFactory) (*preProcessorsContainerFactory, error) {
	err := checkPreProcessorContainerFactoryNilParameters(args)
	if err != nil {
		return nil, err
	}

	return &preProcessorsContainerFactory{
		shardCoordinator:             args.ShardCoordinator,
		store:                        args.Store,
		marshaller:                   args.Marshaller,
		hasher:                       args.Hasher,
		dataPool:                     args.DataPool,
		txProcessor:                  args.TxProcessor,
		accounts:                     args.Accounts,
		requestHandler:               args.RequestHandler,
		economicsFee:                 args.EconomicsFee,
		scResultProcessor:            args.ScResultProcessor,
		gasHandler:                   args.GasHandler,
		blockTracker:                 args.BlockTracker,
		pubkeyConverter:              args.PubkeyConverter,
		blockSizeComputation:         args.BlockSizeComputation,
		balanceComputation:           args.BalanceComputation,
		enableEpochsHandler:          args.EnableEpochsHandler,
		txTypeHandler:                args.TxTypeHandler,
		scheduledTxsExecutionHandler: args.ScheduledTxsExecutionHandler,
		processedMiniBlocksTracker:   args.ProcessedMiniBlocksTracker,
		txExecutionOrderHandler:      args.TxExecutionOrderHandler,
		txPreprocessorCreator:        args.TxPreProcessorCreator,
	}, nil
}

// Create returns a preprocessor container that will hold all preprocessors in the system
func (ppcf *preProcessorsContainerFactory) Create() (process.PreProcessorsContainer, error) {
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

	return container, nil
}

func (ppcf *preProcessorsContainerFactory) createTxPreProcessor() (process.PreProcessor, error) {
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

	return ppcf.txPreprocessorCreator.CreateTxPreProcessor(args)
}

func (ppcf *preProcessorsContainerFactory) createSmartContractResultPreProcessor() (process.PreProcessor, error) {
	scrPreprocessor, err := preprocess.NewSmartContractResultPreprocessor(
		ppcf.dataPool.UnsignedTransactions(),
		ppcf.store,
		ppcf.hasher,
		ppcf.marshaller,
		ppcf.scResultProcessor,
		ppcf.shardCoordinator,
		ppcf.accounts,
		ppcf.requestHandler.RequestUnsignedTransactions,
		ppcf.gasHandler,
		ppcf.economicsFee,
		ppcf.pubkeyConverter,
		ppcf.blockSizeComputation,
		ppcf.balanceComputation,
		ppcf.enableEpochsHandler,
		ppcf.processedMiniBlocksTracker,
		ppcf.txExecutionOrderHandler,
	)

	return scrPreprocessor, err
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppcf *preProcessorsContainerFactory) IsInterfaceNil() bool {
	return ppcf == nil
}

func checkPreProcessorContainerFactoryNilParameters(args ArgPreProcessorsContainerFactory) error {
	if check.IfNil(args.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(args.Store) {
		return process.ErrNilStore
	}
	if check.IfNil(args.Marshaller) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return process.ErrNilHasher
	}
	if check.IfNil(args.DataPool) {
		return process.ErrNilDataPoolHolder
	}
	if check.IfNil(args.TxProcessor) {
		return process.ErrNilTxProcessor
	}
	if check.IfNil(args.Accounts) {
		return process.ErrNilAccountsAdapter
	}
	if check.IfNil(args.RequestHandler) {
		return process.ErrNilRequestHandler
	}
	if check.IfNil(args.EconomicsFee) {
		return process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(args.ScResultProcessor) {
		return process.ErrNilSmartContractResultProcessor
	}
	if check.IfNil(args.GasHandler) {
		return process.ErrNilGasHandler
	}
	if check.IfNil(args.BlockTracker) {
		return process.ErrNilBlockTracker
	}
	if check.IfNil(args.PubkeyConverter) {
		return process.ErrNilPubkeyConverter
	}
	if check.IfNil(args.BlockSizeComputation) {
		return process.ErrNilBlockSizeComputationHandler
	}
	if check.IfNil(args.BalanceComputation) {
		return process.ErrNilBalanceComputationHandler
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return process.ErrNilEnableEpochsHandler
	}
	if check.IfNil(args.TxTypeHandler) {
		return process.ErrNilTxTypeHandler
	}
	if check.IfNil(args.ScheduledTxsExecutionHandler) {
		return process.ErrNilScheduledTxsExecutionHandler
	}
	if check.IfNil(args.ProcessedMiniBlocksTracker) {
		return process.ErrNilProcessedMiniBlocksTracker
	}
	if check.IfNil(args.TxExecutionOrderHandler) {
		return process.ErrNilTxExecutionOrderHandler
	}
	if check.IfNil(args.TxPreProcessorCreator) {
		return errors.ErrNilTxPreProcessorCreator
	}

	return nil
}
