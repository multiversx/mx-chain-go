package api

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/disabled"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	"github.com/multiversx/mx-chain-go/facade"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/external/blockAPI"
	"github.com/multiversx/mx-chain-go/node/external/logs"
	"github.com/multiversx/mx-chain-go/node/external/timemachine/fee"
	"github.com/multiversx/mx-chain-go/node/external/transactionAPI"
	"github.com/multiversx/mx-chain-go/node/trieIterators"
	trieIteratorsFactory "github.com/multiversx/mx-chain-go/node/trieIterators/factory"
	"github.com/multiversx/mx-chain-go/outport/process/alteredaccounts"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/factory/metachain"
	"github.com/multiversx/mx-chain-go/process/factory/shard"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/process/smartContract/builtInFunctions"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks/counters"
	"github.com/multiversx/mx-chain-go/process/txstatus"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/blockInfoProviders"
	disabledState "github.com/multiversx/mx-chain-go/state/disabled"
	factoryState "github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager/evictionWaitingList"
	"github.com/multiversx/mx-chain-go/state/syncer"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	trieFactory "github.com/multiversx/mx-chain-go/trie/factory"
	"github.com/multiversx/mx-chain-go/vm"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	datafield "github.com/multiversx/mx-chain-vm-common-go/parsers/dataField"
)

var log = logger.GetOrCreate("factory")

// ApiResolverArgs holds the argument needed to create an API resolver
type ApiResolverArgs struct {
	Configs              *config.Configs
	CoreComponents       factory.CoreComponentsHolder
	DataComponents       factory.DataComponentsHolder
	StateComponents      factory.StateComponentsHolder
	BootstrapComponents  factory.BootstrapComponentsHolder
	CryptoComponents     factory.CryptoComponentsHolder
	ProcessComponents    factory.ProcessComponentsHolder
	StatusCoreComponents factory.StatusCoreComponentsHolder
	StatusComponents     factory.StatusComponentsHolder
	GasScheduleNotifier  common.GasScheduleNotifierAPI
	Bootstrapper         process.Bootstrapper
	AllowVMQueriesChan   chan struct{}
	ProcessingMode       common.NodeProcessingMode
}

type scQueryServiceArgs struct {
	generalConfig         *config.Config
	epochConfig           *config.EpochConfig
	coreComponents        factory.CoreComponentsHolder
	stateComponents       factory.StateComponentsHolder
	dataComponents        factory.DataComponentsHolder
	processComponents     factory.ProcessComponentsHolder
	statusCoreComponents  factory.StatusCoreComponentsHolder
	gasScheduleNotifier   core.GasScheduleNotifier
	messageSigVerifier    vm.MessageSignVerifier
	systemSCConfig        *config.SystemSmartContractsConfig
	bootstrapper          process.Bootstrapper
	guardedAccountHandler process.GuardedAccountHandler
	allowVMQueriesChan    chan struct{}
	workingDir            string
	processingMode        common.NodeProcessingMode
}

type scQueryElementArgs struct {
	generalConfig         *config.Config
	epochConfig           *config.EpochConfig
	coreComponents        factory.CoreComponentsHolder
	stateComponents       factory.StateComponentsHolder
	dataComponents        factory.DataComponentsHolder
	processComponents     factory.ProcessComponentsHolder
	statusCoreComponents  factory.StatusCoreComponentsHolder
	gasScheduleNotifier   core.GasScheduleNotifier
	messageSigVerifier    vm.MessageSignVerifier
	systemSCConfig        *config.SystemSmartContractsConfig
	bootstrapper          process.Bootstrapper
	guardedAccountHandler process.GuardedAccountHandler
	allowVMQueriesChan    chan struct{}
	workingDir            string
	index                 int
	processingMode        common.NodeProcessingMode
}

