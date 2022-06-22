package factory

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/facade"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/external/blockAPI"
	"github.com/ElrondNetwork/elrond-go/node/external/transactionAPI"
	"github.com/ElrondNetwork/elrond-go/node/trieIterators"
	trieIteratorsFactory "github.com/ElrondNetwork/elrond-go/node/trieIterators/factory"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/process/txstatus"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	vmcommonBuiltInFunctions "github.com/ElrondNetwork/elrond-vm-common/builtInFunctions"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
	datafield "github.com/ElrondNetwork/elrond-vm-common/parsers/dataField"
)

// ApiResolverArgs holds the argument needed to create an API resolver
type ApiResolverArgs struct {
	Configs             *config.Configs
	CoreComponents      CoreComponentsHolder
	DataComponents      DataComponentsHolder
	StateComponents     StateComponentsHolder
	BootstrapComponents BootstrapComponentsHolder
	CryptoComponents    CryptoComponentsHolder
	ProcessComponents   ProcessComponentsHolder
	GasScheduleNotifier core.GasScheduleNotifier
	Bootstrapper        process.Bootstrapper
	AllowVMQueriesChan  chan struct{}
}

type scQueryServiceArgs struct {
	generalConfig       *config.Config
	epochConfig         *config.EpochConfig
	coreComponents      CoreComponentsHolder
	stateComponents     StateComponentsHolder
	dataComponents      DataComponentsHolder
	processComponents   ProcessComponentsHolder
	gasScheduleNotifier core.GasScheduleNotifier
	messageSigVerifier  vm.MessageSignVerifier
	systemSCConfig      *config.SystemSmartContractsConfig
	bootstrapper        process.Bootstrapper
	allowVMQueriesChan  chan struct{}
	workingDir          string
}

type scQueryElementArgs struct {
	generalConfig       *config.Config
	epochConfig         *config.EpochConfig
	coreComponents      CoreComponentsHolder
	stateComponents     StateComponentsHolder
	dataComponents      DataComponentsHolder
	processComponents   ProcessComponentsHolder
	gasScheduleNotifier core.GasScheduleNotifier
	messageSigVerifier  vm.MessageSignVerifier
	systemSCConfig      *config.SystemSmartContractsConfig
	bootstrapper        process.Bootstrapper
	allowVMQueriesChan  chan struct{}
	workingDir          string
	index               int
}

