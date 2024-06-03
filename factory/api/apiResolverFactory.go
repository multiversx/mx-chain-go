package api

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	datafield "github.com/multiversx/mx-chain-vm-common-go/parsers/dataField"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/disabled"
	"github.com/multiversx/mx-chain-go/common/operationmodes"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	"github.com/multiversx/mx-chain-go/facade"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/addressDecoder"
	factoryVm "github.com/multiversx/mx-chain-go/factory/vm"
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
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/process/smartContract/builtInFunctions"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks/counters"
	"github.com/multiversx/mx-chain-go/process/txstatus"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/blockInfoProviders"
	disabledState "github.com/multiversx/mx-chain-go/state/disabled"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager/evictionWaitingList"
	"github.com/multiversx/mx-chain-go/state/syncer"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	trieFactory "github.com/multiversx/mx-chain-go/trie/factory"
	"github.com/multiversx/mx-chain-go/vm"
)

var log = logger.GetOrCreate("factory")

// ApiResolverArgs holds the argument needed to create an API resolver
type ApiResolverArgs struct {
	Configs                        *config.Configs
	CoreComponents                 factory.CoreComponentsHolder
	DataComponents                 factory.DataComponentsHolder
	StateComponents                factory.StateComponentsHolder
	BootstrapComponents            factory.BootstrapComponentsHolder
	CryptoComponents               factory.CryptoComponentsHolder
	ProcessComponents              factory.ProcessComponentsHolder
	StatusCoreComponents           factory.StatusCoreComponentsHolder
	StatusComponents               factory.StatusComponentsHolder
	GasScheduleNotifier            common.GasScheduleNotifierAPI
	Bootstrapper                   process.Bootstrapper
	RunTypeComponents              factory.RunTypeComponentsHolder
	AllowVMQueriesChan             chan struct{}
	ProcessingMode                 common.NodeProcessingMode
	DelegatedListFactoryHandler    trieIteratorsFactory.DelegatedListProcessorFactoryHandler
	DirectStakedListFactoryHandler trieIteratorsFactory.DirectStakedListProcessorFactoryHandler
	TotalStakedValueFactoryHandler trieIteratorsFactory.TotalStakedValueProcessorFactoryHandler
}

type scQueryServiceArgs struct {
	generalConfig              *config.Config
	epochConfig                *config.EpochConfig
	coreComponents             factory.CoreComponentsHolder
	stateComponents            factory.StateComponentsHolder
	dataComponents             factory.DataComponentsHolder
	processComponents          factory.ProcessComponentsHolder
	statusCoreComponents       factory.StatusCoreComponentsHolder
	runTypeComponents          factory.RunTypeComponentsHolder
	gasScheduleNotifier        core.GasScheduleNotifier
	messageSigVerifier         vm.MessageSignVerifier
	systemSCConfig             *config.SystemSmartContractsConfig
	bootstrapper               process.Bootstrapper
	guardedAccountHandler      process.GuardedAccountHandler
	allowVMQueriesChan         chan struct{}
	workingDir                 string
	processingMode             common.NodeProcessingMode
	isInHistoricalBalancesMode bool
}

type scQueryElementArgs struct {
	generalConfig              *config.Config
	epochConfig                *config.EpochConfig
	coreComponents             factory.CoreComponentsHolder
	stateComponents            factory.StateComponentsHolder
	dataComponents             factory.DataComponentsHolder
	processComponents          factory.ProcessComponentsHolder
	statusCoreComponents       factory.StatusCoreComponentsHolder
	runTypeComponents          factory.RunTypeComponentsHolder
	gasScheduleNotifier        core.GasScheduleNotifier
	messageSigVerifier         vm.MessageSignVerifier
	systemSCConfig             *config.SystemSmartContractsConfig
	bootstrapper               process.Bootstrapper
	guardedAccountHandler      process.GuardedAccountHandler
	allowVMQueriesChan         chan struct{}
	workingDir                 string
	index                      int
	processingMode             common.NodeProcessingMode
	isInHistoricalBalancesMode bool
}

