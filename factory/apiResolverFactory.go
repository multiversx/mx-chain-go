package factory

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/parsers"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/facade"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node/external"
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
	"github.com/ElrondNetwork/elrond-go/sharding"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/vm"
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
	workingDir          string
	index               int
}

// CreateApiResolver is able to create an ApiResolver instance that will solve the REST API requests through the node facade
// TODO: refactor to further decrease node's codebase
func CreateApiResolver(args *ApiResolverArgs) (facade.ApiResolver, error) {
	apiWorkingDir := filepath.Join(args.Configs.FlagsConfig.WorkingDir, core.TemporaryPath)
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
		workingDir:          apiWorkingDir,
	}

	scQueryService, _, _, err := createScQueryService(argsSCQuery)
	if err != nil {
		return nil, err
	}

	builtInFuncs, err := createBuiltinFuncs(
		args.GasScheduleNotifier,
		args.CoreComponents.InternalMarshalizer(),
		args.StateComponents.AccountsAdapter(),
		args.BootstrapComponents.ShardCoordinator(),
	)
	if err != nil {
		return nil, err
	}

	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:        args.CoreComponents.AddressPubKeyConverter(),
		ShardCoordinator:       args.ProcessComponents.ShardCoordinator(),
		BuiltInFuncNames:       builtInFuncs.Keys(),
		ArgumentParser:         parsers.NewCallArgsParser(),
		EpochNotifier:          args.CoreComponents.EpochNotifier(),
		RelayedTxV2EnableEpoch: args.Configs.EpochConfig.EnableEpochs.RelayedTransactionsV2EnableEpoch,
	}
	txTypeHandler, err := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	if err != nil {
		return nil, err
	}

	txCostHandler, err := transaction.NewTransactionCostEstimator(
		txTypeHandler,
		args.CoreComponents.EconomicsData(),
		scQueryService,
		args.GasScheduleNotifier,
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
		BlockChain:         args.DataComponents.Blockchain(),
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

	argsApiResolver := external.ArgNodeApiResolver{
		SCQueryService:          scQueryService,
		StatusMetricsHandler:    args.CoreComponents.StatusHandlerUtils().Metrics(),
		TxCostHandler:           txCostHandler,
		TotalStakedValueHandler: totalStakedValueHandler,
		DirectStakedListHandler: directStakedListHandler,
		DelegatedListHandler:    delegatedListHandler,
	}

	return external.NewNodeApiResolver(argsApiResolver)
}

func createScQueryService(
	args *scQueryServiceArgs,
) (process.SCQueryService, process.VirtualMachinesContainerFactory, process.VirtualMachinesContainer, error) {
	numConcurrentVms := args.generalConfig.VirtualMachine.Querying.NumConcurrentVMs
	if numConcurrentVms < 1 {
		return nil, nil, nil, fmt.Errorf("VirtualMachine.Querying.NumConcurrentVms should be a positive number more than 1")
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
		index:               0,
	}

	var err error
	var vmFactory process.VirtualMachinesContainerFactory
	var vmContainer process.VirtualMachinesContainer
	var scQueryService process.SCQueryService

	list := make([]process.SCQueryService, 0, numConcurrentVms)
	for i := 0; i < numConcurrentVms; i++ {
		argsQueryElem.index = i
		scQueryService, vmFactory, vmContainer, err = createScQueryElement(argsQueryElem)
		if err != nil {
			return nil, nil, nil, err
		}

		list = append(list, scQueryService)
	}

	sqQueryDispatcher, err := smartContract.NewScQueryServiceDispatcher(list)
	if err != nil {
		return nil, nil, nil, err
	}

	return sqQueryDispatcher, vmFactory, vmContainer, nil
}

