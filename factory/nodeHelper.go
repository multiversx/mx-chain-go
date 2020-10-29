package factory

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/facade"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
)

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
	economics process.EconomicsHandler,
	messageSigVerifier vm.MessageSignVerifier,
	nodesSetup sharding.GenesisNodesSetupHandler,
	systemSCConfig *config.SystemSmartContractsConfig,
	rater sharding.ChanceComputer,
	epochNotifier process.EpochNotifier,
) (facade.ApiResolver, error) {
	var vmFactory process.VirtualMachinesContainerFactory
	var err error

	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasMap:          gasSchedule,
		MapDNSAddresses: make(map[string]struct{}),
		Marshalizer:     marshalizer,
		Accounts:        accnts,
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
			rater,
			epochNotifier,
		)
		if err != nil {
			return nil, err
		}
	} else {
		vmFactory, err = shard.NewVMContainerFactory(
			config.VirtualMachineConfig,
			economics.MaxGasLimitPerBlock(shardCoordinator.SelfId()),
			gasSchedule,
			argsHook,
			config.GeneralSettings.SCDeployEnableEpoch,
		)
		if err != nil {
			return nil, err
		}
	}

	vmContainer, err := vmFactory.Create()
	if err != nil {
		return nil, err
	}

	err = builtInFunctions.SetPayableHandler(builtInFuncs, vmFactory.BlockChainHookImpl())
	if err != nil {
		return nil, err
	}

	scQueryService, err := smartContract.NewSCQueryService(vmContainer, economics, vmFactory.BlockChainHookImpl(), blockChain)
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

	return external.NewNodeApiResolverWithContainer(scQueryService, statusMetrics, txCostHandler, vmContainer)
}