// CreateApiResolver is able to create an ApiResolver instance that will solve the REST API requests through the node facade
// TODO: refactor to further decrease node's codebase
func CreateApiResolver(args *ApiResolverArgs) (facade.ApiResolver, error) {
	apiWorkingDir := filepath.Join(args.Configs.FlagsConfig.WorkingDir, common.TemporaryPath)

	if check.IfNilReflect(args.DelegatedListFactoryHandler) {
		return nil, factory.ErrNilDelegatedListFactory
	}
	if check.IfNilReflect(args.DirectStakedListFactoryHandler) {
		return nil, factory.ErrNilDirectStakedListFactory
	}
	if check.IfNilReflect(args.TotalStakedValueFactoryHandler) {
		return nil, factory.ErrNilTotalStakedValueFactory
	}

	argsSCQuery := &scQueryServiceArgs{
		generalConfig:              args.Configs.GeneralConfig,
		epochConfig:                args.Configs.EpochConfig,
		coreComponents:             args.CoreComponents,
		dataComponents:             args.DataComponents,
		stateComponents:            args.StateComponents,
		processComponents:          args.ProcessComponents,
		statusCoreComponents:       args.StatusCoreComponents,
		gasScheduleNotifier:        args.GasScheduleNotifier,
		messageSigVerifier:         args.CryptoComponents.MessageSignVerifier(),
		systemSCConfig:             args.Configs.SystemSCConfig,
		bootstrapper:               args.Bootstrapper,
		guardedAccountHandler:      args.BootstrapComponents.GuardedAccountHandler(),
		allowVMQueriesChan:         args.AllowVMQueriesChan,
		workingDir:                 apiWorkingDir,
		processingMode:             args.ProcessingMode,
		isInHistoricalBalancesMode: operationmodes.IsInHistoricalBalancesMode(args.Configs),
		runTypeComponents:          args.RunTypeComponents,
	}

	scQueryService, storageManagers, err := createScQueryService(argsSCQuery)
	if err != nil {
		return nil, err
	}

	pkConverter := args.CoreComponents.AddressPubKeyConverter()
	automaticCrawlerAddressesStrings := args.Configs.GeneralConfig.BuiltInFunctions.AutomaticCrawlerAddresses
	convertedAddresses, errDecode := addressDecoder.DecodeAddresses(pkConverter, automaticCrawlerAddressesStrings)
	if errDecode != nil {
		return nil, errDecode
	}

	dnsV2AddressesStrings := args.Configs.GeneralConfig.BuiltInFunctions.DNSV2Addresses
	convertedDNSV2Addresses, errDecode := addressDecoder.DecodeAddresses(pkConverter, dnsV2AddressesStrings)
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
	totalStakedValueHandler, err := args.TotalStakedValueFactoryHandler.CreateTotalStakedValueProcessorHandler(argsProcessors)
	if err != nil {
		return nil, err
	}

	directStakedListHandler, err := args.DirectStakedListFactoryHandler.CreateDirectStakedListProcessorHandler(argsProcessors)
	if err != nil {
		return nil, err
	}

	delegatedListHandler, err := args.DelegatedListFactoryHandler.CreateDelegatedListProcessorHandler(argsProcessors)
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
		PublicKey:                args.CryptoComponents.PublicKeyString(),
		NodesCoordinator:         args.ProcessComponents.NodesCoordinator(),
		StorageManagers:          storageManagers,
	}

	return external.NewNodeApiResolver(argsApiResolver)
}

