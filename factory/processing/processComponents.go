package processing

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"path/filepath"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/partitioning"
	"github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	nodeFactory "github.com/multiversx/mx-chain-go/cmd/node/factory"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/containers"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/epochProviders"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/resolverscontainer"
	storageResolversContainers "github.com/multiversx/mx-chain-go/dataRetriever/factory/storageResolversContainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/dblookupext"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/metachain"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/epochStart/shardchain"
	errErd "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	mainFactory "github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/disabled"
	"github.com/multiversx/mx-chain-go/fallback"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/checking"
	processGenesis "github.com/multiversx/mx-chain-go/genesis/process"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/block/pendingMb"
	"github.com/multiversx/mx-chain-go/process/block/poolsCleaner"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/multiversx/mx-chain-go/process/factory/interceptorscontainer"
	"github.com/multiversx/mx-chain-go/process/headerCheck"
	"github.com/multiversx/mx-chain-go/process/heartbeat/validator"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/process/receipts"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/process/transactionLog"
	"github.com/multiversx/mx-chain-go/process/txsSender"
	"github.com/multiversx/mx-chain-go/process/txsimulator"
	"github.com/multiversx/mx-chain-go/redundancy"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/networksharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/cache"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/update"
	updateDisabled "github.com/multiversx/mx-chain-go/update/disabled"
	updateFactory "github.com/multiversx/mx-chain-go/update/factory"
	"github.com/multiversx/mx-chain-go/update/trigger"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	vmcommonBuiltInFunctions "github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
)

// timeSpanForBadHeaders is the expiry time for an added block header hash
var timeSpanForBadHeaders = time.Minute * 2

// processComponents struct holds the process components
type processComponents struct {
	nodesCoordinator             nodesCoordinator.NodesCoordinator
	shardCoordinator             sharding.Coordinator
	interceptorsContainer        process.InterceptorsContainer
	resolversFinder              dataRetriever.ResolversFinder
	roundHandler                 consensus.RoundHandler
	epochStartTrigger            epochStart.TriggerHandler
	epochStartNotifier           factory.EpochStartNotifier
	forkDetector                 process.ForkDetector
	blockProcessor               process.BlockProcessor
	blackListHandler             process.TimeCacher
	bootStorer                   process.BootStorer
	headerSigVerifier            process.InterceptedHeaderSigVerifier
	headerIntegrityVerifier      nodeFactory.HeaderIntegrityVerifierHandler
	validatorsStatistics         process.ValidatorStatisticsProcessor
	validatorsProvider           process.ValidatorsProvider
	blockTracker                 process.BlockTracker
	pendingMiniBlocksHandler     process.PendingMiniBlocksHandler
	requestHandler               process.RequestHandler
	txLogsProcessor              process.TransactionLogProcessorDatabase
	headerConstructionValidator  process.HeaderConstructionValidator
	peerShardMapper              process.NetworkShardingCollector
	txSimulatorProcessor         factory.TransactionSimulatorProcessor
	miniBlocksPoolCleaner        process.PoolsCleaner
	txsPoolCleaner               process.PoolsCleaner
	fallbackHeaderValidator      process.FallbackHeaderValidator
	whiteListHandler             process.WhiteListHandler
	whiteListerVerifiedTxs       process.WhiteListHandler
	historyRepository            dblookupext.HistoryRepository
	importStartHandler           update.ImportStartHandler
	requestedItemsHandler        dataRetriever.RequestedItemsHandler
	importHandler                update.ImportHandler
	nodeRedundancyHandler        consensus.NodeRedundancyHandler
	currentEpochProvider         dataRetriever.CurrentNetworkEpochProviderHandler
	vmFactoryForTxSimulator      process.VirtualMachinesContainerFactory
	vmFactoryForProcessing       process.VirtualMachinesContainerFactory
	scheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	txsSender                    process.TxsSenderHandler
	hardforkTrigger              factory.HardforkTrigger
	processedMiniBlocksTracker   process.ProcessedMiniBlocksTracker
	esdtDataStorageForApi        vmcommon.ESDTNFTStorageHandler
	accountsParser               genesis.AccountsParser
	receiptsRepository           mainFactory.ReceiptsRepository
}

// ProcessComponentsFactoryArgs holds the arguments needed to create a process components factory
type ProcessComponentsFactoryArgs struct {
	Config                 config.Config
	EpochConfig            config.EpochConfig
	PrefConfigs            config.PreferencesConfig
	ImportDBConfig         config.ImportDbConfig
	AccountsParser         genesis.AccountsParser
	SmartContractParser    genesis.InitialSmartContractParser
	GasSchedule            core.GasScheduleNotifier
	NodesCoordinator       nodesCoordinator.NodesCoordinator
	RequestedItemsHandler  dataRetriever.RequestedItemsHandler
	WhiteListHandler       process.WhiteListHandler
	WhiteListerVerifiedTxs process.WhiteListHandler
	MaxRating              uint32
	SystemSCConfig         *config.SystemSmartContractsConfig
	Version                string
	ImportStartHandler     update.ImportStartHandler
	WorkingDir             string
	HistoryRepo            dblookupext.HistoryRepository

	Data                 factory.DataComponentsHolder
	CoreData             factory.CoreComponentsHolder
	Crypto               factory.CryptoComponentsHolder
	State                factory.StateComponentsHolder
	Network              factory.NetworkComponentsHolder
	BootstrapComponents  factory.BootstrapComponentsHolder
	StatusComponents     factory.StatusComponentsHolder
	StatusCoreComponents factory.StatusCoreComponentsHolder
}

type processComponentsFactory struct {
	config                 config.Config
	epochConfig            config.EpochConfig
	prefConfigs            config.PreferencesConfig
	importDBConfig         config.ImportDbConfig
	accountsParser         genesis.AccountsParser
	smartContractParser    genesis.InitialSmartContractParser
	gasSchedule            core.GasScheduleNotifier
	nodesCoordinator       nodesCoordinator.NodesCoordinator
	requestedItemsHandler  dataRetriever.RequestedItemsHandler
	whiteListHandler       process.WhiteListHandler
	whiteListerVerifiedTxs process.WhiteListHandler
	maxRating              uint32
	systemSCConfig         *config.SystemSmartContractsConfig
	txLogsProcessor        process.TransactionLogProcessor
	version                string
	importStartHandler     update.ImportStartHandler
	workingDir             string
	historyRepo            dblookupext.HistoryRepository
	epochNotifier          process.EpochNotifier
	importHandler          update.ImportHandler
	esdtNftStorage         vmcommon.ESDTNFTStorageHandler

	data                 factory.DataComponentsHolder
	coreData             factory.CoreComponentsHolder
	crypto               factory.CryptoComponentsHolder
	state                factory.StateComponentsHolder
	network              factory.NetworkComponentsHolder
	bootstrapComponents  factory.BootstrapComponentsHolder
	statusComponents     factory.StatusComponentsHolder
	statusCoreComponents factory.StatusCoreComponentsHolder
}

// NewProcessComponentsFactory will return a new instance of processComponentsFactory
func NewProcessComponentsFactory(args ProcessComponentsFactoryArgs) (*processComponentsFactory, error) {
	err := checkProcessComponentsArgs(args)
	if err != nil {
		return nil, err
	}

	return &processComponentsFactory{
		config:                 args.Config,
		epochConfig:            args.EpochConfig,
		prefConfigs:            args.PrefConfigs,
		importDBConfig:         args.ImportDBConfig,
		accountsParser:         args.AccountsParser,
		smartContractParser:    args.SmartContractParser,
		gasSchedule:            args.GasSchedule,
		nodesCoordinator:       args.NodesCoordinator,
		data:                   args.Data,
		coreData:               args.CoreData,
		crypto:                 args.Crypto,
		state:                  args.State,
		network:                args.Network,
		bootstrapComponents:    args.BootstrapComponents,
		statusComponents:       args.StatusComponents,
		requestedItemsHandler:  args.RequestedItemsHandler,
		whiteListHandler:       args.WhiteListHandler,
		whiteListerVerifiedTxs: args.WhiteListerVerifiedTxs,
		maxRating:              args.MaxRating,
		systemSCConfig:         args.SystemSCConfig,
		version:                args.Version,
		importStartHandler:     args.ImportStartHandler,
		workingDir:             args.WorkingDir,
		historyRepo:            args.HistoryRepo,
		epochNotifier:          args.CoreData.EpochNotifier(),
		statusCoreComponents:   args.StatusCoreComponents,
	}, nil
}

// TODO: Think if it would make sense here to create an array of closable interfaces

