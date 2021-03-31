package metachain

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ process.PreProcessorsContainerFactory = (*preProcessorsContainerFactory)(nil)

type preProcessorsContainerFactory struct {
	shardCoordinator               sharding.Coordinator
	store                          dataRetriever.StorageService
	marshalizer                    marshal.Marshalizer
	hasher                         hashing.Hasher
	dataPool                       dataRetriever.PoolsHolder
	txProcessor                    process.TransactionProcessor
	scResultProcessor              process.SmartContractResultProcessor
	accounts                       state.AccountsAdapter
	requestHandler                 process.RequestHandler
	economicsFee                   process.FeeHandler
	gasHandler                     process.GasHandler
	blockTracker                   preprocess.BlockTracker
	pubkeyConverter                core.PubkeyConverter
	blockSizeComputation           preprocess.BlockSizeComputationHandler
	balanceComputation             preprocess.BalanceComputationHandler
	epochNotifier                  process.EpochNotifier
	scheduledMiniBlocksEnableEpoch uint32
	txTypeHandler                  process.TxTypeHandler
	scheduledTxsExecutionHandler   process.ScheduledTxsExecutionHandler
}

// NewPreProcessorsContainerFactory is responsible for creating a new preProcessors factory object
func NewPreProcessorsContainerFactory(
	shardCoordinator sharding.Coordinator,
	store dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	dataPool dataRetriever.PoolsHolder,
	accounts state.AccountsAdapter,
	requestHandler process.RequestHandler,
	txProcessor process.TransactionProcessor,
	scResultProcessor process.SmartContractResultProcessor,
	economicsFee process.FeeHandler,
	gasHandler process.GasHandler,
	blockTracker preprocess.BlockTracker,
	pubkeyConverter core.PubkeyConverter,
	blockSizeComputation preprocess.BlockSizeComputationHandler,
	balanceComputation preprocess.BalanceComputationHandler,
	epochNotifier process.EpochNotifier,
	scheduledMiniBlocksEnableEpoch uint32,
	txTypeHandler process.TxTypeHandler,
	scheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler,
) (*preProcessorsContainerFactory, error) {

	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(store) {
		return nil, process.ErrNilStore
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(dataPool) {
		return nil, process.ErrNilDataPoolHolder
	}
	if check.IfNil(txProcessor) {
		return nil, process.ErrNilTxProcessor
	}
	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(requestHandler) {
		return nil, process.ErrNilRequestHandler
	}
	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(scResultProcessor) {
		return nil, process.ErrNilSmartContractResultProcessor
	}
	if check.IfNil(gasHandler) {
		return nil, process.ErrNilGasHandler
	}
	if check.IfNil(blockTracker) {
		return nil, process.ErrNilBlockTracker
	}
	if check.IfNil(pubkeyConverter) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(blockSizeComputation) {
		return nil, process.ErrNilBlockSizeComputationHandler
	}
	if check.IfNil(balanceComputation) {
		return nil, process.ErrNilBalanceComputationHandler
	}
	if check.IfNil(epochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}
	if check.IfNil(txTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(scheduledTxsExecutionHandler) {
		return nil, process.ErrNilScheduledTxsExecutionHandler
	}

	return &preProcessorsContainerFactory{
		shardCoordinator:               shardCoordinator,
		store:                          store,
		marshalizer:                    marshalizer,
		hasher:                         hasher,
		dataPool:                       dataPool,
		txProcessor:                    txProcessor,
		accounts:                       accounts,
		requestHandler:                 requestHandler,
		economicsFee:                   economicsFee,
		scResultProcessor:              scResultProcessor,
		gasHandler:                     gasHandler,
		blockTracker:                   blockTracker,
		pubkeyConverter:                pubkeyConverter,
		blockSizeComputation:           blockSizeComputation,
		balanceComputation:             balanceComputation,
		epochNotifier:                  epochNotifier,
		scheduledMiniBlocksEnableEpoch: scheduledMiniBlocksEnableEpoch,
		txTypeHandler:                  txTypeHandler,
		scheduledTxsExecutionHandler:   scheduledTxsExecutionHandler,
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
	txPreprocessor, err := preprocess.NewTransactionPreprocessor(
		ppcm.dataPool.Transactions(),
		ppcm.store,
		ppcm.hasher,
		ppcm.marshalizer,
		ppcm.txProcessor,
		ppcm.shardCoordinator,
		ppcm.accounts,
		ppcm.requestHandler.RequestTransaction,
		ppcm.economicsFee,
		ppcm.gasHandler,
		ppcm.blockTracker,
		block.TxBlock,
		ppcm.pubkeyConverter,
		ppcm.blockSizeComputation,
		ppcm.balanceComputation,
		ppcm.epochNotifier,
		ppcm.scheduledMiniBlocksEnableEpoch,
		ppcm.txTypeHandler,
		ppcm.scheduledTxsExecutionHandler,
	)

	return txPreprocessor, err
}

func (ppcm *preProcessorsContainerFactory) createSmartContractResultPreProcessor() (process.PreProcessor, error) {
	scrPreprocessor, err := preprocess.NewSmartContractResultPreprocessor(
		ppcm.dataPool.UnsignedTransactions(),
		ppcm.store,
		ppcm.hasher,
		ppcm.marshalizer,
		ppcm.scResultProcessor,
		ppcm.shardCoordinator,
		ppcm.accounts,
		ppcm.requestHandler.RequestUnsignedTransactions,
		ppcm.gasHandler,
		ppcm.economicsFee,
		ppcm.pubkeyConverter,
		ppcm.blockSizeComputation,
		ppcm.balanceComputation,
	)

	return scrPreprocessor, err
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppcm *preProcessorsContainerFactory) IsInterfaceNil() bool {
	return ppcm == nil
}