func createScQueryService(
	args *scQueryServiceArgs,
) (process.SCQueryService, []common.StorageManager, error) {
	numConcurrentVms := args.generalConfig.VirtualMachine.Querying.NumConcurrentVMs
	if numConcurrentVms < 1 {
		return nil, nil, fmt.Errorf("VirtualMachine.Querying.NumConcurrentVms should be a positive number more than 1")
	}

	argsQueryElem := &scQueryElementArgs{
		generalConfig:              args.generalConfig,
		epochConfig:                args.epochConfig,
		coreComponents:             args.coreComponents,
		stateComponents:            args.stateComponents,
		dataComponents:             args.dataComponents,
		processComponents:          args.processComponents,
		statusCoreComponents:       args.statusCoreComponents,
		gasScheduleNotifier:        args.gasScheduleNotifier,
		messageSigVerifier:         args.messageSigVerifier,
		systemSCConfig:             args.systemSCConfig,
		bootstrapper:               args.bootstrapper,
		guardedAccountHandler:      args.guardedAccountHandler,
		allowVMQueriesChan:         args.allowVMQueriesChan,
		workingDir:                 args.workingDir,
		index:                      0,
		processingMode:             args.processingMode,
		isInHistoricalBalancesMode: args.isInHistoricalBalancesMode,
		runTypeComponents:          args.runTypeComponents,
	}

	var err error
	var scQueryService process.SCQueryService
	var storageManager common.StorageManager
	storageManagers := make([]common.StorageManager, 0, numConcurrentVms)

	list := make([]process.SCQueryService, 0, numConcurrentVms)
	for i := 0; i < numConcurrentVms; i++ {
		argsQueryElem.index = i
		scQueryService, storageManager, err = createScQueryElement(*argsQueryElem)
		if err != nil {
			return nil, nil, err
		}

		list = append(list, scQueryService)
		storageManagers = append(storageManagers, storageManager)
	}

	sqQueryDispatcher, err := smartContract.NewScQueryServiceDispatcher(list)
	if err != nil {
		return nil, nil, err
	}

	return sqQueryDispatcher, storageManagers, nil
}

func createScQueryElement(
	args scQueryElementArgs,
) (process.SCQueryService, common.StorageManager, error) {
	argsNewSCQueryService, storageMapper, err := createArgsSCQueryService(&args)
	if err != nil {
		return nil, nil, err
	}
	scQueryService, err := smartContract.NewSCQueryService(*argsNewSCQueryService)
	if err != nil {
		return nil, nil, err
	}

	return scQueryService, storageMapper, nil
}