// CreateApiResolver is able to create an ApiResolver instance that will solve the REST API requests through the node facade
// TODO: refactor to further decrease node's codebase
func CreateApiResolver(args *ApiResolverArgs) (facade.ApiResolver, error) {
	apiWorkingDir := filepath.Join(args.Configs.FlagsConfig.WorkingDir, common.TemporaryPath)
	argsSCQuery := &scQueryServiceArgs{
		generalConfig:         args.Configs.GeneralConfig,
		epochConfig:           args.Configs.EpochConfig,
		coreComponents:        args.CoreComponents,
		dataComponents:        args.DataComponents,
		stateComponents:       args.StateComponents,
		processComponents:     args.ProcessComponents,
		statusCoreComponents:  args.StatusCoreComponents,
		gasScheduleNotifier:   args.GasScheduleNotifier,
		messageSigVerifier:    args.CryptoComponents.MessageSignVerifier(),
		systemSCConfig:        args.Configs.SystemSCConfig,
		bootstrapper:          args.Bootstrapper,
		guardedAccountHandler: args.BootstrapComponents.GuardedAccountHandler(),
		allowVMQueriesChan:    args.AllowVMQueriesChan,
		workingDir:            apiWorkingDir,
		processingMode:        args.ProcessingMode,
	}

	scQueryService, err := createScQueryService(argsSCQuery)
	if err != nil {
		return nil, err
	}

	pkConverter := args.CoreComponents.AddressPubKeyConverter()
	automaticCrawlerAddressesStrings := args.Configs.GeneralConfig.BuiltInFunctions.AutomaticCrawlerAddresses
	convertedAddresses, errDecode := factory.DecodeAddresses(pkConverter, automaticCrawlerAddressesStrings)
	if errDecode != nil {
		return nil, errDecode
	}

	dnsV2AddressesStrings := args.Configs.GeneralConfig.BuiltInFunctions.DNSV2Addresses
	convertedDNSV2Addresses, errDecode := factory.DecodeAddresses(pkConverter, dnsV2AddressesStrings)
	if errDecode != nil {
		return nil, errDecode
	}

	builtInFuncFactory, err := createBuiltinFuncs(
		args.GasScheduleNotifier,
		args.CoreComponents.InternalMarshalizer(),
		args.StateComponents.AccountsAdapterAPI(),
		args.BootstrapComponents.ShardCoordinator(),
		args.CoreComponents.EpochNotifier(),
		args.CoreComponents.EnableEpochsHandler(),
		args.BootstrapComponents.GuardedAccountHandler(),
		convertedAddresses,
		args.Configs.GeneralConfig.BuiltInFunctions.MaxNumAddressesInTransferRole,
		convertedDNSV2Addresses,
	)
	if err != nil {
		return nil, err
	}

	esdtTransferParser, err := parsers.NewESDTTransferParser(args.CoreComponents.InternalMarshalizer())
	if err != nil {
		return nil, err
	}

	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:     args.CoreComponents.AddressPubKeyConverter(),
		ShardCoordinator:    args.ProcessComponents.ShardCoordinator(),
		BuiltInFunctions:    builtInFuncFactory.BuiltInFunctionContainer(),
		ArgumentParser:      parsers.NewCallArgsParser(),
		ESDTTransferParser:  esdtTransferParser,
		EnableEpochsHandler: args.CoreComponents.EnableEpochsHandler(),
	}
	txTypeHandler, err := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	if err != nil {
		return nil, err
	}

	accountsWrapper := &trieIterators.AccountsWrapper{
		Mutex:           &sync.Mutex{},
		AccountsAdapter: args.StateComponents.AccountsAdapterAPI(),
	}

	argsProcessors := trieIterators.ArgTrieIteratorProcessor{
		ShardID:            args.BootstrapComponents.ShardCoordinator().SelfId(),
		Accounts:           accountsWrapper,
		PublicKeyConverter: args.CoreComponents.AddressPubKeyConverter(),
		QueryService:       scQueryService,
	}
	totalStakedValueHandler, err := trieIteratorsFactory.CreateTotalStakedValueHandler(argsProcessors)
	if err != nil {
		return nil, err
	}

	directStakedListHandler, err := trieIteratorsFactory.CreateDirectStakedListHandler(argsProcessors)
	if err != nil {
		return nil, err
	}

	delegatedListHandler, err := trieIteratorsFactory.CreateDelegatedListHandler(argsProcessors)
	if err != nil {
		return nil, err
	}

	feeComputer, err := fee.NewFeeComputer(args.CoreComponents.EconomicsData())
	if err != nil {
		return nil, err
	}

	logsFacade, err := createLogsFacade(args)
	if err != nil {
		return nil, err
	}

	argsDataFieldParser := &datafield.ArgsOperationDataFieldParser{
		AddressLength: args.CoreComponents.AddressPubKeyConverter().Len(),
		Marshalizer:   args.CoreComponents.InternalMarshalizer(),
	}
	dataFieldParser, err := datafield.NewOperationDataFieldParser(argsDataFieldParser)
	if err != nil {
		return nil, err
	}

	argsAPITransactionProc := &transactionAPI.ArgAPITransactionProcessor{
		RoundDuration:            args.CoreComponents.GenesisNodesSetup().GetRoundDuration(),
		GenesisTime:              args.CoreComponents.GenesisTime(),
		Marshalizer:              args.CoreComponents.InternalMarshalizer(),
		AddressPubKeyConverter:   args.CoreComponents.AddressPubKeyConverter(),
		ShardCoordinator:         args.ProcessComponents.ShardCoordinator(),
		HistoryRepository:        args.ProcessComponents.HistoryRepository(),
		StorageService:           args.DataComponents.StorageService(),
		DataPool:                 args.DataComponents.Datapool(),
		Uint64ByteSliceConverter: args.CoreComponents.Uint64ByteSliceConverter(),
		FeeComputer:              feeComputer,
		TxTypeHandler:            txTypeHandler,
		LogsFacade:               logsFacade,
		DataFieldParser:          dataFieldParser,
	}
	apiTransactionProcessor, err := transactionAPI.NewAPITransactionProcessor(argsAPITransactionProc)
	if err != nil {
		return nil, err
	}

	apiBlockProcessor, err := createAPIBlockProcessor(args, apiTransactionProcessor)
	if err != nil {
		return nil, err
	}

	apiInternalBlockProcessor, err := createAPIInternalBlockProcessor(args, apiTransactionProcessor)
	if err != nil {
		return nil, err
	}

	argsApiResolver := external.ArgNodeApiResolver{
		SCQueryService:           scQueryService,
		StatusMetricsHandler:     args.StatusCoreComponents.StatusMetrics(),
		APITransactionEvaluator:  args.ProcessComponents.APITransactionEvaluator(),
		TotalStakedValueHandler:  totalStakedValueHandler,
		DirectStakedListHandler:  directStakedListHandler,
		DelegatedListHandler:     delegatedListHandler,
		APITransactionHandler:    apiTransactionProcessor,
		APIBlockHandler:          apiBlockProcessor,
		APIInternalBlockHandler:  apiInternalBlockProcessor,
		GenesisNodesSetupHandler: args.CoreComponents.GenesisNodesSetup(),
		ValidatorPubKeyConverter: args.CoreComponents.ValidatorPubKeyConverter(),
		AccountsParser:           args.ProcessComponents.AccountsParser(),
		GasScheduleNotifier:      args.GasScheduleNotifier,
		ManagedPeersMonitor:      args.StatusComponents.ManagedPeersMonitor(),
		NodesCoordinator:         args.ProcessComponents.NodesCoordinator(),
	}

	return external.NewNodeApiResolver(argsApiResolver)
}

