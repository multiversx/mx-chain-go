package main

// TODO: Add unit tests
// TODO: Create a baseNodeRunner that uses common code from here and nodeRunner.go to avoid duplicated code

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

	"github.com/multiversx/mx-chain-go/api/gin"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/disabled"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/common/goroutines"
	"github.com/multiversx/mx-chain-go/common/ordering"
	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	requesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/resolverscontainer"
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
	"github.com/multiversx/mx-chain-go/factory/runType"
	stateComp "github.com/multiversx/mx-chain-go/factory/state"
	statusComp "github.com/multiversx/mx-chain-go/factory/status"
	"github.com/multiversx/mx-chain-go/factory/statusCore"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/parsing"
	"github.com/multiversx/mx-chain-go/health"
	"github.com/multiversx/mx-chain-go/node"
	"github.com/multiversx/mx-chain-go/node/metrics"
	trieIteratorsFactory "github.com/multiversx/mx-chain-go/node/trieIterators/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/factory/interceptorscontainer"
	"github.com/multiversx/mx-chain-go/process/headerCheck"
	"github.com/multiversx/mx-chain-go/process/interceptors"
	"github.com/multiversx/mx-chain-go/process/rating"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	sovereignConfig "github.com/multiversx/mx-chain-go/sovereignnode/config"
	"github.com/multiversx/mx-chain-go/sovereignnode/dataCodec"
	"github.com/multiversx/mx-chain-go/sovereignnode/incomingHeader"
	"github.com/multiversx/mx-chain-go/state/syncer"
	"github.com/multiversx/mx-chain-go/storage/cache"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	trieStatistics "github.com/multiversx/mx-chain-go/trie/statistics"
	"github.com/multiversx/mx-chain-go/update/trigger"

	"github.com/google/gops/agent"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/closing"
	"github.com/multiversx/mx-chain-core-go/core/throttler"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	hasherFactory "github.com/multiversx/mx-chain-core-go/hashing/factory"
	marshallerFactory "github.com/multiversx/mx-chain-core-go/marshal/factory"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-sovereign-bridge-go/cert"
	factoryBridge "github.com/multiversx/mx-chain-sovereign-bridge-go/client"
	bridgeCfg "github.com/multiversx/mx-chain-sovereign-bridge-go/client/config"
	notifierCfg "github.com/multiversx/mx-chain-sovereign-notifier-go/config"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/factory"
	notifierProcess "github.com/multiversx/mx-chain-sovereign-notifier-go/process"
	"github.com/multiversx/mx-sdk-abi-incubator/golang/abi"
)

var log = logger.GetOrCreate("sovereignNode")

const (
	// TODO: remove this after better handling VM versions switching
	// defaultDelayBeforeScQueriesStartInSec represents the default delay before the sc query processor should start to allow external queries
	defaultDelayBeforeScQueriesStartInSec = 120

	maxTimeToClose = 10 * time.Second
	// SoftRestartMessage is the custom message used when the node does a soft restart operation
	SoftRestartMessage = "Shuffled out - soft restart"
)

// sovereignNodeRunner holds the sovereign node runner configuration and controls running of a node
type sovereignNodeRunner struct {
	configs *sovereignConfig.SovereignConfig
}

// NewSovereignNodeRunner creates a sovereignNodeRunner instance
func NewSovereignNodeRunner(cfgs *sovereignConfig.SovereignConfig) (*sovereignNodeRunner, error) {
	if cfgs == nil {
		return nil, fmt.Errorf("nil configs provided")
	}
	return &sovereignNodeRunner{
		configs: cfgs,
	}, nil
}