// Create will create and return a struct containing process components
func (pcf *processComponentsFactory) Create() (*processComponents, error) {
	currentEpochProvider, err := epochProviders.CreateCurrentEpochProvider(
		pcf.config,
		pcf.coreData.GenesisNodesSetup().GetRoundDuration(),
		pcf.coreData.GenesisTime().Unix(),
		pcf.prefConfigs.FullArchive,
	)
	if err != nil {
		return nil, err
	}

	pcf.epochNotifier.RegisterNotifyHandler(currentEpochProvider)

	fallbackHeaderValidator, err := fallback.NewFallbackHeaderValidator(
		pcf.data.Datapool().Headers(),
		pcf.coreData.InternalMarshalizer(),
		pcf.data.StorageService(),
	)
	if err != nil {
		return nil, err
	}

	argsHeaderSig := &headerCheck.ArgsHeaderSigVerifier{
		Marshalizer:             pcf.coreData.InternalMarshalizer(),
		Hasher:                  pcf.coreData.Hasher(),
		NodesCoordinator:        pcf.nodesCoordinator,
		MultiSigContainer:       pcf.crypto.MultiSignerContainer(),
		SingleSigVerifier:       pcf.crypto.BlockSigner(),
		KeyGen:                  pcf.crypto.BlockSignKeyGen(),
		FallbackHeaderValidator: fallbackHeaderValidator,
	}
	headerSigVerifier, err := headerCheck.NewHeaderSigVerifier(argsHeaderSig)
	if err != nil {
		return nil, err
	}

	// TODO: maybe move PeerShardMapper to network components
	peerShardMapper, err := pcf.prepareNetworkShardingCollector()
	if err != nil {
		return nil, err
	}

	resolversContainerFactory, err := pcf.newResolverContainerFactory(currentEpochProvider)
	if err != nil {
		return nil, err
	}

	resolversContainer, err := resolversContainerFactory.Create()
	if err != nil {
		return nil, err
	}

	resolversFinder, err := containers.NewResolversFinder(resolversContainer, pcf.bootstrapComponents.ShardCoordinator())
	if err != nil {
		return nil, err
	}

	requestHandler, err := requestHandlers.NewResolverRequestHandler(
		resolversFinder,
		pcf.requestedItemsHandler,
		pcf.whiteListHandler,
		common.MaxTxsToRequest,
		pcf.bootstrapComponents.ShardCoordinator().SelfId(),
		time.Second,
	)
	if err != nil {
		return nil, err
	}

	txLogsStorage, err := pcf.data.StorageService().GetStorer(dataRetriever.TxLogsUnit)
	if err != nil {
		return nil, err
	}

	if !pcf.config.LogsAndEvents.SaveInStorageEnabled && pcf.config.DbLookupExtensions.Enabled {
		log.Warn("processComponentsFactory.Create() node will save logs in storage because DbLookupExtensions is enabled")
	}

	saveLogsInStorage := pcf.config.LogsAndEvents.SaveInStorageEnabled || pcf.config.DbLookupExtensions.Enabled
	txLogsProcessor, err := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer:               txLogsStorage,
		Marshalizer:          pcf.coreData.InternalMarshalizer(),
		SaveInStorageEnabled: saveLogsInStorage,
	})
	if err != nil {
		return nil, err
	}

	pcf.txLogsProcessor = txLogsProcessor
	genesisBlocks, initialTxs, err := pcf.generateGenesisHeadersAndApplyInitialBalances()
	if err != nil {
		return nil, err
	}

	genesisAccounts, err := pcf.indexAndReturnGenesisAccounts()
	if err != nil {
		log.Warn("cannot index genesis accounts", "error", err)
	}

	genesisBlock, ok := genesisBlocks[core.MetachainShardId]
	if !ok {
		return nil, errors.New("genesis meta block does not exist")
	}

	genesisMetaBlock, ok := genesisBlock.(data.MetaHeaderHandler)
	if !ok {
		return nil, errors.New("genesis meta block invalid")
	}

	err = pcf.setGenesisHeader(genesisBlocks)
	if err != nil {
		return nil, err
	}

	validatorStatisticsProcessor, err := pcf.newValidatorStatisticsProcessor()
	if err != nil {
		return nil, err
	}

	validatorStatsRootHash, err := validatorStatisticsProcessor.RootHash()
	if err != nil {
		return nil, err
	}

	err = genesisMetaBlock.SetValidatorStatsRootHash(validatorStatsRootHash)
	if err != nil {
		return nil, err
	}

	startEpochNum := pcf.bootstrapComponents.EpochBootstrapParams().Epoch()
	if startEpochNum == 0 {
		err = pcf.indexGenesisBlocks(genesisBlocks, initialTxs, genesisAccounts)
		if err != nil {
			return nil, err
		}
	}

	cacheRefreshDuration := time.Duration(pcf.config.ValidatorStatistics.CacheRefreshIntervalInSec) * time.Second
	argVSP := peer.ArgValidatorsProvider{
		NodesCoordinator:                  pcf.nodesCoordinator,
		StartEpoch:                        startEpochNum,
		EpochStartEventNotifier:           pcf.coreData.EpochStartNotifierWithConfirm(),
		CacheRefreshIntervalDurationInSec: cacheRefreshDuration,
		ValidatorStatistics:               validatorStatisticsProcessor,
		MaxRating:                         pcf.maxRating,
		PubKeyConverter:                   pcf.coreData.ValidatorPubKeyConverter(),
	}

	validatorsProvider, err := peer.NewValidatorsProvider(argVSP)
	if err != nil {
		return nil, err
	}

	epochStartTrigger, err := pcf.newEpochStartTrigger(requestHandler)
	if err != nil {
		return nil, err
	}

	requestHandler.SetEpoch(epochStartTrigger.Epoch())

	err = dataRetriever.SetEpochHandlerToHdrResolver(resolversContainer, epochStartTrigger)
	if err != nil {
		return nil, err
	}

	log.Debug("Validator stats created", "validatorStatsRootHash", validatorStatsRootHash)

	err = pcf.prepareGenesisBlock(genesisBlocks)
	if err != nil {
		return nil, err
	}

	bootStr, err := pcf.data.StorageService().GetStorer(dataRetriever.BootstrapUnit)
	if err != nil {
		return nil, err
	}

	bootStorer, err := bootstrapStorage.NewBootstrapStorer(pcf.coreData.InternalMarshalizer(), bootStr)
	if err != nil {
		return nil, err
	}

	argsHeaderValidator := block.ArgsHeaderValidator{
		Hasher:      pcf.coreData.Hasher(),
		Marshalizer: pcf.coreData.InternalMarshalizer(),
	}
	headerValidator, err := block.NewHeaderValidator(argsHeaderValidator)
	if err != nil {
		return nil, err
	}

	blockTracker, err := pcf.newBlockTracker(
		headerValidator,
		requestHandler,
		genesisBlocks,
	)
	if err != nil {
		return nil, err
	}

	argsMiniBlocksPoolsCleaner := poolsCleaner.ArgMiniBlocksPoolsCleaner{
		ArgBasePoolsCleaner: poolsCleaner.ArgBasePoolsCleaner{
			RoundHandler:                   pcf.coreData.RoundHandler(),
			ShardCoordinator:               pcf.bootstrapComponents.ShardCoordinator(),
			MaxRoundsToKeepUnprocessedData: pcf.config.PoolsCleanersConfig.MaxRoundsToKeepUnprocessedMiniBlocks,
		},
		MiniblocksPool: pcf.data.Datapool().MiniBlocks(),
	}
	mbsPoolsCleaner, err := poolsCleaner.NewMiniBlocksPoolsCleaner(argsMiniBlocksPoolsCleaner)
	if err != nil {
		return nil, err
	}

	mbsPoolsCleaner.StartCleaning()

	argsBasePoolsCleaner := poolsCleaner.ArgTxsPoolsCleaner{
		ArgBasePoolsCleaner: poolsCleaner.ArgBasePoolsCleaner{
			RoundHandler:                   pcf.coreData.RoundHandler(),
			ShardCoordinator:               pcf.bootstrapComponents.ShardCoordinator(),
			MaxRoundsToKeepUnprocessedData: pcf.config.PoolsCleanersConfig.MaxRoundsToKeepUnprocessedTransactions,
		},
		AddressPubkeyConverter: pcf.coreData.AddressPubKeyConverter(),
		DataPool:               pcf.data.Datapool(),
	}
	txsPoolsCleaner, err := poolsCleaner.NewTxsPoolsCleaner(argsBasePoolsCleaner)
	if err != nil {
		return nil, err
	}

	txsPoolsCleaner.StartCleaning()

	_, err = track.NewMiniBlockTrack(
		pcf.data.Datapool(),
		pcf.bootstrapComponents.ShardCoordinator(),
		pcf.whiteListHandler,
	)
	if err != nil {
		return nil, err
	}

	hardforkTrigger, err := pcf.createHardforkTrigger(epochStartTrigger)
	if err != nil {
		return nil, err
	}

	interceptorContainerFactory, blackListHandler, err := pcf.newInterceptorContainerFactory(
		headerSigVerifier,
		pcf.bootstrapComponents.HeaderIntegrityVerifier(),
		blockTracker,
		epochStartTrigger,
		requestHandler,
		peerShardMapper,
		hardforkTrigger,
	)
	if err != nil {
		return nil, err
	}

	// TODO refactor all these factory calls
	interceptorsContainer, err := interceptorContainerFactory.Create()
	if err != nil {
		return nil, err
	}

	exportFactoryHandler, err := pcf.createExportFactoryHandler(
		headerValidator,
		requestHandler,
		resolversFinder,
		interceptorsContainer,
		headerSigVerifier,
		blockTracker,
	)
	if err != nil {
		return nil, err
	}

	err = hardforkTrigger.SetExportFactoryHandler(exportFactoryHandler)
	if err != nil {
		return nil, err
	}

	var pendingMiniBlocksHandler process.PendingMiniBlocksHandler
	pendingMiniBlocksHandler, err = pendingMb.NewNilPendingMiniBlocks()
	if err != nil {
		return nil, err
	}
	if pcf.bootstrapComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		pendingMiniBlocksHandler, err = pendingMb.NewPendingMiniBlocks()
		if err != nil {
			return nil, err
		}
	}

	forkDetector, err := pcf.newForkDetector(blackListHandler, blockTracker)
	if err != nil {
		return nil, err
	}

	vmOutputCacherConfig := storageFactory.GetCacherFromConfig(pcf.config.VMOutputCacher)
	vmOutputCacher, err := storageunit.NewCache(vmOutputCacherConfig)
	if err != nil {
		return nil, err
	}

	txSimulatorProcessorArgs := &txsimulator.ArgsTxSimulator{
		AddressPubKeyConverter: pcf.coreData.AddressPubKeyConverter(),
		ShardCoordinator:       pcf.bootstrapComponents.ShardCoordinator(),
		VMOutputCacher:         vmOutputCacher,
		Hasher:                 pcf.coreData.Hasher(),
		Marshalizer:            pcf.coreData.InternalMarshalizer(),
	}

	scheduledSCRSStorer, err := pcf.data.StorageService().GetStorer(dataRetriever.ScheduledSCRsUnit)
	if err != nil {
		return nil, err
	}

	scheduledTxsExecutionHandler, err := preprocess.NewScheduledTxsExecution(
		&disabled.TxProcessor{},
		&disabled.TxCoordinator{},
		scheduledSCRSStorer,
		pcf.coreData.InternalMarshalizer(),
		pcf.coreData.Hasher(),
		pcf.bootstrapComponents.ShardCoordinator(),
	)
	if err != nil {
		return nil, err
	}

	esdtDataStorageArgs := vmcommonBuiltInFunctions.ArgsNewESDTDataStorage{
		Accounts:              pcf.state.AccountsAdapterAPI(),
		GlobalSettingsHandler: disabled.NewDisabledGlobalSettingHandler(),
		Marshalizer:           pcf.coreData.InternalMarshalizer(),
		ShardCoordinator:      pcf.bootstrapComponents.ShardCoordinator(),
		EnableEpochsHandler:   pcf.coreData.EnableEpochsHandler(),
	}
	pcf.esdtNftStorage, err = vmcommonBuiltInFunctions.NewESDTDataStorage(esdtDataStorageArgs)
	if err != nil {
		return nil, err
	}

	processedMiniBlocksTracker := processedMb.NewProcessedMiniBlocksTracker()

	receiptsRepository, err := receipts.NewReceiptsRepository(receipts.ArgsNewReceiptsRepository{
		Store:      pcf.data.StorageService(),
		Marshaller: pcf.coreData.InternalMarshalizer(),
		Hasher:     pcf.coreData.Hasher(),
	})
	if err != nil {
		return nil, err
	}

	blockProcessorComponents, err := pcf.newBlockProcessor(
		requestHandler,
		forkDetector,
		epochStartTrigger,
		bootStorer,
		validatorStatisticsProcessor,
		headerValidator,
		blockTracker,
		pendingMiniBlocksHandler,
		txSimulatorProcessorArgs,
		pcf.coreData.WasmVMChangeLocker(),
		scheduledTxsExecutionHandler,
		processedMiniBlocksTracker,
		receiptsRepository,
	)
	if err != nil {
		return nil, err
	}

	conversionBase := 10
	genesisNodePrice, ok := big.NewInt(0).SetString(pcf.systemSCConfig.StakingSystemSCConfig.GenesisNodePrice, conversionBase)
	if !ok {
		return nil, errors.New("invalid genesis node price")
	}

	nodesSetupChecker, err := checking.NewNodesSetupChecker(
		pcf.accountsParser,
		genesisNodePrice,
		pcf.coreData.ValidatorPubKeyConverter(),
		pcf.crypto.BlockSignKeyGen(),
	)
	if err != nil {
		return nil, err
	}

	err = nodesSetupChecker.Check(pcf.coreData.GenesisNodesSetup().AllInitialNodes())
	if err != nil {
		return nil, err
	}

	txSimulator, err := txsimulator.NewTransactionSimulator(*txSimulatorProcessorArgs)
	if err != nil {
		return nil, err
	}

	observerBLSPrivateKey, observerBLSPublicKey := pcf.crypto.BlockSignKeyGen().GeneratePair()
	observerBLSPublicKeyBuff, err := observerBLSPublicKey.ToByteArray()
	if err != nil {
		return nil, fmt.Errorf("error generating observerBLSPublicKeyBuff, %w", err)
	} else {
		log.Debug("generated BLS private key for redundancy handler. This key will be used on heartbeat messages "+
			"if the node is in backup mode and the main node is active", "hex public key", observerBLSPublicKeyBuff)
	}

	nodeRedundancyArg := redundancy.ArgNodeRedundancy{
		RedundancyLevel:    pcf.prefConfigs.RedundancyLevel,
		Messenger:          pcf.network.NetworkMessenger(),
		ObserverPrivateKey: observerBLSPrivateKey,
	}
	nodeRedundancyHandler, err := redundancy.NewNodeRedundancy(nodeRedundancyArg)
	if err != nil {
		return nil, err
	}

	dataPacker, err := partitioning.NewSimpleDataPacker(pcf.coreData.InternalMarshalizer())
	if err != nil {
		return nil, err
	}
	args := txsSender.ArgsTxsSenderWithAccumulator{
		Marshaller:        pcf.coreData.InternalMarshalizer(),
		ShardCoordinator:  pcf.bootstrapComponents.ShardCoordinator(),
		NetworkMessenger:  pcf.network.NetworkMessenger(),
		AccumulatorConfig: pcf.config.Antiflood.TxAccumulator,
		DataPacker:        dataPacker,
	}
	txsSenderWithAccumulator, err := txsSender.NewTxsSenderWithAccumulator(args)
	if err != nil {
		return nil, err
	}

	return &processComponents{
		nodesCoordinator:             pcf.nodesCoordinator,
		shardCoordinator:             pcf.bootstrapComponents.ShardCoordinator(),
		interceptorsContainer:        interceptorsContainer,
		resolversFinder:              resolversFinder,
		roundHandler:                 pcf.coreData.RoundHandler(),
		forkDetector:                 forkDetector,
		blockProcessor:               blockProcessorComponents.blockProcessor,
		epochStartTrigger:            epochStartTrigger,
		epochStartNotifier:           pcf.coreData.EpochStartNotifierWithConfirm(),
		blackListHandler:             blackListHandler,
		bootStorer:                   bootStorer,
		headerSigVerifier:            headerSigVerifier,
		validatorsStatistics:         validatorStatisticsProcessor,
		validatorsProvider:           validatorsProvider,
		blockTracker:                 blockTracker,
		pendingMiniBlocksHandler:     pendingMiniBlocksHandler,
		requestHandler:               requestHandler,
		txLogsProcessor:              txLogsProcessor,
		headerConstructionValidator:  headerValidator,
		headerIntegrityVerifier:      pcf.bootstrapComponents.HeaderIntegrityVerifier(),
		peerShardMapper:              peerShardMapper,
		txSimulatorProcessor:         txSimulator,
		miniBlocksPoolCleaner:        mbsPoolsCleaner,
		txsPoolCleaner:               txsPoolsCleaner,
		fallbackHeaderValidator:      fallbackHeaderValidator,
		whiteListHandler:             pcf.whiteListHandler,
		whiteListerVerifiedTxs:       pcf.whiteListerVerifiedTxs,
		historyRepository:            pcf.historyRepo,
		importStartHandler:           pcf.importStartHandler,
		requestedItemsHandler:        pcf.requestedItemsHandler,
		importHandler:                pcf.importHandler,
		nodeRedundancyHandler:        nodeRedundancyHandler,
		currentEpochProvider:         currentEpochProvider,
		vmFactoryForTxSimulator:      blockProcessorComponents.vmFactoryForTxSimulate,
		vmFactoryForProcessing:       blockProcessorComponents.vmFactoryForProcessing,
		scheduledTxsExecutionHandler: scheduledTxsExecutionHandler,
		txsSender:                    txsSenderWithAccumulator,
		hardforkTrigger:              hardforkTrigger,
		processedMiniBlocksTracker:   processedMiniBlocksTracker,
		esdtDataStorageForApi:        pcf.esdtNftStorage,
		accountsParser:               pcf.accountsParser,
		receiptsRepository:           receiptsRepository,
	}, nil
}

