package factory

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/accumulator"
	"github.com/ElrondNetwork/elrond-go/core/dblookupext"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/facade"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/nodeDebugFactory"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	factory3 "github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/blackList"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
	factory2 "github.com/ElrondNetwork/elrond-go/update/factory"
	"github.com/ElrondNetwork/elrond-go/update/trigger"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
)

func CreateHardForkTrigger(
	config *config.Config,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	coreData CoreComponentsHolder,
	stateComponents StateComponentsHolder,
	data DataComponentsHolder,
	crypto CryptoComponentsHolder,
	process ProcessComponentsHolder,
	network NetworkComponentsHolder,
	whiteListRequest process.WhiteListHandler,
	whiteListerVerifiedTxs process.WhiteListHandler,
	chanStopNodeProcess chan endProcess.ArgEndProcess,
	epochNotifier factory.EpochStartNotifier,
	importStartHandler update.ImportStartHandler,
	nodesSetup update.GenesisNodesSetupHandler,
	workingDir string,
) (node.HardforkTrigger, error) {

	selfPubKeyBytes := crypto.PublicKeyBytes()
	triggerPubKeyBytes, err := coreData.ValidatorPubKeyConverter().Decode(config.Hardfork.PublicKeyToListenFrom)
	if err != nil {
		return nil, fmt.Errorf("%w while decoding HardforkConfig.PublicKeyToListenFrom", err)
	}

	accountsDBs := make(map[state.AccountsDbIdentifier]state.AccountsAdapter)
	accountsDBs[state.UserAccountsState] = stateComponents.AccountsAdapter()
	accountsDBs[state.PeerAccountsState] = stateComponents.PeerAccounts()
	hardForkConfig := config.Hardfork
	exportFolder := filepath.Join(workingDir, hardForkConfig.ImportFolder)
	argsExporter := factory2.ArgsExporter{
		CoreComponents:           coreData,
		CryptoComponents:         crypto,
		HeaderValidator:          process.HeaderConstructionValidator(),
		DataPool:                 data.Datapool(),
		StorageService:           data.StorageService(),
		RequestHandler:           process.RequestHandler(),
		ShardCoordinator:         shardCoordinator,
		Messenger:                network.NetworkMessenger(),
		ActiveAccountsDBs:        accountsDBs,
		ExistingResolvers:        process.ResolversFinder(),
		ExportFolder:             exportFolder,
		ExportTriesStorageConfig: hardForkConfig.ExportTriesStorageConfig,
		ExportStateStorageConfig: hardForkConfig.ExportStateStorageConfig,
		ExportStateKeysConfig:    hardForkConfig.ExportKeysStorageConfig,
		WhiteListHandler:         whiteListRequest,
		WhiteListerVerifiedTxs:   whiteListerVerifiedTxs,
		InterceptorsContainer:    process.InterceptorsContainer(),
		NodesCoordinator:         nodesCoordinator,
		HeaderSigVerifier:        process.HeaderSigVerifier(),
		HeaderIntegrityVerifier:  process.HeaderIntegrityVerifier(),
		MaxTrieLevelInMemory:     config.StateTriesConfig.MaxStateTrieLevelInMemory,
		InputAntifloodHandler:    network.InputAntiFloodHandler(),
		OutputAntifloodHandler:   network.OutputAntiFloodHandler(),
		ValidityAttester:         process.BlockTracker(),
		Rounder:                  process.Rounder(),
		GenesisNodesSetupHandler: nodesSetup,
	}
	hardForkExportFactory, err := factory2.NewExportHandlerFactory(argsExporter)
	if err != nil {
		return nil, err
	}

	atArgumentParser := smartContract.NewArgumentParser()
	argTrigger := trigger.ArgHardforkTrigger{
		TriggerPubKeyBytes:        triggerPubKeyBytes,
		SelfPubKeyBytes:           selfPubKeyBytes,
		Enabled:                   config.Hardfork.EnableTrigger,
		EnabledAuthenticated:      config.Hardfork.EnableTriggerFromP2P,
		ArgumentParser:            atArgumentParser,
		EpochProvider:             process.EpochStartTrigger(),
		ExportFactoryHandler:      hardForkExportFactory,
		ChanStopNodeProcess:       chanStopNodeProcess,
		EpochConfirmedNotifier:    epochNotifier,
		CloseAfterExportInMinutes: config.Hardfork.CloseAfterExportInMinutes,
		ImportStartHandler:        importStartHandler,
	}
	hardforkTrigger, err := trigger.NewTrigger(argTrigger)
	if err != nil {
		return nil, err
	}

	return hardforkTrigger, nil
}

