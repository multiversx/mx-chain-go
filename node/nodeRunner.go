package node

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/google/gops/agent"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/closing"
	"github.com/multiversx/mx-chain-core-go/core/throttler"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	outportCore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/api/gin"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/disabled"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/common/goroutines"
	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	dbLookupFactory "github.com/multiversx/mx-chain-go/dblookupext/factory"
	"github.com/multiversx/mx-chain-go/facade"
	"github.com/multiversx/mx-chain-go/facade/initial"
	mainFactory "github.com/multiversx/mx-chain-go/factory"
	apiComp "github.com/multiversx/mx-chain-go/factory/api"
	bootstrapComp "github.com/multiversx/mx-chain-go/factory/bootstrap"
	consensusComp "github.com/multiversx/mx-chain-go/factory/consensus"
	coreComp "github.com/multiversx/mx-chain-go/factory/core"
	cryptoComp "github.com/multiversx/mx-chain-go/factory/crypto"
	dataComp "github.com/multiversx/mx-chain-go/factory/data"
	heartbeatComp "github.com/multiversx/mx-chain-go/factory/heartbeat"
	networkComp "github.com/multiversx/mx-chain-go/factory/network"
	processComp "github.com/multiversx/mx-chain-go/factory/processing"
	stateComp "github.com/multiversx/mx-chain-go/factory/state"
	statusComp "github.com/multiversx/mx-chain-go/factory/status"
	"github.com/multiversx/mx-chain-go/factory/statusCore"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/parsing"
	"github.com/multiversx/mx-chain-go/health"
	"github.com/multiversx/mx-chain-go/node/metrics"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/interceptors"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state/syncer"
	"github.com/multiversx/mx-chain-go/storage/cache"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	trieStatistics "github.com/multiversx/mx-chain-go/trie/statistics"
	"github.com/multiversx/mx-chain-go/update/trigger"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type nextOperationForNode int

const (
	// TODO: remove this after better handling VM versions switching
	// defaultDelayBeforeScQueriesStartInSec represents the default delay before the sc query processor should start to allow external queries
	defaultDelayBeforeScQueriesStartInSec = 120

	maxTimeToClose = 10 * time.Second
	// SoftRestartMessage is the custom message used when the node does a soft restart operation
	SoftRestartMessage = "Shuffled out - soft restart"

	nextOperationShouldRestart nextOperationForNode = 1
	nextOperationShouldStop    nextOperationForNode = 2
)

// nodeRunner holds the node runner configuration and controls running of a node
type nodeRunner struct {
	configs *config.Configs
}

// NewNodeRunner creates a nodeRunner instance
func NewNodeRunner(cfgs *config.Configs) (*nodeRunner, error) {
	if cfgs == nil {
		return nil, fmt.Errorf("nil configs provided")
	}

	return &nodeRunner{
		configs: cfgs,
	}, nil
}

// Start creates and starts the managed components
func (nr *nodeRunner) Start() error {
	configs := nr.configs
	flagsConfig := configs.FlagsConfig
	configurationPaths := configs.ConfigurationPathsHolder
	chanStopNodeProcess := make(chan endProcess.ArgEndProcess, 1)

	enableGopsIfNeeded(flagsConfig.EnableGops)

	var err error
	configurationPaths.Nodes, err = nr.getNodesFileName()
	if err != nil {
		return err
	}

	log.Debug("config", "file", configurationPaths.Nodes)
	log.Debug("config", "file", configurationPaths.Genesis)

	log.Info("starting node", "version", flagsConfig.Version, "pid", os.Getpid())

	err = cleanupStorageIfNecessary(flagsConfig.DbDir, flagsConfig.CleanupStorage)
	if err != nil {
		return err
	}

	printEnableEpochs(nr.configs)

	core.DumpGoRoutinesToLog(0, log)

	err = nr.startShufflingProcessLoop(chanStopNodeProcess)
	if err != nil {
		return err
	}

	return nil
}

func printEnableEpochs(configs *config.Configs) {
	var readEpochFor = func(flag string) string {
		return fmt.Sprintf("read enable epoch for %s", flag)
	}

	enableEpochs := configs.EpochConfig.EnableEpochs

	log.Debug(readEpochFor("sc deploy"), "epoch", enableEpochs.SCDeployEnableEpoch)
	log.Debug(readEpochFor("built in functions"), "epoch", enableEpochs.BuiltInFunctionsEnableEpoch)
	log.Debug(readEpochFor("relayed transactions"), "epoch", enableEpochs.RelayedTransactionsEnableEpoch)
	log.Debug(readEpochFor("penalized too much gas"), "epoch", enableEpochs.PenalizedTooMuchGasEnableEpoch)
	log.Debug(readEpochFor("switch jail waiting"), "epoch", enableEpochs.SwitchJailWaitingEnableEpoch)
	log.Debug(readEpochFor("switch hysteresis for min nodes"), "epoch", enableEpochs.SwitchHysteresisForMinNodesEnableEpoch)
	log.Debug(readEpochFor("below signed threshold"), "epoch", enableEpochs.BelowSignedThresholdEnableEpoch)
	log.Debug(readEpochFor("transaction signed with tx hash"), "epoch", enableEpochs.TransactionSignedWithTxHashEnableEpoch)
	log.Debug(readEpochFor("meta protection"), "epoch", enableEpochs.MetaProtectionEnableEpoch)
	log.Debug(readEpochFor("ahead of time gas usage"), "epoch", enableEpochs.AheadOfTimeGasUsageEnableEpoch)
	log.Debug(readEpochFor("gas price modifier"), "epoch", enableEpochs.GasPriceModifierEnableEpoch)
	log.Debug(readEpochFor("repair callback"), "epoch", enableEpochs.RepairCallbackEnableEpoch)
	log.Debug(readEpochFor("max nodes change"), "epoch", enableEpochs.MaxNodesChangeEnableEpoch)
	log.Debug(readEpochFor("block gas and fees re-check"), "epoch", enableEpochs.BlockGasAndFeesReCheckEnableEpoch)
	log.Debug(readEpochFor("staking v2 epoch"), "epoch", enableEpochs.StakingV2EnableEpoch)
	log.Debug(readEpochFor("stake"), "epoch", enableEpochs.StakeEnableEpoch)
	log.Debug(readEpochFor("double key protection"), "epoch", enableEpochs.DoubleKeyProtectionEnableEpoch)
	log.Debug(readEpochFor("esdt"), "epoch", enableEpochs.ESDTEnableEpoch)
	log.Debug(readEpochFor("governance"), "epoch", enableEpochs.GovernanceEnableEpoch)
	log.Debug(readEpochFor("delegation manager"), "epoch", enableEpochs.DelegationManagerEnableEpoch)
	log.Debug(readEpochFor("delegation smart contract"), "epoch", enableEpochs.DelegationSmartContractEnableEpoch)
	log.Debug(readEpochFor("correct last unjailed"), "epoch", enableEpochs.CorrectLastUnjailedEnableEpoch)
	log.Debug(readEpochFor("balance waiting lists"), "epoch", enableEpochs.BalanceWaitingListsEnableEpoch)
	log.Debug(readEpochFor("relayed transactions v2"), "epoch", enableEpochs.RelayedTransactionsV2EnableEpoch)
	log.Debug(readEpochFor("unbond tokens v2"), "epoch", enableEpochs.UnbondTokensV2EnableEpoch)
	log.Debug(readEpochFor("save jailed always"), "epoch", enableEpochs.SaveJailedAlwaysEnableEpoch)
	log.Debug(readEpochFor("validator to delegation"), "epoch", enableEpochs.ValidatorToDelegationEnableEpoch)
	log.Debug(readEpochFor("re-delegate below minimum check"), "epoch", enableEpochs.ReDelegateBelowMinCheckEnableEpoch)
	log.Debug(readEpochFor("waiting waiting list"), "epoch", enableEpochs.WaitingListFixEnableEpoch)
	log.Debug(readEpochFor("increment SCR nonce in multi transfer"), "epoch", enableEpochs.IncrementSCRNonceInMultiTransferEnableEpoch)
	log.Debug(readEpochFor("esdt and NFT multi transfer"), "epoch", enableEpochs.ESDTMultiTransferEnableEpoch)
	log.Debug(readEpochFor("contract global mint and burn"), "epoch", enableEpochs.GlobalMintBurnDisableEpoch)
	log.Debug(readEpochFor("contract transfer role"), "epoch", enableEpochs.ESDTTransferRoleEnableEpoch)
	log.Debug(readEpochFor("built in functions on metachain"), "epoch", enableEpochs.BuiltInFunctionOnMetaEnableEpoch)
	log.Debug(readEpochFor("compute rewards checkpoint on delegation"), "epoch", enableEpochs.ComputeRewardCheckpointEnableEpoch)
	log.Debug(readEpochFor("esdt NFT create on multiple shards"), "epoch", enableEpochs.ESDTNFTCreateOnMultiShardEnableEpoch)
	log.Debug(readEpochFor("SCR size invariant check"), "epoch", enableEpochs.SCRSizeInvariantCheckEnableEpoch)
	log.Debug(readEpochFor("backward compatibility flag for save key value"), "epoch", enableEpochs.BackwardCompSaveKeyValueEnableEpoch)
	log.Debug(readEpochFor("meta ESDT, financial SFT"), "epoch", enableEpochs.MetaESDTSetEnableEpoch)
	log.Debug(readEpochFor("add tokens to delegation"), "epoch", enableEpochs.AddTokensToDelegationEnableEpoch)
	log.Debug(readEpochFor("multi ESDT transfer on callback"), "epoch", enableEpochs.MultiESDTTransferFixOnCallBackOnEnableEpoch)
	log.Debug(readEpochFor("optimize gas used in cross mini blocks"), "epoch", enableEpochs.OptimizeGasUsedInCrossMiniBlocksEnableEpoch)
	log.Debug(readEpochFor("correct first queued"), "epoch", enableEpochs.CorrectFirstQueuedEpoch)
	log.Debug(readEpochFor("fix out of gas return code"), "epoch", enableEpochs.FixOOGReturnCodeEnableEpoch)
	log.Debug(readEpochFor("remove non updated storage"), "epoch", enableEpochs.RemoveNonUpdatedStorageEnableEpoch)
	log.Debug(readEpochFor("delete delegator data after claim rewards"), "epoch", enableEpochs.DeleteDelegatorAfterClaimRewardsEnableEpoch)
	log.Debug(readEpochFor("optimize nft metadata store"), "epoch", enableEpochs.OptimizeNFTStoreEnableEpoch)
	log.Debug(readEpochFor("create nft through execute on destination by caller"), "epoch", enableEpochs.CreateNFTThroughExecByCallerEnableEpoch)
	log.Debug(readEpochFor("payable by smart contract"), "epoch", enableEpochs.IsPayableBySCEnableEpoch)
	log.Debug(readEpochFor("cleanup informative only SCRs"), "epoch", enableEpochs.CleanUpInformativeSCRsEnableEpoch)
	log.Debug(readEpochFor("storage API cost optimization"), "epoch", enableEpochs.StorageAPICostOptimizationEnableEpoch)
	log.Debug(readEpochFor("transform to multi shard create on esdt"), "epoch", enableEpochs.TransformToMultiShardCreateEnableEpoch)
	log.Debug(readEpochFor("esdt: enable epoch for esdt register and set all roles function"), "epoch", enableEpochs.ESDTRegisterAndSetAllRolesEnableEpoch)
	log.Debug(readEpochFor("scheduled mini blocks"), "epoch", enableEpochs.ScheduledMiniBlocksEnableEpoch)
	log.Debug(readEpochFor("correct jailed not unstaked if empty queue"), "epoch", enableEpochs.CorrectJailedNotUnstakedEmptyQueueEpoch)
	log.Debug(readEpochFor("do not return old block in blockchain hook"), "epoch", enableEpochs.DoNotReturnOldBlockInBlockchainHookEnableEpoch)
	log.Debug(readEpochFor("scr size invariant check on built in"), "epoch", enableEpochs.SCRSizeInvariantOnBuiltInResultEnableEpoch)
	log.Debug(readEpochFor("correct check on tokenID for transfer role"), "epoch", enableEpochs.CheckCorrectTokenIDForTransferRoleEnableEpoch)
	log.Debug(readEpochFor("disable check value on exec by caller"), "epoch", enableEpochs.DisableExecByCallerEnableEpoch)
	log.Debug(readEpochFor("fail execution on every wrong API call"), "epoch", enableEpochs.FailExecutionOnEveryAPIErrorEnableEpoch)
	log.Debug(readEpochFor("managed crypto API in wasm vm"), "epoch", enableEpochs.ManagedCryptoAPIsEnableEpoch)
	log.Debug(readEpochFor("refactor contexts"), "epoch", enableEpochs.RefactorContextEnableEpoch)
	log.Debug(readEpochFor("mini block partial execution"), "epoch", enableEpochs.MiniBlockPartialExecutionEnableEpoch)
	log.Debug(readEpochFor("fix async callback arguments list"), "epoch", enableEpochs.FixAsyncCallBackArgsListEnableEpoch)
	log.Debug(readEpochFor("fix old token liquidity"), "epoch", enableEpochs.FixOldTokenLiquidityEnableEpoch)
	log.Debug(readEpochFor("set sender in eei output transfer"), "epoch", enableEpochs.SetSenderInEeiOutputTransferEnableEpoch)
	log.Debug(readEpochFor("refactor peers mini blocks"), "epoch", enableEpochs.RefactorPeersMiniBlocksEnableEpoch)
	log.Debug(readEpochFor("runtime memstore limit"), "epoch", enableEpochs.RuntimeMemStoreLimitEnableEpoch)
	log.Debug(readEpochFor("max blockchainhook counters"), "epoch", enableEpochs.MaxBlockchainHookCountersEnableEpoch)
	gasSchedule := configs.EpochConfig.GasSchedule

	log.Debug(readEpochFor("gas schedule directories paths"), "epoch", gasSchedule.GasScheduleByEpochs)
}