func (pcf *processComponentsFactory) newValidatorStatisticsProcessor() (process.ValidatorStatisticsProcessor, error) {

	storageService := pcf.data.StorageService()

	var peerDataPool peer.DataPool = pcf.data.Datapool()
	if pcf.bootstrapComponents.ShardCoordinator().SelfId() < pcf.bootstrapComponents.ShardCoordinator().NumberOfShards() {
		peerDataPool = pcf.data.Datapool()
	}

	hardforkConfig := pcf.config.Hardfork
	ratingEnabledEpoch := uint32(0)
	if hardforkConfig.AfterHardFork {
		ratingEnabledEpoch = hardforkConfig.StartEpoch + hardforkConfig.ValidatorGracePeriodInEpochs
	}
	arguments := peer.ArgValidatorStatisticsProcessor{
		PeerAdapter:                          pcf.state.PeerAccounts(),
		PubkeyConv:                           pcf.coreData.ValidatorPubKeyConverter(),
		NodesCoordinator:                     pcf.nodesCoordinator,
		ShardCoordinator:                     pcf.bootstrapComponents.ShardCoordinator(),
		DataPool:                             peerDataPool,
		StorageService:                       storageService,
		Marshalizer:                          pcf.coreData.InternalMarshalizer(),
		Rater:                                pcf.coreData.Rater(),
		MaxComputableRounds:                  pcf.config.GeneralSettings.MaxComputableRounds,
		MaxConsecutiveRoundsOfRatingDecrease: pcf.config.GeneralSettings.MaxConsecutiveRoundsOfRatingDecrease,
		RewardsHandler:                       pcf.coreData.EconomicsData(),
		NodesSetup:                           pcf.coreData.GenesisNodesSetup(),
		RatingEnableEpoch:                    ratingEnabledEpoch,
		GenesisNonce:                         pcf.data.Blockchain().GetGenesisHeader().GetNonce(),
		EnableEpochsHandler:                  pcf.coreData.EnableEpochsHandler(),
	}

	validatorStatisticsProcessor, err := peer.NewValidatorStatisticsProcessor(arguments)
	if err != nil {
		return nil, err
	}

	return validatorStatisticsProcessor, nil
}