// CreateApiResolver is able to create an ApiResolver instance that will solve the REST API requests through the node facade
// TODO: refactor to further decrease node's codebase
func CreateApiResolver(args *ApiResolverArgs) (facade.ApiResolver, error) {
	apiWorkingDir := filepath.Join(args.Configs.FlagsConfig.WorkingDir, common.TemporaryPath)
	argsSCQuery := &scQueryServiceArgs{
		generalConfig:       args.Configs.GeneralConfig,
		epochConfig:         args.Configs.EpochConfig,
		coreComponents:      args.CoreComponents,
		dataComponents:      args.DataComponents,
		stateComponents:     args.StateComponents,
		processComponents:   args.ProcessComponents,
		gasScheduleNotifier: args.GasScheduleNotifier,
		messageSigVerifier:  args.CryptoComponents.MessageSignVerifier(),
		systemSCConfig:      args.Configs.SystemSCConfig,
		bootstrapper:        args.Bootstrapper,
		allowVMQueriesChan:  args.AllowVMQueriesChan,
		workingDir:          apiWorkingDir,
	}

	scQueryService, err := createScQueryService(argsSCQuery)
	if err != nil {
		return nil, err
	}

	pubKeyConverter := args.CoreComponents.AddressPubKeyConverter()
	convertedAddress, err := pubKeyConverter.Decode(args.Configs.GeneralConfig.BuiltInFunctions.AutomaticCrawlerAddress)
	if err != nil {
		return nil, err
	}
	builtInFuncs, _, _, err := createBuiltinFuncs(
		args.GasScheduleNotifier,
		args.CoreComponents.InternalMarshalizer(),
		args.StateComponents.AccountsAdapterAPI(),
		args.BootstrapComponents.ShardCoordinator(),
		args.CoreComponents.EnableEpochsHandler(),
		convertedAddress,
	)
	if err != nil {
		return nil, err
	}

	esdtTransferParser, err := parsers.NewESDTTransferParser(args.CoreComponents.InternalMarshalizer())
	if err != nil {
		return nil, err
	}

	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:    args.CoreComponents.AddressPubKeyConverter(),
		ShardCoordinator:   args.ProcessComponents.ShardCoordinator(),
		BuiltInFunctions:   builtInFuncs,
		ArgumentParser:     parsers.NewCallArgsParser(),
		ESDTTransferParser: esdtTransferParser,
	}
	txTypeHandler, err := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	if err != nil {
		return nil, err
	}

	txCostHandler, err := transaction.NewTransactionCostEstimator(
		txTypeHandler,
		args.CoreComponents.EconomicsData(),
		args.ProcessComponents.TransactionSimulatorProcessor(),
		args.StateComponents.AccountsAdapterAPI(),
		args.ProcessComponents.ShardCoordinator(),
		args.CoreComponents.EnableEpochsHandler(),
	)
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

	argsDataFieldParser := &datafield.ArgsOperationDataFieldParser{
		AddressLength:    args.CoreComponents.AddressPubKeyConverter().Len(),
		Marshalizer:      args.CoreComponents.InternalMarshalizer(),
		ShardCoordinator: args.ProcessComponents.ShardCoordinator(),
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
		StatusMetricsHandler:     args.CoreComponents.StatusHandlerUtils().Metrics(),
		TxCostHandler:            txCostHandler,
		TotalStakedValueHandler:  totalStakedValueHandler,
		DirectStakedListHandler:  directStakedListHandler,
		DelegatedListHandler:     delegatedListHandler,
		APITransactionHandler:    apiTransactionProcessor,
		APIBlockHandler:          apiBlockProcessor,
		APIInternalBlockHandler:  apiInternalBlockProcessor,
		GenesisNodesSetupHandler: args.CoreComponents.GenesisNodesSetup(),
		ValidatorPubKeyConverter: args.CoreComponents.ValidatorPubKeyConverter(),
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
		generalConfig:       args.generalConfig,
		epochConfig:         args.epochConfig,
		coreComponents:      args.coreComponents,
		dataComponents:      args.dataComponents,
		stateComponents:     args.stateComponents,
		processComponents:   args.processComponents,
		gasScheduleNotifier: args.gasScheduleNotifier,
		messageSigVerifier:  args.messageSigVerifier,
		systemSCConfig:      args.systemSCConfig,
		workingDir:          args.workingDir,
		bootstrapper:        args.bootstrapper,
		allowVMQueriesChan:  args.allowVMQueriesChan,
		index:               0,
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
	var vmFactory process.VirtualMachinesContainerFactory
	var err error

	pubKeyConverter := args.coreComponents.AddressPubKeyConverter()
	convertedAddress, err := pubKeyConverter.Decode(args.generalConfig.BuiltInFunctions.AutomaticCrawlerAddress)
	if err != nil {
		return nil, err
	}
	builtInFuncs, nftStorageHandler, globalSettingsHandler, err := createBuiltinFuncs(
		args.gasScheduleNotifier,
		args.coreComponents.InternalMarshalizer(),
		args.stateComponents.AccountsAdapterAPI(),
		args.processComponents.ShardCoordinator(),
		args.coreComponents.EnableEpochsHandler(),
		convertedAddress,
	)
	if err != nil {
		return nil, err
	}

	cacherCfg := storageFactory.GetCacherFromConfig(args.generalConfig.SmartContractDataPool)
	smartContractsCache, err := storageUnit.NewCache(cacherCfg)
	if err != nil {
		return nil, err
	}

	scStorage := args.generalConfig.SmartContractsStorageForSCQuery
	scStorage.DB.FilePath += fmt.Sprintf("%d", args.index)
	argsHook := hooks.ArgBlockChainHook{
		Accounts:              args.stateComponents.AccountsAdapterAPI(),
		PubkeyConv:            args.coreComponents.AddressPubKeyConverter(),
		StorageService:        args.dataComponents.StorageService(),
		BlockChain:            args.dataComponents.Blockchain(),
		ShardCoordinator:      args.processComponents.ShardCoordinator(),
		Marshalizer:           args.coreComponents.InternalMarshalizer(),
		Uint64Converter:       args.coreComponents.Uint64ByteSliceConverter(),
		BuiltInFunctions:      builtInFuncs,
		NFTStorageHandler:     nftStorageHandler,
		GlobalSettingsHandler: globalSettingsHandler,
		DataPool:              args.dataComponents.Datapool(),
		ConfigSCStorage:       scStorage,
		CompiledSCPool:        smartContractsCache,
		WorkingDir:            args.workingDir,
		EnableEpochsHandler:   args.coreComponents.EnableEpochsHandler(),
		NilCompiledSCStore:    true,
	}

	if args.processComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		blockChainHookImpl, errBlockChainHook := hooks.NewBlockChainHookImpl(argsHook)
		if errBlockChainHook != nil {
			return nil, errBlockChainHook
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
			ChanceComputer:      args.coreComponents.Rater(),
			ShardCoordinator:    args.processComponents.ShardCoordinator(),
			EnableEpochsHandler: args.coreComponents.EnableEpochsHandler(),
		}
		vmFactory, err = metachain.NewVMContainerFactory(argsNewVmFactory)
		if err != nil {
			return nil, err
		}
	} else {
		queryVirtualMachineConfig := args.generalConfig.VirtualMachine.Querying.VirtualMachineConfig
		esdtTransferParser, errParser := parsers.NewESDTTransferParser(args.coreComponents.InternalMarshalizer())
		if errParser != nil {
			return nil, err
		}

		blockChainHookImpl, errBlockChainHook := hooks.NewBlockChainHookImpl(argsHook)
		if errBlockChainHook != nil {
			return nil, errBlockChainHook
		}

		argsNewVMFactory := shard.ArgVMContainerFactory{
			BlockChainHook:      blockChainHookImpl,
			BuiltInFunctions:    argsHook.BuiltInFunctions,
			Config:              queryVirtualMachineConfig,
			BlockGasLimit:       args.coreComponents.EconomicsData().MaxGasLimitPerBlock(args.processComponents.ShardCoordinator().SelfId()),
			GasSchedule:         args.gasScheduleNotifier,
			EpochNotifier:       args.coreComponents.EpochNotifier(),
			EnableEpochsHandler: args.coreComponents.EnableEpochsHandler(),
			ArwenChangeLocker:   args.coreComponents.ArwenChangeLocker(),
			ESDTTransferParser:  esdtTransferParser,
		}

		log.Debug("apiResolver: enable epoch for sc deploy", "epoch", args.epochConfig.EnableEpochs.SCDeployEnableEpoch)
		log.Debug("apiResolver: enable epoch for ahead of time gas usage", "epoch", args.epochConfig.EnableEpochs.AheadOfTimeGasUsageEnableEpoch)
		log.Debug("apiResolver: enable epoch for repair callback", "epoch", args.epochConfig.EnableEpochs.RepairCallbackEnableEpoch)

		vmFactory, err = shard.NewVMContainerFactory(argsNewVMFactory)
		if err != nil {
			return nil, err
		}
	}

	vmContainer, err := vmFactory.Create()
	if err != nil {
		return nil, err
	}

	err = vmcommonBuiltInFunctions.SetPayableHandler(builtInFuncs, vmFactory.BlockChainHookImpl())
	if err != nil {
		return nil, err
	}

	argsNewSCQueryService := smartContract.ArgsNewSCQueryService{
		VmContainer:              vmContainer,
		EconomicsFee:             args.coreComponents.EconomicsData(),
		BlockChainHook:           vmFactory.BlockChainHookImpl(),
		BlockChain:               args.dataComponents.Blockchain(),
		ArwenChangeLocker:        args.coreComponents.ArwenChangeLocker(),
		Bootstrapper:             args.bootstrapper,
		AllowExternalQueriesChan: args.allowVMQueriesChan,
		MaxGasLimitPerQuery:      args.generalConfig.VirtualMachine.GasConfig.MaxGasPerVmQuery,
	}

	return smartContract.NewSCQueryService(argsNewSCQueryService)
}