func (nr *nodeRunner) startShufflingProcessLoop(
	chanStopNodeProcess chan endProcess.ArgEndProcess,
) error {
	for {
		log.Debug("\n\n====================Starting managedComponents creation================================")

		shouldStop, err := nr.executeOneComponentCreationCycle(chanStopNodeProcess)
		if shouldStop {
			return err
		}

		nr.shuffleOutStatsAndGC()
	}
}

func (nr *nodeRunner) shuffleOutStatsAndGC() {
	debugConfig := nr.configs.GeneralConfig.Debug.ShuffleOut

	extraMessage := ""
	if debugConfig.CallGCWhenShuffleOut {
		extraMessage = " before running GC"
	}
	if debugConfig.ExtraPrintsOnShuffleOut {
		log.Debug("node statistics"+extraMessage, statistics.GetRuntimeStatistics()...)
	}
	if debugConfig.CallGCWhenShuffleOut {
		log.Debug("running runtime.GC()")
		runtime.GC()
	}
	shouldPrintAnotherNodeStatistics := debugConfig.CallGCWhenShuffleOut && debugConfig.ExtraPrintsOnShuffleOut
	if shouldPrintAnotherNodeStatistics {
		log.Debug("node statistics after running GC", statistics.GetRuntimeStatistics()...)
	}

	nr.doProfileOnShuffleOut()
}

func (nr *nodeRunner) doProfileOnShuffleOut() {
	debugConfig := nr.configs.GeneralConfig.Debug.ShuffleOut
	shouldDoProfile := debugConfig.DoProfileOnShuffleOut && nr.configs.FlagsConfig.UseHealthService
	if !shouldDoProfile {
		return
	}

	log.Debug("running profile job")
	parentPath := filepath.Join(nr.configs.FlagsConfig.WorkingDir, nr.configs.GeneralConfig.Health.FolderPath)
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	err := health.WriteMemoryUseInfo(stats, time.Now(), parentPath, "softrestart")
	log.LogIfError(err)
}