func (pcf *processComponentsFactory) newEpochStartTrigger(requestHandler epochStart.RequestHandler) (epochStart.TriggerHandler, error) {
	if pcf.bootstrapComponents.ShardCoordinator().SelfId() < pcf.bootstrapComponents.ShardCoordinator().NumberOfShards() {
		argsHeaderValidator := block.ArgsHeaderValidator{
			Hasher:      pcf.coreData.Hasher(),
			Marshalizer: pcf.coreData.InternalMarshalizer(),
		}
		headerValidator, err := block.NewHeaderValidator(argsHeaderValidator)
		if err != nil {
			return nil, err
		}

		argsPeerMiniBlockSyncer := shardchain.ArgPeerMiniBlockSyncer{
			MiniBlocksPool:     pcf.data.Datapool().MiniBlocks(),
			ValidatorsInfoPool: pcf.data.Datapool().ValidatorsInfo(),
			RequestHandler:     requestHandler,
		}

		peerMiniBlockSyncer, err := shardchain.NewPeerMiniBlockSyncer(argsPeerMiniBlockSyncer)
		if err != nil {
			return nil, err
		}

		argEpochStart := &shardchain.ArgsShardEpochStartTrigger{
			Marshalizer:          pcf.coreData.InternalMarshalizer(),
			Hasher:               pcf.coreData.Hasher(),
			HeaderValidator:      headerValidator,
			Uint64Converter:      pcf.coreData.Uint64ByteSliceConverter(),
			DataPool:             pcf.data.Datapool(),
			Storage:              pcf.data.StorageService(),
			RequestHandler:       requestHandler,
			Epoch:                pcf.bootstrapComponents.EpochBootstrapParams().Epoch(),
			EpochStartNotifier:   pcf.coreData.EpochStartNotifierWithConfirm(),
			Validity:             process.MetaBlockValidity,
			Finality:             process.BlockFinality,
			PeerMiniBlocksSyncer: peerMiniBlockSyncer,
			RoundHandler:         pcf.coreData.RoundHandler(),
			AppStatusHandler:     pcf.statusCoreComponents.AppStatusHandler(),
			EnableEpochsHandler:  pcf.coreData.EnableEpochsHandler(),
		}
		epochStartTrigger, err := shardchain.NewEpochStartTrigger(argEpochStart)
		if err != nil {
			return nil, errors.New("error creating new start of epoch trigger" + err.Error())
		}

		return epochStartTrigger, nil
	}

	if pcf.bootstrapComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		argEpochStart := &metachain.ArgsNewMetaEpochStartTrigger{
			GenesisTime:        time.Unix(pcf.coreData.GenesisNodesSetup().GetStartTime(), 0),
			Settings:           &pcf.config.EpochStartConfig,
			Epoch:              pcf.bootstrapComponents.EpochBootstrapParams().Epoch(),
			EpochStartRound:    pcf.data.Blockchain().GetGenesisHeader().GetRound(),
			EpochStartNotifier: pcf.coreData.EpochStartNotifierWithConfirm(),
			Storage:            pcf.data.StorageService(),
			Marshalizer:        pcf.coreData.InternalMarshalizer(),
			Hasher:             pcf.coreData.Hasher(),
			AppStatusHandler:   pcf.statusCoreComponents.AppStatusHandler(),
			DataPool:           pcf.data.Datapool(),
		}
		epochStartTrigger, err := metachain.NewEpochStartTrigger(argEpochStart)
		if err != nil {
			return nil, errors.New("error creating new start of epoch trigger" + err.Error())
		}

		return epochStartTrigger, nil
	}

	return nil, errors.New("error creating new start of epoch trigger because of invalid shard id")
}

func (pcf *processComponentsFactory) generateGenesisHeadersAndApplyInitialBalances() (map[uint32]data.HeaderHandler, map[uint32]*genesis.IndexingData, error) {
	genesisVmConfig := pcf.config.VirtualMachine.Execution
	conversionBase := 10
	genesisNodePrice, ok := big.NewInt(0).SetString(pcf.systemSCConfig.StakingSystemSCConfig.GenesisNodePrice, conversionBase)
	if !ok {
		return nil, nil, errors.New("invalid genesis node price")
	}

	arg := processGenesis.ArgsGenesisBlockCreator{
		Core:                 pcf.coreData,
		Data:                 pcf.data,
		GenesisTime:          uint64(pcf.coreData.GenesisNodesSetup().GetStartTime()),
		StartEpochNum:        pcf.bootstrapComponents.EpochBootstrapParams().Epoch(),
		Accounts:             pcf.state.AccountsAdapter(),
		InitialNodesSetup:    pcf.coreData.GenesisNodesSetup(),
		Economics:            pcf.coreData.EconomicsData(),
		ShardCoordinator:     pcf.bootstrapComponents.ShardCoordinator(),
		AccountsParser:       pcf.accountsParser,
		SmartContractParser:  pcf.smartContractParser,
		ValidatorAccounts:    pcf.state.PeerAccounts(),
		GasSchedule:          pcf.gasSchedule,
		VirtualMachineConfig: genesisVmConfig,
		TxLogsProcessor:      pcf.txLogsProcessor,
		HardForkConfig:       pcf.config.Hardfork,
		TrieStorageManagers:  pcf.state.TrieStorageManagers(),
		SystemSCConfig:       *pcf.systemSCConfig,
		ImportStartHandler:   pcf.importStartHandler,
		WorkingDir:           pcf.workingDir,
		BlockSignKeyGen:      pcf.crypto.BlockSignKeyGen(),
		GenesisString:        pcf.config.GeneralSettings.GenesisString,
		GenesisNodePrice:     genesisNodePrice,
		EpochConfig:          &pcf.epochConfig,
	}

	gbc, err := processGenesis.NewGenesisBlockCreator(arg)
	if err != nil {
		return nil, nil, err
	}
	pcf.importHandler = gbc.ImportHandler()

	genesisBlocks, err := gbc.CreateGenesisBlocks()
	if err != nil {
		return nil, nil, err
	}
	indexingData := gbc.GetIndexingData()

	return genesisBlocks, indexingData, nil
}

func (pcf *processComponentsFactory) indexAndReturnGenesisAccounts() (map[string]*outport.AlteredAccount, error) {
	if !pcf.statusComponents.OutportHandler().HasDrivers() {
		return map[string]*outport.AlteredAccount{}, nil
	}

	rootHash, err := pcf.state.AccountsAdapter().RootHash()
	if err != nil {
		return nil, err
	}

	leavesChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	err = pcf.state.AccountsAdapter().GetAllLeaves(leavesChannels, context.Background(), rootHash)
	if err != nil {
		return nil, err
	}

	genesisAccounts := make(map[string]*outport.AlteredAccount, 0)
	for leaf := range leavesChannels.LeavesChan {
		userAccount, errUnmarshal := pcf.unmarshalUserAccount(leaf.Key(), leaf.Value())
		if errUnmarshal != nil {
			log.Debug("cannot unmarshal genesis user account. it may be a code leaf", "error", errUnmarshal)
			continue
		}

		encodedAddress := pcf.coreData.AddressPubKeyConverter().Encode(userAccount.AddressBytes())
		genesisAccounts[encodedAddress] = &outport.AlteredAccount{
			AdditionalData: &outport.AdditionalAccountData{
				BalanceChanged: true,
			},
			Address: encodedAddress,
			Balance: userAccount.GetBalance().String(),
			Nonce:   userAccount.GetNonce(),
			Tokens:  nil,
		}
	}

	err = leavesChannels.ErrChan.ReadFromChanNonBlocking()
	if err != nil {
		return nil, err
	}

	shardID := pcf.bootstrapComponents.ShardCoordinator().SelfId()
	pcf.statusComponents.OutportHandler().SaveAccounts(uint64(pcf.coreData.GenesisNodesSetup().GetStartTime()), genesisAccounts, shardID)
	return genesisAccounts, nil
}

func (pcf *processComponentsFactory) unmarshalUserAccount(address []byte, userAccountsBytes []byte) (state.UserAccountHandler, error) {
	userAccount, err := state.NewUserAccount(address)
	if err != nil {
		return nil, err
	}
	err = pcf.coreData.InternalMarshalizer().Unmarshal(userAccount, userAccountsBytes)
	if err != nil {
		return nil, err
	}

	return userAccount, nil
}

func (pcf *processComponentsFactory) setGenesisHeader(genesisBlocks map[uint32]data.HeaderHandler) error {
	genesisBlock, ok := genesisBlocks[pcf.bootstrapComponents.ShardCoordinator().SelfId()]
	if !ok {
		return errors.New("genesis block does not exist")
	}

	err := pcf.data.Blockchain().SetGenesisHeader(genesisBlock)
	if err != nil {
		return err
	}

	return nil
}

func (pcf *processComponentsFactory) prepareGenesisBlock(
	genesisBlocks map[uint32]data.HeaderHandler,
) error {
	genesisBlock, ok := genesisBlocks[pcf.bootstrapComponents.ShardCoordinator().SelfId()]
	if !ok {
		return errors.New("genesis block does not exist")
	}

	genesisBlockHash, err := core.CalculateHash(pcf.coreData.InternalMarshalizer(), pcf.coreData.Hasher(), genesisBlock)
	if err != nil {
		return err
	}

	err = pcf.data.Blockchain().SetGenesisHeader(genesisBlock)
	if err != nil {
		return err
	}

	pcf.data.Blockchain().SetGenesisHeaderHash(genesisBlockHash)
	nonceToByteSlice := pcf.coreData.Uint64ByteSliceConverter().ToByteSlice(genesisBlock.GetNonce())

	err = pcf.saveGenesisHeaderToStorage(genesisBlock, genesisBlockHash, nonceToByteSlice)
	if err != nil {
		return err
	}

	return nil
}

func (pcf *processComponentsFactory) saveGenesisHeaderToStorage(
	genesisBlock data.HeaderHandler,
	genesisBlockHash []byte,
	nonceToByteSlice []byte,
) error {
	marshalledBlock, err := pcf.coreData.InternalMarshalizer().Marshal(genesisBlock)
	if err != nil {
		return err
	}

	if pcf.bootstrapComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		pcf.saveMetaBlock(genesisBlockHash, marshalledBlock, nonceToByteSlice)
	} else {
		pcf.saveShardBlock(genesisBlockHash, marshalledBlock, nonceToByteSlice, genesisBlock.GetShardID())
	}

	return nil
}

func (pcf *processComponentsFactory) saveMetaBlock(genesisBlockHash []byte, marshalledBlock []byte, nonceToByteSlice []byte) {
	errNotCritical := pcf.data.StorageService().Put(dataRetriever.MetaBlockUnit, genesisBlockHash, marshalledBlock)
	if errNotCritical != nil {
		log.Error("error storing genesis metablock", "error", errNotCritical.Error())
	}
	errNotCritical = pcf.data.StorageService().Put(dataRetriever.MetaHdrNonceHashDataUnit, nonceToByteSlice, genesisBlockHash)
	if errNotCritical != nil {
		log.Error("error storing genesis metablock (nonce-hash)", "error", errNotCritical.Error())
	}
}