func createScQueryService(
	args *scQueryServiceArgs,
) (process.SCQueryService, error) {
	numConcurrentVms := args.generalConfig.VirtualMachine.Querying.NumConcurrentVMs
	if numConcurrentVms < 1 {
		return nil, fmt.Errorf("VirtualMachine.Querying.NumConcurrentVms should be a positive number more than 1")
	}

	argsQueryElem := &scQueryElementArgs{
		generalConfig:         args.generalConfig,
		epochConfig:           args.epochConfig,
		coreComponents:        args.coreComponents,
		stateComponents:       args.stateComponents,
		dataComponents:        args.dataComponents,
		processComponents:     args.processComponents,
		statusCoreComponents:  args.statusCoreComponents,
		gasScheduleNotifier:   args.gasScheduleNotifier,
		messageSigVerifier:    args.messageSigVerifier,
		systemSCConfig:        args.systemSCConfig,
		bootstrapper:          args.bootstrapper,
		guardedAccountHandler: args.guardedAccountHandler,
		allowVMQueriesChan:    args.allowVMQueriesChan,
		workingDir:            args.workingDir,
		index:                 0,
		processingMode:        args.processingMode,
	}

	var err error
	var scQueryService process.SCQueryService

	list := make([]process.SCQueryService, 0, numConcurrentVms)
	for i := 0; i < numConcurrentVms; i++ {
		argsQueryElem.index = i
		scQueryService, err = createScQueryElement(argsQueryElem)
		if err != nil {
			return nil, err
		}

		list = append(list, scQueryService)
	}

	sqQueryDispatcher, err := smartContract.NewScQueryServiceDispatcher(list)
	if err != nil {
		return nil, err
	}

	return sqQueryDispatcher, nil
}

