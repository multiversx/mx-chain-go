package runType

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos/sposFactory"
	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	requesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/resolverscontainer"
	storageRequestFactory "github.com/multiversx/mx-chain-go/dataRetriever/factory/storageRequestersContainer/factory"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/processing/api"
	factoryVm "github.com/multiversx/mx-chain-go/factory/vm"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/checking"
	processComp "github.com/multiversx/mx-chain-go/genesis/process"
	"github.com/multiversx/mx-chain-go/node/external/transactionAPI"
	trieIteratorsFactory "github.com/multiversx/mx-chain-go/node/trieIterators/factory"
	"github.com/multiversx/mx-chain-go/process"
	processBlock "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/factory/interceptorscontainer"
	"github.com/multiversx/mx-chain-go/process/factory/shard/data"
	"github.com/multiversx/mx-chain-go/process/headerCheck"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/process/scToProtocol"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	processSync "github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/process/sync/storageBootstrap"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	syncerFactory "github.com/multiversx/mx-chain-go/state/syncer/factory"
	"github.com/multiversx/mx-chain-go/storage/latestData"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

var _ factory.ComponentHandler = (*managedRunTypeComponents)(nil)
var _ factory.RunTypeComponentsHandler = (*managedRunTypeComponents)(nil)
var _ factory.RunTypeComponentsHolder = (*managedRunTypeComponents)(nil)

type managedRunTypeComponents struct {
	*runTypeComponents
	factory              runTypeComponentsCreator
	mutRunTypeComponents sync.RWMutex
}

// NewManagedRunTypeComponents returns a news instance of managedRunTypeComponents
func NewManagedRunTypeComponents(rcf runTypeComponentsCreator) (*managedRunTypeComponents, error) {
	if rcf == nil {
		return nil, errors.ErrNilRunTypeComponentsFactory
	}

	return &managedRunTypeComponents{
		runTypeComponents: nil,
		factory:           rcf,
	}, nil
}

// Create will create the managed components
func (mrc *managedRunTypeComponents) Create() error {
	rtc, err := mrc.factory.Create()
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrRunTypeComponentsFactoryCreate, err)
	}

	mrc.mutRunTypeComponents.Lock()
	mrc.runTypeComponents = rtc
	mrc.mutRunTypeComponents.Unlock()

	return nil
}

// Close will close all underlying subcomponents
func (mrc *managedRunTypeComponents) Close() error {
	mrc.mutRunTypeComponents.Lock()
	defer mrc.mutRunTypeComponents.Unlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	err := mrc.runTypeComponents.Close()
	if err != nil {
		return err
	}
	mrc.runTypeComponents = nil

	return nil
}