func (pcf *processComponentsFactory) saveShardBlock(genesisBlockHash []byte, marshalledBlock []byte, nonceToByteSlice []byte, shardID uint32) {
	errNotCritical := pcf.data.StorageService().Put(dataRetriever.BlockHeaderUnit, genesisBlockHash, marshalledBlock)
	if errNotCritical != nil {
		log.Error("error storing genesis shardblock", "error", errNotCritical.Error())
	}

	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardID)
	errNotCritical = pcf.data.StorageService().Put(hdrNonceHashDataUnit, nonceToByteSlice, genesisBlockHash)
	if errNotCritical != nil {
		log.Error("error storing genesis shard header (nonce-hash)", "error", errNotCritical.Error())
	}
}

func getGenesisBlockForShard(miniBlocks []*dataBlock.MiniBlock, shardId uint32) *dataBlock.Body {
	var indexMiniBlocks = make([]*dataBlock.MiniBlock, 0)

	for _, miniBlock := range miniBlocks {
		if miniBlock.GetSenderShardID() == shardId ||
			miniBlock.GetReceiverShardID() == shardId {
			indexMiniBlocks = append(indexMiniBlocks, miniBlock)
		}
	}

	genesisMiniBlocks := &dataBlock.Body{
		MiniBlocks: indexMiniBlocks,
	}

	return genesisMiniBlocks
}

func getGenesisIntraShardMiniblocks(miniBlocks []*dataBlock.MiniBlock) []*dataBlock.MiniBlock {
	intraShardMiniBlocks := make([]*dataBlock.MiniBlock, 0)

	for _, miniBlock := range miniBlocks {
		if miniBlock.GetReceiverShardID() == miniBlock.GetSenderShardID() {
			intraShardMiniBlocks = append(intraShardMiniBlocks, miniBlock)
		}
	}

	return intraShardMiniBlocks
}

func (pcf *processComponentsFactory) createGenesisMiniBlockHandlers(miniBlocks []*dataBlock.MiniBlock) ([]data.MiniBlockHeaderHandler, error) {
	miniBlockHeaderHandlers := make([]data.MiniBlockHeaderHandler, len(miniBlocks))

	for i := 0; i < len(miniBlocks); i++ {
		txCount := len(miniBlocks[i].GetTxHashes())

		miniBlockHash, err := core.CalculateHash(pcf.coreData.InternalMarshalizer(), pcf.coreData.Hasher(), miniBlocks[i])
		if err != nil {
			return nil, err
		}

		miniBlockHeader := &dataBlock.MiniBlockHeader{
			Hash:            miniBlockHash,
			SenderShardID:   miniBlocks[i].GetSenderShardID(),
			ReceiverShardID: miniBlocks[i].GetReceiverShardID(),
			TxCount:         uint32(txCount),
			Type:            miniBlocks[i].GetType(),
		}

		err = miniBlockHeader.SetProcessingType(int32(dataBlock.Normal))
		if err != nil {
			return nil, err
		}
		err = miniBlockHeader.SetConstructionState(int32(dataBlock.Final))
		if err != nil {
			return nil, err
		}

		miniBlockHeaderHandlers[i] = miniBlockHeader
	}

	return miniBlockHeaderHandlers, nil
}

func (pcf *processComponentsFactory) indexGenesisBlocks(
	genesisBlocks map[uint32]data.HeaderHandler,
	initialIndexingData map[uint32]*genesis.IndexingData,
	alteredAccounts map[string]*outport.AlteredAccount,
) error {
	currentShardId := pcf.bootstrapComponents.ShardCoordinator().SelfId()
	originalGenesisBlockHeader := genesisBlocks[currentShardId]
	genesisBlockHeader := originalGenesisBlockHeader.ShallowClone()

	genesisBlockHash, err := core.CalculateHash(pcf.coreData.InternalMarshalizer(), pcf.coreData.Hasher(), genesisBlockHeader)
	if err != nil {
		return err
	}

	miniBlocks, txsPoolPerShard, errGenerate := pcf.accountsParser.GenerateInitialTransactions(pcf.bootstrapComponents.ShardCoordinator(), initialIndexingData)
	if errGenerate != nil {
		return errGenerate
	}

	intraShardMiniBlocks := getGenesisIntraShardMiniblocks(miniBlocks)
	genesisBody := getGenesisBlockForShard(miniBlocks, currentShardId)

	if pcf.statusComponents.OutportHandler().HasDrivers() {
		log.Info("indexGenesisBlocks(): indexer.SaveBlock", "hash", genesisBlockHash)

		// manually add the genesis minting address as it is not exist in the trie
		genesisAddress := pcf.accountsParser.GenesisMintingAddress()
		alteredAccounts[genesisAddress] = &outport.AlteredAccount{
			Address: genesisAddress,
			Balance: "0",
		}

		_ = genesisBlockHeader.SetTxCount(uint32(len(txsPoolPerShard[currentShardId].Txs)))

		arg := &outport.ArgsSaveBlockData{
			HeaderHash: genesisBlockHash,
			Body:       genesisBody,
			Header:     genesisBlockHeader,
			HeaderGasConsumption: outport.HeaderGasConsumption{
				GasProvided:    0,
				GasRefunded:    0,
				GasPenalized:   0,
				MaxGasPerBlock: pcf.coreData.EconomicsData().MaxGasLimitPerBlock(currentShardId),
			},
			TransactionsPool: txsPoolPerShard[currentShardId],
			AlteredAccounts:  alteredAccounts,
		}
		pcf.statusComponents.OutportHandler().SaveBlock(arg)
	}

	log.Info("indexGenesisBlocks(): historyRepo.RecordBlock", "shardID", currentShardId, "hash", genesisBlockHash)
	if txsPoolPerShard[currentShardId] != nil {
		err = pcf.historyRepo.RecordBlock(
			genesisBlockHash,
			originalGenesisBlockHeader,
			genesisBody,
			unwrapTxs(txsPoolPerShard[currentShardId].Scrs),
			unwrapTxs(txsPoolPerShard[currentShardId].Receipts),
			intraShardMiniBlocks,
			txsPoolPerShard[currentShardId].Logs)
		if err != nil {
			return err
		}
	}

	err = pcf.saveGenesisMiniBlocksToStorage(miniBlocks)
	if err != nil {
		return err
	}

	if txsPoolPerShard[currentShardId] != nil {
		err = pcf.saveGenesisTxsToStorage(unwrapTxs(txsPoolPerShard[currentShardId].Txs))
		if err != nil {
			return err
		}
	}

	nonceByHashDataUnit := dataRetriever.GetHdrNonceHashDataUnit(currentShardId)
	nonceAsBytes := pcf.coreData.Uint64ByteSliceConverter().ToByteSlice(genesisBlockHeader.GetNonce())
	err = pcf.data.StorageService().Put(nonceByHashDataUnit, nonceAsBytes, genesisBlockHash)
	if err != nil {
		return err
	}

	err = pcf.saveAlteredGenesisHeaderToStorage(
		genesisBlockHeader,
		genesisBlockHash,
		genesisBody,
		intraShardMiniBlocks,
		txsPoolPerShard)
	if err != nil {
		return err
	}

	return nil
}