func createBuiltinFuncs(
	gasScheduleNotifier core.GasScheduleNotifier,
	marshalizer marshal.Marshalizer,
	accnts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	enableEpochsHandler vmcommon.EnableEpochsHandler,
	automaticCrawlerAddress []byte,
) (vmcommon.BuiltInFunctionContainer, vmcommon.SimpleESDTNFTStorageHandler, vmcommon.ESDTGlobalSettingsHandler, error) {
	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasSchedule:             gasScheduleNotifier,
		MapDNSAddresses:         make(map[string]struct{}),
		Marshalizer:             marshalizer,
		Accounts:                accnts,
		ShardCoordinator:        shardCoordinator,
		EnableEpochsHandler:     enableEpochsHandler,
		AutomaticCrawlerAddress: automaticCrawlerAddress,
	}
	return builtInFunctions.CreateBuiltInFuncContainerAndNFTStorageHandler(argsBuiltIn)
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

	blockApiArgs := &blockAPI.ArgAPIBlockProcessor{
		SelfShardID:              args.ProcessComponents.ShardCoordinator().SelfId(),
		Store:                    args.DataComponents.StorageService(),
		Marshalizer:              args.CoreComponents.InternalMarshalizer(),
		Uint64ByteSliceConverter: args.CoreComponents.Uint64ByteSliceConverter(),
		HistoryRepo:              args.ProcessComponents.HistoryRepository(),
		TxUnmarshaller:           apiTransactionHandler,
		StatusComputer:           statusComputer,
		AddressPubkeyConverter:   args.CoreComponents.AddressPubKeyConverter(),
		Hasher:                   args.CoreComponents.Hasher(),
	}

	return blockApiArgs, nil
}