func createScQueryElement(
	args *scQueryElementArgs,
) (process.SCQueryService, error) {
	var err error

	pkConverter := args.coreComponents.AddressPubKeyConverter()
	automaticCrawlerAddressesStrings := args.generalConfig.BuiltInFunctions.AutomaticCrawlerAddresses
	convertedAddresses, errDecode := factory.DecodeAddresses(pkConverter, automaticCrawlerAddressesStrings)
	if errDecode != nil {
		return nil, errDecode
	}

	dnsV2AddressesStrings := args.generalConfig.BuiltInFunctions.DNSV2Addresses
	convertedDNSV2Addresses, errDecode := factory.DecodeAddresses(pkConverter, dnsV2AddressesStrings)
	if errDecode != nil {
		return nil, errDecode
	}

	builtInFuncFactory, err := createBuiltinFuncs(
		args.gasScheduleNotifier,
		args.coreComponents.InternalMarshalizer(),
		args.stateComponents.AccountsAdapterAPI(),
		args.processComponents.ShardCoordinator(),
		args.coreComponents.EpochNotifier(),
		args.coreComponents.EnableEpochsHandler(),
		args.guardedAccountHandler,
		convertedAddresses,
		args.generalConfig.BuiltInFunctions.MaxNumAddressesInTransferRole,
		convertedDNSV2Addresses,
	)
	if err != nil {
		return nil, err
	}

	cacherCfg := storageFactory.GetCacherFromConfig(args.generalConfig.SmartContractDataPool)
	smartContractsCache, err := storageunit.NewCache(cacherCfg)
	if err != nil {
		return nil, err
	}

	scStorage := args.generalConfig.SmartContractsStorageForSCQuery
	scStorage.DB.FilePath += fmt.Sprintf("%d", args.index)
	argsHook := hooks.ArgBlockChainHook{
		PubkeyConv:               args.coreComponents.AddressPubKeyConverter(),
		StorageService:           args.dataComponents.StorageService(),
		ShardCoordinator:         args.processComponents.ShardCoordinator(),
		Marshalizer:              args.coreComponents.InternalMarshalizer(),
		Uint64Converter:          args.coreComponents.Uint64ByteSliceConverter(),
		BuiltInFunctions:         builtInFuncFactory.BuiltInFunctionContainer(),
		NFTStorageHandler:        builtInFuncFactory.NFTStorageHandler(),
		GlobalSettingsHandler:    builtInFuncFactory.ESDTGlobalSettingsHandler(),
		DataPool:                 args.dataComponents.Datapool(),
		ConfigSCStorage:          scStorage,
		CompiledSCPool:           smartContractsCache,
		WorkingDir:               args.workingDir,
		EpochNotifier:            args.coreComponents.EpochNotifier(),
		EnableEpochsHandler:      args.coreComponents.EnableEpochsHandler(),
		NilCompiledSCStore:       true,
		GasSchedule:              args.gasScheduleNotifier,
		Counter:                  counters.NewDisabledCounter(),
		MissingTrieNodesNotifier: syncer.NewMissingTrieNodesNotifier(),
		PersisterFactory:         args.coreComponents.PersisterFactory(),
	}

	var apiBlockchain data.ChainHandler
	var vmFactory process.VirtualMachinesContainerFactory
	maxGasForVmQueries := args.generalConfig.VirtualMachine.GasConfig.ShardMaxGasPerVmQuery
	if args.processComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		maxGasForVmQueries = args.generalConfig.VirtualMachine.GasConfig.MetaMaxGasPerVmQuery
		apiBlockchain, vmFactory, err = createMetaVmContainerFactory(args, argsHook)
	} else {
		apiBlockchain, vmFactory, err = createShardVmContainerFactory(args, argsHook)
	}
	if err != nil {
		return nil, err
	}

	log.Debug("maximum gas per VM Query", "value", maxGasForVmQueries)

	vmContainer, err := vmFactory.Create()
	if err != nil {
		return nil, err
	}

	err = vmFactory.BlockChainHookImpl().SetVMContainer(vmContainer)
	if err != nil {
		return nil, err
	}

	err = builtInFuncFactory.SetPayableHandler(vmFactory.BlockChainHookImpl())
	if err != nil {
		return nil, err
	}

	argsNewSCQueryService := smartContract.ArgsNewSCQueryService{
		VmContainer:              vmContainer,
		EconomicsFee:             args.coreComponents.EconomicsData(),
		BlockChainHook:           vmFactory.BlockChainHookImpl(),
		MainBlockChain:           args.dataComponents.Blockchain(),
		APIBlockChain:            apiBlockchain,
		WasmVMChangeLocker:       args.coreComponents.WasmVMChangeLocker(),
		Bootstrapper:             args.bootstrapper,
		AllowExternalQueriesChan: args.allowVMQueriesChan,
		MaxGasLimitPerQuery:      maxGasForVmQueries,
		HistoryRepository:        args.processComponents.HistoryRepository(),
		ShardCoordinator:         args.processComponents.ShardCoordinator(),
		StorageService:           args.dataComponents.StorageService(),
		Marshaller:               args.coreComponents.InternalMarshalizer(),
		Hasher:                   args.coreComponents.Hasher(),
		Uint64ByteSliceConverter: args.coreComponents.Uint64ByteSliceConverter(),
	}

	return smartContract.NewSCQueryService(argsNewSCQueryService)
}