func (pcf *processComponentsFactory) saveAlteredGenesisHeaderToStorage(
	genesisBlockHeader data.HeaderHandler,
	genesisBlockHash []byte,
	genesisBody *dataBlock.Body,
	intraShardMiniBlocks []*dataBlock.MiniBlock,
	txsPoolPerShard map[uint32]*outport.Pool,
) error {
	currentShardId := pcf.bootstrapComponents.ShardCoordinator().SelfId()

	genesisMiniBlockHeaderHandlers, err := pcf.createGenesisMiniBlockHandlers(genesisBody.GetMiniBlocks())
	if err != nil {
		return err
	}

	nonceAsBytes := pcf.coreData.Uint64ByteSliceConverter().ToByteSlice(genesisBlockHeader.GetNonce())
	nonceAsBytes = append(nonceAsBytes, []byte(common.GenesisStorageSuffix)...)
	err = genesisBlockHeader.SetMiniBlockHeaderHandlers(genesisMiniBlockHeaderHandlers)
	if err != nil {
		return err
	}

	genesisBlockHash = append(genesisBlockHash, []byte(common.GenesisStorageSuffix)...)
	err = pcf.saveGenesisHeaderToStorage(genesisBlockHeader, genesisBlockHash, nonceAsBytes)
	if err != nil {
		return err
	}

	if txsPoolPerShard[currentShardId] != nil {
		err = pcf.historyRepo.RecordBlock(
			genesisBlockHash,
			genesisBlockHeader,
			genesisBody,
			unwrapTxs(txsPoolPerShard[currentShardId].Scrs),
			unwrapTxs(txsPoolPerShard[currentShardId].Receipts),
			intraShardMiniBlocks,
			txsPoolPerShard[currentShardId].Logs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pcf *processComponentsFactory) saveGenesisTxsToStorage(txs map[string]data.TransactionHandler) error {
	for txHash, tx := range txs {
		marshalledTx, err := pcf.coreData.InternalMarshalizer().Marshal(tx)
		if err != nil {
			return err
		}

		err = pcf.data.StorageService().Put(dataRetriever.TransactionUnit, []byte(txHash), marshalledTx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pcf *processComponentsFactory) saveGenesisMiniBlocksToStorage(miniBlocks []*dataBlock.MiniBlock) error {
	for _, miniBlock := range miniBlocks {
		marshalizedMiniBlock, err := pcf.coreData.InternalMarshalizer().Marshal(miniBlock)
		if err != nil {
			return err
		}

		miniBlockHash, err := core.CalculateHash(pcf.coreData.InternalMarshalizer(), pcf.coreData.Hasher(), miniBlock)
		if err != nil {
			return err
		}

		err = pcf.data.StorageService().Put(dataRetriever.MiniBlockUnit, miniBlockHash, marshalizedMiniBlock)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pcf *processComponentsFactory) newBlockTracker(
	headerValidator process.HeaderConstructionValidator,
	requestHandler process.RequestHandler,
	genesisBlocks map[uint32]data.HeaderHandler,
) (process.BlockTracker, error) {
	argBaseTracker := track.ArgBaseTracker{
		Hasher:           pcf.coreData.Hasher(),
		HeaderValidator:  headerValidator,
		Marshalizer:      pcf.coreData.InternalMarshalizer(),
		RequestHandler:   requestHandler,
		RoundHandler:     pcf.coreData.RoundHandler(),
		ShardCoordinator: pcf.bootstrapComponents.ShardCoordinator(),
		Store:            pcf.data.StorageService(),
		StartHeaders:     genesisBlocks,
		PoolsHolder:      pcf.data.Datapool(),
		WhitelistHandler: pcf.whiteListHandler,
		FeeHandler:       pcf.coreData.EconomicsData(),
	}

	if pcf.bootstrapComponents.ShardCoordinator().SelfId() < pcf.bootstrapComponents.ShardCoordinator().NumberOfShards() {
		arguments := track.ArgShardTracker{
			ArgBaseTracker: argBaseTracker,
		}

		return track.NewShardBlockTrack(arguments)
	}

	if pcf.bootstrapComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		arguments := track.ArgMetaTracker{
			ArgBaseTracker: argBaseTracker,
		}

		return track.NewMetaBlockTrack(arguments)
	}

	return nil, errors.New("could not create block tracker")
}

// -- Resolvers container Factory begin
func (pcf *processComponentsFactory) newResolverContainerFactory(
	currentEpochProvider dataRetriever.CurrentNetworkEpochProviderHandler,
) (dataRetriever.ResolversContainerFactory, error) {

	if pcf.importDBConfig.IsImportDBMode {
		log.Debug("starting with storage resolvers", "path", pcf.importDBConfig.ImportDBWorkingDir)
		return pcf.newStorageResolver()
	}

	payloadValidator, err := validator.NewPeerAuthenticationPayloadValidator(pcf.config.HeartbeatV2.HeartbeatExpiryTimespanInSec)
	if err != nil {
		return nil, err
	}

	if pcf.bootstrapComponents.ShardCoordinator().SelfId() < pcf.bootstrapComponents.ShardCoordinator().NumberOfShards() {
		return pcf.newShardResolverContainerFactory(currentEpochProvider, payloadValidator)
	}
	if pcf.bootstrapComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		return pcf.newMetaResolverContainerFactory(currentEpochProvider, payloadValidator)
	}

	return nil, errors.New("could not create interceptor and resolver container factory")
}

func (pcf *processComponentsFactory) newShardResolverContainerFactory(
	currentEpochProvider dataRetriever.CurrentNetworkEpochProviderHandler,
	payloadValidator process.PeerAuthenticationPayloadValidator,
) (dataRetriever.ResolversContainerFactory, error) {

	dataPacker, err := partitioning.NewSimpleDataPacker(pcf.coreData.InternalMarshalizer())
	if err != nil {
		return nil, err
	}

	resolversContainerFactoryArgs := resolverscontainer.FactoryArgs{
		ShardCoordinator:            pcf.bootstrapComponents.ShardCoordinator(),
		Messenger:                   pcf.network.NetworkMessenger(),
		Store:                       pcf.data.StorageService(),
		Marshalizer:                 pcf.coreData.InternalMarshalizer(),
		DataPools:                   pcf.data.Datapool(),
		Uint64ByteSliceConverter:    pcf.coreData.Uint64ByteSliceConverter(),
		DataPacker:                  dataPacker,
		TriesContainer:              pcf.state.TriesContainer(),
		SizeCheckDelta:              pcf.config.Marshalizer.SizeCheckDelta,
		InputAntifloodHandler:       pcf.network.InputAntiFloodHandler(),
		OutputAntifloodHandler:      pcf.network.OutputAntiFloodHandler(),
		NumConcurrentResolvingJobs:  pcf.config.Antiflood.NumConcurrentResolverJobs,
		IsFullHistoryNode:           pcf.prefConfigs.FullArchive,
		CurrentNetworkEpochProvider: currentEpochProvider,
		ResolverConfig:              pcf.config.Resolvers,
		PreferredPeersHolder:        pcf.network.PreferredPeersHolderHandler(),
		PeersRatingHandler:          pcf.network.PeersRatingHandler(),
		PayloadValidator:            payloadValidator,
	}
	resolversContainerFactory, err := resolverscontainer.NewShardResolversContainerFactory(resolversContainerFactoryArgs)
	if err != nil {
		return nil, err
	}

	return resolversContainerFactory, nil
}

func (pcf *processComponentsFactory) newMetaResolverContainerFactory(
	currentEpochProvider dataRetriever.CurrentNetworkEpochProviderHandler,
	payloadValidator process.PeerAuthenticationPayloadValidator,
) (dataRetriever.ResolversContainerFactory, error) {

	dataPacker, err := partitioning.NewSimpleDataPacker(pcf.coreData.InternalMarshalizer())
	if err != nil {
		return nil, err
	}

	resolversContainerFactoryArgs := resolverscontainer.FactoryArgs{
		ShardCoordinator:            pcf.bootstrapComponents.ShardCoordinator(),
		Messenger:                   pcf.network.NetworkMessenger(),
		Store:                       pcf.data.StorageService(),
		Marshalizer:                 pcf.coreData.InternalMarshalizer(),
		DataPools:                   pcf.data.Datapool(),
		Uint64ByteSliceConverter:    pcf.coreData.Uint64ByteSliceConverter(),
		DataPacker:                  dataPacker,
		TriesContainer:              pcf.state.TriesContainer(),
		SizeCheckDelta:              pcf.config.Marshalizer.SizeCheckDelta,
		InputAntifloodHandler:       pcf.network.InputAntiFloodHandler(),
		OutputAntifloodHandler:      pcf.network.OutputAntiFloodHandler(),
		NumConcurrentResolvingJobs:  pcf.config.Antiflood.NumConcurrentResolverJobs,
		IsFullHistoryNode:           pcf.prefConfigs.FullArchive,
		CurrentNetworkEpochProvider: currentEpochProvider,
		ResolverConfig:              pcf.config.Resolvers,
		PreferredPeersHolder:        pcf.network.PreferredPeersHolderHandler(),
		PeersRatingHandler:          pcf.network.PeersRatingHandler(),
		PayloadValidator:            payloadValidator,
	}
	resolversContainerFactory, err := resolverscontainer.NewMetaResolversContainerFactory(resolversContainerFactoryArgs)
	if err != nil {
		return nil, err
	}
	return resolversContainerFactory, nil
}

func (pcf *processComponentsFactory) newInterceptorContainerFactory(
	headerSigVerifier process.InterceptedHeaderSigVerifier,
	headerIntegrityVerifier nodeFactory.HeaderIntegrityVerifierHandler,
	validityAttester process.ValidityAttester,
	epochStartTrigger process.EpochStartTriggerHandler,
	requestHandler process.RequestHandler,
	peerShardMapper *networksharding.PeerShardMapper,
	hardforkTrigger factory.HardforkTrigger,
) (process.InterceptorsContainerFactory, process.TimeCacher, error) {
	if pcf.bootstrapComponents.ShardCoordinator().SelfId() < pcf.bootstrapComponents.ShardCoordinator().NumberOfShards() {
		return pcf.newShardInterceptorContainerFactory(
			headerSigVerifier,
			headerIntegrityVerifier,
			validityAttester,
			epochStartTrigger,
			requestHandler,
			peerShardMapper,
			hardforkTrigger,
		)
	}
	if pcf.bootstrapComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		return pcf.newMetaInterceptorContainerFactory(
			headerSigVerifier,
			headerIntegrityVerifier,
			validityAttester,
			epochStartTrigger,
			requestHandler,
			peerShardMapper,
			hardforkTrigger,
		)
	}

	return nil, nil, errors.New("could not create interceptor container factory")
}

func (pcf *processComponentsFactory) newStorageResolver() (dataRetriever.ResolversContainerFactory, error) {
	pathManager, err := storageFactory.CreatePathManager(
		storageFactory.ArgCreatePathManager{
			WorkingDir: pcf.importDBConfig.ImportDBWorkingDir,
			ChainID:    pcf.coreData.ChainID(),
		},
	)
	if err != nil {
		return nil, err
	}

	manualEpochStartNotifier := notifier.NewManualEpochStartNotifier()
	defer func() {
		// we need to call this after we wired all the notified components
		if pcf.importDBConfig.IsImportDBMode {
			manualEpochStartNotifier.NewEpoch(pcf.bootstrapComponents.EpochBootstrapParams().Epoch() + 1)
		}
	}()

	storageServiceCreator, err := storageFactory.NewStorageServiceFactory(
		storageFactory.StorageServiceFactoryArgs{
			Config:                        pcf.config,
			PrefsConfig:                   pcf.prefConfigs,
			ShardCoordinator:              pcf.bootstrapComponents.ShardCoordinator(),
			PathManager:                   pathManager,
			EpochStartNotifier:            manualEpochStartNotifier,
			NodeTypeProvider:              pcf.coreData.NodeTypeProvider(),
			CurrentEpoch:                  pcf.bootstrapComponents.EpochBootstrapParams().Epoch(),
			StorageType:                   storageFactory.ProcessStorageService,
			CreateTrieEpochRootHashStorer: false,
			NodeProcessingMode:            common.GetNodeProcessingMode(&pcf.importDBConfig),
		},
	)
	if err != nil {
		return nil, err
	}

	if pcf.bootstrapComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		store, errStore := storageServiceCreator.CreateForMeta()
		if errStore != nil {
			return nil, errStore
		}

		return pcf.createStorageResolversForMeta(
			store,
			manualEpochStartNotifier,
		)
	}

	store, err := storageServiceCreator.CreateForShard()
	if err != nil {
		return nil, err
	}

	return pcf.createStorageResolversForShard(
		store,
		manualEpochStartNotifier,
	)
}

func (pcf *processComponentsFactory) createStorageResolversForMeta(
	store dataRetriever.StorageService,
	manualEpochStartNotifier dataRetriever.ManualEpochStartNotifier,
) (dataRetriever.ResolversContainerFactory, error) {
	dataPacker, err := partitioning.NewSimpleDataPacker(pcf.coreData.InternalMarshalizer())
	if err != nil {
		return nil, err
	}

	resolversContainerFactoryArgs := storageResolversContainers.FactoryArgs{
		GeneralConfig:            pcf.config,
		ShardIDForTries:          pcf.importDBConfig.ImportDBTargetShardID,
		ChainID:                  pcf.coreData.ChainID(),
		WorkingDirectory:         pcf.importDBConfig.ImportDBWorkingDir,
		Hasher:                   pcf.coreData.Hasher(),
		ShardCoordinator:         pcf.bootstrapComponents.ShardCoordinator(),
		Messenger:                pcf.network.NetworkMessenger(),
		Store:                    store,
		Marshalizer:              pcf.coreData.InternalMarshalizer(),
		Uint64ByteSliceConverter: pcf.coreData.Uint64ByteSliceConverter(),
		DataPacker:               dataPacker,
		ManualEpochStartNotifier: manualEpochStartNotifier,
		ChanGracefullyClose:      pcf.coreData.ChanStopNodeProcess(),
	}
	resolversContainerFactory, err := storageResolversContainers.NewMetaResolversContainerFactory(resolversContainerFactoryArgs)
	if err != nil {
		return nil, err
	}

	return resolversContainerFactory, nil
}

func (pcf *processComponentsFactory) createStorageResolversForShard(
	store dataRetriever.StorageService,
	manualEpochStartNotifier dataRetriever.ManualEpochStartNotifier,
) (dataRetriever.ResolversContainerFactory, error) {
	dataPacker, err := partitioning.NewSimpleDataPacker(pcf.coreData.InternalMarshalizer())
	if err != nil {
		return nil, err
	}

	resolversContainerFactoryArgs := storageResolversContainers.FactoryArgs{
		GeneralConfig:            pcf.config,
		ShardIDForTries:          pcf.importDBConfig.ImportDBTargetShardID,
		ChainID:                  pcf.coreData.ChainID(),
		WorkingDirectory:         pcf.importDBConfig.ImportDBWorkingDir,
		Hasher:                   pcf.coreData.Hasher(),
		ShardCoordinator:         pcf.bootstrapComponents.ShardCoordinator(),
		Messenger:                pcf.network.NetworkMessenger(),
		Store:                    store,
		Marshalizer:              pcf.coreData.InternalMarshalizer(),
		Uint64ByteSliceConverter: pcf.coreData.Uint64ByteSliceConverter(),
		DataPacker:               dataPacker,
		ManualEpochStartNotifier: manualEpochStartNotifier,
		ChanGracefullyClose:      pcf.coreData.ChanStopNodeProcess(),
	}
	resolversContainerFactory, err := storageResolversContainers.NewShardResolversContainerFactory(resolversContainerFactoryArgs)
	if err != nil {
		return nil, err
	}

	return resolversContainerFactory, nil
}

func (pcf *processComponentsFactory) newShardInterceptorContainerFactory(
	headerSigVerifier process.InterceptedHeaderSigVerifier,
	headerIntegrityVerifier nodeFactory.HeaderIntegrityVerifierHandler,
	validityAttester process.ValidityAttester,
	epochStartTrigger process.EpochStartTriggerHandler,
	requestHandler process.RequestHandler,
	peerShardMapper *networksharding.PeerShardMapper,
	hardforkTrigger factory.HardforkTrigger,
) (process.InterceptorsContainerFactory, process.TimeCacher, error) {
	headerBlackList := cache.NewTimeCache(timeSpanForBadHeaders)
	shardInterceptorsContainerFactoryArgs := interceptorscontainer.CommonInterceptorsContainerFactoryArgs{
		CoreComponents:               pcf.coreData,
		CryptoComponents:             pcf.crypto,
		Accounts:                     pcf.state.AccountsAdapter(),
		ShardCoordinator:             pcf.bootstrapComponents.ShardCoordinator(),
		NodesCoordinator:             pcf.nodesCoordinator,
		Messenger:                    pcf.network.NetworkMessenger(),
		Store:                        pcf.data.StorageService(),
		DataPool:                     pcf.data.Datapool(),
		MaxTxNonceDeltaAllowed:       common.MaxTxNonceDeltaAllowed,
		TxFeeHandler:                 pcf.coreData.EconomicsData(),
		BlockBlackList:               headerBlackList,
		HeaderSigVerifier:            headerSigVerifier,
		HeaderIntegrityVerifier:      headerIntegrityVerifier,
		ValidityAttester:             validityAttester,
		EpochStartTrigger:            epochStartTrigger,
		WhiteListHandler:             pcf.whiteListHandler,
		WhiteListerVerifiedTxs:       pcf.whiteListerVerifiedTxs,
		AntifloodHandler:             pcf.network.InputAntiFloodHandler(),
		ArgumentsParser:              smartContract.NewArgumentParser(),
		PreferredPeersHolder:         pcf.network.PreferredPeersHolderHandler(),
		SizeCheckDelta:               pcf.config.Marshalizer.SizeCheckDelta,
		RequestHandler:               requestHandler,
		PeerSignatureHandler:         pcf.crypto.PeerSignatureHandler(),
		SignaturesHandler:            pcf.network.NetworkMessenger(),
		HeartbeatExpiryTimespanInSec: pcf.config.HeartbeatV2.HeartbeatExpiryTimespanInSec,
		PeerShardMapper:              peerShardMapper,
		HardforkTrigger:              hardforkTrigger,
	}

	interceptorContainerFactory, err := interceptorscontainer.NewShardInterceptorsContainerFactory(shardInterceptorsContainerFactoryArgs)
	if err != nil {
		return nil, nil, err
	}

	return interceptorContainerFactory, headerBlackList, nil
}

func (pcf *processComponentsFactory) newMetaInterceptorContainerFactory(
	headerSigVerifier process.InterceptedHeaderSigVerifier,
	headerIntegrityVerifier nodeFactory.HeaderIntegrityVerifierHandler,
	validityAttester process.ValidityAttester,
	epochStartTrigger process.EpochStartTriggerHandler,
	requestHandler process.RequestHandler,
	peerShardMapper *networksharding.PeerShardMapper,
	hardforkTrigger factory.HardforkTrigger,
) (process.InterceptorsContainerFactory, process.TimeCacher, error) {
	headerBlackList := cache.NewTimeCache(timeSpanForBadHeaders)
	metaInterceptorsContainerFactoryArgs := interceptorscontainer.CommonInterceptorsContainerFactoryArgs{
		CoreComponents:               pcf.coreData,
		CryptoComponents:             pcf.crypto,
		ShardCoordinator:             pcf.bootstrapComponents.ShardCoordinator(),
		NodesCoordinator:             pcf.nodesCoordinator,
		Messenger:                    pcf.network.NetworkMessenger(),
		Store:                        pcf.data.StorageService(),
		DataPool:                     pcf.data.Datapool(),
		Accounts:                     pcf.state.AccountsAdapter(),
		MaxTxNonceDeltaAllowed:       common.MaxTxNonceDeltaAllowed,
		TxFeeHandler:                 pcf.coreData.EconomicsData(),
		BlockBlackList:               headerBlackList,
		HeaderSigVerifier:            headerSigVerifier,
		HeaderIntegrityVerifier:      headerIntegrityVerifier,
		ValidityAttester:             validityAttester,
		EpochStartTrigger:            epochStartTrigger,
		WhiteListHandler:             pcf.whiteListHandler,
		WhiteListerVerifiedTxs:       pcf.whiteListerVerifiedTxs,
		AntifloodHandler:             pcf.network.InputAntiFloodHandler(),
		ArgumentsParser:              smartContract.NewArgumentParser(),
		SizeCheckDelta:               pcf.config.Marshalizer.SizeCheckDelta,
		PreferredPeersHolder:         pcf.network.PreferredPeersHolderHandler(),
		RequestHandler:               requestHandler,
		PeerSignatureHandler:         pcf.crypto.PeerSignatureHandler(),
		SignaturesHandler:            pcf.network.NetworkMessenger(),
		HeartbeatExpiryTimespanInSec: pcf.config.HeartbeatV2.HeartbeatExpiryTimespanInSec,
		PeerShardMapper:              peerShardMapper,
		HardforkTrigger:              hardforkTrigger,
	}

	interceptorContainerFactory, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(metaInterceptorsContainerFactoryArgs)
	if err != nil {
		return nil, nil, err
	}

	return interceptorContainerFactory, headerBlackList, nil
}

func (pcf *processComponentsFactory) newForkDetector(
	headerBlackList process.TimeCacher,
	blockTracker process.BlockTracker,
) (process.ForkDetector, error) {
	if pcf.bootstrapComponents.ShardCoordinator().SelfId() < pcf.bootstrapComponents.ShardCoordinator().NumberOfShards() {
		return sync.NewShardForkDetector(pcf.coreData.RoundHandler(), headerBlackList, blockTracker, pcf.coreData.GenesisNodesSetup().GetStartTime())
	}
	if pcf.bootstrapComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		return sync.NewMetaForkDetector(pcf.coreData.RoundHandler(), headerBlackList, blockTracker, pcf.coreData.GenesisNodesSetup().GetStartTime())
	}

	return nil, errors.New("could not create fork detector")
}