func (nr *nodeRunner) executeOneComponentCreationCycle(
	chanStopNodeProcess chan endProcess.ArgEndProcess,
) (bool, error) {
	goRoutinesNumberStart := runtime.NumGoroutine()
	configs := nr.configs
	flagsConfig := configs.FlagsConfig
	configurationPaths := configs.ConfigurationPathsHolder

	log.Debug("creating healthService")
	healthService := nr.createHealthService(flagsConfig)

	log.Debug("creating core components")
	managedCoreComponents, err := nr.CreateManagedCoreComponents(
		chanStopNodeProcess,
	)
	if err != nil {
		return true, err
	}

	log.Debug("creating status core components")
	managedStatusCoreComponents, err := nr.CreateManagedStatusCoreComponents(managedCoreComponents)
	if err != nil {
		return true, err
	}

	log.Debug("creating crypto components")
	managedCryptoComponents, err := nr.CreateManagedCryptoComponents(managedCoreComponents)
	if err != nil {
		return true, err
	}

	log.Debug("creating network components")
	managedNetworkComponents, err := nr.CreateManagedNetworkComponents(managedCoreComponents, managedStatusCoreComponents, managedCryptoComponents)
	if err != nil {
		return true, err
	}

	log.Debug("creating disabled API services")
	webServerHandler, err := nr.createHttpServer(managedStatusCoreComponents)
	if err != nil {
		return true, err
	}

	log.Debug("creating bootstrap components")
	managedBootstrapComponents, err := nr.CreateManagedBootstrapComponents(managedStatusCoreComponents, managedCoreComponents, managedCryptoComponents, managedNetworkComponents)
	if err != nil {
		return true, err
	}

	nr.logInformation(managedCoreComponents, managedCryptoComponents, managedBootstrapComponents)

	log.Debug("creating data components")
	managedDataComponents, err := nr.CreateManagedDataComponents(managedStatusCoreComponents, managedCoreComponents, managedBootstrapComponents, managedCryptoComponents)
	if err != nil {
		return true, err
	}

	log.Debug("creating state components")
	managedStateComponents, err := nr.CreateManagedStateComponents(
		managedCoreComponents,
		managedDataComponents,
		managedStatusCoreComponents,
	)
	if err != nil {
		return true, err
	}

	log.Debug("creating metrics")
	// this should be called before setting the storer (done in the managedDataComponents creation)
	err = nr.createMetrics(managedStatusCoreComponents, managedCoreComponents, managedCryptoComponents, managedBootstrapComponents)
	if err != nil {
		return true, err
	}

	log.Debug("registering components in healthService")
	nr.registerDataComponentsInHealthService(healthService, managedDataComponents)

	nodesShufflerOut, err := bootstrapComp.CreateNodesShuffleOut(
		managedCoreComponents.GenesisNodesSetup(),
		configs.GeneralConfig.EpochStartConfig,
		managedCoreComponents.ChanStopNodeProcess(),
	)
	if err != nil {
		return true, err
	}

	bootstrapStorer, err := managedDataComponents.StorageService().GetStorer(dataRetriever.BootstrapUnit)
	if err != nil {
		return true, err
	}

	log.Debug("creating nodes coordinator")
	nodesCoordinatorInstance, err := bootstrapComp.CreateNodesCoordinator(
		nodesShufflerOut,
		managedCoreComponents.GenesisNodesSetup(),
		configs.PreferencesConfig.Preferences,
		managedCoreComponents.EpochStartNotifierWithConfirm(),
		managedCryptoComponents.PublicKey(),
		managedCoreComponents.InternalMarshalizer(),
		managedCoreComponents.Hasher(),
		managedCoreComponents.Rater(),
		bootstrapStorer,
		managedCoreComponents.NodesShuffler(),
		managedBootstrapComponents.ShardCoordinator().SelfId(),
		managedBootstrapComponents.EpochBootstrapParams(),
		managedBootstrapComponents.EpochBootstrapParams().Epoch(),
		managedCoreComponents.ChanStopNodeProcess(),
		managedCoreComponents.NodeTypeProvider(),
		managedCoreComponents.EnableEpochsHandler(),
		managedDataComponents.Datapool().CurrentEpochValidatorInfo(),
		nodesCoordinator.NewIndexHashedNodesCoordinatorWithRaterFactory(),
	)
	if err != nil {
		return true, err
	}

	log.Debug("starting status pooling components")
	managedStatusComponents, err := nr.CreateManagedStatusComponents(
		managedStatusCoreComponents,
		managedCoreComponents,
		managedNetworkComponents,
		managedBootstrapComponents,
		managedStateComponents,
		nodesCoordinatorInstance,
		configs.ImportDbConfig.IsImportDBMode,
	)
	if err != nil {
		return true, err
	}

	argsGasScheduleNotifier := forking.ArgsNewGasScheduleNotifier{
		GasScheduleConfig:  configs.EpochConfig.GasSchedule,
		ConfigDir:          configurationPaths.GasScheduleDirectoryName,
		EpochNotifier:      managedCoreComponents.EpochNotifier(),
		WasmVMChangeLocker: managedCoreComponents.WasmVMChangeLocker(),
	}
	gasScheduleNotifier, err := forking.NewGasScheduleNotifier(argsGasScheduleNotifier)
	if err != nil {
		return true, err
	}

	log.Debug("creating process components")
	managedProcessComponents, err := nr.CreateManagedProcessComponents(
		managedCoreComponents,
		managedCryptoComponents,
		managedNetworkComponents,
		managedBootstrapComponents,
		managedStateComponents,
		managedDataComponents,
		managedStatusComponents,
		managedStatusCoreComponents,
		gasScheduleNotifier,
		nodesCoordinatorInstance,
	)
	if err != nil {
		return true, err
	}

	err = addSyncersToAccountsDB(
		configs.GeneralConfig,
		managedCoreComponents,
		managedDataComponents,
		managedStateComponents,
		managedBootstrapComponents,
		managedProcessComponents,
	)
	if err != nil {
		return true, err
	}

	hardforkTrigger := managedProcessComponents.HardforkTrigger()
	err = hardforkTrigger.AddCloser(nodesShufflerOut)
	if err != nil {
		return true, fmt.Errorf("%w when adding nodeShufflerOut in hardForkTrigger", err)
	}

	err = managedStatusComponents.SetForkDetector(managedProcessComponents.ForkDetector())
	if err != nil {
		return true, err
	}

	err = managedStatusComponents.StartPolling()
	if err != nil {
		return true, err
	}

	log.Debug("starting node... executeOneComponentCreationCycle")

	managedConsensusComponents, err := nr.CreateManagedConsensusComponents(
		managedCoreComponents,
		managedNetworkComponents,
		managedCryptoComponents,
		managedDataComponents,
		managedStateComponents,
		managedStatusComponents,
		managedProcessComponents,
		managedStatusCoreComponents,
	)
	if err != nil {
		return true, err
	}

	managedHeartbeatV2Components, err := nr.CreateManagedHeartbeatV2Components(
		managedBootstrapComponents,
		managedCoreComponents,
		managedNetworkComponents,
		managedCryptoComponents,
		managedDataComponents,
		managedProcessComponents,
		managedStatusCoreComponents,
	)

	if err != nil {
		return true, err
	}

	log.Debug("creating node structure")
	currentNode, err := CreateNode(
		configs.GeneralConfig,
		managedStatusCoreComponents,
		managedBootstrapComponents,
		managedCoreComponents,
		managedCryptoComponents,
		managedDataComponents,
		managedNetworkComponents,
		managedProcessComponents,
		managedStateComponents,
		managedStatusComponents,
		managedHeartbeatV2Components,
		managedConsensusComponents,
		flagsConfig.BootstrapRoundIndex,
		configs.ImportDbConfig.IsImportDBMode,
	)
	if err != nil {
		return true, err
	}

	if managedBootstrapComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		log.Debug("activating nodesCoordinator's validators indexing")
		indexValidatorsListIfNeeded(
			managedStatusComponents.OutportHandler(),
			nodesCoordinatorInstance,
			managedProcessComponents.EpochStartTrigger().Epoch(),
		)
	}

	// this channel will trigger the moment when the sc query service should be able to process VM Query requests
	allowExternalVMQueriesChan := make(chan struct{})

	log.Debug("updating the API service after creating the node facade")
	facadeInstance, err := nr.createApiFacade(currentNode, webServerHandler, gasScheduleNotifier, allowExternalVMQueriesChan)
	if err != nil {
		return true, err
	}

	log.Info("application is now running")

	delayInSecBeforeAllowingVmQueries := configs.GeneralConfig.WebServerAntiflood.VmQueryDelayAfterStartInSec
	if delayInSecBeforeAllowingVmQueries == 0 {
		log.Warn("WebServerAntiflood.VmQueryDelayAfterStartInSec value not set. will use default", "default", defaultDelayBeforeScQueriesStartInSec)
		delayInSecBeforeAllowingVmQueries = defaultDelayBeforeScQueriesStartInSec
	}
	// TODO: remove this and treat better the VM versions switching
	go func(statusHandler core.AppStatusHandler) {
		time.Sleep(time.Duration(delayInSecBeforeAllowingVmQueries) * time.Second)
		close(allowExternalVMQueriesChan)
		statusHandler.SetStringValue(common.MetricAreVMQueriesReady, strconv.FormatBool(true))
	}(managedStatusCoreComponents.AppStatusHandler())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	nextOperation := waitForSignal(
		sigs,
		managedCoreComponents.ChanStopNodeProcess(),
		healthService,
		facadeInstance,
		webServerHandler,
		currentNode,
		goRoutinesNumberStart,
	)

	return nextOperation == nextOperationShouldStop, nil
}

func addSyncersToAccountsDB(
	config *config.Config,
	coreComponents mainFactory.CoreComponentsHolder,
	dataComponents mainFactory.DataComponentsHolder,
	stateComponents mainFactory.StateComponentsHolder,
	bootstrapComponents mainFactory.BootstrapComponentsHolder,
	processComponents mainFactory.ProcessComponentsHolder,
) error {
	selfId := bootstrapComponents.ShardCoordinator().SelfId()
	if selfId == core.MetachainShardId {
		stateSyncer, err := getValidatorAccountSyncer(
			config,
			coreComponents,
			dataComponents,
			stateComponents,
			processComponents,
		)
		if err != nil {
			return err
		}

		err = stateComponents.PeerAccounts().SetSyncer(stateSyncer)
		if err != nil {
			return err
		}

		err = stateComponents.PeerAccounts().StartSnapshotIfNeeded()
		if err != nil {
			return err
		}
	}

	stateSyncer, err := getUserAccountSyncer(
		config,
		coreComponents,
		dataComponents,
		stateComponents,
		bootstrapComponents,
		processComponents,
	)
	if err != nil {
		return err
	}
	err = stateComponents.AccountsAdapter().SetSyncer(stateSyncer)
	if err != nil {
		return err
	}

	return stateComponents.AccountsAdapter().StartSnapshotIfNeeded()
}