func createMetaVmContainerFactory(args *scQueryElementArgs, argsHook hooks.ArgBlockChainHook) (data.ChainHandler, process.VirtualMachinesContainerFactory, error) {
	apiBlockchain, err := blockchain.NewMetaChain(disabled.NewAppStatusHandler())
	if err != nil {
		return nil, nil, err
	}

	accountsAdapterApi, err := createNewAccountsAdapterApi(args, apiBlockchain)
	if err != nil {
		return nil, nil, err
	}

	argsHook.BlockChain = apiBlockchain
	argsHook.Accounts = accountsAdapterApi

	blockChainHookImpl, errBlockChainHook := hooks.NewBlockChainHookImpl(argsHook)
	if errBlockChainHook != nil {
		return nil, nil, errBlockChainHook
	}

	argsNewVmFactory := metachain.ArgsNewVMContainerFactory{
		BlockChainHook:      blockChainHookImpl,
		PubkeyConv:          argsHook.PubkeyConv,
		Economics:           args.coreComponents.EconomicsData(),
		MessageSignVerifier: args.messageSigVerifier,
		GasSchedule:         args.gasScheduleNotifier,
		NodesConfigProvider: args.coreComponents.GenesisNodesSetup(),
		Hasher:              args.coreComponents.Hasher(),
		Marshalizer:         args.coreComponents.InternalMarshalizer(),
		SystemSCConfig:      args.systemSCConfig,
		ValidatorAccountsDB: args.stateComponents.PeerAccounts(),
		UserAccountsDB:      args.stateComponents.AccountsAdapterAPI(),
		ChanceComputer:      args.coreComponents.Rater(),
		ShardCoordinator:    args.processComponents.ShardCoordinator(),
		EnableEpochsHandler: args.coreComponents.EnableEpochsHandler(),
	}
	vmFactory, err := metachain.NewVMContainerFactory(argsNewVmFactory)
	if err != nil {
		return nil, nil, err
	}

	return apiBlockchain, vmFactory, nil
}

