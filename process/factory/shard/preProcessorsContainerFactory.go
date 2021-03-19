package shard

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
	pubkeyConverter                core.PubkeyConverter
	txProcessor                    process.TransactionProcessor
	scProcessor                    process.SmartContractProcessor
	scResultProcessor              process.SmartContractResultProcessor
	rewardsTxProcessor             process.RewardTransactionProcessor
	accounts                       state.AccountsAdapter
	requestHandler                 process.RequestHandler
	economicsFee                   process.FeeHandler
	gasHandler                     process.GasHandler
	blockTracker                   preprocess.BlockTracker
	blockSizeComputation           preprocess.BlockSizeComputationHandler
	balanceComputation             preprocess.BalanceComputationHandler
	epochNotifier                  process.EpochNotifier
	scheduledMiniBlocksEnableEpoch uint32
	txTypeHandler                  process.TxTypeHandler
	scheduledTxsExecutionHandler   preprocess.ScheduledTxsExecutionHandler
}

// NewPreProcessorsContainerFactory is responsible for creating a new preProcessors factory object
func NewPreProcessorsContainerFactory(
	shardCoordinator sharding.Coordinator,
	store dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	dataPool dataRetriever.PoolsHolder,
	pubkeyConverter core.PubkeyConverter,
	accounts state.AccountsAdapter,
	requestHandler process.RequestHandler,
	txProcessor process.TransactionProcessor,
	scProcessor process.SmartContractProcessor,
	scResultProcessor process.SmartContractResultProcessor,
	rewardsTxProcessor process.RewardTransactionProcessor,
	economicsFee process.FeeHandler,
	gasHandler process.GasHandler,
	blockTracker preprocess.BlockTracker,
	blockSizeComputation preprocess.BlockSizeComputationHandler,
	balanceComputation preprocess.BalanceComputationHandler,
	epochNotifier process.EpochNotifier,
	scheduledMiniBlocksEnableEpoch uint32,
	txTypeHandler process.TxTypeHandler,
	scheduledTxsExecutionHandler preprocess.ScheduledTxsExecutionHandler,
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
	if check.IfNil(pubkeyConverter) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(txProcessor) {
		return nil, process.ErrNilTxProcessor
	}
	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(scProcessor) {
		return nil, process.ErrNilSmartContractProcessor
	}
	if check.IfNil(scResultProcessor) {
		return nil, process.ErrNilSmartContractResultProcessor
	}
	if check.IfNil(rewardsTxProcessor) {
		return nil, process.ErrNilRewardsTxProcessor
	}
	if check.IfNil(requestHandler) {
		return nil, process.ErrNilRequestHandler
	}
	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(gasHandler) {
		return nil, process.ErrNilGasHandler
	}
	if check.IfNil(blockTracker) {
		return nil, process.ErrNilBlockTracker
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
		pubkeyConverter:                pubkeyConverter,
		txProcessor:                    txProcessor,
		accounts:                       accounts,
		scProcessor:                    scProcessor,
		scResultProcessor:              scResultProcessor,
		rewardsTxProcessor:             rewardsTxProcessor,
		requestHandler:                 requestHandler,
		economicsFee:                   economicsFee,
		gasHandler:                     gasHandler,
		blockTracker:                   blockTracker,
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

func (ppcm *preProcessorsContainerFactory) createRewardsTransactionPreProcessor() (process.PreProcessor, error) {
	rewardTxPreprocessor, err := preprocess.NewRewardTxPreprocessor(
		ppcm.dataPool.RewardTransactions(),
		ppcm.store,
		ppcm.hasher,
		ppcm.marshalizer,
		ppcm.rewardsTxProcessor,
		ppcm.shardCoordinator,
		ppcm.accounts,
		ppcm.requestHandler.RequestRewardTransactions,
		ppcm.gasHandler,
		ppcm.pubkeyConverter,
		ppcm.blockSizeComputation,
		ppcm.balanceComputation,
	)

	return rewardTxPreprocessor, err
}

func (ppcm *preProcessorsContainerFactory) createValidatorInfoPreProcessor() (process.PreProcessor, error) {
	validatorInfoPreprocessor, err := preprocess.NewValidatorInfoPreprocessor(
		ppcm.hasher,
		ppcm.marshalizer,
		ppcm.blockSizeComputation,
	)

	return validatorInfoPreprocessor, err
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppcm *preProcessorsContainerFactory) IsInterfaceNil() bool {
	return ppcm == nil
}