func CreateApiResolver(
	config *config.Config,
	accnts state.AccountsAdapter,
	validatorAccounts state.AccountsAdapter,
	pubkeyConv core.PubkeyConverter,
	storageService dataRetriever.StorageService,
	blockChain data.ChainHandler,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
	shardCoordinator sharding.Coordinator,
	statusMetrics external.StatusMetricsHandler,
	gasSchedule map[string]map[string]uint64,
	economics *economics.EconomicsData,
	messageSigVerifier vm.MessageSignVerifier,
	nodesSetup sharding.GenesisNodesSetupHandler,
	systemSCConfig *config.SystemSmartContractsConfig,
) (facade.ApiResolver, error) {
	var vmFactory process.VirtualMachinesContainerFactory
	var err error

	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasMap:          gasSchedule,
		MapDNSAddresses: make(map[string]struct{}),
		Marshalizer:     marshalizer,
	}
	builtInFuncs, err := builtInFunctions.CreateBuiltInFunctionContainer(argsBuiltIn)
	if err != nil {
		return nil, err
	}

	argsHook := hooks.ArgBlockChainHook{
		Accounts:         accnts,
		PubkeyConv:       pubkeyConv,
		StorageService:   storageService,
		BlockChain:       blockChain,
		ShardCoordinator: shardCoordinator,
		Marshalizer:      marshalizer,
		Uint64Converter:  uint64Converter,
		BuiltInFunctions: builtInFuncs,
	}

	if shardCoordinator.SelfId() == core.MetachainShardId {
		vmFactory, err = metachain.NewVMContainerFactory(
			argsHook,
			economics,
			messageSigVerifier,
			gasSchedule,
			nodesSetup,
			hasher,
			marshalizer,
			systemSCConfig,
			validatorAccounts,
		)
		if err != nil {
			return nil, err
		}
	} else {
		vmFactory, err = shard.NewVMContainerFactory(
			config.VirtualMachineConfig,
			economics.MaxGasLimitPerBlock(shardCoordinator.SelfId()),
			gasSchedule,
			argsHook)
		if err != nil {
			return nil, err
		}
	}

	vmContainer, err := vmFactory.Create()
	if err != nil {
		return nil, err
	}

	scQueryService, err := smartContract.NewSCQueryService(vmContainer, economics)
	if err != nil {
		return nil, err
	}

	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:  pubkeyConv,
		ShardCoordinator: shardCoordinator,
		BuiltInFuncNames: builtInFuncs.Keys(),
		ArgumentParser:   parsers.NewCallArgsParser(),
	}
	txTypeHandler, err := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	if err != nil {
		return nil, err
	}

	txCostHandler, err := transaction.NewTransactionCostEstimator(txTypeHandler, economics, scQueryService, gasSchedule)
	if err != nil {
		return nil, err
	}

	return external.NewNodeApiResolver(scQueryService, statusMetrics, txCostHandler)
}

// prepareOpenTopics will set to the anti flood handler the topics for which
// the node can receive messages from others than validators
func prepareOpenTopics(
	antiflood P2PAntifloodHandler,
	shardCoordinator sharding.Coordinator,
) {
	selfID := shardCoordinator.SelfId()
	if selfID == core.MetachainShardId {
		antiflood.SetTopicsForAll(core.HeartbeatTopic)
		return
	}

	selfShardTxTopic := factory3.TransactionTopic + core.CommunicationIdentifierBetweenShards(selfID, selfID)
	antiflood.SetTopicsForAll(core.HeartbeatTopic, selfShardTxTopic)
}

func getConsensusGroupSize(nodesConfig NodesSetupHandler, shardCoordinator sharding.Coordinator) (uint32, error) {
	if shardCoordinator.SelfId() == core.MetachainShardId {
		return nodesConfig.GetMetaConsensusGroupSize(), nil
	}
	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return nodesConfig.GetShardConsensusGroupSize(), nil
	}

	return 0, state.ErrUnknownShardId
}