func getUserAccountSyncer(
	config *config.Config,
	coreComponents mainFactory.CoreComponentsHolder,
	dataComponents mainFactory.DataComponentsHolder,
	stateComponents mainFactory.StateComponentsHolder,
	bootstrapComponents mainFactory.BootstrapComponentsHolder,
	processComponents mainFactory.ProcessComponentsHolder,
) (process.AccountsDBSyncer, error) {
	maxTrieLevelInMemory := config.StateTriesConfig.MaxStateTrieLevelInMemory
	userTrie := stateComponents.TriesContainer().Get([]byte(dataRetriever.UserAccountsUnit.String()))
	storageManager := userTrie.GetStorageManager()

	thr, err := throttler.NewNumGoRoutinesThrottler(int32(config.TrieSync.NumConcurrentTrieSyncers))
	if err != nil {
		return nil, err
	}

	args := syncer.ArgsNewUserAccountsSyncer{
		ArgsNewBaseAccountsSyncer: getBaseAccountSyncerArgs(
			config,
			coreComponents,
			dataComponents,
			processComponents,
			storageManager,
			maxTrieLevelInMemory,
		),
		ShardId:                bootstrapComponents.ShardCoordinator().SelfId(),
		Throttler:              thr,
		AddressPubKeyConverter: coreComponents.AddressPubKeyConverter(),
	}

	return syncer.NewUserAccountsSyncer(args)
}

func getValidatorAccountSyncer(
	config *config.Config,
	coreComponents mainFactory.CoreComponentsHolder,
	dataComponents mainFactory.DataComponentsHolder,
	stateComponents mainFactory.StateComponentsHolder,
	processComponents mainFactory.ProcessComponentsHolder,
) (process.AccountsDBSyncer, error) {
	maxTrieLevelInMemory := config.StateTriesConfig.MaxPeerTrieLevelInMemory
	peerTrie := stateComponents.TriesContainer().Get([]byte(dataRetriever.PeerAccountsUnit.String()))
	storageManager := peerTrie.GetStorageManager()

	args := syncer.ArgsNewValidatorAccountsSyncer{
		ArgsNewBaseAccountsSyncer: getBaseAccountSyncerArgs(
			config,
			coreComponents,
			dataComponents,
			processComponents,
			storageManager,
			maxTrieLevelInMemory,
		),
	}

	return syncer.NewValidatorAccountsSyncer(args)
}

func getBaseAccountSyncerArgs(
	config *config.Config,
	coreComponents mainFactory.CoreComponentsHolder,
	dataComponents mainFactory.DataComponentsHolder,
	processComponents mainFactory.ProcessComponentsHolder,
	storageManager common.StorageManager,
	maxTrieLevelInMemory uint,
) syncer.ArgsNewBaseAccountsSyncer {
	return syncer.ArgsNewBaseAccountsSyncer{
		Hasher:                            coreComponents.Hasher(),
		Marshalizer:                       coreComponents.InternalMarshalizer(),
		TrieStorageManager:                storageManager,
		RequestHandler:                    processComponents.RequestHandler(),
		Timeout:                           common.TimeoutGettingTrieNodes,
		Cacher:                            dataComponents.Datapool().TrieNodes(),
		MaxTrieLevelInMemory:              maxTrieLevelInMemory,
		MaxHardCapForMissingNodes:         config.TrieSync.MaxHardCapForMissingNodes,
		TrieSyncerVersion:                 config.TrieSync.TrieSyncerVersion,
		CheckNodesOnDisk:                  true,
		UserAccountsSyncStatisticsHandler: trieStatistics.NewTrieSyncStatistics(),
		AppStatusHandler:                  disabled.NewAppStatusHandler(),
		EnableEpochsHandler:               coreComponents.EnableEpochsHandler(),
	}
}

func (nr *nodeRunner) createApiFacade(
	currentNode *Node,
	upgradableHttpServer shared.UpgradeableHttpServerHandler,
	gasScheduleNotifier common.GasScheduleNotifierAPI,
	allowVMQueriesChan chan struct{},
) (closing.Closer, error) {
	configs := nr.configs

	log.Debug("creating api resolver structure")

	apiResolverArgs := &apiComp.ApiResolverArgs{
		Configs:              configs,
		CoreComponents:       currentNode.coreComponents,
		DataComponents:       currentNode.dataComponents,
		StateComponents:      currentNode.stateComponents,
		BootstrapComponents:  currentNode.bootstrapComponents,
		CryptoComponents:     currentNode.cryptoComponents,
		ProcessComponents:    currentNode.processComponents,
		StatusCoreComponents: currentNode.statusCoreComponents,
		GasScheduleNotifier:  gasScheduleNotifier,
		Bootstrapper:         currentNode.consensusComponents.Bootstrapper(),
		AllowVMQueriesChan:   allowVMQueriesChan,
		ChainRunType:         common.ChainRunTypeRegular,
	}

	apiResolver, err := apiComp.CreateApiResolver(apiResolverArgs)
	if err != nil {
		return nil, err
	}

	log.Debug("creating multiversx node facade")

	flagsConfig := configs.FlagsConfig

	argNodeFacade := facade.ArgNodeFacade{
		Node:                   currentNode,
		ApiResolver:            apiResolver,
		RestAPIServerDebugMode: flagsConfig.EnableRestAPIServerDebugMode,
		WsAntifloodConfig:      configs.GeneralConfig.WebServerAntiflood,
		FacadeConfig: config.FacadeConfig{
			RestApiInterface: flagsConfig.RestApiInterface,
			PprofEnabled:     flagsConfig.EnablePprof,
		},
		ApiRoutesConfig: *configs.ApiRoutesConfig,
		AccountsState:   currentNode.stateComponents.AccountsAdapter(),
		PeerState:       currentNode.stateComponents.PeerAccounts(),
		Blockchain:      currentNode.dataComponents.Blockchain(),
	}

	ef, err := facade.NewNodeFacade(argNodeFacade)
	if err != nil {
		return nil, fmt.Errorf("%w while creating NodeFacade", err)
	}

	ef.SetSyncer(currentNode.coreComponents.SyncTimer())

	err = upgradableHttpServer.UpdateFacade(ef)
	if err != nil {
		return nil, err
	}

	log.Debug("updated node facade")

	log.Trace("starting background services")

	return ef, nil
}

func (nr *nodeRunner) createHttpServer(managedStatusCoreComponents mainFactory.StatusCoreComponentsHolder) (shared.UpgradeableHttpServerHandler, error) {
	if check.IfNil(managedStatusCoreComponents) {
		return nil, ErrNilStatusHandler
	}
	initialFacade, err := initial.NewInitialNodeFacade(nr.configs.FlagsConfig.RestApiInterface, nr.configs.FlagsConfig.EnablePprof, managedStatusCoreComponents.StatusMetrics())
	if err != nil {
		return nil, err
	}

	httpServerArgs := gin.ArgsNewWebServer{
		Facade:          initialFacade,
		ApiConfig:       *nr.configs.ApiRoutesConfig,
		AntiFloodConfig: nr.configs.GeneralConfig.WebServerAntiflood,
	}

	httpServerWrapper, err := gin.NewGinWebServerHandler(httpServerArgs)
	if err != nil {
		return nil, err
	}

	err = httpServerWrapper.StartHttpServer()
	if err != nil {
		return nil, err
	}

	return httpServerWrapper, nil
}

func (nr *nodeRunner) createMetrics(
	statusCoreComponents mainFactory.StatusCoreComponentsHolder,
	coreComponents mainFactory.CoreComponentsHolder,
	cryptoComponents mainFactory.CryptoComponentsHolder,
	bootstrapComponents mainFactory.BootstrapComponentsHolder,
) error {
	err := metrics.InitMetrics(
		statusCoreComponents.AppStatusHandler(),
		cryptoComponents.PublicKeyString(),
		bootstrapComponents.NodeType(),
		bootstrapComponents.ShardCoordinator(),
		coreComponents.GenesisNodesSetup(),
		nr.configs.FlagsConfig.Version,
		nr.configs.EconomicsConfig,
		nr.configs.GeneralConfig.EpochStartConfig.RoundsPerEpoch,
		coreComponents.MinTransactionVersion(),
	)

	if err != nil {
		return err
	}

	metrics.SaveStringMetric(statusCoreComponents.AppStatusHandler(), common.MetricNodeDisplayName, nr.configs.PreferencesConfig.Preferences.NodeDisplayName)
	metrics.SaveStringMetric(statusCoreComponents.AppStatusHandler(), common.MetricRedundancyLevel, fmt.Sprintf("%d", nr.configs.PreferencesConfig.Preferences.RedundancyLevel))
	metrics.SaveStringMetric(statusCoreComponents.AppStatusHandler(), common.MetricRedundancyIsMainActive, common.MetricValueNA)
	metrics.SaveStringMetric(statusCoreComponents.AppStatusHandler(), common.MetricChainId, coreComponents.ChainID())
	metrics.SaveUint64Metric(statusCoreComponents.AppStatusHandler(), common.MetricGasPerDataByte, coreComponents.EconomicsData().GasPerDataByte())
	metrics.SaveUint64Metric(statusCoreComponents.AppStatusHandler(), common.MetricMinGasPrice, coreComponents.EconomicsData().MinGasPrice())
	metrics.SaveUint64Metric(statusCoreComponents.AppStatusHandler(), common.MetricMinGasLimit, coreComponents.EconomicsData().MinGasLimit())
	metrics.SaveUint64Metric(statusCoreComponents.AppStatusHandler(), common.MetricExtraGasLimitGuardedTx, coreComponents.EconomicsData().ExtraGasLimitGuardedTx())
	metrics.SaveStringMetric(statusCoreComponents.AppStatusHandler(), common.MetricRewardsTopUpGradientPoint, coreComponents.EconomicsData().RewardsTopUpGradientPoint().String())
	metrics.SaveStringMetric(statusCoreComponents.AppStatusHandler(), common.MetricTopUpFactor, fmt.Sprintf("%g", coreComponents.EconomicsData().RewardsTopUpFactor()))
	metrics.SaveStringMetric(statusCoreComponents.AppStatusHandler(), common.MetricGasPriceModifier, fmt.Sprintf("%g", coreComponents.EconomicsData().GasPriceModifier()))
	metrics.SaveUint64Metric(statusCoreComponents.AppStatusHandler(), common.MetricMaxGasPerTransaction, coreComponents.EconomicsData().MaxGasLimitPerTx())
	if nr.configs.PreferencesConfig.Preferences.FullArchive {
		metrics.SaveStringMetric(statusCoreComponents.AppStatusHandler(), common.MetricPeerType, core.ObserverPeer.String())
		metrics.SaveStringMetric(statusCoreComponents.AppStatusHandler(), common.MetricPeerSubType, core.FullHistoryObserver.String())
	}

	return nil
}