// Start creates and starts the managed components
func (snr *sovereignNodeRunner) Start() error {
	configs := snr.configs
	flagsConfig := configs.FlagsConfig
	configurationPaths := configs.ConfigurationPathsHolder
	chanStopNodeProcess := make(chan endProcess.ArgEndProcess, 1)

	enableGopsIfNeeded(flagsConfig.EnableGops)

	var err error
	configurationPaths.Nodes, err = snr.getNodesFileName()
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

	printEnableEpochs(snr.configs.Configs)

	core.DumpGoRoutinesToLog(0, log)

	err = snr.startShufflingProcessLoop(chanStopNodeProcess)
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
	log.Debug(readEpochFor("increment SCR nonce in multi transfer"), "epoch", enableEpochs.IncrementSCRNonceInMultiTransferEnableEpoch)
	log.Debug(readEpochFor("esdt and NFT multi transfer"), "epoch", enableEpochs.ESDTMultiTransferEnableEpoch)
	log.Debug(readEpochFor("contract global mint and burn"), "epoch", enableEpochs.GlobalMintBurnDisableEpoch)
	log.Debug(readEpochFor("contract transfer role"), "epoch", enableEpochs.ESDTTransferRoleEnableEpoch)
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

func (snr *sovereignNodeRunner) startShufflingProcessLoop(
	chanStopNodeProcess chan endProcess.ArgEndProcess,
) error {
	for {
		log.Debug("\n\n====================Starting managedComponents creation================================")

		shouldStop, err := snr.executeOneComponentCreationCycle(chanStopNodeProcess)
		if shouldStop {
			return err
		}

		snr.shuffleOutStatsAndGC()
	}
}

func (snr *sovereignNodeRunner) shuffleOutStatsAndGC() {
	debugConfig := snr.configs.GeneralConfig.Debug.ShuffleOut

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

	snr.doProfileOnShuffleOut()
}

func (snr *sovereignNodeRunner) doProfileOnShuffleOut() {
	debugConfig := snr.configs.GeneralConfig.Debug.ShuffleOut
	shouldDoProfile := debugConfig.DoProfileOnShuffleOut && snr.configs.FlagsConfig.UseHealthService
	if !shouldDoProfile {
		return
	}

	log.Debug("running profile job")
	parentPath := filepath.Join(snr.configs.FlagsConfig.WorkingDir, snr.configs.GeneralConfig.Health.FolderPath)
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	err := health.WriteMemoryUseInfo(stats, time.Now(), parentPath, "softrestart")
	log.LogIfError(err)
}

func (snr *sovereignNodeRunner) executeOneComponentCreationCycle(
	chanStopNodeProcess chan endProcess.ArgEndProcess,
) (bool, error) {
	goRoutinesNumberStart := runtime.NumGoroutine()
	configs := snr.configs
	flagsConfig := configs.FlagsConfig
	configurationPaths := configs.ConfigurationPathsHolder

	log.Debug("creating healthService")
	healthService := snr.createHealthService(flagsConfig)

	log.Debug("creating core components")
	managedCoreComponents, err := snr.CreateManagedCoreComponents(
		chanStopNodeProcess,
	)
	if err != nil {
		return true, err
	}

	log.Debug("creating args for runType components")
	argsSovereignRunTypeComponents, err := snr.CreateArgsRunTypeComponents()
	if err != nil {
		return true, err
	}

	log.Debug("creating runType components")
	managedRunTypeComponents, err := snr.CreateManagedRunTypeComponents(managedCoreComponents, *argsSovereignRunTypeComponents)
	if err != nil {
		return true, err
	}

	log.Debug("creating status core components")
	managedStatusCoreComponents, err := snr.CreateManagedStatusCoreComponents(managedCoreComponents)
	if err != nil {
		return true, err
	}

	log.Debug("creating crypto components")
	managedCryptoComponents, err := snr.CreateManagedCryptoComponents(managedCoreComponents)
	if err != nil {
		return true, err
	}

	log.Debug("creating network components")
	managedNetworkComponents, err := snr.CreateManagedNetworkComponents(managedCoreComponents, managedStatusCoreComponents, managedCryptoComponents)
	if err != nil {
		return true, err
	}

	log.Debug("creating disabled API services")
	webServerHandler, err := snr.createHttpServer(managedStatusCoreComponents)
	if err != nil {
		return true, err
	}

	log.Debug("creating bootstrap components")
	managedBootstrapComponents, err := snr.CreateManagedBootstrapComponents(managedStatusCoreComponents, managedCoreComponents, managedCryptoComponents, managedNetworkComponents, managedRunTypeComponents)
	if err != nil {
		return true, err
	}

	snr.logInformation(managedCoreComponents, managedCryptoComponents, managedBootstrapComponents)

	log.Debug("creating data components")
	managedDataComponents, err := snr.CreateManagedDataComponents(managedStatusCoreComponents, managedCoreComponents, managedBootstrapComponents, managedCryptoComponents, managedRunTypeComponents)
	if err != nil {
		return true, err
	}

	log.Debug("creating state components")
	managedStateComponents, err := snr.CreateManagedStateComponents(
		managedCoreComponents,
		managedDataComponents,
		managedStatusCoreComponents,
		managedRunTypeComponents,
	)
	if err != nil {
		return true, err
	}

	log.Debug("creating metrics")
	// this should be called before setting the storer (done in the managedDataComponents creation)
	err = snr.createMetrics(managedStatusCoreComponents, managedCoreComponents, managedCryptoComponents, managedBootstrapComponents)
	if err != nil {
		return true, err
	}

	log.Debug("registering components in healthService")
	snr.registerDataComponentsInHealthService(healthService, managedDataComponents)

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
		managedBootstrapComponents.NodesCoordinatorRegistryFactory(),
		nodesCoordinator.NewSovereignIndexHashedNodesCoordinatorWithRaterFactory(),
	)
	if err != nil {
		return true, err
	}

	log.Debug("starting status pooling components")
	managedStatusComponents, err := snr.CreateManagedStatusComponents(
		managedStatusCoreComponents,
		managedCoreComponents,
		managedNetworkComponents,
		managedBootstrapComponents,
		managedStateComponents,
		nodesCoordinatorInstance,
		configs.ImportDbConfig.IsImportDBMode,
		managedCryptoComponents,
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

	incomingHeaderHandler, err := createIncomingHeaderProcessor(
		&configs.SovereignExtraConfig.NotifierConfig,
		managedDataComponents.Datapool(),
		configs.SovereignExtraConfig.MainChainNotarization.MainChainNotarizationStartRound,
		managedRunTypeComponents,
	)

	managedProcessComponents, err := snr.CreateManagedProcessComponents(
		managedRunTypeComponents,
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
		incomingHeaderHandler,
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

	outGoingBridgeOpHandler, err := factoryBridge.CreateClient(&bridgeCfg.ClientConfig{
		GRPCHost: snr.configs.SovereignExtraConfig.OutGoingBridge.GRPCHost,
		GRPCPort: snr.configs.SovereignExtraConfig.OutGoingBridge.GRPCPort,
		CertificateCfg: cert.FileCfg{
			CertFile: snr.configs.SovereignExtraConfig.OutGoingBridgeCertificate.CertificatePath,
			PkFile:   snr.configs.SovereignExtraConfig.OutGoingBridgeCertificate.CertificatePkPath,
		},
	})
	if err != nil {
		return true, err
	}

	managedConsensusComponents, err := snr.CreateManagedConsensusComponents(
		managedCoreComponents,
		managedNetworkComponents,
		managedCryptoComponents,
		managedDataComponents,
		managedStateComponents,
		managedStatusComponents,
		managedProcessComponents,
		managedStatusCoreComponents,
		outGoingBridgeOpHandler,
		managedRunTypeComponents,
	)
	if err != nil {
		return true, err
	}

	managedHeartbeatV2Components, err := snr.CreateManagedHeartbeatV2Components(
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

	sovereignWsReceiver, err := createSovereignWsReceiver(
		&configs.SovereignExtraConfig.NotifierConfig,
		incomingHeaderHandler,
	)
	if err != nil {
		return true, err
	}

	log.Debug("creating node structure")

	extraOptionNotifierReceiver := func(n *node.Node) error {
		n.AddClosableComponent(sovereignWsReceiver)
		return nil
	}
	extraOptionOutGoingBridgeSender := func(n *node.Node) error {
		n.AddClosableComponent(outGoingBridgeOpHandler)
		return nil
	}
	nodeHandler, err := node.CreateNode(
		configs.GeneralConfig,
		managedRunTypeComponents,
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
		node.NewSovereignNodeFactory(),
		extraOptionNotifierReceiver,
		extraOptionOutGoingBridgeSender,
	)
	if err != nil {
		return true, err
	}

	// this channel will trigger the moment when the sc query service should be able to process VM Query requests
	allowExternalVMQueriesChan := make(chan struct{})

	log.Debug("updating the API service after creating the node facade")
	ef, err := snr.createApiFacade(nodeHandler, webServerHandler, gasScheduleNotifier, allowExternalVMQueriesChan)
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

	err = waitForSignal(
		sigs,
		managedCoreComponents.ChanStopNodeProcess(),
		healthService,
		ef,
		webServerHandler,
		nodeHandler,
		goRoutinesNumberStart,
	)
	if err != nil {
		return true, nil
	}

	return false, nil
}

func addSyncersToAccountsDB(
	config *config.Config,
	coreComponents mainFactory.CoreComponentsHolder,
	dataComponents mainFactory.DataComponentsHolder,
	stateComponents mainFactory.StateComponentsHolder,
	bootstrapComponents mainFactory.BootstrapComponentsHolder,
	processComponents mainFactory.ProcessComponentsHolder,
) error {
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

func (snr *sovereignNodeRunner) createApiFacade(
	nodeHandler node.NodeHandler,
	upgradableHttpServer shared.UpgradeableHttpServerHandler,
	gasScheduleNotifier common.GasScheduleNotifierAPI,
	allowVMQueriesChan chan struct{},
) (closing.Closer, error) {
	configs := snr.configs

	log.Debug("creating api resolver structure")

	apiResolverArgs := &apiComp.ApiResolverArgs{
		Configs:                        configs.Configs,
		CoreComponents:                 nodeHandler.GetCoreComponents(),
		DataComponents:                 nodeHandler.GetDataComponents(),
		StateComponents:                nodeHandler.GetStateComponents(),
		BootstrapComponents:            nodeHandler.GetBootstrapComponents(),
		CryptoComponents:               nodeHandler.GetCryptoComponents(),
		ProcessComponents:              nodeHandler.GetProcessComponents(),
		StatusCoreComponents:           nodeHandler.GetStatusCoreComponents(),
		GasScheduleNotifier:            gasScheduleNotifier,
		Bootstrapper:                   nodeHandler.GetConsensusComponents().Bootstrapper(),
		RunTypeComponents:              nodeHandler.GetRunTypeComponents(),
		AllowVMQueriesChan:             allowVMQueriesChan,
		StatusComponents:               nodeHandler.GetStatusComponents(),
		DelegatedListFactoryHandler:    trieIteratorsFactory.NewSovereignDelegatedListProcessorFactory(),
		DirectStakedListFactoryHandler: trieIteratorsFactory.NewSovereignDirectStakedListProcessorFactory(),
		TotalStakedValueFactoryHandler: trieIteratorsFactory.NewSovereignTotalStakedValueProcessorFactory(),
	}

	apiResolver, err := apiComp.CreateApiResolver(apiResolverArgs)
	if err != nil {
		return nil, err
	}

	log.Debug("creating multiversx node facade")

	flagsConfig := configs.FlagsConfig

	argNodeFacade := facade.ArgNodeFacade{
		Node:                   nodeHandler,
		ApiResolver:            apiResolver,
		RestAPIServerDebugMode: flagsConfig.EnableRestAPIServerDebugMode,
		WsAntifloodConfig:      configs.GeneralConfig.WebServerAntiflood,
		FacadeConfig: config.FacadeConfig{
			RestApiInterface: flagsConfig.RestApiInterface,
			PprofEnabled:     flagsConfig.EnablePprof,
		},
		ApiRoutesConfig: *configs.ApiRoutesConfig,
		AccountsState:   nodeHandler.GetStateComponents().AccountsAdapter(),
		PeerState:       nodeHandler.GetStateComponents().PeerAccounts(),
		Blockchain:      nodeHandler.GetDataComponents().Blockchain(),
	}

	ef, err := facade.NewNodeFacade(argNodeFacade)
	if err != nil {
		return nil, fmt.Errorf("%w while creating NodeFacade", err)
	}

	ef.SetSyncer(nodeHandler.GetCoreComponents().SyncTimer())

	err = upgradableHttpServer.UpdateFacade(ef)
	if err != nil {
		return nil, err
	}

	log.Debug("updated node facade")

	log.Trace("starting background services")

	return ef, nil
}

func (snr *sovereignNodeRunner) createHttpServer(managedStatusCoreComponents mainFactory.StatusCoreComponentsHolder) (shared.UpgradeableHttpServerHandler, error) {
	if check.IfNil(managedStatusCoreComponents) {
		return nil, node.ErrNilCoreComponents
	}

	argsInitialNodeFacade := initial.ArgInitialNodeFacade{
		ApiInterface:                snr.configs.FlagsConfig.RestApiInterface,
		PprofEnabled:                snr.configs.FlagsConfig.EnablePprof,
		P2PPrometheusMetricsEnabled: snr.configs.FlagsConfig.P2PPrometheusMetricsEnabled,
		StatusMetricsHandler:        managedStatusCoreComponents.StatusMetrics(),
	}
	initialFacade, err := initial.NewInitialNodeFacade(argsInitialNodeFacade)
	if err != nil {
		return nil, err
	}

	httpServerArgs := gin.ArgsNewWebServer{
		Facade:          initialFacade,
		ApiConfig:       *snr.configs.ApiRoutesConfig,
		AntiFloodConfig: snr.configs.GeneralConfig.WebServerAntiflood,
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

func (snr *sovereignNodeRunner) createMetrics(
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
		snr.configs.FlagsConfig.Version,
		snr.configs.EconomicsConfig,
		snr.configs.GeneralConfig.EpochStartConfig.RoundsPerEpoch,
		coreComponents.MinTransactionVersion(),
	)

	if err != nil {
		return err
	}

	metrics.SaveStringMetric(statusCoreComponents.AppStatusHandler(), common.MetricNodeDisplayName, snr.configs.PreferencesConfig.Preferences.NodeDisplayName)
	metrics.SaveStringMetric(statusCoreComponents.AppStatusHandler(), common.MetricRedundancyLevel, fmt.Sprintf("%d", snr.configs.PreferencesConfig.Preferences.RedundancyLevel))
	metrics.SaveStringMetric(statusCoreComponents.AppStatusHandler(), common.MetricRedundancyIsMainActive, common.MetricValueNA)
	metrics.SaveStringMetric(statusCoreComponents.AppStatusHandler(), common.MetricChainId, coreComponents.ChainID())
	metrics.SaveUint64Metric(statusCoreComponents.AppStatusHandler(), common.MetricGasPerDataByte, coreComponents.EconomicsData().GasPerDataByte())
	metrics.SaveUint64Metric(statusCoreComponents.AppStatusHandler(), common.MetricMinGasPrice, coreComponents.EconomicsData().MinGasPrice())
	metrics.SaveUint64Metric(statusCoreComponents.AppStatusHandler(), common.MetricMinGasLimit, coreComponents.EconomicsData().MinGasLimit())
	metrics.SaveStringMetric(statusCoreComponents.AppStatusHandler(), common.MetricRewardsTopUpGradientPoint, coreComponents.EconomicsData().RewardsTopUpGradientPoint().String())
	metrics.SaveStringMetric(statusCoreComponents.AppStatusHandler(), common.MetricTopUpFactor, fmt.Sprintf("%g", coreComponents.EconomicsData().RewardsTopUpFactor()))
	metrics.SaveStringMetric(statusCoreComponents.AppStatusHandler(), common.MetricGasPriceModifier, fmt.Sprintf("%g", coreComponents.EconomicsData().GasPriceModifier()))
	metrics.SaveUint64Metric(statusCoreComponents.AppStatusHandler(), common.MetricMaxGasPerTransaction, coreComponents.EconomicsData().MaxGasLimitPerTx())
	metrics.SaveUint64Metric(statusCoreComponents.AppStatusHandler(), common.MetricExtraGasLimitGuardedTx, coreComponents.EconomicsData().ExtraGasLimitGuardedTx())
	if snr.configs.PreferencesConfig.Preferences.FullArchive {
		metrics.SaveStringMetric(statusCoreComponents.AppStatusHandler(), common.MetricPeerType, core.ObserverPeer.String())
		metrics.SaveStringMetric(statusCoreComponents.AppStatusHandler(), common.MetricPeerSubType, core.FullHistoryObserver.String())
	}

	return nil
}

func (snr *sovereignNodeRunner) createHealthService(flagsConfig *config.ContextFlagsConfig) node.HealthService {
	healthService := health.NewHealthService(snr.configs.GeneralConfig.Health, flagsConfig.WorkingDir)
	if flagsConfig.UseHealthService {
		healthService.Start()
	}

	return healthService
}

func (snr *sovereignNodeRunner) registerDataComponentsInHealthService(healthService node.HealthService, dataComponents mainFactory.DataComponentsHolder) {
	healthService.RegisterComponent(dataComponents.Datapool().Transactions())
	healthService.RegisterComponent(dataComponents.Datapool().UnsignedTransactions())
	healthService.RegisterComponent(dataComponents.Datapool().RewardTransactions())
}

// CreateManagedConsensusComponents is the managed consensus components factory
func (snr *sovereignNodeRunner) CreateManagedConsensusComponents(
	coreComponents mainFactory.CoreComponentsHolder,
	networkComponents mainFactory.NetworkComponentsHolder,
	cryptoComponents mainFactory.CryptoComponentsHolder,
	dataComponents mainFactory.DataComponentsHolder,
	stateComponents mainFactory.StateComponentsHolder,
	statusComponents mainFactory.StatusComponentsHolder,
	processComponents mainFactory.ProcessComponentsHolder,
	statusCoreComponents mainFactory.StatusCoreComponentsHolder,
	outGoingBridgeOpHandler bls.BridgeOperationsHandler,
	runTypeComponents mainFactory.RunTypeComponentsHolder,
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

	extraSignersHolder, err := createOutGoingTxDataSigners(cryptoComponents.ConsensusSigningHandler())
	if err != nil {
		return nil, err
	}

	sovSubRoundEndCreator, err := bls.NewSovereignSubRoundEndCreator(runTypeComponents.OutGoingOperationsPoolHandler(), outGoingBridgeOpHandler)
	if err != nil {
		return nil, err
	}

	consensusArgs := consensusComp.ConsensusComponentsFactoryArgs{
		Config:                *snr.configs.GeneralConfig,
		BootstrapRoundIndex:   snr.configs.FlagsConfig.BootstrapRoundIndex,
		CoreComponents:        coreComponents,
		NetworkComponents:     networkComponents,
		CryptoComponents:      cryptoComponents,
		DataComponents:        dataComponents,
		ProcessComponents:     processComponents,
		StateComponents:       stateComponents,
		StatusComponents:      statusComponents,
		StatusCoreComponents:  statusCoreComponents,
		ScheduledProcessor:    scheduledProcessor,
		IsInImportMode:        snr.configs.ImportDbConfig.IsImportDBMode,
		ShouldDisableWatchdog: snr.configs.FlagsConfig.DisableConsensusWatchdog,
		RunTypeComponents:     runTypeComponents,
		ExtraSignersHolder:    extraSignersHolder,
		SubRoundEndV2Creator:  sovSubRoundEndCreator,
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

func createOutGoingTxDataSigners(signingHandler consensus.SigningHandler) (bls.ExtraSignersHolder, error) {
	extraSignerHandler := signingHandler.ShallowClone()
	startRoundExtraSignersHolder := bls.NewSubRoundStartExtraSignersHolder()
	startRoundExtraSigner, err := bls.NewSovereignSubRoundStartOutGoingTxData(extraSignerHandler)
	if err != nil {
		return nil, err
	}
	err = startRoundExtraSignersHolder.RegisterExtraSigningHandler(startRoundExtraSigner)
	if err != nil {
		return nil, err
	}

	signRoundExtraSignersHolder := bls.NewSubRoundSignatureExtraSignersHolder()
	signRoundExtraSigner, err := bls.NewSovereignSubRoundSignatureOutGoingTxData(extraSignerHandler)
	if err != nil {
		return nil, err
	}
	err = signRoundExtraSignersHolder.RegisterExtraSigningHandler(signRoundExtraSigner)
	if err != nil {
		return nil, err
	}

	endRoundExtraSignersHolder := bls.NewSubRoundEndExtraSignersHolder()
	endRoundExtraSigner, err := bls.NewSovereignSubRoundEndOutGoingTxData(extraSignerHandler)
	if err != nil {
		return nil, err
	}
	err = endRoundExtraSignersHolder.RegisterExtraSigningHandler(endRoundExtraSigner)
	if err != nil {
		return nil, err
	}

	return bls.NewExtraSignersHolder(
		startRoundExtraSignersHolder,
		signRoundExtraSignersHolder,
		endRoundExtraSignersHolder)
}

// CreateManagedHeartbeatV2Components is the managed heartbeatV2 components factory
func (snr *sovereignNodeRunner) CreateManagedHeartbeatV2Components(
	bootstrapComponents mainFactory.BootstrapComponentsHolder,
	coreComponents mainFactory.CoreComponentsHolder,
	networkComponents mainFactory.NetworkComponentsHolder,
	cryptoComponents mainFactory.CryptoComponentsHolder,
	dataComponents mainFactory.DataComponentsHolder,
	processComponents mainFactory.ProcessComponentsHolder,
	statusCoreComponents mainFactory.StatusCoreComponentsHolder,
) (mainFactory.HeartbeatV2ComponentsHandler, error) {
	heartbeatV2Args := heartbeatComp.ArgHeartbeatV2ComponentsFactory{
		Config:               *snr.configs.GeneralConfig,
		Prefs:                *snr.configs.PreferencesConfig,
		BaseVersion:          snr.configs.FlagsConfig.BaseVersion,
		AppVersion:           snr.configs.FlagsConfig.Version,
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
	ef closing.Closer,
	httpServer shared.UpgradeableHttpServerHandler,
	nodeHandler node.NodeHandler,
	goRoutinesNumberStart int,
) error {
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
		closeAllComponents(healthService, ef, httpServer, nodeHandler, chanCloseComponents)
	}()

	select {
	case <-chanCloseComponents:
		log.Debug("Closed all components gracefully")
	case <-time.After(maxTimeToClose):
		log.Warn("force closing the node",
			"error", "closeAllComponents did not finish on time",
			"stack", goroutines.GetGoRoutines())

		return fmt.Errorf("did NOT close all components gracefully")
	}

	if wrongConfig {
		// hang the node's process because it cannot continue with the current configuration and a restart doesn't
		// change this behaviour
		for {
			log.Error("wrong configuration. stopped processing", "description", wrongConfigDescription)
			time.Sleep(1 * time.Minute)
		}
	}

	if reshuffled {
		log.Info("=============================" + SoftRestartMessage + "==================================")
		core.DumpGoRoutinesToLog(goRoutinesNumberStart, log)

		return nil
	}

	return fmt.Errorf("not reshuffled, closing")
}

func (snr *sovereignNodeRunner) logInformation(
	coreComponents mainFactory.CoreComponentsHolder,
	cryptoComponents mainFactory.CryptoComponentsHolder,
	bootstrapComponents mainFactory.BootstrapComponentsHolder,
) {
	log.Info("Bootstrap", "epoch", bootstrapComponents.EpochBootstrapParams().Epoch())
	if bootstrapComponents.EpochBootstrapParams().NodesConfig() != nil {
		log.Info("the epoch from nodesConfig is",
			"epoch", bootstrapComponents.EpochBootstrapParams().NodesConfig().GetCurrentEpoch())
	}

	var shardIdString = core.GetShardIDString(bootstrapComponents.ShardCoordinator().SelfId())
	logger.SetCorrelationShard(shardIdString)

	sessionInfoFileOutput := fmt.Sprintf("%s:%s\n%s:%s\n%s:%v\n%s:%s\n%s:%v\n",
		"PkBlockSign", cryptoComponents.PublicKeyString(),
		"ShardId", shardIdString,
		"TotalShards", bootstrapComponents.ShardCoordinator().NumberOfShards(),
		"AppVersion", snr.configs.FlagsConfig.Version,
		"GenesisTimeStamp", coreComponents.GenesisTime().Unix(),
	)

	sessionInfoFileOutput += "\nStarted with parameters:\n"
	sessionInfoFileOutput += snr.configs.FlagsConfig.SessionInfoFileOutput

	snr.logSessionInformation(snr.configs.FlagsConfig.WorkingDir, sessionInfoFileOutput, coreComponents)
}

func (snr *sovereignNodeRunner) getNodesFileName() (string, error) {
	flagsConfig := snr.configs.FlagsConfig
	configurationPaths := snr.configs.ConfigurationPathsHolder
	nodesFileName := configurationPaths.Nodes

	exportFolder := filepath.Join(flagsConfig.WorkingDir, snr.configs.GeneralConfig.Hardfork.ImportFolder)
	if snr.configs.GeneralConfig.Hardfork.AfterHardFork {
		exportFolderNodesSetupPath := filepath.Join(exportFolder, common.NodesSetupJsonFileName)
		if !core.FileExists(exportFolderNodesSetupPath) {
			return "", fmt.Errorf("cannot find %s in the export folder", common.NodesSetupJsonFileName)
		}

		nodesFileName = exportFolderNodesSetupPath
	}
	return nodesFileName, nil
}

// CreateManagedStatusComponents is the managed status components factory
func (snr *sovereignNodeRunner) CreateManagedStatusComponents(
	managedStatusCoreComponents mainFactory.StatusCoreComponentsHolder,
	managedCoreComponents mainFactory.CoreComponentsHolder,
	managedNetworkComponents mainFactory.NetworkComponentsHolder,
	managedBootstrapComponents mainFactory.BootstrapComponentsHolder,
	managedStateComponents mainFactory.StateComponentsHolder,
	nodesCoordinator nodesCoordinator.NodesCoordinator,
	isInImportMode bool,
	cryptoComponents mainFactory.CryptoComponentsHolder,
) (mainFactory.StatusComponentsHandler, error) {
	statArgs := statusComp.StatusComponentsFactoryArgs{
		Config:               *snr.configs.GeneralConfig,
		ExternalConfig:       *snr.configs.ExternalConfig,
		EconomicsConfig:      *snr.configs.EconomicsConfig,
		ShardCoordinator:     managedBootstrapComponents.ShardCoordinator(),
		NodesCoordinator:     nodesCoordinator,
		EpochStartNotifier:   managedCoreComponents.EpochStartNotifierWithConfirm(),
		CoreComponents:       managedCoreComponents,
		NetworkComponents:    managedNetworkComponents,
		StateComponents:      managedStateComponents,
		IsInImportMode:       isInImportMode,
		StatusCoreComponents: managedStatusCoreComponents,
		CryptoComponents:     cryptoComponents,
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

func (snr *sovereignNodeRunner) logSessionInformation(
	workingDir string,
	sessionInfoFileOutput string,
	coreComponents mainFactory.CoreComponentsHolder,
) {
	statsFolder := filepath.Join(workingDir, common.DefaultStatsPath)
	configurationPaths := snr.configs.ConfigurationPathsHolder
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
			configurationPaths.MainP2p,
			configurationPaths.FullArchiveP2p,
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
func (snr *sovereignNodeRunner) CreateManagedProcessComponents(
	runTypeComponents mainFactory.RunTypeComponentsHolder,
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
	incomingHeaderHandler process.IncomingHeaderSubscriber,
) (mainFactory.ProcessComponentsHandler, error) {
	configs := snr.configs
	configurationPaths := snr.configs.ConfigurationPathsHolder
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
	sovereignAccountsParser, err := parsing.NewSovereignAccountsParser(accountsParser)
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

	extraHeaderSigVerifierHolder := headerCheck.NewExtraHeaderSigVerifierHolder()
	sovHeaderSigVerifier, err := headerCheck.NewSovereignHeaderSigVerifier(cryptoComponents.BlockSigner())
	if err != nil {
		return nil, err
	}

	err = extraHeaderSigVerifierHolder.RegisterExtraHeaderSigVerifier(sovHeaderSigVerifier)
	if err != nil {
		return nil, err
	}

	processArgs := processComp.ProcessComponentsFactoryArgs{
		Config:                                *configs.GeneralConfig,
		EpochConfig:                           *configs.EpochConfig,
		RoundConfig:                           *configs.RoundConfig,
		PrefConfigs:                           *configs.PreferencesConfig,
		ImportDBConfig:                        *configs.ImportDbConfig,
		AccountsParser:                        sovereignAccountsParser,
		SmartContractParser:                   smartContractParser,
		GasSchedule:                           gasScheduleNotifier,
		NodesCoordinator:                      nodesCoordinator,
		Data:                                  dataComponents,
		CoreData:                              coreComponents,
		Crypto:                                cryptoComponents,
		State:                                 stateComponents,
		Network:                               networkComponents,
		BootstrapComponents:                   bootstrapComponents,
		StatusComponents:                      statusComponents,
		StatusCoreComponents:                  statusCoreComponents,
		RequestedItemsHandler:                 requestedItemsHandler,
		WhiteListHandler:                      whiteListRequest,
		WhiteListerVerifiedTxs:                whiteListerVerifiedTxs,
		MaxRating:                             configs.RatingsConfig.General.MaxRating,
		SystemSCConfig:                        configs.SystemSCConfig,
		ImportStartHandler:                    importStartHandler,
		HistoryRepo:                           historyRepository,
		FlagsConfig:                           *configs.FlagsConfig,
		TxExecutionOrderHandler:               ordering.NewOrderedCollection(),
		RunTypeComponents:                     runTypeComponents,
		GenesisMetaBlockChecker:               processComp.NewSovereignGenesisMetaBlockChecker(),
		RequesterContainerFactoryCreator:      requesterscontainer.NewSovereignShardRequestersContainerFactoryCreator(),
		IncomingHeaderSubscriber:              incomingHeaderHandler,
		InterceptorsContainerFactoryCreator:   interceptorscontainer.NewSovereignShardInterceptorsContainerFactoryCreator(),
		ShardResolversContainerFactoryCreator: resolverscontainer.NewSovereignShardResolversContainerFactoryCreator(),
		TxPreProcessorCreator:                 preprocess.NewSovereignTxPreProcessorCreator(),
		ExtraHeaderSigVerifierHolder:          extraHeaderSigVerifierHolder,
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
func (snr *sovereignNodeRunner) CreateManagedDataComponents(
	statusCoreComponents mainFactory.StatusCoreComponentsHolder,
	coreComponents mainFactory.CoreComponentsHolder,
	bootstrapComponents mainFactory.BootstrapComponentsHolder,
	crypto mainFactory.CryptoComponentsHolder,
	runTypeComponents mainFactory.RunTypeComponentsHolder,
) (mainFactory.DataComponentsHandler, error) {
	configs := snr.configs
	storerEpoch := bootstrapComponents.EpochBootstrapParams().Epoch()
	if !configs.GeneralConfig.StoragePruning.Enabled {
		// TODO: refactor this as when the pruning storer is disabled, the default directory path is Epoch_0
		// and it should be Epoch_ALL or something similar
		storerEpoch = 0
	}

	dataArgs := dataComp.DataComponentsFactoryArgs{
		Config:                          *configs.GeneralConfig,
		PrefsConfig:                     configs.PreferencesConfig.Preferences,
		ShardCoordinator:                bootstrapComponents.ShardCoordinator(),
		Core:                            coreComponents,
		StatusCore:                      statusCoreComponents,
		Crypto:                          crypto,
		CurrentEpoch:                    storerEpoch,
		CreateTrieEpochRootHashStorer:   configs.ImportDbConfig.ImportDbSaveTrieEpochRootHash,
		FlagsConfigs:                    *configs.FlagsConfig,
		NodeProcessingMode:              common.GetNodeProcessingMode(snr.configs.ImportDbConfig),
		AdditionalStorageServiceCreator: runTypeComponents.AdditionalStorageServiceCreator(),
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
func (snr *sovereignNodeRunner) CreateManagedStateComponents(
	coreComponents mainFactory.CoreComponentsHolder,
	dataComponents mainFactory.DataComponentsHandler,
	statusCoreComponents mainFactory.StatusCoreComponentsHolder,
	runTypeComponents mainFactory.RunTypeComponentsHolder,
) (mainFactory.StateComponentsHandler, error) {
	stateArgs := stateComp.StateComponentsFactoryArgs{
		Config:                   *snr.configs.GeneralConfig,
		Core:                     coreComponents,
		StatusCore:               statusCoreComponents,
		StorageService:           dataComponents.StorageService(),
		ProcessingMode:           common.GetNodeProcessingMode(snr.configs.ImportDbConfig),
		ShouldSerializeSnapshots: snr.configs.FlagsConfig.SerializeSnapshots,
		ChainHandler:             dataComponents.Blockchain(),
		AccountsCreator:          runTypeComponents.AccountsCreator(),
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
func (snr *sovereignNodeRunner) CreateManagedBootstrapComponents(
	statusCoreComponents mainFactory.StatusCoreComponentsHolder,
	coreComponents mainFactory.CoreComponentsHolder,
	cryptoComponents mainFactory.CryptoComponentsHolder,
	networkComponents mainFactory.NetworkComponentsHolder,
	runTypeComponents mainFactory.RunTypeComponentsHolder,
) (mainFactory.BootstrapComponentsHandler, error) {

	bootstrapComponentsFactoryArgs := bootstrapComp.BootstrapComponentsFactoryArgs{
		Config:                           *snr.configs.GeneralConfig,
		PrefConfig:                       *snr.configs.PreferencesConfig,
		ImportDbConfig:                   *snr.configs.ImportDbConfig,
		FlagsConfig:                      *snr.configs.FlagsConfig,
		WorkingDir:                       snr.configs.FlagsConfig.DbDir,
		CoreComponents:                   coreComponents,
		CryptoComponents:                 cryptoComponents,
		NetworkComponents:                networkComponents,
		StatusCoreComponents:             statusCoreComponents,
		RunTypeComponents:                runTypeComponents,
		NodesCoordinatorWithRaterFactory: nodesCoordinator.NewSovereignIndexHashedNodesCoordinatorWithRaterFactory(),
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
func (snr *sovereignNodeRunner) CreateManagedNetworkComponents(
	coreComponents mainFactory.CoreComponentsHolder,
	statusCoreComponents mainFactory.StatusCoreComponentsHolder,
	cryptoComponents mainFactory.CryptoComponentsHolder,
) (mainFactory.NetworkComponentsHandler, error) {
	networkComponentsFactoryArgs := networkComp.NetworkComponentsFactoryArgs{
		MainP2pConfig:         *snr.configs.MainP2pConfig,
		FullArchiveP2pConfig:  *snr.configs.FullArchiveP2pConfig,
		MainConfig:            *snr.configs.GeneralConfig,
		RatingsConfig:         *snr.configs.RatingsConfig,
		StatusHandler:         statusCoreComponents.AppStatusHandler(),
		Marshalizer:           coreComponents.InternalMarshalizer(),
		Syncer:                coreComponents.SyncTimer(),
		PreferredPeersSlices:  snr.configs.PreferencesConfig.Preferences.PreferredConnections,
		BootstrapWaitTime:     common.TimeToWaitForP2PBootstrap,
		NodeOperationMode:     common.NormalOperation,
		ConnectionWatcherType: snr.configs.PreferencesConfig.Preferences.ConnectionWatcherType,
		CryptoComponents:      cryptoComponents,
	}
	if snr.configs.ImportDbConfig.IsImportDBMode {
		networkComponentsFactoryArgs.BootstrapWaitTime = 0
	}
	if snr.configs.PreferencesConfig.Preferences.FullArchive {
		networkComponentsFactoryArgs.NodeOperationMode = common.FullArchiveMode
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
func (snr *sovereignNodeRunner) CreateManagedCoreComponents(
	chanStopNodeProcess chan endProcess.ArgEndProcess,
) (mainFactory.CoreComponentsHandler, error) {
	coreArgs := coreComp.CoreComponentsFactoryArgs{
		Config:                   *snr.configs.GeneralConfig,
		ConfigPathsHolder:        *snr.configs.ConfigurationPathsHolder,
		EpochConfig:              *snr.configs.EpochConfig,
		RoundConfig:              *snr.configs.RoundConfig,
		ImportDbConfig:           *snr.configs.ImportDbConfig,
		RatingsConfig:            *snr.configs.RatingsConfig,
		EconomicsConfig:          *snr.configs.EconomicsConfig,
		NodesFilename:            snr.configs.ConfigurationPathsHolder.Nodes,
		WorkingDirectory:         snr.configs.FlagsConfig.DbDir,
		ChanStopNodeProcess:      chanStopNodeProcess,
		GenesisNodesSetupFactory: sharding.NewSovereignGenesisNodesSetupFactory(),
		RatingsDataFactory:       rating.NewSovereignRatingsDataFactory(),
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
func (snr *sovereignNodeRunner) CreateManagedStatusCoreComponents(
	coreComponents mainFactory.CoreComponentsHolder,
) (mainFactory.StatusCoreComponentsHandler, error) {
	args := statusCore.StatusCoreComponentsFactoryArgs{
		Config:          *snr.configs.GeneralConfig,
		EpochConfig:     *snr.configs.EpochConfig,
		RoundConfig:     *snr.configs.RoundConfig,
		RatingsConfig:   *snr.configs.RatingsConfig,
		EconomicsConfig: *snr.configs.EconomicsConfig,
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
func (snr *sovereignNodeRunner) CreateManagedCryptoComponents(
	coreComponents mainFactory.CoreComponentsHolder,
) (mainFactory.CryptoComponentsHandler, error) {
	configs := snr.configs
	validatorKeyPemFileName := configs.ConfigurationPathsHolder.ValidatorKey
	allValidatorKeysPemFileName := configs.ConfigurationPathsHolder.AllValidatorKeys
	cryptoComponentsHandlerArgs := cryptoComp.CryptoComponentsFactoryArgs{
		ValidatorKeyPemFileName:              validatorKeyPemFileName,
		AllValidatorKeysPemFileName:          allValidatorKeysPemFileName,
		SkIndex:                              configs.FlagsConfig.ValidatorKeyIndex,
		Config:                               *configs.GeneralConfig,
		PrefsConfig:                          *configs.PreferencesConfig,
		CoreComponentsHolder:                 coreComponents,
		ActivateBLSPubKeyMessageVerification: configs.SystemSCConfig.StakingSystemSCConfig.ActivateBLSPubKeyMessageVerification,
		KeyLoader:                            core.NewKeyLoader(),
		ImportModeNoSigCheck:                 configs.ImportDbConfig.ImportDbNoSigCheckFlag,
		IsInImportMode:                       configs.ImportDbConfig.IsImportDBMode,
		EnableEpochs:                         configs.EpochConfig.EnableEpochs,
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

func (snr *sovereignNodeRunner) CreateArgsRunTypeComponents() (*runType.ArgsSovereignRunTypeComponents, error) {
	sovereignCfg := snr.configs.SovereignExtraConfig

	codec := abi.NewDefaultCodec()
	argsDataCodec := dataCodec.ArgsDataCodec{
		Serializer: abi.NewSerializer(codec),
	}

	dataCodecHandler, err := dataCodec.NewDataCodec(argsDataCodec)
	if err != nil {
		return nil, err
	}

	topicsCheckerHandler := incomingHeader.NewTopicsChecker()

	return &runType.ArgsSovereignRunTypeComponents{
		Config:        *sovereignCfg,
		DataCodec:     dataCodecHandler,
		TopicsChecker: topicsCheckerHandler,
	}, nil
}

// CreateManagedRunTypeComponents creates the managed runType components
func (snr *sovereignNodeRunner) CreateManagedRunTypeComponents(coreComp mainFactory.CoreComponentsHandler, args runType.ArgsSovereignRunTypeComponents) (mainFactory.RunTypeComponentsHandler, error) {
	runTypeComponentsFactory, err := runType.NewRunTypeComponentsFactory(coreComp)
	if err != nil {
		return nil, fmt.Errorf("NewRunTypeComponentsFactory failed: %w", err)
	}

	sovereignRunTypeComponentsFactory, err := runType.NewSovereignRunTypeComponentsFactory(
		runTypeComponentsFactory,
		args,
	)
	if err != nil {
		return nil, fmt.Errorf("NewSovereignRunTypeComponentsFactory failed: %w", err)
	}

	managedRunTypeComponents, err := runType.NewManagedRunTypeComponents(sovereignRunTypeComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedRunTypeComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedRunTypeComponents, nil
}

func closeAllComponents(
	healthService io.Closer,
	facade mainFactory.Closer,
	httpServer shared.UpgradeableHttpServerHandler,
	node node.NodeHandler,
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

// TODO: add some unit tests
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

func createIncomingHeaderProcessor(
	config *config.NotifierConfig,
	dataPool dataRetriever.PoolsHolder,
	mainChainNotarizationStartRound uint64,
	runTypeComponents mainFactory.RunTypeComponentsHolder,
) (process.IncomingHeaderSubscriber, error) {
	marshaller, err := marshallerFactory.NewMarshalizer(config.WebSocketConfig.MarshallerType)
	if err != nil {
		return nil, err
	}
	hasher, err := hasherFactory.NewHasher(config.WebSocketConfig.HasherType)
	if err != nil {
		return nil, err
	}

	argsIncomingHeaderHandler := incomingHeader.ArgsIncomingHeaderProcessor{
		HeadersPool:                     dataPool.Headers(),
		TxPool:                          dataPool.UnsignedTransactions(),
		Marshaller:                      marshaller,
		Hasher:                          hasher,
		MainChainNotarizationStartRound: mainChainNotarizationStartRound,
		OutGoingOperationsPool:          runTypeComponents.OutGoingOperationsPoolHandler(),
		DataCodec:                       runTypeComponents.DataCodecHandler(),
		TopicsChecker:                   runTypeComponents.TopicsCheckerHandler(),
	}

	return incomingHeader.NewIncomingHeaderProcessor(argsIncomingHeaderHandler)
}

func createSovereignWsReceiver(
	config *config.NotifierConfig,
	incomingHeaderHandler process.IncomingHeaderSubscriber,
) (notifierProcess.WSClient, error) {
	argsNotifier := factory.ArgsCreateSovereignNotifier{
		MarshallerType:   config.WebSocketConfig.MarshallerType,
		SubscribedEvents: getNotifierSubscribedEvents(config.SubscribedEvents),
		HasherType:       config.WebSocketConfig.HasherType,
	}

	sovereignNotifier, err := factory.CreateSovereignNotifier(argsNotifier)
	if err != nil {
		return nil, err
	}

	err = sovereignNotifier.RegisterHandler(incomingHeaderHandler)
	if err != nil {
		return nil, err
	}

	argsWsReceiver := factory.ArgsWsClientReceiverNotifier{
		WebSocketConfig: notifierCfg.WebSocketConfig{
			Url:                config.WebSocketConfig.Url,
			MarshallerType:     config.WebSocketConfig.MarshallerType,
			Mode:               config.WebSocketConfig.Mode,
			RetryDuration:      config.WebSocketConfig.RetryDuration,
			WithAcknowledge:    config.WebSocketConfig.WithAcknowledge,
			BlockingAckOnError: config.WebSocketConfig.BlockingAckOnError,
			AcknowledgeTimeout: config.WebSocketConfig.AcknowledgeTimeout,
			Version:            config.WebSocketConfig.Version,
		},
		SovereignNotifier: sovereignNotifier,
	}

	return factory.CreateWsClientReceiverNotifier(argsWsReceiver)
}

func getNotifierSubscribedEvents(events []config.SubscribedEvent) []notifierCfg.SubscribedEvent {
	ret := make([]notifierCfg.SubscribedEvent, len(events))

	for idx, event := range events {
		ret[idx] = notifierCfg.SubscribedEvent{
			Identifier: event.Identifier,
			Addresses:  event.Addresses,
		}
	}

	return ret
}