func createShardVmContainerFactory(args *scQueryElementArgs, argsHook hooks.ArgBlockChainHook) (data.ChainHandler, process.VirtualMachinesContainerFactory, error) {
	apiBlockchain, err := blockchain.NewBlockChain(disabled.NewAppStatusHandler())
	if err != nil {
		return nil, nil, err
	}

	accountsAdapterApi, err := createNewAccountsAdapterApi(args, apiBlockchain)
	if err != nil {
		return nil, nil, err
	}

	argsHook.BlockChain = apiBlockchain
	argsHook.Accounts = accountsAdapterApi

	queryVirtualMachineConfig := args.generalConfig.VirtualMachine.Querying.VirtualMachineConfig
	esdtTransferParser, errParser := parsers.NewESDTTransferParser(args.coreComponents.InternalMarshalizer())
	if errParser != nil {
		return nil, nil, errParser
	}

	blockChainHookImpl, errBlockChainHook := hooks.NewBlockChainHookImpl(argsHook)
	if errBlockChainHook != nil {
		return nil, nil, errBlockChainHook
	}

	argsNewVMFactory := shard.ArgVMContainerFactory{
		BlockChainHook:      blockChainHookImpl,
		BuiltInFunctions:    argsHook.BuiltInFunctions,
		Config:              queryVirtualMachineConfig,
		BlockGasLimit:       args.coreComponents.EconomicsData().MaxGasLimitPerBlock(args.processComponents.ShardCoordinator().SelfId()),
		GasSchedule:         args.gasScheduleNotifier,
		EpochNotifier:       args.coreComponents.EpochNotifier(),
		EnableEpochsHandler: args.coreComponents.EnableEpochsHandler(),
		WasmVMChangeLocker:  args.coreComponents.WasmVMChangeLocker(),
		ESDTTransferParser:  esdtTransferParser,
		Hasher:              args.coreComponents.Hasher(),
	}

	log.Debug("apiResolver: enable epoch for sc deploy", "epoch", args.epochConfig.EnableEpochs.SCDeployEnableEpoch)
	log.Debug("apiResolver: enable epoch for ahead of time gas usage", "epoch", args.epochConfig.EnableEpochs.AheadOfTimeGasUsageEnableEpoch)
	log.Debug("apiResolver: enable epoch for repair callback", "epoch", args.epochConfig.EnableEpochs.RepairCallbackEnableEpoch)

	vmFactory, err := shard.NewVMContainerFactory(argsNewVMFactory)
	if err != nil {
		return nil, nil, err
	}

	return apiBlockchain, vmFactory, nil
}