func (nr *nodeRunner) createHealthService(flagsConfig *config.ContextFlagsConfig) HealthService {
	healthService := health.NewHealthService(nr.configs.GeneralConfig.Health, flagsConfig.WorkingDir)
	if flagsConfig.UseHealthService {
		healthService.Start()
	}

	return healthService
}

func (nr *nodeRunner) registerDataComponentsInHealthService(healthService HealthService, dataComponents mainFactory.DataComponentsHolder) {
	healthService.RegisterComponent(dataComponents.Datapool().Transactions())
	healthService.RegisterComponent(dataComponents.Datapool().UnsignedTransactions())
	healthService.RegisterComponent(dataComponents.Datapool().RewardTransactions())
}

// CreateManagedConsensusComponents is the managed consensus components factory
func (nr *nodeRunner) CreateManagedConsensusComponents(
	coreComponents mainFactory.CoreComponentsHolder,
	networkComponents mainFactory.NetworkComponentsHolder,
	cryptoComponents mainFactory.CryptoComponentsHolder,
	dataComponents mainFactory.DataComponentsHolder,
	stateComponents mainFactory.StateComponentsHolder,
	statusComponents mainFactory.StatusComponentsHolder,
	processComponents mainFactory.ProcessComponentsHolder,
	statusCoreComponents mainFactory.StatusCoreComponentsHolder,
) (mainFactory.ConsensusComponentsHandler, error) {
	scheduledProcessorArgs := spos.ScheduledProcessorWrapperArgs{
		SyncTimer:                coreComponents.SyncTimer(),
		Processor:                processComponents.BlockProcessor(),
		RoundTimeDurationHandler: coreComponents.RoundHandler(),
	}

	scheduledProcessor, err := spos.NewScheduledProcessorWrapper(scheduledProcessorArgs)
	if err != nil {
		return nil, err
	}

	consensusArgs := consensusComp.ConsensusComponentsFactoryArgs{
		Config:                *nr.configs.GeneralConfig,
		FlagsConfig:           *nr.configs.FlagsConfig,
		BootstrapRoundIndex:   nr.configs.FlagsConfig.BootstrapRoundIndex,
		CoreComponents:        coreComponents,
		NetworkComponents:     networkComponents,
		CryptoComponents:      cryptoComponents,
		DataComponents:        dataComponents,
		ProcessComponents:     processComponents,
		StateComponents:       stateComponents,
		StatusComponents:      statusComponents,
		StatusCoreComponents:  statusCoreComponents,
		ScheduledProcessor:    scheduledProcessor,
		IsInImportMode:        nr.configs.ImportDbConfig.IsImportDBMode,
		ShouldDisableWatchdog: nr.configs.FlagsConfig.DisableConsensusWatchdog,
		ConsensusModel:        consensus.ConsensusModelV1,
		ChainRunType:          common.ChainRunTypeRegular,
	}

	consensusFactory, err := consensusComp.NewConsensusComponentsFactory(consensusArgs)
	if err != nil {
		return nil, fmt.Errorf("NewConsensusComponentsFactory failed: %w", err)
	}

	managedConsensusComponents, err := consensusComp.NewManagedConsensusComponents(consensusFactory)
	if err != nil {
		return nil, err
	}

	err = managedConsensusComponents.Create()
	if err != nil {
		return nil, err
	}
	return managedConsensusComponents, nil
}