func CreateNode(
	config *config.Config,
	preferencesConfig *config.Preferences,
	nodesConfig NodesSetupHandler,
	bootstrapComponents BootstrapComponentsHandler,
	coreComponents CoreComponentsHandler,
	cryptoComponents CryptoComponentsHandler,
	dataComponents DataComponentsHandler,
	networkComponents NetworkComponentsHandler,
	processComponents ProcessComponentsHandler,
	stateComponents StateComponentsHandler,
	statusComponents StatusComponentsHandler,
	bootstrapRoundIndex uint64,
	version string,
	requestedItemsHandler dataRetriever.RequestedItemsHandler,
	whiteListRequest process.WhiteListHandler,
	whiteListerVerifiedTxs process.WhiteListHandler,
	chanStopNodeProcess chan endProcess.ArgEndProcess,
	hardForkTrigger node.HardforkTrigger,
	historyRepository dblookupext.HistoryRepository,
) (*node.Node, error) {
	var err error
	var consensusGroupSize uint32
	consensusGroupSize, err = getConsensusGroupSize(nodesConfig, processComponents.ShardCoordinator())
	if err != nil {
		return nil, err
	}

	var txAccumulator node.Accumulator
	txAccumulatorConfig := config.Antiflood.TxAccumulator
	txAccumulator, err = accumulator.NewTimeAccumulator(
		time.Duration(txAccumulatorConfig.MaxAllowedTimeInMilliseconds)*time.Millisecond,
		time.Duration(txAccumulatorConfig.MaxDeviationTimeInMilliseconds)*time.Millisecond,
	)
	if err != nil {
		return nil, err
	}

	prepareOpenTopics(networkComponents.InputAntiFloodHandler(), processComponents.ShardCoordinator())

	peerDenialEvaluator, err := blackList.NewPeerDenialEvaluator(
		networkComponents.PeerBlackListHandler(),
		networkComponents.PubKeyCacher(),
		processComponents.PeerShardMapper(),
	)
	if err != nil {
		return nil, err
	}

	err = networkComponents.NetworkMessenger().SetPeerDenialEvaluator(peerDenialEvaluator)
	if err != nil {
		return nil, err
	}

	genesisTime := time.Unix(nodesConfig.GetStartTime(), 0)
	heartbeatArgs := HeartbeatComponentsFactoryArgs{
		Config:            *config,
		Prefs:             *preferencesConfig,
		AppVersion:        version,
		GenesisTime:       genesisTime,
		HardforkTrigger:   hardForkTrigger,
		CoreComponents:    coreComponents,
		DataComponents:    dataComponents,
		NetworkComponents: networkComponents,
		CryptoComponents:  cryptoComponents,
		ProcessComponents: processComponents,
	}

	heartbeatComponentsFactory, err := NewHeartbeatComponentsFactory(heartbeatArgs)
	if err != nil {
		return nil, fmt.Errorf("NewHeartbeatComponentsFactory failed: %w", err)
	}

	managedHeartbeatComponents, err := NewManagedHeartbeatComponents(heartbeatComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedHeartbeatComponents.Create()
	if err != nil {
		return nil, err
	}

	var nd *node.Node
	nd, err = node.NewNode(
		node.WithBootstrapComponents(bootstrapComponents),
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithNetworkComponents(networkComponents),
		node.WithProcessComponents(processComponents),
		node.WithCryptoComponents(cryptoComponents),
		node.WithStateComponents(stateComponents),
		node.WithStatusComponents(statusComponents),
		node.WithInitialNodesPubKeys(nodesConfig.InitialNodesPubKeys()),
		node.WithRoundDuration(nodesConfig.GetRoundDuration()),
		node.WithConsensusGroupSize(int(consensusGroupSize)),
		node.WithGenesisTime(genesisTime),
		node.WithConsensusType(config.Consensus.Type),
		node.WithBootstrapRoundIndex(bootstrapRoundIndex),
		node.WithPeerDenialEvaluator(peerDenialEvaluator),
		node.WithRequestedItemsHandler(requestedItemsHandler),
		node.WithTxAccumulator(txAccumulator),
		node.WithHardforkTrigger(hardForkTrigger),
		node.WithWhiteListHandler(whiteListRequest),
		node.WithWhiteListHandlerVerified(whiteListerVerifiedTxs),
		node.WithSignatureSize(config.ValidatorPubkeyConverter.SignatureLength),
		node.WithPublicKeySize(config.ValidatorPubkeyConverter.Length),
		node.WithNodeStopChannel(chanStopNodeProcess),
		node.WithHistoryRepository(historyRepository),
	)
	if err != nil {
		return nil, errors.New("error creating node: " + err.Error())
	}

	if processComponents.ShardCoordinator().SelfId() < processComponents.ShardCoordinator().NumberOfShards() {
		err = nd.CreateShardedStores()
		if err != nil {
			return nil, err
		}
	}

	err = nodeDebugFactory.CreateInterceptedDebugHandler(
		nd,
		processComponents.InterceptorsContainer(),
		processComponents.ResolversFinder(),
		config.Debug.InterceptorResolver,
	)
	if err != nil {
		return nil, err
	}

	return nd, nil
}