func createArgsSCQueryService(args *scQueryElementArgs) (*smartContract.ArgsNewSCQueryService, common.StorageManager, error) {
	var err error

	selfShardID := args.processComponents.ShardCoordinator().SelfId()

	pkConverter := args.coreComponents.AddressPubKeyConverter()
	automaticCrawlerAddressesStrings := args.generalConfig.BuiltInFunctions.AutomaticCrawlerAddresses
	convertedAddresses, errDecode := addressDecoder.DecodeAddresses(pkConverter, automaticCrawlerAddressesStrings)
	if errDecode != nil {
		return nil, nil, errDecode
	}

	dnsV2AddressesStrings := args.generalConfig.BuiltInFunctions.DNSV2Addresses
	convertedDNSV2Addresses, errDecode := addressDecoder.DecodeAddresses(pkConverter, dnsV2AddressesStrings)
	if errDecode != nil {
		return nil, nil, errDecode
	}

	apiBlockchain, err := createBlockchainForScQuery(selfShardID)
	if err != nil {
		return nil, nil, err
	}

	accountsAdapterApi, _, err := createNewAccountsAdapterApi(args, apiBlockchain)
	if err != nil {
		return nil, nil, err
	}

	builtInFuncFactory, err := createBuiltinFuncs(
		args.gasScheduleNotifier,
		args.coreComponents.InternalMarshalizer(),
		accountsAdapterApi,
		args.processComponents.ShardCoordinator(),
		args.coreComponents.EpochNotifier(),
		args.coreComponents.EnableEpochsHandler(),
		args.guardedAccountHandler,
		convertedAddresses,
		args.generalConfig.BuiltInFunctions.MaxNumAddressesInTransferRole,
		convertedDNSV2Addresses,
	)
	if err != nil {
		return nil, nil, err
	}

	cacherCfg := storageFactory.GetCacherFromConfig(args.generalConfig.SmartContractDataPool)
	smartContractsCache, err := storageunit.NewCache(cacherCfg)
	if err != nil {
		return nil, nil, err
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
		Accounts:                 accountsAdapterApi,
		BlockChain:               apiBlockchain,
	}

	var vmContainer process.VirtualMachinesContainer
	var vmFactory process.VirtualMachinesContainerFactory
	var storageManager common.StorageManager
	maxGasForVmQueries := args.generalConfig.VirtualMachine.GasConfig.ShardMaxGasPerVmQuery
	if selfShardID == core.MetachainShardId {
		maxGasForVmQueries = args.generalConfig.VirtualMachine.GasConfig.MetaMaxGasPerVmQuery

		argsHook.BlockChain, err = blockchain.NewMetaChain(disabled.NewAppStatusHandler())
		if err != nil {
			return nil, nil, err
		}

		argsHook.Accounts, storageManager, err = createNewAccountsAdapterApi(args, argsHook.BlockChain)
		if err != nil {
			return nil, nil, err
		}

		argsNewVmContainerFactory := factoryVm.ArgsVmContainerFactory{
			PubkeyConv:                 argsHook.PubkeyConv,
			Economics:                  args.coreComponents.EconomicsData(),
			MessageSignVerifier:        args.messageSigVerifier,
			GasSchedule:                args.gasScheduleNotifier,
			NodesConfigProvider:        args.coreComponents.GenesisNodesSetup(),
			Hasher:                     args.coreComponents.Hasher(),
			Marshalizer:                args.coreComponents.InternalMarshalizer(),
			SystemSCConfig:             args.systemSCConfig,
			ValidatorAccountsDB:        args.stateComponents.PeerAccounts(),
			UserAccountsDB:             args.stateComponents.AccountsAdapterAPI(),
			ChanceComputer:             args.coreComponents.Rater(),
			ShardCoordinator:           args.processComponents.ShardCoordinator(),
			EnableEpochsHandler:        args.coreComponents.EnableEpochsHandler(),
			NodesCoordinator:           args.processComponents.NodesCoordinator(),
			IsInHistoricalBalancesMode: args.isInHistoricalBalancesMode,
		}

		vmContainer, vmFactory, err = args.runTypeComponents.VmContainerMetaFactoryCreator().CreateVmContainerFactory(argsHook, argsNewVmContainerFactory)
	} else {
		argsHook.BlockChain, err = blockchain.NewBlockChain(disabled.NewAppStatusHandler())
		if err != nil {
			return nil, nil, err
		}

		argsHook.Accounts, storageManager, err = createNewAccountsAdapterApi(args, argsHook.BlockChain)
		if err != nil {
			return nil, nil, err
		}

		queryVirtualMachineConfig := args.generalConfig.VirtualMachine.Querying.VirtualMachineConfig
		esdtTransferParser, errParser := parsers.NewESDTTransferParser(args.coreComponents.InternalMarshalizer())
		if errParser != nil {
			return nil, nil, errParser
		}

		argsNewVmContainerFactory := factoryVm.ArgsVmContainerFactory{
			BuiltInFunctions:    argsHook.BuiltInFunctions,
			Config:              queryVirtualMachineConfig,
			BlockGasLimit:       args.coreComponents.EconomicsData().MaxGasLimitPerBlock(args.processComponents.ShardCoordinator().SelfId()),
			GasSchedule:         args.gasScheduleNotifier,
			EpochNotifier:       args.coreComponents.EpochNotifier(),
			EnableEpochsHandler: args.coreComponents.EnableEpochsHandler(),
			WasmVMChangeLocker:  args.coreComponents.WasmVMChangeLocker(),
			ESDTTransferParser:  esdtTransferParser,
			Hasher:              args.coreComponents.Hasher(),
			PubkeyConv:          args.coreComponents.AddressPubKeyConverter(),
			Economics:           args.coreComponents.EconomicsData(),
			MessageSignVerifier: args.messageSigVerifier,
			NodesConfigProvider: args.coreComponents.GenesisNodesSetup(),
			Marshalizer:         args.coreComponents.InternalMarshalizer(),
			SystemSCConfig:      args.systemSCConfig,
			ValidatorAccountsDB: args.stateComponents.PeerAccounts(),
			UserAccountsDB:      args.stateComponents.AccountsAdapterAPI(),
			ChanceComputer:      args.coreComponents.Rater(),
			ShardCoordinator:    args.processComponents.ShardCoordinator(),
			NodesCoordinator:    args.processComponents.NodesCoordinator(),
		}

		vmContainer, vmFactory, err = args.runTypeComponents.VmContainerShardFactoryCreator().CreateVmContainerFactory(argsHook, argsNewVmContainerFactory)
	}

	if err != nil {
		return nil, nil, err
	}
	log.Debug("maximum gas per VM Query", "value", maxGasForVmQueries)

	err = vmFactory.BlockChainHookImpl().SetVMContainer(vmContainer)
	if err != nil {
		return nil, nil, err
	}

	err = builtInFuncFactory.SetPayableHandler(vmFactory.BlockChainHookImpl())
	if err != nil {
		return nil, nil, err
	}

	return &smartContract.ArgsNewSCQueryService{
		VmContainer:              vmContainer,
		EconomicsFee:             args.coreComponents.EconomicsData(),
		BlockChainHook:           vmFactory.BlockChainHookImpl(),
		MainBlockChain:           args.dataComponents.Blockchain(),
		APIBlockChain:            argsHook.BlockChain,
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
	}, storageManager, nil
}

