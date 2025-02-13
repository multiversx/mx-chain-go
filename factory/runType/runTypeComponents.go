package runType

import (
	"fmt"
	"math/big"

	nodeFactory "github.com/multiversx/mx-chain-go/cmd/node/factory"
	"github.com/multiversx/mx-chain-go/common/disabled"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/broadcastFactory"
	"github.com/multiversx/mx-chain-go/consensus/spos/sposFactory"
	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	requesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/resolverscontainer"
	storageRequestFactory "github.com/multiversx/mx-chain-go/dataRetriever/factory/storageRequestersContainer/factory"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap"
	"github.com/multiversx/mx-chain-go/epochStart/metachain"
	"github.com/multiversx/mx-chain-go/errors"
	mainFactory "github.com/multiversx/mx-chain-go/factory"
	factoryBlock "github.com/multiversx/mx-chain-go/factory/block"
	"github.com/multiversx/mx-chain-go/factory/epochStartTrigger"
	"github.com/multiversx/mx-chain-go/factory/processing/api"
	"github.com/multiversx/mx-chain-go/factory/processing/dataRetriever"
	factoryVm "github.com/multiversx/mx-chain-go/factory/vm"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/checking"
	"github.com/multiversx/mx-chain-go/genesis/parsing"
	processGenesis "github.com/multiversx/mx-chain-go/genesis/process"
	"github.com/multiversx/mx-chain-go/node/external/transactionAPI"
	trieIteratorsFactory "github.com/multiversx/mx-chain-go/node/trieIterators/factory"
	outportFactory "github.com/multiversx/mx-chain-go/outport/process/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
	processBlock "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/factory/interceptorscontainer"
	"github.com/multiversx/mx-chain-go/process/factory/shard"
	"github.com/multiversx/mx-chain-go/process/factory/shard/data"
	"github.com/multiversx/mx-chain-go/process/headerCheck"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/process/scToProtocol"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/processProxy"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/process/sync/storageBootstrap"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/sharding"
	nodesCoord "github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/factory"
	syncerFactory "github.com/multiversx/mx-chain-go/state/syncer/factory"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/latestData"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	updateFactory "github.com/multiversx/mx-chain-go/update/factory/creator"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

// ArgsRunTypeComponents struct holds the arguments for run type component
type ArgsRunTypeComponents struct {
	CoreComponents   process.CoreComponentsHolder
	CryptoComponents process.CryptoComponentsHolder
	Configs          config.Configs
	InitialAccounts  []genesis.InitialAccountHandler
}

type runTypeComponentsFactory struct {
	coreComponents   process.CoreComponentsHolder
	cryptoComponents process.CryptoComponentsHolder
	configs          config.Configs
	initialAccounts  []genesis.InitialAccountHandler
}

// runTypeComponents struct holds the components needed for a run type
type runTypeComponents struct {
	blockChainHookHandlerCreator            hooks.BlockChainHookHandlerCreator
	epochStartBootstrapperCreator           bootstrap.EpochStartBootstrapperCreator
	bootstrapperFromStorageCreator          storageBootstrap.BootstrapperFromStorageCreator
	bootstrapperCreator                     storageBootstrap.BootstrapperCreator
	blockProcessorCreator                   processBlock.BlockProcessorCreator
	forkDetectorCreator                     sync.ForkDetectorCreator
	blockTrackerCreator                     track.BlockTrackerCreator
	requestHandlerCreator                   requestHandlers.RequestHandlerCreator
	headerValidatorCreator                  processBlock.HeaderValidatorCreator
	scheduledTxsExecutionCreator            preprocess.ScheduledTxsExecutionCreator
	transactionCoordinatorCreator           coordinator.TransactionCoordinatorCreator
	validatorStatisticsProcessorCreator     peer.ValidatorStatisticsProcessorCreator
	additionalStorageServiceCreator         process.AdditionalStorageServiceCreator
	scProcessorCreator                      scrCommon.SCProcessorCreator
	scResultPreProcessorCreator             preprocess.SmartContractResultPreProcessorCreator
	consensusModel                          consensus.ConsensusModel
	vmContainerMetaFactory                  factoryVm.VmContainerCreator
	vmContainerShardFactory                 factoryVm.VmContainerCreator
	accountsParser                          genesis.AccountsParser
	accountsCreator                         state.AccountFactory
	vmContextCreator                        systemSmartContracts.VMContextCreatorHandler
	outGoingOperationsPoolHandler           sovereignBlock.OutGoingOperationsPool
	dataCodecHandler                        sovereign.DataCodecHandler
	topicsCheckerHandler                    sovereign.TopicsCheckerHandler
	shardCoordinatorCreator                 sharding.ShardCoordinatorFactory
	nodesCoordinatorWithRaterFactoryCreator nodesCoord.NodesCoordinatorWithRaterFactory
	requestersContainerFactoryCreator       requesterscontainer.RequesterContainerFactoryCreator
	interceptorsContainerFactoryCreator     interceptorscontainer.InterceptorsContainerFactoryCreator
	shardResolversContainerFactoryCreator   resolverscontainer.ShardResolversContainerFactoryCreator
	txPreProcessorCreator                   preprocess.TxPreProcessorCreator
	extraHeaderSigVerifierHolder            headerCheck.ExtraHeaderSigVerifierHolder
	genesisBlockCreatorFactory              processGenesis.GenesisBlockCreatorFactory
	genesisMetaBlockCheckerCreator          processGenesis.GenesisMetaBlockChecker
	nodesSetupCheckerFactory                checking.NodesSetupCheckerFactory
	epochStartTriggerFactory                mainFactory.EpochStartTriggerFactoryHandler
	latestDataProviderFactory               latestData.LatestDataProviderFactory
	scToProtocolFactory                     scToProtocol.StakingToPeerFactoryHandler
	validatorInfoCreatorFactory             mainFactory.ValidatorInfoCreatorFactory
	apiProcessorCompsCreatorHandler         api.ApiProcessorCompsCreatorHandler
	endOfEpochEconomicsFactoryHandler       mainFactory.EndOfEpochEconomicsFactoryHandler
	rewardsCreatorFactory                   mainFactory.RewardsCreatorFactory
	systemSCProcessorFactory                mainFactory.SystemSCProcessorFactory
	preProcessorsContainerFactoryCreator    data.PreProcessorsContainerFactoryCreator
	dataRetrieverContainersSetter           mainFactory.DataRetrieverContainersSetter
	shardMessengerFactory                   sposFactory.BroadCastShardMessengerFactoryHandler
	exportHandlerFactoryCreator             mainFactory.ExportHandlerFactoryCreator
	validatorAccountsSyncerFactoryHandler   syncerFactory.ValidatorAccountsSyncerFactoryHandler
	shardRequestersContainerCreatorHandler  storageRequestFactory.ShardRequestersContainerCreatorHandler
	apiRewardTxHandler                      transactionAPI.APIRewardTxHandler
	outportDataProviderFactory              mainFactory.OutportDataProviderFactoryHandler
	delegatedListFactoryHandler             trieIteratorsFactory.DelegatedListProcessorFactoryHandler
	directStakedListFactoryHandler          trieIteratorsFactory.DirectStakedListProcessorFactoryHandler
	totalStakedValueFactoryHandler          trieIteratorsFactory.TotalStakedValueProcessorFactoryHandler
	versionedHeaderFactory                  genesis.VersionedHeaderFactory
}

// NewRunTypeComponentsFactory will return a new instance of runTypeComponentsFactory
func NewRunTypeComponentsFactory(args ArgsRunTypeComponents) (*runTypeComponentsFactory, error) {
	if check.IfNil(args.CoreComponents) {
		return nil, errors.ErrNilCoreComponents
	}
	if check.IfNil(args.CryptoComponents) {
		return nil, errors.ErrNilCryptoComponents
	}
	if args.InitialAccounts == nil {
		return nil, errors.ErrNilInitialAccounts
	}

	return &runTypeComponentsFactory{
		coreComponents:   args.CoreComponents,
		cryptoComponents: args.CryptoComponents,
		configs:          args.Configs,
		initialAccounts:  args.InitialAccounts,
	}, nil
}

// Create creates the runType components
func (rcf *runTypeComponentsFactory) Create() (*runTypeComponents, error) {
	vmContextCreator := systemSmartContracts.NewVMContextCreator()
	vmContainerMetaCreator, err := factoryVm.NewVmContainerMetaFactory(vmContextCreator)
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewVmContainerMetaFactory failed: %w", err)
	}

	totalSupply, ok := big.NewInt(0).SetString(rcf.configs.EconomicsConfig.GlobalSettings.GenesisTotalSupply, 10)
	if !ok {
		return nil, fmt.Errorf("can not parse total suply from economics.toml, %s is not a valid value",
			rcf.configs.EconomicsConfig.GlobalSettings.GenesisTotalSupply)
	}

	accountsParserArgs := genesis.AccountsParserArgs{
		InitialAccounts: rcf.initialAccounts,
		EntireSupply:    totalSupply,
		MinterAddress:   rcf.configs.EconomicsConfig.GlobalSettings.GenesisMintingSenderAddress,
		PubkeyConverter: rcf.coreComponents.AddressPubKeyConverter(),
		KeyGenerator:    rcf.cryptoComponents.TxSignKeyGen(),
		Hasher:          rcf.coreComponents.Hasher(),
		Marshalizer:     rcf.coreComponents.InternalMarshalizer(),
	}
	accountsParser, err := parsing.NewAccountsParser(accountsParserArgs)
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewAccountsParser failed: %w", err)
	}

	accountsCreator, err := factory.NewAccountCreator(factory.ArgsAccountCreator{
		Hasher:              rcf.coreComponents.Hasher(),
		Marshaller:          rcf.coreComponents.InternalMarshalizer(),
		EnableEpochsHandler: rcf.coreComponents.EnableEpochsHandler(),
	})
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewAccountCreator failed: %w", err)
	}
	apiRewardTxHandler, err := transactionAPI.NewAPIRewardsHandler(rcf.coreComponents.AddressPubKeyConverter())
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewAPIRewardsHandler failed: %w", err)
	}

	headerVersionHandler, err := rcf.createHeaderVersionHandler()
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - createHeaderVersionHandler failed: %w", err)
	}

	versionedHeaderFactory, err := factoryBlock.NewShardHeaderFactory(headerVersionHandler)
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewShardHeaderFactory failed: %w", err)
	}

	return &runTypeComponents{
		blockChainHookHandlerCreator:            hooks.NewBlockChainHookFactory(),
		epochStartBootstrapperCreator:           bootstrap.NewEpochStartBootstrapperFactory(),
		bootstrapperFromStorageCreator:          storageBootstrap.NewShardStorageBootstrapperFactory(),
		bootstrapperCreator:                     storageBootstrap.NewShardBootstrapFactory(),
		blockProcessorCreator:                   block.NewShardBlockProcessorFactory(),
		forkDetectorCreator:                     sync.NewShardForkDetectorFactory(),
		blockTrackerCreator:                     track.NewShardBlockTrackerFactory(),
		requestHandlerCreator:                   requestHandlers.NewResolverRequestHandlerFactory(),
		headerValidatorCreator:                  block.NewShardHeaderValidatorFactory(),
		scheduledTxsExecutionCreator:            preprocess.NewShardScheduledTxsExecutionFactory(),
		transactionCoordinatorCreator:           coordinator.NewShardTransactionCoordinatorFactory(),
		validatorStatisticsProcessorCreator:     peer.NewValidatorStatisticsProcessorFactory(),
		additionalStorageServiceCreator:         storageFactory.NewShardAdditionalStorageServiceFactory(),
		scProcessorCreator:                      processProxy.NewSCProcessProxyFactory(),
		scResultPreProcessorCreator:             preprocess.NewSmartContractResultPreProcessorFactory(),
		consensusModel:                          consensus.ConsensusModelV1,
		vmContainerMetaFactory:                  vmContainerMetaCreator,
		vmContainerShardFactory:                 factoryVm.NewVmContainerShardFactory(),
		accountsParser:                          accountsParser,
		accountsCreator:                         accountsCreator,
		vmContextCreator:                        vmContextCreator,
		outGoingOperationsPoolHandler:           disabled.NewDisabledOutGoingOperationPool(),
		dataCodecHandler:                        disabled.NewDisabledDataCodec(),
		topicsCheckerHandler:                    disabled.NewDisabledTopicsChecker(),
		shardCoordinatorCreator:                 sharding.NewMultiShardCoordinatorFactory(),
		nodesCoordinatorWithRaterFactoryCreator: nodesCoord.NewIndexHashedNodesCoordinatorWithRaterFactory(),
		requestersContainerFactoryCreator:       requesterscontainer.NewShardRequestersContainerFactoryCreator(),
		interceptorsContainerFactoryCreator:     interceptorscontainer.NewShardInterceptorsContainerFactoryCreator(),
		shardResolversContainerFactoryCreator:   resolverscontainer.NewShardResolversContainerFactoryCreator(),
		txPreProcessorCreator:                   preprocess.NewTxPreProcessorCreator(),
		extraHeaderSigVerifierHolder:            headerCheck.NewExtraHeaderSigVerifierHolder(),
		genesisBlockCreatorFactory:              processGenesis.NewGenesisBlockCreatorFactory(),
		genesisMetaBlockCheckerCreator:          processGenesis.NewGenesisMetaBlockChecker(),
		nodesSetupCheckerFactory:                checking.NewNodesSetupCheckerFactory(),
		epochStartTriggerFactory:                epochStartTrigger.NewEpochStartTriggerFactory(),
		latestDataProviderFactory:               latestData.NewLatestDataProviderFactory(),
		scToProtocolFactory:                     scToProtocol.NewStakingToPeerFactory(),
		validatorInfoCreatorFactory:             metachain.NewValidatorInfoCreatorFactory(),
		apiProcessorCompsCreatorHandler:         api.NewAPIProcessorCompsCreator(),
		endOfEpochEconomicsFactoryHandler:       metachain.NewEconomicsFactory(),
		rewardsCreatorFactory:                   metachain.NewRewardsCreatorFactory(),
		systemSCProcessorFactory:                metachain.NewSysSCFactory(),
		preProcessorsContainerFactoryCreator:    shard.NewPreProcessorContainerFactoryCreator(),
		dataRetrieverContainersSetter:           dataRetriever.NewDataRetrieverContainerSetter(),
		shardMessengerFactory:                   broadcastFactory.NewShardChainMessengerFactory(),
		exportHandlerFactoryCreator:             updateFactory.NewExportHandlerFactoryCreator(),
		validatorAccountsSyncerFactoryHandler:   syncerFactory.NewValidatorAccountsSyncerFactory(),
		shardRequestersContainerCreatorHandler:  storageRequestFactory.NewShardRequestersContainerCreator(),
		apiRewardTxHandler:                      apiRewardTxHandler,
		outportDataProviderFactory:              outportFactory.NewOutportDataProviderFactory(),
		delegatedListFactoryHandler:             trieIteratorsFactory.NewDelegatedListProcessorFactory(),
		directStakedListFactoryHandler:          trieIteratorsFactory.NewDirectStakedListProcessorFactory(),
		totalStakedValueFactoryHandler:          trieIteratorsFactory.NewTotalStakedListProcessorFactory(),
		versionedHeaderFactory:                  versionedHeaderFactory,
	}, nil
}

func (rcf *runTypeComponentsFactory) createHeaderVersionHandler() (nodeFactory.HeaderVersionHandler, error) {
	cacheConfig := storageFactory.GetCacherFromConfig(rcf.configs.GeneralConfig.Versions.Cache)
	cache, err := storageunit.NewCache(cacheConfig)
	if err != nil {
		return nil, err
	}

	return factoryBlock.NewHeaderVersionHandler(
		rcf.configs.GeneralConfig.Versions.VersionsByEpochs,
		rcf.configs.GeneralConfig.Versions.DefaultVersion,
		cache,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rc *runTypeComponentsFactory) IsInterfaceNil() bool {
	return rc == nil
}

// Close does nothing
func (rc *runTypeComponents) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rc *runTypeComponents) IsInterfaceNil() bool {
	return rc == nil
}