func createNewAccountsAdapterApi(args *scQueryElementArgs, chainHandler data.ChainHandler) (state.AccountsAdapterAPI, error) {
	argsAccCreator := factoryState.ArgsAccountCreator{
		Hasher:              args.coreComponents.Hasher(),
		Marshaller:          args.coreComponents.InternalMarshalizer(),
		EnableEpochsHandler: args.coreComponents.EnableEpochsHandler(),
	}
	accountFactory, err := factoryState.NewAccountCreator(argsAccCreator)
	if err != nil {
		return nil, err
	}

	storagePruning, err := newStoragePruningManager(args)
	if err != nil {
		return nil, err
	}
	storageService := args.dataComponents.StorageService()
	trieStorer, err := storageService.GetStorer(dataRetriever.UserAccountsUnit)
	if err != nil {
		return nil, err
	}

	trieFactoryArgs := trieFactory.TrieFactoryArgs{
		Marshalizer:              args.coreComponents.InternalMarshalizer(),
		Hasher:                   args.coreComponents.Hasher(),
		PathManager:              args.coreComponents.PathHandler(),
		TrieStorageManagerConfig: args.generalConfig.TrieStorageManagerConfig,
	}
	trFactory, err := trieFactory.NewTrieFactory(trieFactoryArgs)
	if err != nil {
		return nil, err
	}

	trieCreatorArgs := trieFactory.TrieCreateArgs{
		MainStorer:          trieStorer,
		PruningEnabled:      args.generalConfig.StateTriesConfig.AccountsStatePruningEnabled,
		MaxTrieLevelInMem:   args.generalConfig.StateTriesConfig.MaxStateTrieLevelInMemory,
		SnapshotsEnabled:    args.generalConfig.StateTriesConfig.SnapshotsEnabled,
		IdleProvider:        args.coreComponents.ProcessStatusHandler(),
		Identifier:          dataRetriever.UserAccountsUnit.String(),
		EnableEpochsHandler: args.coreComponents.EnableEpochsHandler(),
		StatsCollector:      args.statusCoreComponents.StateStatsHandler(),
	}
	_, merkleTrie, err := trFactory.Create(trieCreatorArgs)
	if err != nil {
		return nil, err
	}

	argsAPIAccountsDB := state.ArgsAccountsDB{
		Trie:                  merkleTrie,
		Hasher:                args.coreComponents.Hasher(),
		Marshaller:            args.coreComponents.InternalMarshalizer(),
		AccountFactory:        accountFactory,
		StoragePruningManager: storagePruning,
		AddressConverter:      args.coreComponents.AddressPubKeyConverter(),
		SnapshotsManager:      disabledState.NewDisabledSnapshotsManager(),
	}

	provider, err := blockInfoProviders.NewCurrentBlockInfo(chainHandler)
	if err != nil {
		return nil, err
	}

	accounts, err := state.NewAccountsDB(argsAPIAccountsDB)
	if err != nil {
		return nil, err
	}

	return state.NewAccountsDBApi(accounts, provider)
}

func newStoragePruningManager(args *scQueryElementArgs) (state.StoragePruningManager, error) {
	argsMemEviction := evictionWaitingList.MemoryEvictionWaitingListArgs{
		RootHashesSize: args.generalConfig.EvictionWaitingList.RootHashesSize,
		HashesSize:     args.generalConfig.EvictionWaitingList.HashesSize,
	}
	trieEvictionWaitingList, err := evictionWaitingList.NewMemoryEvictionWaitingList(argsMemEviction)
	if err != nil {
		return nil, err
	}

	storagePruning, err := storagePruningManager.NewStoragePruningManager(
		trieEvictionWaitingList,
		args.generalConfig.TrieStorageManagerConfig.PruningBufferLen,
	)
	if err != nil {
		return nil, err
	}

	return storagePruning, nil
}

func createBuiltinFuncs(
	gasScheduleNotifier core.GasScheduleNotifier,
	marshalizer marshal.Marshalizer,
	accnts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	epochNotifier vmcommon.EpochNotifier,
	enableEpochsHandler vmcommon.EnableEpochsHandler,
	guardedAccountHandler vmcommon.GuardedAccountHandler,
	automaticCrawlerAddresses [][]byte,
	maxNumAddressesInTransferRole uint32,
	dnsV2Addresses [][]byte,
) (vmcommon.BuiltInFunctionFactory, error) {
	mapDNSV2Addresses := make(map[string]struct{})
	for _, address := range dnsV2Addresses {
		mapDNSV2Addresses[string(address)] = struct{}{}
	}

	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasSchedule:               gasScheduleNotifier,
		MapDNSAddresses:           make(map[string]struct{}),
		MapDNSV2Addresses:         mapDNSV2Addresses,
		Marshalizer:               marshalizer,
		Accounts:                  accnts,
		ShardCoordinator:          shardCoordinator,
		EpochNotifier:             epochNotifier,
		EnableEpochsHandler:       enableEpochsHandler,
		GuardedAccountHandler:     guardedAccountHandler,
		AutomaticCrawlerAddresses: automaticCrawlerAddresses,
		MaxNumNodesInTransferRole: maxNumAddressesInTransferRole,
	}
	return builtInFunctions.CreateBuiltInFunctionsFactory(argsBuiltIn)
}