func createBlockchainForScQuery(selfShardID uint32) (data.ChainHandler, error) {
	isMetachain := selfShardID == core.MetachainShardId
	if isMetachain {
		return blockchain.NewMetaChain(disabled.NewAppStatusHandler())
	}

	return blockchain.NewBlockChain(disabled.NewAppStatusHandler())
}

func createNewAccountsAdapterApi(args *scQueryElementArgs, chainHandler data.ChainHandler) (state.AccountsAdapterAPI, common.StorageManager, error) {
	storagePruning, err := newStoragePruningManager(*args)
	if err != nil {
		return nil, nil, err
	}
	storageService := args.dataComponents.StorageService()
	trieStorer, err := storageService.GetStorer(dataRetriever.UserAccountsUnit)
	if err != nil {
		return nil, nil, err
	}

	trieFactoryArgs := trieFactory.TrieFactoryArgs{
		Marshalizer:              args.coreComponents.InternalMarshalizer(),
		Hasher:                   args.coreComponents.Hasher(),
		PathManager:              args.coreComponents.PathHandler(),
		TrieStorageManagerConfig: args.generalConfig.TrieStorageManagerConfig,
	}
	trFactory, err := trieFactory.NewTrieFactory(trieFactoryArgs)
	if err != nil {
		return nil, nil, err
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
	trieStorageManager, merkleTrie, err := trFactory.Create(trieCreatorArgs)
	if err != nil {
		return nil, nil, err
	}

	argsAPIAccountsDB := state.ArgsAccountsDB{
		Trie:                  merkleTrie,
		Hasher:                args.coreComponents.Hasher(),
		Marshaller:            args.coreComponents.InternalMarshalizer(),
		AccountFactory:        args.runTypeComponents.AccountsCreator(),
		StoragePruningManager: storagePruning,
		AddressConverter:      args.coreComponents.AddressPubKeyConverter(),
		SnapshotsManager:      disabledState.NewDisabledSnapshotsManager(),
	}

	provider, err := blockInfoProviders.NewCurrentBlockInfo(chainHandler)
	if err != nil {
		return nil, nil, err
	}

	accounts, err := state.NewAccountsDB(argsAPIAccountsDB)
	if err != nil {
		return nil, nil, err
	}

	accountsDB, err := state.NewAccountsDBApi(accounts, provider)

	return accountsDB, trieStorageManager, err
}

func newStoragePruningManager(args scQueryElementArgs) (state.StoragePruningManager, error) {
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