// CreateManagedHeartbeatV2Components is the managed heartbeatV2 components factory
func (nr *nodeRunner) CreateManagedHeartbeatV2Components(
	bootstrapComponents mainFactory.BootstrapComponentsHolder,
	coreComponents mainFactory.CoreComponentsHolder,
	networkComponents mainFactory.NetworkComponentsHolder,
	cryptoComponents mainFactory.CryptoComponentsHolder,
	dataComponents mainFactory.DataComponentsHolder,
	processComponents mainFactory.ProcessComponentsHolder,
	statusCoreComponents mainFactory.StatusCoreComponentsHolder,
) (mainFactory.HeartbeatV2ComponentsHandler, error) {
	heartbeatV2Args := heartbeatComp.ArgHeartbeatV2ComponentsFactory{
		Config:               *nr.configs.GeneralConfig,
		Prefs:                *nr.configs.PreferencesConfig,
		BaseVersion:          nr.configs.FlagsConfig.BaseVersion,
		AppVersion:           nr.configs.FlagsConfig.Version,
		BootstrapComponents:  bootstrapComponents,
		CoreComponents:       coreComponents,
		DataComponents:       dataComponents,
		NetworkComponents:    networkComponents,
		CryptoComponents:     cryptoComponents,
		ProcessComponents:    processComponents,
		StatusCoreComponents: statusCoreComponents,
	}

	heartbeatV2ComponentsFactory, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(heartbeatV2Args)
	if err != nil {
		return nil, fmt.Errorf("NewHeartbeatV2ComponentsFactory failed: %w", err)
	}

	managedHeartbeatV2Components, err := heartbeatComp.NewManagedHeartbeatV2Components(heartbeatV2ComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedHeartbeatV2Components.Create()
	if err != nil {
		return nil, err
	}
	return managedHeartbeatV2Components, nil
}

func waitForSignal(
	sigs chan os.Signal,
	chanStopNodeProcess chan endProcess.ArgEndProcess,
	healthService closing.Closer,
	facade closing.Closer,
	httpServer shared.UpgradeableHttpServerHandler,
	currentNode *Node,
	goRoutinesNumberStart int,
) nextOperationForNode {
	var sig endProcess.ArgEndProcess
	reshuffled := false
	wrongConfig := false
	wrongConfigDescription := ""

	select {
	case <-sigs:
		log.Info("terminating at user's signal...")
	case sig = <-chanStopNodeProcess:
		log.Info("terminating at internal stop signal", "reason", sig.Reason, "description", sig.Description)
		if sig.Reason == common.ShuffledOut {
			reshuffled = true
		}
		if sig.Reason == common.WrongConfiguration {
			wrongConfig = true
			wrongConfigDescription = sig.Description
		}
	}

	chanCloseComponents := make(chan struct{})
	go func() {
		closeAllComponents(healthService, facade, httpServer, currentNode, chanCloseComponents)
	}()

	select {
	case <-chanCloseComponents:
		log.Debug("Closed all components gracefully")
	case <-time.After(maxTimeToClose):
		log.Warn("force closing the node",
			"error", "closeAllComponents did not finish on time",
			"stack", goroutines.GetGoRoutines())

		return nextOperationShouldStop
	}

	if wrongConfig {
		// hang the node's process because it cannot continue with the current configuration and a restart doesn't
		// change this behaviour
		for {
			log.Error("wrong configuration. stopped the processing and left the node unclosed", "description", wrongConfigDescription)
			time.Sleep(1 * time.Minute)
		}
	}

	if reshuffled {
		log.Info("=============================" + SoftRestartMessage + "==================================")
		core.DumpGoRoutinesToLog(goRoutinesNumberStart, log)

		return nextOperationShouldRestart
	}

	return nextOperationShouldStop
}

func (nr *nodeRunner) logInformation(
	coreComponents mainFactory.CoreComponentsHolder,
	cryptoComponents mainFactory.CryptoComponentsHolder,
	bootstrapComponents mainFactory.BootstrapComponentsHolder,
) {
	log.Info("Bootstrap", "epoch", bootstrapComponents.EpochBootstrapParams().Epoch())
	if bootstrapComponents.EpochBootstrapParams().NodesConfig() != nil {
		log.Info("the epoch from nodesConfig is",
			"epoch", bootstrapComponents.EpochBootstrapParams().NodesConfig().CurrentEpoch)
	}

	var shardIdString = core.GetShardIDString(bootstrapComponents.ShardCoordinator().SelfId())
	logger.SetCorrelationShard(shardIdString)

	sessionInfoFileOutput := fmt.Sprintf("%s:%s\n%s:%s\n%s:%v\n%s:%s\n%s:%v\n",
		"PkBlockSign", cryptoComponents.PublicKeyString(),
		"ShardId", shardIdString,
		"TotalShards", bootstrapComponents.ShardCoordinator().NumberOfShards(),
		"AppVersion", nr.configs.FlagsConfig.Version,
		"GenesisTimeStamp", coreComponents.GenesisTime().Unix(),
	)

	sessionInfoFileOutput += "\nStarted with parameters:\n"
	sessionInfoFileOutput += nr.configs.FlagsConfig.SessionInfoFileOutput

	nr.logSessionInformation(nr.configs.FlagsConfig.WorkingDir, sessionInfoFileOutput, coreComponents)
}

func (nr *nodeRunner) getNodesFileName() (string, error) {
	flagsConfig := nr.configs.FlagsConfig
	configurationPaths := nr.configs.ConfigurationPathsHolder
	nodesFileName := configurationPaths.Nodes

	exportFolder := filepath.Join(flagsConfig.WorkingDir, nr.configs.GeneralConfig.Hardfork.ImportFolder)
	if nr.configs.GeneralConfig.Hardfork.AfterHardFork {
		exportFolderNodesSetupPath := filepath.Join(exportFolder, common.NodesSetupJsonFileName)
		if !core.FileExists(exportFolderNodesSetupPath) {
			return "", fmt.Errorf("cannot find %s in the export folder", common.NodesSetupJsonFileName)
		}

		nodesFileName = exportFolderNodesSetupPath
	}
	return nodesFileName, nil
}

// CreateManagedStatusComponents is the managed status components factory
func (nr *nodeRunner) CreateManagedStatusComponents(
	managedStatusCoreComponents mainFactory.StatusCoreComponentsHolder,
	managedCoreComponents mainFactory.CoreComponentsHolder,
	managedNetworkComponents mainFactory.NetworkComponentsHolder,
	managedBootstrapComponents mainFactory.BootstrapComponentsHolder,
	managedStateComponents mainFactory.StateComponentsHolder,
	nodesCoordinator nodesCoordinator.NodesCoordinator,
	isInImportMode bool,
) (mainFactory.StatusComponentsHandler, error) {
	statArgs := statusComp.StatusComponentsFactoryArgs{
		Config:               *nr.configs.GeneralConfig,
		ExternalConfig:       *nr.configs.ExternalConfig,
		EconomicsConfig:      *nr.configs.EconomicsConfig,
		ShardCoordinator:     managedBootstrapComponents.ShardCoordinator(),
		NodesCoordinator:     nodesCoordinator,
		EpochStartNotifier:   managedCoreComponents.EpochStartNotifierWithConfirm(),
		CoreComponents:       managedCoreComponents,
		NetworkComponents:    managedNetworkComponents,
		StateComponents:      managedStateComponents,
		IsInImportMode:       isInImportMode,
		StatusCoreComponents: managedStatusCoreComponents,
	}

	statusComponentsFactory, err := statusComp.NewStatusComponentsFactory(statArgs)
	if err != nil {
		return nil, fmt.Errorf("NewStatusComponentsFactory failed: %w", err)
	}

	managedStatusComponents, err := statusComp.NewManagedStatusComponents(statusComponentsFactory)
	if err != nil {
		return nil, err
	}
	err = managedStatusComponents.Create()
	if err != nil {
		return nil, err
	}
	return managedStatusComponents, nil
}

func (nr *nodeRunner) logSessionInformation(
	workingDir string,
	sessionInfoFileOutput string,
	coreComponents mainFactory.CoreComponentsHolder,
) {
	statsFolder := filepath.Join(workingDir, common.DefaultStatsPath)
	configurationPaths := nr.configs.ConfigurationPathsHolder
	copyConfigToStatsFolder(
		statsFolder,
		configurationPaths.GasScheduleDirectoryName,
		[]string{
			configurationPaths.ApiRoutes,
			configurationPaths.MainConfig,
			configurationPaths.Economics,
			configurationPaths.Epoch,
			configurationPaths.RoundActivation,
			configurationPaths.External,
			configurationPaths.Genesis,
			configurationPaths.SmartContracts,
			configurationPaths.Nodes,
			configurationPaths.P2p,
			configurationPaths.Preferences,
			configurationPaths.Ratings,
			configurationPaths.SystemSC,
		})

	statsFile := filepath.Join(statsFolder, "session.info")
	err := ioutil.WriteFile(statsFile, []byte(sessionInfoFileOutput), core.FileModeReadWrite)
	log.LogIfError(err)

	computedRatingsDataStr := createStringFromRatingsData(coreComponents.RatingsData())
	log.Debug("rating data", "rating", computedRatingsDataStr)
}

// CreateManagedProcessComponents is the managed process components factory
func (nr *nodeRunner) CreateManagedProcessComponents(
	coreComponents mainFactory.CoreComponentsHolder,
	cryptoComponents mainFactory.CryptoComponentsHolder,
	networkComponents mainFactory.NetworkComponentsHolder,
	bootstrapComponents mainFactory.BootstrapComponentsHolder,
	stateComponents mainFactory.StateComponentsHolder,
	dataComponents mainFactory.DataComponentsHolder,
	statusComponents mainFactory.StatusComponentsHolder,
	statusCoreComponents mainFactory.StatusCoreComponentsHolder,
	gasScheduleNotifier core.GasScheduleNotifier,
	nodesCoordinator nodesCoordinator.NodesCoordinator,
) (mainFactory.ProcessComponentsHandler, error) {
	configs := nr.configs
	configurationPaths := nr.configs.ConfigurationPathsHolder
	importStartHandler, err := trigger.NewImportStartHandler(filepath.Join(configs.FlagsConfig.DbDir, common.DefaultDBPath), configs.FlagsConfig.Version)
	if err != nil {
		return nil, err
	}

	totalSupply, ok := big.NewInt(0).SetString(configs.EconomicsConfig.GlobalSettings.GenesisTotalSupply, 10)
	if !ok {
		return nil, fmt.Errorf("can not parse total suply from economics.toml, %s is not a valid value",
			configs.EconomicsConfig.GlobalSettings.GenesisTotalSupply)
	}

	mintingSenderAddress := configs.EconomicsConfig.GlobalSettings.GenesisMintingSenderAddress

	args := genesis.AccountsParserArgs{
		GenesisFilePath: configurationPaths.Genesis,
		EntireSupply:    totalSupply,
		MinterAddress:   mintingSenderAddress,
		PubkeyConverter: coreComponents.AddressPubKeyConverter(),
		KeyGenerator:    cryptoComponents.TxSignKeyGen(),
		Hasher:          coreComponents.Hasher(),
		Marshalizer:     coreComponents.InternalMarshalizer(),
	}

	accountsParser, err := parsing.NewAccountsParser(args)
	if err != nil {
		return nil, err
	}

	smartContractParser, err := parsing.NewSmartContractsParser(
		configurationPaths.SmartContracts,
		coreComponents.AddressPubKeyConverter(),
		cryptoComponents.TxSignKeyGen(),
	)
	if err != nil {
		return nil, err
	}

	historyRepoFactoryArgs := &dbLookupFactory.ArgsHistoryRepositoryFactory{
		SelfShardID:              bootstrapComponents.ShardCoordinator().SelfId(),
		Config:                   configs.GeneralConfig.DbLookupExtensions,
		Hasher:                   coreComponents.Hasher(),
		Marshalizer:              coreComponents.InternalMarshalizer(),
		Store:                    dataComponents.StorageService(),
		Uint64ByteSliceConverter: coreComponents.Uint64ByteSliceConverter(),
	}
	historyRepositoryFactory, err := dbLookupFactory.NewHistoryRepositoryFactory(historyRepoFactoryArgs)
	if err != nil {
		return nil, err
	}

	whiteListCache, err := storageunit.NewCache(storageFactory.GetCacherFromConfig(configs.GeneralConfig.WhiteListPool))
	if err != nil {
		return nil, err
	}
	whiteListRequest, err := interceptors.NewWhiteListDataVerifier(whiteListCache)
	if err != nil {
		return nil, err
	}

	whiteListerVerifiedTxs, err := createWhiteListerVerifiedTxs(configs.GeneralConfig)
	if err != nil {
		return nil, err
	}

	historyRepository, err := historyRepositoryFactory.Create()
	if err != nil {
		return nil, err
	}

	log.Trace("creating time cache for requested items components")
	// TODO consider lowering this (perhaps to 1 second) and use a common const
	requestedItemsHandler := cache.NewTimeCache(
		time.Duration(uint64(time.Millisecond) * coreComponents.GenesisNodesSetup().GetRoundDuration()))

	processArgs := processComp.ProcessComponentsFactoryArgs{
		Config:                 *configs.GeneralConfig,
		EpochConfig:            *configs.EpochConfig,
		PrefConfigs:            *configs.PreferencesConfig,
		ImportDBConfig:         *configs.ImportDbConfig,
		AccountsParser:         accountsParser,
		SmartContractParser:    smartContractParser,
		GasSchedule:            gasScheduleNotifier,
		NodesCoordinator:       nodesCoordinator,
		Data:                   dataComponents,
		CoreData:               coreComponents,
		Crypto:                 cryptoComponents,
		State:                  stateComponents,
		Network:                networkComponents,
		BootstrapComponents:    bootstrapComponents,
		StatusComponents:       statusComponents,
		StatusCoreComponents:   statusCoreComponents,
		RequestedItemsHandler:  requestedItemsHandler,
		WhiteListHandler:       whiteListRequest,
		WhiteListerVerifiedTxs: whiteListerVerifiedTxs,
		MaxRating:              configs.RatingsConfig.General.MaxRating,
		SystemSCConfig:         configs.SystemSCConfig,
		ImportStartHandler:     importStartHandler,
		HistoryRepo:            historyRepository,
		FlagsConfig:            *configs.FlagsConfig,
		ChainRunType:           common.ChainRunTypeRegular,
	}
	processComponentsFactory, err := processComp.NewProcessComponentsFactory(processArgs)
	if err != nil {
		return nil, fmt.Errorf("NewProcessComponentsFactory failed: %w", err)
	}

	managedProcessComponents, err := processComp.NewManagedProcessComponents(processComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedProcessComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedProcessComponents, nil
}

// CreateManagedDataComponents is the managed data components factory
func (nr *nodeRunner) CreateManagedDataComponents(
	statusCoreComponents mainFactory.StatusCoreComponentsHolder,
	coreComponents mainFactory.CoreComponentsHolder,
	bootstrapComponents mainFactory.BootstrapComponentsHolder,
	crypto mainFactory.CryptoComponentsHolder,
) (mainFactory.DataComponentsHandler, error) {
	configs := nr.configs
	storerEpoch := bootstrapComponents.EpochBootstrapParams().Epoch()
	if !configs.GeneralConfig.StoragePruning.Enabled {
		// TODO: refactor this as when the pruning storer is disabled, the default directory path is Epoch_0
		// and it should be Epoch_ALL or something similar
		storerEpoch = 0
	}

	dataArgs := dataComp.DataComponentsFactoryArgs{
		Config:                        *configs.GeneralConfig,
		PrefsConfig:                   configs.PreferencesConfig.Preferences,
		ShardCoordinator:              bootstrapComponents.ShardCoordinator(),
		Core:                          coreComponents,
		StatusCore:                    statusCoreComponents,
		Crypto:                        crypto,
		CurrentEpoch:                  storerEpoch,
		CreateTrieEpochRootHashStorer: configs.ImportDbConfig.ImportDbSaveTrieEpochRootHash,
		FlagsConfigs:                  *configs.FlagsConfig,
		NodeProcessingMode:            common.GetNodeProcessingMode(nr.configs.ImportDbConfig),
		ChainRunType:                  common.ChainRunTypeRegular,
	}

	dataComponentsFactory, err := dataComp.NewDataComponentsFactory(dataArgs)
	if err != nil {
		return nil, fmt.Errorf("NewDataComponentsFactory failed: %w", err)
	}
	managedDataComponents, err := dataComp.NewManagedDataComponents(dataComponentsFactory)
	if err != nil {
		return nil, err
	}
	err = managedDataComponents.Create()
	if err != nil {
		return nil, err
	}

	statusMetricsStorer, err := managedDataComponents.StorageService().GetStorer(dataRetriever.StatusMetricsUnit)
	if err != nil {
		return nil, err
	}

	err = statusCoreComponents.PersistentStatusHandler().SetStorage(statusMetricsStorer)

	if err != nil {
		return nil, err
	}

	return managedDataComponents, nil
}

// CreateManagedStateComponents is the managed state components factory
func (nr *nodeRunner) CreateManagedStateComponents(
	coreComponents mainFactory.CoreComponentsHolder,
	dataComponents mainFactory.DataComponentsHandler,
	statusCoreComponents mainFactory.StatusCoreComponentsHolder,
) (mainFactory.StateComponentsHandler, error) {
	stateArgs := stateComp.StateComponentsFactoryArgs{
		Config:                   *nr.configs.GeneralConfig,
		Core:                     coreComponents,
		StatusCore:               statusCoreComponents,
		StorageService:           dataComponents.StorageService(),
		ProcessingMode:           common.GetNodeProcessingMode(nr.configs.ImportDbConfig),
		ShouldSerializeSnapshots: nr.configs.FlagsConfig.SerializeSnapshots,
		SnapshotsEnabled:         nr.configs.FlagsConfig.SnapshotsEnabled,
		ChainHandler:             dataComponents.Blockchain(),
	}

	stateComponentsFactory, err := stateComp.NewStateComponentsFactory(stateArgs)
	if err != nil {
		return nil, fmt.Errorf("NewStateComponentsFactory failed: %w", err)
	}

	managedStateComponents, err := stateComp.NewManagedStateComponents(stateComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedStateComponents.Create()
	if err != nil {
		return nil, err
	}
	return managedStateComponents, nil
}

// CreateManagedBootstrapComponents is the managed bootstrap components factory
func (nr *nodeRunner) CreateManagedBootstrapComponents(
	statusCoreComponents mainFactory.StatusCoreComponentsHolder,
	coreComponents mainFactory.CoreComponentsHolder,
	cryptoComponents mainFactory.CryptoComponentsHolder,
	networkComponents mainFactory.NetworkComponentsHolder,
) (mainFactory.BootstrapComponentsHandler, error) {

	bootstrapComponentsFactoryArgs := bootstrapComp.BootstrapComponentsFactoryArgs{
		Config:                           *nr.configs.GeneralConfig,
		PrefConfig:                       *nr.configs.PreferencesConfig,
		ImportDbConfig:                   *nr.configs.ImportDbConfig,
		FlagsConfig:                      *nr.configs.FlagsConfig,
		WorkingDir:                       nr.configs.FlagsConfig.DbDir,
		CoreComponents:                   coreComponents,
		CryptoComponents:                 cryptoComponents,
		NetworkComponents:                networkComponents,
		StatusCoreComponents:             statusCoreComponents,
		ChainRunType:                     common.ChainRunTypeRegular,
		NodesCoordinatorWithRaterFactory: nodesCoordinator.NewIndexHashedNodesCoordinatorWithRaterFactory(),
	}

	bootstrapComponentsFactory, err := bootstrapComp.NewBootstrapComponentsFactory(bootstrapComponentsFactoryArgs)
	if err != nil {
		return nil, fmt.Errorf("NewBootstrapComponentsFactory failed: %w", err)
	}

	managedBootstrapComponents, err := bootstrapComp.NewManagedBootstrapComponents(bootstrapComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedBootstrapComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedBootstrapComponents, nil
}

// CreateManagedNetworkComponents is the managed network components factory
func (nr *nodeRunner) CreateManagedNetworkComponents(
	coreComponents mainFactory.CoreComponentsHolder,
	statusCoreComponents mainFactory.StatusCoreComponentsHolder,
	cryptoComponents mainFactory.CryptoComponentsHolder,
) (mainFactory.NetworkComponentsHandler, error) {
	networkComponentsFactoryArgs := networkComp.NetworkComponentsFactoryArgs{
		P2pConfig:             *nr.configs.P2pConfig,
		MainConfig:            *nr.configs.GeneralConfig,
		RatingsConfig:         *nr.configs.RatingsConfig,
		StatusHandler:         statusCoreComponents.AppStatusHandler(),
		Marshalizer:           coreComponents.InternalMarshalizer(),
		Syncer:                coreComponents.SyncTimer(),
		PreferredPeersSlices:  nr.configs.PreferencesConfig.Preferences.PreferredConnections,
		BootstrapWaitTime:     common.TimeToWaitForP2PBootstrap,
		NodeOperationMode:     p2p.NormalOperation,
		ConnectionWatcherType: nr.configs.PreferencesConfig.Preferences.ConnectionWatcherType,
		CryptoComponents:      cryptoComponents,
	}
	if nr.configs.ImportDbConfig.IsImportDBMode {
		networkComponentsFactoryArgs.BootstrapWaitTime = 0
	}
	if nr.configs.PreferencesConfig.Preferences.FullArchive {
		networkComponentsFactoryArgs.NodeOperationMode = p2p.FullArchiveMode
	}

	networkComponentsFactory, err := networkComp.NewNetworkComponentsFactory(networkComponentsFactoryArgs)
	if err != nil {
		return nil, fmt.Errorf("NewNetworkComponentsFactory failed: %w", err)
	}

	managedNetworkComponents, err := networkComp.NewManagedNetworkComponents(networkComponentsFactory)
	if err != nil {
		return nil, err
	}
	err = managedNetworkComponents.Create()
	if err != nil {
		return nil, err
	}
	return managedNetworkComponents, nil
}

// CreateManagedCoreComponents is the managed core components factory
func (nr *nodeRunner) CreateManagedCoreComponents(
	chanStopNodeProcess chan endProcess.ArgEndProcess,
) (mainFactory.CoreComponentsHandler, error) {
	coreArgs := coreComp.CoreComponentsFactoryArgs{
		Config:                   *nr.configs.GeneralConfig,
		ConfigPathsHolder:        *nr.configs.ConfigurationPathsHolder,
		EpochConfig:              *nr.configs.EpochConfig,
		RoundConfig:              *nr.configs.RoundConfig,
		ImportDbConfig:           *nr.configs.ImportDbConfig,
		RatingsConfig:            *nr.configs.RatingsConfig,
		EconomicsConfig:          *nr.configs.EconomicsConfig,
		NodesFilename:            nr.configs.ConfigurationPathsHolder.Nodes,
		WorkingDirectory:         nr.configs.FlagsConfig.DbDir,
		ChanStopNodeProcess:      chanStopNodeProcess,
		GenesisNodesSetupFactory: sharding.NewGenesisNodesSetupFactory(),
	}

	coreComponentsFactory, err := coreComp.NewCoreComponentsFactory(coreArgs)
	if err != nil {
		return nil, fmt.Errorf("NewCoreComponentsFactory failed: %w", err)
	}

	managedCoreComponents, err := coreComp.NewManagedCoreComponents(coreComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedCoreComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedCoreComponents, nil
}

// CreateManagedStatusCoreComponents is the managed status core components factory
func (nr *nodeRunner) CreateManagedStatusCoreComponents(
	coreComponents mainFactory.CoreComponentsHolder,
) (mainFactory.StatusCoreComponentsHandler, error) {
	args := statusCore.StatusCoreComponentsFactoryArgs{
		Config:          *nr.configs.GeneralConfig,
		EpochConfig:     *nr.configs.EpochConfig,
		RoundConfig:     *nr.configs.RoundConfig,
		RatingsConfig:   *nr.configs.RatingsConfig,
		EconomicsConfig: *nr.configs.EconomicsConfig,
		CoreComp:        coreComponents,
	}

	statusCoreComponentsFactory, err := statusCore.NewStatusCoreComponentsFactory(args)
	if err != nil {
		return nil, err
	}
	managedStatusCoreComponents, err := statusCore.NewManagedStatusCoreComponents(statusCoreComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedStatusCoreComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedStatusCoreComponents, nil
}

// CreateManagedCryptoComponents is the managed crypto components factory
func (nr *nodeRunner) CreateManagedCryptoComponents(
	coreComponents mainFactory.CoreComponentsHolder,
) (mainFactory.CryptoComponentsHandler, error) {
	configs := nr.configs
	validatorKeyPemFileName := configs.ConfigurationPathsHolder.ValidatorKey
	allValidatorKeysPemFileName := configs.ConfigurationPathsHolder.AllValidatorKeys
	cryptoComponentsHandlerArgs := cryptoComp.CryptoComponentsFactoryArgs{
		ValidatorKeyPemFileName:              validatorKeyPemFileName,
		AllValidatorKeysPemFileName:          allValidatorKeysPemFileName,
		SkIndex:                              configs.FlagsConfig.ValidatorKeyIndex,
		Config:                               *configs.GeneralConfig,
		CoreComponentsHolder:                 coreComponents,
		ActivateBLSPubKeyMessageVerification: configs.SystemSCConfig.StakingSystemSCConfig.ActivateBLSPubKeyMessageVerification,
		KeyLoader:                            core.NewKeyLoader(),
		ImportModeNoSigCheck:                 configs.ImportDbConfig.ImportDbNoSigCheckFlag,
		IsInImportMode:                       configs.ImportDbConfig.IsImportDBMode,
		EnableEpochs:                         configs.EpochConfig.EnableEpochs,
		NoKeyProvided:                        configs.FlagsConfig.NoKeyProvided,
		P2pKeyPemFileName:                    configs.ConfigurationPathsHolder.P2pKey,
	}

	cryptoComponentsFactory, err := cryptoComp.NewCryptoComponentsFactory(cryptoComponentsHandlerArgs)
	if err != nil {
		return nil, fmt.Errorf("NewCryptoComponentsFactory failed: %w", err)
	}

	managedCryptoComponents, err := cryptoComp.NewManagedCryptoComponents(cryptoComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedCryptoComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedCryptoComponents, nil
}

func closeAllComponents(
	healthService io.Closer,
	facade mainFactory.Closer,
	httpServer shared.UpgradeableHttpServerHandler,
	node *Node,
	chanCloseComponents chan struct{},
) {
	log.Debug("closing health service...")
	err := healthService.Close()
	log.LogIfError(err)

	log.Debug("closing http server")
	log.LogIfError(httpServer.Close())

	log.Debug("closing facade")
	log.LogIfError(facade.Close())

	log.Debug("closing node")
	log.LogIfError(node.Close())

	chanCloseComponents <- struct{}{}
}

func createStringFromRatingsData(ratingsData process.RatingsInfoHandler) string {
	metaChainStepHandler := ratingsData.MetaChainRatingsStepHandler()
	shardChainHandler := ratingsData.ShardChainRatingsStepHandler()
	computedRatingsDataStr := fmt.Sprintf(
		"meta:\n"+
			"ProposerIncrease=%v\n"+
			"ProposerDecrease=%v\n"+
			"ValidatorIncrease=%v\n"+
			"ValidatorDecrease=%v\n\n"+
			"shard:\n"+
			"ProposerIncrease=%v\n"+
			"ProposerDecrease=%v\n"+
			"ValidatorIncrease=%v\n"+
			"ValidatorDecrease=%v",
		metaChainStepHandler.ProposerIncreaseRatingStep(),
		metaChainStepHandler.ProposerDecreaseRatingStep(),
		metaChainStepHandler.ValidatorIncreaseRatingStep(),
		metaChainStepHandler.ValidatorDecreaseRatingStep(),
		shardChainHandler.ProposerIncreaseRatingStep(),
		shardChainHandler.ProposerDecreaseRatingStep(),
		shardChainHandler.ValidatorIncreaseRatingStep(),
		shardChainHandler.ValidatorDecreaseRatingStep(),
	)
	return computedRatingsDataStr
}

func cleanupStorageIfNecessary(workingDir string, cleanupStorage bool) error {
	if !cleanupStorage {
		return nil
	}

	dbPath := filepath.Join(
		workingDir,
		common.DefaultDBPath)
	log.Trace("cleaning storage", "path", dbPath)

	return os.RemoveAll(dbPath)
}

func copyConfigToStatsFolder(statsFolder string, gasScheduleDirectory string, configs []string) {
	err := os.MkdirAll(statsFolder, os.ModePerm)
	log.LogIfError(err)

	newGasScheduleDirectory := path.Join(statsFolder, filepath.Base(gasScheduleDirectory))
	err = copyDirectory(gasScheduleDirectory, newGasScheduleDirectory)
	log.LogIfError(err)

	for _, configFile := range configs {
		copySingleFile(statsFolder, configFile)
	}
}

func copyDirectory(source string, destination string) error {
	fileDescriptors, err := ioutil.ReadDir(source)
	if err != nil {
		return err
	}

	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	err = os.MkdirAll(destination, sourceInfo.Mode())
	if err != nil {
		return err
	}

	for _, fd := range fileDescriptors {
		srcFilePath := path.Join(source, fd.Name())
		if fd.IsDir() {
			dstFilePath := path.Join(destination, filepath.Base(srcFilePath))
			err = copyDirectory(srcFilePath, dstFilePath)
			log.LogIfError(err)
		} else {
			copySingleFile(destination, srcFilePath)
		}
	}
	return nil
}

func copySingleFile(destinationDirectory string, sourceFile string) {
	fileName := filepath.Base(sourceFile)

	source, err := core.OpenFile(sourceFile)
	if err != nil {
		return
	}
	defer func() {
		err = source.Close()
		if err != nil {
			log.Warn("copySingleFile", "Could not close file", source.Name(), "error", err.Error())
		}
	}()

	destPath := filepath.Join(destinationDirectory, fileName)
	destination, err := os.Create(destPath)
	if err != nil {
		return
	}
	defer func() {
		err = destination.Close()
		if err != nil {
			log.Warn("copySingleFile", "Could not close file", source.Name(), "error", err.Error())
		}
	}()

	_, err = io.Copy(destination, source)
	if err != nil {
		log.Warn("copySingleFile", "Could not copy file", source.Name(), "error", err.Error())
	}
}

func indexValidatorsListIfNeeded(
	outportHandler outport.OutportHandler,
	coordinator nodesCoordinator.NodesCoordinator,
	epoch uint32,
) {
	if !outportHandler.HasDrivers() {
		return
	}

	validatorsPubKeys, err := coordinator.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		log.Warn("GetAllEligibleValidatorPublicKeys for epoch 0 failed", "error", err)
	}

	if len(validatorsPubKeys) > 0 {
		outportHandler.SaveValidatorsPubKeys(&outportCore.ValidatorsPubKeys{
			ShardValidatorsPubKeys: outportCore.ConvertPubKeys(validatorsPubKeys),
			Epoch:                  epoch,
		})
	}
}

func enableGopsIfNeeded(gopsEnabled bool) {
	if gopsEnabled {
		if err := agent.Listen(agent.Options{}); err != nil {
			log.Error("failure to init gops", "error", err.Error())
		}
	}

	log.Trace("gops", "enabled", gopsEnabled)
}

func createWhiteListerVerifiedTxs(generalConfig *config.Config) (process.WhiteListHandler, error) {
	whiteListCacheVerified, err := storageunit.NewCache(storageFactory.GetCacherFromConfig(generalConfig.WhiteListerVerifiedTxs))
	if err != nil {
		return nil, err
	}
	return interceptors.NewWhiteListDataVerifier(whiteListCacheVerified)
}