func createScQueryElement(
	args *scQueryElementArgs,
) (process.SCQueryService, process.VirtualMachinesContainerFactory, process.VirtualMachinesContainer, error) {
	var vmFactory process.VirtualMachinesContainerFactory
	var err error

	builtInFuncs, err := createBuiltinFuncs(
		args.gasScheduleNotifier,
		args.coreComponents.InternalMarshalizer(),
		args.stateComponents.AccountsAdapter(),
		args.processComponents.ShardCoordinator(),
	)
	if err != nil {
		return nil, nil, nil, err
	}

	cacherCfg := storageFactory.GetCacherFromConfig(args.generalConfig.SmartContractDataPool)
	smartContractsCache, err := storageUnit.NewCache(cacherCfg)
	if err != nil {
		return nil, nil, nil, err
	}

	scStorage := args.generalConfig.SmartContractsStorageForSCQuery
	scStorage.DB.FilePath += fmt.Sprintf("%d", args.index)
	argsHook := hooks.ArgBlockChainHook{
		Accounts:           args.stateComponents.AccountsAdapter(),
		PubkeyConv:         args.coreComponents.AddressPubKeyConverter(),
		StorageService:     args.dataComponents.StorageService(),
		BlockChain:         args.dataComponents.Blockchain(),
		ShardCoordinator:   args.processComponents.ShardCoordinator(),
		Marshalizer:        args.coreComponents.InternalMarshalizer(),
		Uint64Converter:    args.coreComponents.Uint64ByteSliceConverter(),
		BuiltInFunctions:   builtInFuncs,
		DataPool:           args.dataComponents.Datapool(),
		ConfigSCStorage:    scStorage,
		CompiledSCPool:     smartContractsCache,
		WorkingDir:         args.workingDir,
		NilCompiledSCStore: true,
	}

	if args.processComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		argsNewVmFactory := metachain.ArgsNewVMContainerFactory{
			ArgBlockChainHook:   argsHook,
			Economics:           args.coreComponents.EconomicsData(),
			MessageSignVerifier: args.messageSigVerifier,
			GasSchedule:         args.gasScheduleNotifier,
			NodesConfigProvider: args.coreComponents.GenesisNodesSetup(),
			Hasher:              args.coreComponents.Hasher(),
			Marshalizer:         args.coreComponents.InternalMarshalizer(),
			SystemSCConfig:      args.systemSCConfig,
			ValidatorAccountsDB: args.stateComponents.PeerAccounts(),
			ChanceComputer:      args.coreComponents.Rater(),
			EpochNotifier:       args.coreComponents.EpochNotifier(),
			EpochConfig:         args.epochConfig,
		}
		vmFactory, err = metachain.NewVMContainerFactory(argsNewVmFactory)
		if err != nil {
			return nil, nil, nil, err
		}
	} else {
		queryVirtualMachineConfig := args.generalConfig.VirtualMachine.Querying.VirtualMachineConfig
		queryVirtualMachineConfig.OutOfProcessEnabled = true
		argsNewVMFactory := shard.ArgVMContainerFactory{
			Config:                         queryVirtualMachineConfig,
			BlockGasLimit:                  args.coreComponents.EconomicsData().MaxGasLimitPerBlock(args.processComponents.ShardCoordinator().SelfId()),
			GasSchedule:                    args.gasScheduleNotifier,
			ArgBlockChainHook:              argsHook,
			DeployEnableEpoch:              args.epochConfig.EnableEpochs.SCDeployEnableEpoch,
			AheadOfTimeGasUsageEnableEpoch: args.epochConfig.EnableEpochs.AheadOfTimeGasUsageEnableEpoch,
			ArwenV3EnableEpoch:             args.epochConfig.EnableEpochs.RepairCallbackEnableEpoch,
			ArwenESDTFunctionsEnableEpoch:  args.epochConfig.EnableEpochs.ArwenESDTFunctionsEnableEpoch,
		}

		log.Debug("apiResolver: enable epoch for sc deploy", "epoch", args.epochConfig.EnableEpochs.SCDeployEnableEpoch)
		log.Debug("apiResolver: enable epoch for ahead of time gas usage", "epoch", args.epochConfig.EnableEpochs.AheadOfTimeGasUsageEnableEpoch)
		log.Debug("apiResolver: enable epoch for repair callback", "epoch", args.epochConfig.EnableEpochs.RepairCallbackEnableEpoch)
		log.Debug("apiResolver: enable epoch for ESDT functions", "epoch", args.epochConfig.EnableEpochs.ArwenESDTFunctionsEnableEpoch)

		vmFactory, err = shard.NewVMContainerFactory(argsNewVMFactory)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	vmContainer, err := vmFactory.Create()
	if err != nil {
		return nil, nil, nil, err
	}

	err = builtInFunctions.SetPayableHandler(builtInFuncs, vmFactory.BlockChainHookImpl())
	if err != nil {
		return nil, nil, nil, err
	}

	scQueryService, err := smartContract.NewSCQueryService(
		vmContainer,
		args.coreComponents.EconomicsData(),
		vmFactory.BlockChainHookImpl(),
		args.dataComponents.Blockchain(),
	)

	return scQueryService, vmFactory, vmContainer, err
}

func createBuiltinFuncs(
	gasScheduleNotifier core.GasScheduleNotifier,
	marshalizer marshal.Marshalizer,
	accnts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
) (process.BuiltInFunctionContainer, error) {
	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasSchedule:      gasScheduleNotifier,
		MapDNSAddresses:  make(map[string]struct{}),
		Marshalizer:      marshalizer,
		Accounts:         accnts,
		ShardCoordinator: shardCoordinator,
	}
	builtInFuncFactory, err := builtInFunctions.NewBuiltInFunctionsFactory(argsBuiltIn)
	if err != nil {
		return nil, err
	}

	return builtInFuncFactory.CreateBuiltInFunctionContainer()
}