// CheckSubcomponents verifies all subcomponents
func (mrc *managedRunTypeComponents) CheckSubcomponents() error {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return errors.ErrNilRunTypeComponents
	}
	if check.IfNil(mrc.blockProcessorCreator) {
		return errors.ErrNilBlockProcessorCreator
	}
	if check.IfNil(mrc.blockChainHookHandlerCreator) {
		return errors.ErrNilBlockChainHookHandlerCreator
	}
	if check.IfNil(mrc.bootstrapperFromStorageCreator) {
		return errors.ErrNilBootstrapperFromStorageCreator
	}
	if check.IfNil(mrc.bootstrapperCreator) {
		return errors.ErrNilBootstrapperCreator
	}
	if check.IfNil(mrc.blockTrackerCreator) {
		return errors.ErrNilBlockTrackerCreator
	}
	if check.IfNil(mrc.epochStartBootstrapperCreator) {
		return errors.ErrNilEpochStartBootstrapperCreator
	}
	if check.IfNil(mrc.forkDetectorCreator) {
		return errors.ErrNilForkDetectorCreator
	}
	if check.IfNil(mrc.headerValidatorCreator) {
		return errors.ErrNilHeaderValidatorCreator
	}
	if check.IfNil(mrc.requestHandlerCreator) {
		return errors.ErrNilRequestHandlerCreator
	}
	if check.IfNil(mrc.scheduledTxsExecutionCreator) {
		return errors.ErrNilScheduledTxsExecutionCreator
	}
	if check.IfNil(mrc.transactionCoordinatorCreator) {
		return errors.ErrNilTransactionCoordinatorCreator
	}
	if check.IfNil(mrc.validatorStatisticsProcessorCreator) {
		return errors.ErrNilValidatorStatisticsProcessorCreator
	}
	if check.IfNil(mrc.additionalStorageServiceCreator) {
		return errors.ErrNilAdditionalStorageServiceCreator
	}
	if check.IfNil(mrc.scProcessorCreator) {
		return errors.ErrNilSCProcessorCreator
	}
	if check.IfNil(mrc.scResultPreProcessorCreator) {
		return errors.ErrNilSCResultsPreProcessorCreator
	}
	if check.IfNil(mrc.vmContextCreator) {
		return errors.ErrNilVMContextCreator
	}
	if mrc.consensusModel == consensus.ConsensusModelInvalid {
		return errors.ErrInvalidConsensusModel
	}
	if check.IfNil(mrc.vmContainerMetaFactory) {
		return errors.ErrNilVmContainerMetaFactoryCreator
	}
	if check.IfNil(mrc.vmContainerShardFactory) {
		return errors.ErrNilVmContainerShardFactoryCreator
	}
	if check.IfNil(mrc.accountsParser) {
		return errors.ErrNilAccountsParser
	}
	if check.IfNil(mrc.accountsCreator) {
		return errors.ErrNilAccountsCreator
	}
	if check.IfNil(mrc.outGoingOperationsPoolHandler) {
		return errors.ErrNilOutGoingOperationsPool
	}
	if check.IfNil(mrc.dataCodecHandler) {
		return errors.ErrNilDataCodec
	}
	if check.IfNil(mrc.topicsCheckerHandler) {
		return errors.ErrNilTopicsChecker
	}
	if check.IfNil(mrc.shardCoordinatorCreator) {
		return errors.ErrNilShardCoordinatorFactory
	}
	if check.IfNil(mrc.nodesCoordinatorWithRaterFactoryCreator) {
		return errors.ErrNilNodesCoordinatorFactory
	}
	if check.IfNil(mrc.requestersContainerFactoryCreator) {
		return errors.ErrNilRequesterContainerFactoryCreator
	}
	if check.IfNil(mrc.interceptorsContainerFactoryCreator) {
		return errors.ErrNilInterceptorsContainerFactoryCreator
	}
	if check.IfNil(mrc.shardResolversContainerFactoryCreator) {
		return errors.ErrNilShardResolversContainerFactoryCreator
	}
	if check.IfNil(mrc.txPreProcessorCreator) {
		return errors.ErrNilTxPreProcessorCreator
	}
	if check.IfNil(mrc.extraHeaderSigVerifierHolder) {
		return errors.ErrNilExtraHeaderSigVerifierHolder
	}
	if check.IfNil(mrc.genesisBlockCreatorFactory) {
		return errors.ErrNilGenesisBlockFactory
	}
	if check.IfNil(mrc.genesisMetaBlockCheckerCreator) {
		return errors.ErrNilGenesisMetaBlockChecker
	}
	if check.IfNil(mrc.epochStartTriggerFactory) {
		return errors.ErrNilEpochStartTriggerFactory
	}
	if check.IfNil(mrc.latestDataProviderFactory) {
		return errors.ErrNilLatestDataProviderFactory
	}
	if check.IfNil(mrc.scToProtocolFactory) {
		return errors.ErrNilStakingToPeerFactory
	}
	if check.IfNil(mrc.validatorInfoCreatorFactory) {
		return errors.ErrNilValidatorInfoCreatorFactory
	}
	if check.IfNil(mrc.apiProcessorCompsCreatorHandler) {
		return errors.ErrNilAPIProcessorCompsCreator
	}
	if check.IfNil(mrc.endOfEpochEconomicsFactoryHandler) {
		return errors.ErrNilEndOfEpochEconomicsFactory
	}
	if check.IfNil(mrc.rewardsCreatorFactory) {
		return errors.ErrNilRewardsFactory
	}
	if check.IfNil(mrc.systemSCProcessorFactory) {
		return errors.ErrNilSysSCFactory
	}
	if check.IfNil(mrc.preProcessorsContainerFactoryCreator) {
		return errors.ErrNilPreProcessorsContainerFactoryCreator
	}
	if check.IfNil(mrc.dataRetrieverContainersSetter) {
		return errors.ErrNilDataRetrieverContainersSetter
	}
	if check.IfNil(mrc.shardMessengerFactory) {
		return errors.ErrNilBroadCastShardMessengerFactoryHandler
	}
	if check.IfNil(mrc.exportHandlerFactoryCreator) {
		return errors.ErrNilExportHandlerFactoryCreator
	}
	if check.IfNil(mrc.validatorAccountsSyncerFactoryHandler) {
		return errors.ErrNilValidatorAccountsDBSyncerFactory
	}
	if check.IfNil(mrc.shardRequestersContainerCreatorHandler) {
		return errors.ErrNilShardRequestersContainerCreatorHandler
	}
	if check.IfNil(mrc.apiRewardTxHandler) {
		return errors.ErrNilAPIRewardsHandler
	}
	if check.IfNil(mrc.outportDataProviderFactory) {
		return errors.ErrNilOutportDataProviderFactory
	}
	if check.IfNil(mrc.delegatedListFactoryHandler) {
		return factory.ErrNilDelegatedListFactory
	}
	if check.IfNil(mrc.directStakedListFactoryHandler) {
		return factory.ErrNilDirectStakedListFactory
	}
	if check.IfNil(mrc.totalStakedValueFactoryHandler) {
		return factory.ErrNilTotalStakedValueFactory
	}
	if check.IfNil(mrc.versionedHeaderFactory) {
		return process.ErrNilVersionedHeaderFactory
	}

	return nil
}