// PrepareNetworkShardingCollector will create the network sharding collector and apply it to the network messenger
func (pcf *processComponentsFactory) prepareNetworkShardingCollector() (*networksharding.PeerShardMapper, error) {
	networkShardingCollector, err := createNetworkShardingCollector(
		&pcf.config,
		pcf.nodesCoordinator,
		pcf.network.PreferredPeersHolderHandler(),
	)
	if err != nil {
		return nil, err
	}

	localID := pcf.network.NetworkMessenger().ID()
	networkShardingCollector.UpdatePeerIDInfo(localID, pcf.crypto.PublicKeyBytes(), pcf.bootstrapComponents.ShardCoordinator().SelfId())

	err = pcf.network.NetworkMessenger().SetPeerShardResolver(networkShardingCollector)
	if err != nil {
		return nil, err
	}

	err = pcf.network.InputAntiFloodHandler().SetPeerValidatorMapper(networkShardingCollector)
	if err != nil {
		return nil, err
	}

	return networkShardingCollector, nil
}

func (pcf *processComponentsFactory) createExportFactoryHandler(
	headerValidator epochStart.HeaderValidator,
	requestHandler process.RequestHandler,
	resolversFinder dataRetriever.ResolversFinder,
	interceptorsContainer process.InterceptorsContainer,
	headerSigVerifier process.InterceptedHeaderSigVerifier,
	blockTracker process.ValidityAttester,
) (update.ExportFactoryHandler, error) {

	hardforkConfig := pcf.config.Hardfork
	accountsDBs := make(map[state.AccountsDbIdentifier]state.AccountsAdapter)
	accountsDBs[state.UserAccountsState] = pcf.state.AccountsAdapter()
	accountsDBs[state.PeerAccountsState] = pcf.state.PeerAccounts()
	exportFolder := filepath.Join(pcf.workingDir, hardforkConfig.ImportFolder)
	argsExporter := updateFactory.ArgsExporter{
		CoreComponents:            pcf.coreData,
		CryptoComponents:          pcf.crypto,
		StatusCoreComponents:      pcf.statusCoreComponents,
		HeaderValidator:           headerValidator,
		DataPool:                  pcf.data.Datapool(),
		StorageService:            pcf.data.StorageService(),
		RequestHandler:            requestHandler,
		ShardCoordinator:          pcf.bootstrapComponents.ShardCoordinator(),
		Messenger:                 pcf.network.NetworkMessenger(),
		ActiveAccountsDBs:         accountsDBs,
		ExistingResolvers:         resolversFinder,
		ExportFolder:              exportFolder,
		ExportTriesStorageConfig:  hardforkConfig.ExportTriesStorageConfig,
		ExportStateStorageConfig:  hardforkConfig.ExportStateStorageConfig,
		ExportStateKeysConfig:     hardforkConfig.ExportKeysStorageConfig,
		MaxTrieLevelInMemory:      pcf.config.StateTriesConfig.MaxStateTrieLevelInMemory,
		WhiteListHandler:          pcf.whiteListHandler,
		WhiteListerVerifiedTxs:    pcf.whiteListerVerifiedTxs,
		InterceptorsContainer:     interceptorsContainer,
		NodesCoordinator:          pcf.nodesCoordinator,
		HeaderSigVerifier:         headerSigVerifier,
		HeaderIntegrityVerifier:   pcf.bootstrapComponents.HeaderIntegrityVerifier(),
		ValidityAttester:          blockTracker,
		InputAntifloodHandler:     pcf.network.InputAntiFloodHandler(),
		OutputAntifloodHandler:    pcf.network.OutputAntiFloodHandler(),
		RoundHandler:              pcf.coreData.RoundHandler(),
		InterceptorDebugConfig:    pcf.config.Debug.InterceptorResolver,
		MaxHardCapForMissingNodes: pcf.config.TrieSync.MaxHardCapForMissingNodes,
		NumConcurrentTrieSyncers:  pcf.config.TrieSync.NumConcurrentTrieSyncers,
		TrieSyncerVersion:         pcf.config.TrieSync.TrieSyncerVersion,
		PeersRatingHandler:        pcf.network.PeersRatingHandler(),
	}
	return updateFactory.NewExportHandlerFactory(argsExporter)
}