func createAPIBlockProcessor(args *ApiResolverArgs, apiTransactionHandler external.APITransactionHandler) (blockAPI.APIBlockHandler, error) {
	blockApiArgs, err := createAPIBlockProcessorArgs(args, apiTransactionHandler)
	if err != nil {
		return nil, err
	}

	return blockAPI.CreateAPIBlockProcessor(blockApiArgs)
}

func createAPIInternalBlockProcessor(args *ApiResolverArgs, apiTransactionHandler external.APITransactionHandler) (blockAPI.APIInternalBlockHandler, error) {
	blockApiArgs, err := createAPIBlockProcessorArgs(args, apiTransactionHandler)
	if err != nil {
		return nil, err
	}

	return blockAPI.CreateAPIInternalBlockProcessor(blockApiArgs)
}

func createAPIBlockProcessorArgs(args *ApiResolverArgs, apiTransactionHandler external.APITransactionHandler) (*blockAPI.ArgAPIBlockProcessor, error) {
	statusComputer, err := txstatus.NewStatusComputer(
		args.ProcessComponents.ShardCoordinator().SelfId(),
		args.CoreComponents.Uint64ByteSliceConverter(),
		args.DataComponents.StorageService(),
	)
	if err != nil {
		return nil, errors.New("error creating transaction status computer " + err.Error())
	}

	logsFacade, err := createLogsFacade(args)
	if err != nil {
		return nil, err
	}

	alteredAccountsProvider, err := alteredaccounts.NewAlteredAccountsProvider(alteredaccounts.ArgsAlteredAccountsProvider{
		ShardCoordinator:       args.ProcessComponents.ShardCoordinator(),
		AddressConverter:       args.CoreComponents.AddressPubKeyConverter(),
		AccountsDB:             args.StateComponents.AccountsAdapterAPI(),
		EsdtDataStorageHandler: args.ProcessComponents.ESDTDataStorageHandlerForAPI(),
	})
	if err != nil {
		return nil, err
	}

	blockApiArgs := &blockAPI.ArgAPIBlockProcessor{
		SelfShardID:                  args.ProcessComponents.ShardCoordinator().SelfId(),
		Store:                        args.DataComponents.StorageService(),
		Marshalizer:                  args.CoreComponents.InternalMarshalizer(),
		Uint64ByteSliceConverter:     args.CoreComponents.Uint64ByteSliceConverter(),
		HistoryRepo:                  args.ProcessComponents.HistoryRepository(),
		APITransactionHandler:        apiTransactionHandler,
		StatusComputer:               statusComputer,
		AddressPubkeyConverter:       args.CoreComponents.AddressPubKeyConverter(),
		Hasher:                       args.CoreComponents.Hasher(),
		LogsFacade:                   logsFacade,
		ReceiptsRepository:           args.ProcessComponents.ReceiptsRepository(),
		AlteredAccountsProvider:      alteredAccountsProvider,
		AccountsRepository:           args.StateComponents.AccountsRepository(),
		ScheduledTxsExecutionHandler: args.ProcessComponents.ScheduledTxsExecutionHandler(),
		EnableEpochsHandler:          args.CoreComponents.EnableEpochsHandler(),
	}

	return blockApiArgs, nil
}

func createLogsFacade(args *ApiResolverArgs) (factory.LogsFacade, error) {
	return logs.NewLogsFacade(logs.ArgsNewLogsFacade{
		StorageService:  args.DataComponents.StorageService(),
		Marshaller:      args.CoreComponents.InternalMarshalizer(),
		PubKeyConverter: args.CoreComponents.AddressPubKeyConverter(),
	})
}