// AdditionalStorageServiceCreator returns the additional storage service creator
func (mrc *managedRunTypeComponents) AdditionalStorageServiceCreator() process.AdditionalStorageServiceCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.additionalStorageServiceCreator
}

// BlockProcessorCreator returns the block processor creator
func (mrc *managedRunTypeComponents) BlockProcessorCreator() processBlock.BlockProcessorCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.blockProcessorCreator
}

// BlockChainHookHandlerCreator returns the blockchain hook handler creator
func (mrc *managedRunTypeComponents) BlockChainHookHandlerCreator() hooks.BlockChainHookHandlerCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.blockChainHookHandlerCreator
}

// BootstrapperFromStorageCreator returns the bootstrapper from storage creator
func (mrc *managedRunTypeComponents) BootstrapperFromStorageCreator() storageBootstrap.BootstrapperFromStorageCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.bootstrapperFromStorageCreator
}

// BootstrapperCreator returns the bootstrapper creator
func (mrc *managedRunTypeComponents) BootstrapperCreator() storageBootstrap.BootstrapperCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.bootstrapperCreator
}

// BlockTrackerCreator returns the block tracker creator
func (mrc *managedRunTypeComponents) BlockTrackerCreator() track.BlockTrackerCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.blockTrackerCreator
}

// EpochStartBootstrapperCreator returns the epoch start bootstrapper creator
func (mrc *managedRunTypeComponents) EpochStartBootstrapperCreator() bootstrap.EpochStartBootstrapperCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.epochStartBootstrapperCreator
}

// ForkDetectorCreator returns the fork detector creator
func (mrc *managedRunTypeComponents) ForkDetectorCreator() processSync.ForkDetectorCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.forkDetectorCreator
}

// HeaderValidatorCreator returns the header validator creator
func (mrc *managedRunTypeComponents) HeaderValidatorCreator() processBlock.HeaderValidatorCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.headerValidatorCreator
}

// RequestHandlerCreator returns the request handler creator
func (mrc *managedRunTypeComponents) RequestHandlerCreator() requestHandlers.RequestHandlerCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.requestHandlerCreator
}

// ScheduledTxsExecutionCreator returns the scheduled transactions execution creator
func (mrc *managedRunTypeComponents) ScheduledTxsExecutionCreator() preprocess.ScheduledTxsExecutionCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.scheduledTxsExecutionCreator
}

// TransactionCoordinatorCreator returns the transaction coordinator creator
func (mrc *managedRunTypeComponents) TransactionCoordinatorCreator() coordinator.TransactionCoordinatorCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.transactionCoordinatorCreator
}

// ValidatorStatisticsProcessorCreator returns the validator statistics processor creator
func (mrc *managedRunTypeComponents) ValidatorStatisticsProcessorCreator() peer.ValidatorStatisticsProcessorCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.validatorStatisticsProcessorCreator
}

// SCProcessorCreator returns the smart contract processor creator
func (mrc *managedRunTypeComponents) SCProcessorCreator() scrCommon.SCProcessorCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.scProcessorCreator
}

// SCResultsPreProcessorCreator returns the smart contract result pre-processor creator
func (mrc *managedRunTypeComponents) SCResultsPreProcessorCreator() preprocess.SmartContractResultPreProcessorCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.scResultPreProcessorCreator
}

// ConsensusModel returns the consensus model
func (mrc *managedRunTypeComponents) ConsensusModel() consensus.ConsensusModel {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return consensus.ConsensusModelInvalid
	}

	return mrc.runTypeComponents.consensusModel
}

// VmContainerMetaFactoryCreator returns the vm container meta factory creator
func (mrc *managedRunTypeComponents) VmContainerMetaFactoryCreator() factoryVm.VmContainerCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.vmContainerMetaFactory
}