func (pcf *processComponentsFactory) createHardforkTrigger(epochStartTrigger update.EpochHandler) (factory.HardforkTrigger, error) {
	hardforkConfig := pcf.config.Hardfork
	selfPubKeyBytes := pcf.crypto.PublicKeyBytes()
	triggerPubKeyBytes, err := pcf.coreData.ValidatorPubKeyConverter().Decode(hardforkConfig.PublicKeyToListenFrom)
	if err != nil {
		return nil, fmt.Errorf("%w while decoding HardforkConfig.PublicKeyToListenFrom", err)
	}

	argTrigger := trigger.ArgHardforkTrigger{
		TriggerPubKeyBytes:        triggerPubKeyBytes,
		SelfPubKeyBytes:           selfPubKeyBytes,
		Enabled:                   hardforkConfig.EnableTrigger,
		EnabledAuthenticated:      hardforkConfig.EnableTriggerFromP2P,
		ArgumentParser:            smartContract.NewArgumentParser(),
		EpochProvider:             epochStartTrigger,
		ExportFactoryHandler:      &updateDisabled.ExportFactoryHandler{},
		ChanStopNodeProcess:       pcf.coreData.ChanStopNodeProcess(),
		EpochConfirmedNotifier:    pcf.coreData.EpochStartNotifierWithConfirm(),
		CloseAfterExportInMinutes: hardforkConfig.CloseAfterExportInMinutes,
		ImportStartHandler:        pcf.importStartHandler,
		RoundHandler:              pcf.coreData.RoundHandler(),
	}

	return trigger.NewTrigger(argTrigger)
}

func createNetworkShardingCollector(
	config *config.Config,
	nodesCoordinator nodesCoordinator.NodesCoordinator,
	preferredPeersHolder factory.PreferredPeersHolderHandler,
) (*networksharding.PeerShardMapper, error) {

	cacheConfig := config.PublicKeyPeerId
	cachePkPid, err := createCache(cacheConfig)
	if err != nil {
		return nil, err
	}

	cacheConfig = config.PublicKeyShardId
	cachePkShardID, err := createCache(cacheConfig)
	if err != nil {
		return nil, err
	}

	cacheConfig = config.PeerIdShardId
	cachePidShardID, err := createCache(cacheConfig)
	if err != nil {
		return nil, err
	}

	arg := networksharding.ArgPeerShardMapper{
		PeerIdPkCache:         cachePkPid,
		FallbackPkShardCache:  cachePkShardID,
		FallbackPidShardCache: cachePidShardID,
		NodesCoordinator:      nodesCoordinator,
		PreferredPeersHolder:  preferredPeersHolder,
	}
	psm, err := networksharding.NewPeerShardMapper(arg)
	if err != nil {
		return nil, err
	}

	return psm, nil
}

func createCache(cacheConfig config.CacheConfig) (storage.Cacher, error) {
	return storageunit.NewCache(storageFactory.GetCacherFromConfig(cacheConfig))
}

func checkProcessComponentsArgs(args ProcessComponentsFactoryArgs) error {
	baseErrMessage := "error creating process components"
	if check.IfNil(args.AccountsParser) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilAccountsParser)
	}
	if check.IfNil(args.SmartContractParser) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilSmartContractParser)
	}
	if args.GasSchedule == nil {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilGasSchedule)
	}
	if check.IfNil(args.NodesCoordinator) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilNodesCoordinator)
	}
	if check.IfNil(args.Data) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilDataComponentsHolder)
	}
	if check.IfNil(args.CoreData) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilCoreComponentsHolder)
	}
	if args.CoreData.EconomicsData() == nil {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilEconomicsData)
	}
	if check.IfNil(args.CoreData.RoundHandler()) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilRoundHandler)
	}
	if check.IfNil(args.Crypto) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilCryptoComponentsHolder)
	}
	if check.IfNil(args.State) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilStateComponentsHolder)
	}
	if check.IfNil(args.Network) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilNetworkComponentsHolder)
	}
	if check.IfNil(args.RequestedItemsHandler) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilRequestedItemHandler)
	}
	if check.IfNil(args.WhiteListHandler) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilWhiteListHandler)
	}
	if check.IfNil(args.WhiteListerVerifiedTxs) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilWhiteListVerifiedTxs)
	}
	if check.IfNil(args.CoreData.EpochStartNotifierWithConfirm()) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilEpochStartNotifier)
	}
	if check.IfNil(args.CoreData.Rater()) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilRater)
	}
	if check.IfNil(args.CoreData.RatingsData()) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilRatingData)
	}
	if check.IfNil(args.CoreData.ValidatorPubKeyConverter()) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilPubKeyConverter)
	}
	if args.SystemSCConfig == nil {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilSystemSCConfig)
	}
	if check.IfNil(args.CoreData.EpochNotifier()) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilEpochNotifier)
	}
	if check.IfNil(args.CoreData.EnableEpochsHandler()) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilEnableEpochsHandler)
	}
	if check.IfNil(args.BootstrapComponents) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilBootstrapComponentsHolder)
	}
	if check.IfNil(args.BootstrapComponents.ShardCoordinator()) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilShardCoordinator)
	}
	if check.IfNil(args.StatusComponents) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilStatusComponentsHolder)
	}
	if check.IfNil(args.StatusCoreComponents) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilStatusCoreComponents)
	}
	if check.IfNil(args.StatusCoreComponents.AppStatusHandler()) {
		return fmt.Errorf("%s: %w", baseErrMessage, errErd.ErrNilAppStatusHandler)
	}

	return nil
}

// Close closes all underlying components that need closing
func (pc *processComponents) Close() error {
	if !check.IfNil(pc.blockProcessor) {
		log.LogIfError(pc.blockProcessor.Close())
	}
	if !check.IfNil(pc.validatorsProvider) {
		log.LogIfError(pc.validatorsProvider.Close())
	}
	if !check.IfNil(pc.miniBlocksPoolCleaner) {
		log.LogIfError(pc.miniBlocksPoolCleaner.Close())
	}
	if !check.IfNil(pc.txsPoolCleaner) {
		log.LogIfError(pc.txsPoolCleaner.Close())
	}
	if !check.IfNil(pc.epochStartTrigger) {
		log.LogIfError(pc.epochStartTrigger.Close())
	}
	if !check.IfNil(pc.importHandler) {
		log.LogIfError(pc.importHandler.Close())
	}
	if !check.IfNil(pc.interceptorsContainer) {
		log.LogIfError(pc.interceptorsContainer.Close())
	}
	if !check.IfNil(pc.vmFactoryForTxSimulator) {
		log.LogIfError(pc.vmFactoryForTxSimulator.Close())
	}
	if !check.IfNil(pc.vmFactoryForProcessing) {
		log.LogIfError(pc.vmFactoryForProcessing.Close())
	}
	if !check.IfNil(pc.txsSender) {
		log.LogIfError(pc.txsSender.Close())
	}

	return nil
}

func unwrapTxs(txs map[string]data.TransactionHandlerWithGasUsedAndFee) map[string]data.TransactionHandler {
	output := make(map[string]data.TransactionHandler)
	for hash, wrappedTx := range txs {
		output[hash] = wrappedTx.GetTxHandler()
	}

	return output
}
