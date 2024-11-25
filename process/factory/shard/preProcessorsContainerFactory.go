package shard

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
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
	marshalizer                  marshal.Marshalizer
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
	enableEpochsHandler common.EnableEpochsHandler,
	txTypeHandler process.TxTypeHandler,
	scheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler,
	processedMiniBlocksTracker process.ProcessedMiniBlocksTracker,
	txExecutionOrderHandler common.TxExecutionOrderHandler,
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
	if check.IfNil(enableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}
	if check.IfNil(txTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(scheduledTxsExecutionHandler) {
		return nil, process.ErrNilScheduledTxsExecutionHandler
	}
	if check.IfNil(processedMiniBlocksTracker) {
		return nil, process.ErrNilProcessedMiniBlocksTracker
	}
	if check.IfNil(txExecutionOrderHandler) {
		return nil, process.ErrNilTxExecutionOrderHandler
	}

	return &preProcessorsContainerFactory{
		shardCoordinator:             shardCoordinator,
		store:                        store,
		marshalizer:                  marshalizer,
		hasher:                       hasher,
		dataPool:                     dataPool,
		pubkeyConverter:              pubkeyConverter,
		txProcessor:                  txProcessor,
		accounts:                     accounts,
		scProcessor:                  scProcessor,
		scResultProcessor:            scResultProcessor,
		rewardsTxProcessor:           rewardsTxProcessor,
		requestHandler:               requestHandler,
		economicsFee:                 economicsFee,
		gasHandler:                   gasHandler,
		blockTracker:                 blockTracker,
		blockSizeComputation:         blockSizeComputation,
		balanceComputation:           balanceComputation,
		enableEpochsHandler:          enableEpochsHandler,
		txTypeHandler:                txTypeHandler,
		scheduledTxsExecutionHandler: scheduledTxsExecutionHandler,
		processedMiniBlocksTracker:   processedMiniBlocksTracker,
		txExecutionOrderHandler:      txExecutionOrderHandler,
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
		TxDataPool:                   ppcm.dataPool.Transactions(),
		Store:                        ppcm.store,
		Hasher:                       ppcm.hasher,
		Marshalizer:                  ppcm.marshalizer,
		TxProcessor:                  ppcm.txProcessor,
		ShardCoordinator:             ppcm.shardCoordinator,
		Accounts:                     ppcm.accounts,
		OnRequestTransaction:         ppcm.requestHandler.RequestTransaction,
		EconomicsFee:                 ppcm.economicsFee,
		GasHandler:                   ppcm.gasHandler,
		BlockTracker:                 ppcm.blockTracker,
		BlockType:                    block.TxBlock,
		PubkeyConverter:              ppcm.pubkeyConverter,
		BlockSizeComputation:         ppcm.blockSizeComputation,
		BalanceComputation:           ppcm.balanceComputation,
		EnableEpochsHandler:          ppcm.enableEpochsHandler,
		TxTypeHandler:                ppcm.txTypeHandler,
		ScheduledTxsExecutionHandler: ppcm.scheduledTxsExecutionHandler,
		ProcessedMiniBlocksTracker:   ppcm.processedMiniBlocksTracker,
		TxExecutionOrderHandler:      ppcm.txExecutionOrderHandler,
	}

	txPreprocessor, err := preprocess.NewTransactionPreprocessor(args)

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
		ppcm.enableEpochsHandler,
		ppcm.processedMiniBlocksTracker,
		ppcm.txExecutionOrderHandler,
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
		ppcm.processedMiniBlocksTracker,
		ppcm.txExecutionOrderHandler,
	)

	return rewardTxPreprocessor, err
}

func (ppcm *preProcessorsContainerFactory) createValidatorInfoPreProcessor() (process.PreProcessor, error) {
	validatorInfoPreprocessor, err := preprocess.NewValidatorInfoPreprocessor(
		ppcm.hasher,
		ppcm.marshalizer,
		ppcm.blockSizeComputation,
		ppcm.dataPool.ValidatorsInfo(),
		ppcm.store,
		ppcm.enableEpochsHandler,
	)

	return validatorInfoPreprocessor, err
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppcm *preProcessorsContainerFactory) IsInterfaceNil() bool {
	return ppcm == nil
}