// VmContainerShardFactoryCreator returns the vm container shard factory creator
func (mrc *managedRunTypeComponents) VmContainerShardFactoryCreator() factoryVm.VmContainerCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.vmContainerShardFactory
}

// AccountsParser returns the accounts parser
func (mrc *managedRunTypeComponents) AccountsParser() genesis.AccountsParser {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.accountsParser
}

// AccountsCreator returns the accounts factory
func (mrc *managedRunTypeComponents) AccountsCreator() state.AccountFactory {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.accountsCreator
}

// VMContextCreator returns the vm context creator
func (mrc *managedRunTypeComponents) VMContextCreator() systemSmartContracts.VMContextCreatorHandler {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.vmContextCreator
}

// OutGoingOperationsPoolHandler return the outgoing operations pool factory
func (mrc *managedRunTypeComponents) OutGoingOperationsPoolHandler() sovereignBlock.OutGoingOperationsPool {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.outGoingOperationsPoolHandler
}

// DataCodecHandler returns the data codec factory
func (mrc *managedRunTypeComponents) DataCodecHandler() sovereign.DataCodecHandler {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.dataCodecHandler
}

// TopicsCheckerHandler returns the topics checker factory
func (mrc *managedRunTypeComponents) TopicsCheckerHandler() sovereign.TopicsCheckerHandler {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.topicsCheckerHandler
}

// ShardCoordinatorCreator returns the shard coordinator factory
func (mrc *managedRunTypeComponents) ShardCoordinatorCreator() sharding.ShardCoordinatorFactory {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.shardCoordinatorCreator
}

// NodesCoordinatorWithRaterCreator returns the nodes coordinator factory
func (mrc *managedRunTypeComponents) NodesCoordinatorWithRaterCreator() nodesCoordinator.NodesCoordinatorWithRaterFactory {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.nodesCoordinatorWithRaterFactoryCreator
}

// RequestersContainerFactoryCreator returns the shard coordinator factory
func (mrc *managedRunTypeComponents) RequestersContainerFactoryCreator() requesterscontainer.RequesterContainerFactoryCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.requestersContainerFactoryCreator
}

// InterceptorsContainerFactoryCreator returns the shard interceptors container factory
func (mrc *managedRunTypeComponents) InterceptorsContainerFactoryCreator() interceptorscontainer.InterceptorsContainerFactoryCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.interceptorsContainerFactoryCreator
}

// ShardResolversContainerFactoryCreator returns the shard resolvers container factory
func (mrc *managedRunTypeComponents) ShardResolversContainerFactoryCreator() resolverscontainer.ShardResolversContainerFactoryCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.shardResolversContainerFactoryCreator
}

// TxPreProcessorCreator returns the tx pre processor factory
func (mrc *managedRunTypeComponents) TxPreProcessorCreator() preprocess.TxPreProcessorCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.txPreProcessorCreator
}

// ExtraHeaderSigVerifierHolder returns the extra header sig verifier holder
func (mrc *managedRunTypeComponents) ExtraHeaderSigVerifierHolder() headerCheck.ExtraHeaderSigVerifierHolder {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.extraHeaderSigVerifierHolder
}

// GenesisBlockCreatorFactory returns the genesis block factory
func (mrc *managedRunTypeComponents) GenesisBlockCreatorFactory() processComp.GenesisBlockCreatorFactory {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.genesisBlockCreatorFactory
}

// GenesisMetaBlockCheckerCreator returns the genesis meta block checker creator
func (mrc *managedRunTypeComponents) GenesisMetaBlockCheckerCreator() processComp.GenesisMetaBlockChecker {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.genesisMetaBlockCheckerCreator
}

// NodesSetupCheckerFactory returns the nodes setup checker factory
func (mrc *managedRunTypeComponents) NodesSetupCheckerFactory() checking.NodesSetupCheckerFactory {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.nodesSetupCheckerFactory
}

// EpochStartTriggerFactory returns the epoch start trigger factory
func (mrc *managedRunTypeComponents) EpochStartTriggerFactory() factory.EpochStartTriggerFactoryHandler {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.epochStartTriggerFactory
}

// LatestDataProviderFactory returns the latest data provider factory
func (mrc *managedRunTypeComponents) LatestDataProviderFactory() latestData.LatestDataProviderFactory {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.latestDataProviderFactory
}

// StakingToPeerFactory returns the staking to peer factory
func (mrc *managedRunTypeComponents) StakingToPeerFactory() scToProtocol.StakingToPeerFactoryHandler {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.scToProtocolFactory
}

// ValidatorInfoCreatorFactory returns the validator info creator factory
func (mrc *managedRunTypeComponents) ValidatorInfoCreatorFactory() factory.ValidatorInfoCreatorFactory {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.validatorInfoCreatorFactory
}

// ApiProcessorCompsCreatorHandler returns the api processor components creator handler
func (mrc *managedRunTypeComponents) ApiProcessorCompsCreatorHandler() api.ApiProcessorCompsCreatorHandler {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.apiProcessorCompsCreatorHandler
}

// EndOfEpochEconomicsFactoryHandler returns the end of epoch economics factory handler
func (mrc *managedRunTypeComponents) EndOfEpochEconomicsFactoryHandler() factory.EndOfEpochEconomicsFactoryHandler {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.endOfEpochEconomicsFactoryHandler
}

// RewardsCreatorFactory returns the rewards creator factory
func (mrc *managedRunTypeComponents) RewardsCreatorFactory() factory.RewardsCreatorFactory {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.rewardsCreatorFactory
}

// SystemSCProcessorFactory returns the sys sc processor factory
func (mrc *managedRunTypeComponents) SystemSCProcessorFactory() factory.SystemSCProcessorFactory {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.systemSCProcessorFactory
}

// PreProcessorsContainerFactoryCreator returns the pre-processors container factory creator
func (mrc *managedRunTypeComponents) PreProcessorsContainerFactoryCreator() data.PreProcessorsContainerFactoryCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.preProcessorsContainerFactoryCreator
}

// DataRetrieverContainersSetter returns the data retriever containers setter
func (mrc *managedRunTypeComponents) DataRetrieverContainersSetter() factory.DataRetrieverContainersSetter {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.dataRetrieverContainersSetter
}

// BroadCastShardMessengerFactoryHandler returns the shard broadcast messenger factory
func (mrc *managedRunTypeComponents) BroadCastShardMessengerFactoryHandler() sposFactory.BroadCastShardMessengerFactoryHandler {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.shardMessengerFactory
}

// ExportHandlerFactoryCreator returns the export handler factory creator
func (mrc *managedRunTypeComponents) ExportHandlerFactoryCreator() factory.ExportHandlerFactoryCreator {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.exportHandlerFactoryCreator
}

// ValidatorAccountsSyncerFactoryHandler returns validator accounts syncer factory handler
func (mrc *managedRunTypeComponents) ValidatorAccountsSyncerFactoryHandler() syncerFactory.ValidatorAccountsSyncerFactoryHandler {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.validatorAccountsSyncerFactoryHandler
}

// ShardRequestersContainerCreatorHandler returns shard requesters container creator handler
func (mrc *managedRunTypeComponents) ShardRequestersContainerCreatorHandler() storageRequestFactory.ShardRequestersContainerCreatorHandler {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.shardRequestersContainerCreatorHandler
}

// APIRewardsTxHandler returns api rewards tx handler
func (mrc *managedRunTypeComponents) APIRewardsTxHandler() transactionAPI.APIRewardTxHandler {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.apiRewardTxHandler
}

// OutportDataProviderFactory returns the outport data provider factory
func (mrc *managedRunTypeComponents) OutportDataProviderFactory() factory.OutportDataProviderFactoryHandler {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.outportDataProviderFactory
}

// DelegatedListFactoryHandler returns delegated list factory handler
func (mrc *managedRunTypeComponents) DelegatedListFactoryHandler() trieIteratorsFactory.DelegatedListProcessorFactoryHandler {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.delegatedListFactoryHandler
}

// DirectStakedListFactoryHandler returns direct staked list factory handler
func (mrc *managedRunTypeComponents) DirectStakedListFactoryHandler() trieIteratorsFactory.DirectStakedListProcessorFactoryHandler {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.directStakedListFactoryHandler
}

// TotalStakedValueFactoryHandler returns total staked value factory handler
func (mrc *managedRunTypeComponents) TotalStakedValueFactoryHandler() trieIteratorsFactory.TotalStakedValueProcessorFactoryHandler {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.totalStakedValueFactoryHandler
}

// VersionedHeaderFactory returns the versioned header factory
func (mrc *managedRunTypeComponents) VersionedHeaderFactory() genesis.VersionedHeaderFactory {
	mrc.mutRunTypeComponents.RLock()
	defer mrc.mutRunTypeComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.versionedHeaderFactory
}

// IsInterfaceNil returns true if the interface is nil
func (mrc *managedRunTypeComponents) IsInterfaceNil() bool {
	return mrc == nil
}

// String returns the name of the component
func (mrc *managedRunTypeComponents) String() string {
	return factory.RunTypeComponentsName
}
