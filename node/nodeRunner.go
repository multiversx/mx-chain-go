package node

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/api/gin"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/cmd/node/metrics"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/closing"
	dbLookupFactory "github.com/ElrondNetwork/elrond-go/core/dblookupext/factory"
	"github.com/ElrondNetwork/elrond-go/core/forking"
	"github.com/ElrondNetwork/elrond-go/core/logging"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/facade"
	"github.com/ElrondNetwork/elrond-go/facade/disabled"
	mainFactory "github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/genesis/parsing"
	"github.com/ElrondNetwork/elrond-go/health"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	"github.com/ElrondNetwork/elrond-go/sharding"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/trigger"
	"github.com/google/gops/agent"
)

const (
	defaultLogsPath = "logs"
	logFilePrefix   = "elrond-go"
	maxTimeToClose  = 10 * time.Second
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

	log.Info("starting node", "version", flagsConfig.Version, "pid", os.Getpid())
	log.Debug("config", "file", configurationPaths.Genesis)

	enableGopsIfNeeded(flagsConfig.EnableGops)

	fileLogging, err := nr.attachFileLogger()
	if err != nil {
		return err
	}

	configurationPaths.Nodes, err = nr.getNodesFileName()
	if err != nil {
		return err
	}

	log.Debug("config", "file", configurationPaths.Nodes)

	err = cleanupStorageIfNecessary(flagsConfig.WorkingDir, flagsConfig.CleanupStorage)
	if err != nil {
		return err
	}

	printEnableEpochs(nr.configs)

	logGoroutinesNumber(0)

	err = nr.startShufflingProcessLoop(chanStopNodeProcess)
	if err != nil {
		return err
	}

	log.Debug("closing node")
	if !check.IfNil(fileLogging) {
		err = fileLogging.Close()
		log.LogIfError(err)
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
	log.Debug(readEpochFor("staking v2 epoch"), "epoch", enableEpochs.StakingV2Epoch)
	log.Debug(readEpochFor("stake"), "epoch", enableEpochs.StakeEnableEpoch)
	log.Debug(readEpochFor("double key protection"), "epoch", enableEpochs.DoubleKeyProtectionEnableEpoch)
	log.Debug(readEpochFor("esdt"), "epoch", enableEpochs.ESDTEnableEpoch)
	log.Debug(readEpochFor("governance"), "epoch", enableEpochs.GovernanceEnableEpoch)
	log.Debug(readEpochFor("delegation manager"), "epoch", enableEpochs.DelegationManagerEnableEpoch)
	log.Debug(readEpochFor("delegation smart contract"), "epoch", enableEpochs.DelegationSmartContractEnableEpoch)

	gasSchedule := configs.EpochConfig.GasSchedule

	log.Debug(readEpochFor("gas schedule directories paths"), "epoch", gasSchedule.GasScheduleByEpochs)
}

func (nr *nodeRunner) startShufflingProcessLoop(
	chanStopNodeProcess1 chan endProcess.ArgEndProcess,
) error {
	configs := nr.configs
	flagsConfig := configs.FlagsConfig
	configurationPaths := configs.ConfigurationPathsHolder
	for {
		goRoutinesNumberStart := runtime.NumGoroutine()

		log.Debug("\n\n====================Starting managedComponents creation================================")

		log.Debug("creating core components")
		managedCoreComponents, err := nr.CreateManagedCoreComponents(
			chanStopNodeProcess1,
		)
		if err != nil {
			return err
		}

		log.Debug("creating crypto components")
		managedCryptoComponents, err := nr.CreateManagedCryptoComponents(managedCoreComponents)
		if err != nil {
			return err
		}

		log.Debug("creating network components")
		managedNetworkComponents, err := nr.CreateManagedNetworkComponents(managedCoreComponents)
		if err != nil {
			return err
		}

		log.Debug("creating disabled API services")
		httpServerWrapper, err := nr.createInitialHttpServer()
		if err != nil {
			return err
		}

		log.Debug("creating bootstrap components")
		managedBootstrapComponents, err := nr.CreateManagedBootstrapComponents(managedCoreComponents, managedCryptoComponents, managedNetworkComponents)
		if err != nil {
			return err
		}

		nr.logInformation(managedCoreComponents, managedCryptoComponents, managedBootstrapComponents)

		log.Debug("creating state components")
		managedStateComponents, err := nr.CreateManagedStateComponents(managedCoreComponents, managedBootstrapComponents)
		if err != nil {
			return err
		}

		log.Debug("creating data components")
		managedDataComponents, err := nr.CreateManagedDataComponents(managedCoreComponents, managedBootstrapComponents)
		if err != nil {
			return err
		}

		log.Debug("creating healthService")
		healthService := nr.createHealthService(flagsConfig, managedDataComponents)

		log.Trace("creating metrics")
		err = nr.createMetrics(managedCoreComponents, managedCryptoComponents, managedBootstrapComponents)
		if err != nil {
			return err
		}

		nodesShufflerOut, err := mainFactory.CreateNodesShuffleOut(
			managedCoreComponents.GenesisNodesSetup(),
			configs.GeneralConfig.EpochStartConfig,
			managedCoreComponents.ChanStopNodeProcess(),
		)
		if err != nil {
			return err
		}

		log.Debug("creating nodes coordinator")
		nodesCoordinator, err := mainFactory.CreateNodesCoordinator(
			nodesShufflerOut,
			managedCoreComponents.GenesisNodesSetup(),
			configs.PreferencesConfig.Preferences,
			managedCoreComponents.EpochStartNotifierWithConfirm(),
			managedCryptoComponents.PublicKey(),
			managedCoreComponents.InternalMarshalizer(),
			managedCoreComponents.Hasher(),
			managedCoreComponents.Rater(),
			managedDataComponents.StorageService().GetStorer(dataRetriever.BootstrapUnit),
			managedCoreComponents.NodesShuffler(),
			managedBootstrapComponents.ShardCoordinator().SelfId(),
			managedBootstrapComponents.EpochBootstrapParams(),
			managedBootstrapComponents.EpochBootstrapParams().Epoch(),
		)
		if err != nil {
			return err
		}

		log.Trace("starting status pooling components")
		managedStatusComponents, err := nr.CreateManagedStatusComponents(
			managedCoreComponents,
			managedNetworkComponents,
			managedBootstrapComponents,
			managedDataComponents,
			managedStateComponents,
			nodesCoordinator,
			configurationPaths.ElasticSearchTemplatesPath,
			configs.ImportDbConfig.IsImportDBMode,
		)
		if err != nil {
			return err
		}

		argsGasScheduleNotifier := forking.ArgsNewGasScheduleNotifier{
			GasScheduleConfig: configs.EpochConfig.GasSchedule,
			ConfigDir:         configurationPaths.GasScheduleDirectoryName,
			EpochNotifier:     managedCoreComponents.EpochNotifier(),
		}
		gasScheduleNotifier, err := forking.NewGasScheduleNotifier(argsGasScheduleNotifier)
		if err != nil {
			return err
		}

		log.Trace("creating process components")
		managedProcessComponents, err := nr.CreateManagedProcessComponents(
			managedCoreComponents,
			managedCryptoComponents,
			managedNetworkComponents,
			managedBootstrapComponents,
			managedStateComponents,
			managedDataComponents,
			managedStatusComponents,
			gasScheduleNotifier,
			nodesCoordinator,
		)
		if err != nil {
			return err
		}

		managedStatusComponents.SetForkDetector(managedProcessComponents.ForkDetector())
		err = managedStatusComponents.StartPolling()
		if err != nil {
			return err
		}

		elasticIndexer := managedStatusComponents.ElasticIndexer()
		if !elasticIndexer.IsNilIndexer() {
			elasticIndexer.SetTxLogsProcessor(managedProcessComponents.TxLogsProcessor())
			managedProcessComponents.TxLogsProcessor().EnableLogToBeSavedInCache()
		}

		log.Debug("starting node...")

		managedConsensusComponents, err := nr.CreateManagedConsensusComponents(
			managedCoreComponents,
			managedNetworkComponents,
			managedCryptoComponents,
			managedBootstrapComponents,
			managedDataComponents,
			managedStateComponents,
			managedStatusComponents,
			managedProcessComponents,
			nodesCoordinator,
			nodesShufflerOut,
		)
		if err != nil {
			return err
		}

		managedHeartbeatComponents, err := nr.CreateManagedHeartbeatComponents(
			managedCoreComponents,
			managedNetworkComponents,
			managedCryptoComponents,
			managedDataComponents,
			managedProcessComponents,
			managedConsensusComponents.HardforkTrigger(),
			managedProcessComponents.NodeRedundancyHandler(),
		)

		if err != nil {
			return err
		}

		log.Trace("creating node structure")
		currentNode, err := CreateNode(
			configs.GeneralConfig,
			managedBootstrapComponents,
			managedCoreComponents,
			managedCryptoComponents,
			managedDataComponents,
			managedNetworkComponents,
			managedProcessComponents,
			managedStateComponents,
			managedStatusComponents,
			managedHeartbeatComponents,
			managedConsensusComponents,
			flagsConfig.BootstrapRoundIndex,
			configs.ImportDbConfig.IsImportDBMode,
		)
		if err != nil {
			return err
		}

		if managedBootstrapComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
			log.Trace("activating nodesCoordinator's validators indexing")
			indexValidatorsListIfNeeded(
				elasticIndexer,
				nodesCoordinator,
				managedProcessComponents.EpochStartTrigger().Epoch(),
			)
		}

		log.Debug("updating the API service after creating the node facade")
		ef, err := nr.createApiFacade(currentNode, httpServerWrapper, gasScheduleNotifier)
		if err != nil {
			return err
		}

		log.Info("application is now running")
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		err = waitForSignal(sigs, managedCoreComponents.ChanStopNodeProcess(), healthService, ef, httpServerWrapper, currentNode, goRoutinesNumberStart)
		if err != nil {
			break
		}
	}
	return nil
}

func (nr *nodeRunner) createApiFacade(
	currentNode *Node,
	upgradableHttpServer shared.UpgradeableHttpServerHandler,
	gasScheduleNotifier core.GasScheduleNotifier,
) (closing.Closer, error) {
	configs := nr.configs

	log.Trace("creating api resolver structure")

	apiResolverArgs := &mainFactory.ApiResolverArgs{
		Configs:             configs,
		CoreComponents:      currentNode.coreComponents,
		DataComponents:      currentNode.dataComponents,
		StateComponents:     currentNode.stateComponents,
		BootstrapComponents: currentNode.bootstrapComponents,
		CryptoComponents:    currentNode.cryptoComponents,
		ProcessComponents:   currentNode.processComponents,
		GasScheduleNotifier: gasScheduleNotifier,
	}

	apiResolver, err := mainFactory.CreateApiResolver(apiResolverArgs)
	if err != nil {
		return nil, err
	}

	log.Trace("creating elrond node facade")

	flagsConfig := configs.FlagsConfig

	argNodeFacade := facade.ArgNodeFacade{
		Node:                   currentNode,
		ApiResolver:            apiResolver,
		TxSimulatorProcessor:   currentNode.processComponents.TransactionSimulatorProcessor(),
		RestAPIServerDebugMode: flagsConfig.EnableRestAPIServerDebugMode,
		WsAntifloodConfig:      configs.GeneralConfig.Antiflood.WebServer,
		FacadeConfig: config.FacadeConfig{
			RestApiInterface: flagsConfig.RestApiInterface,
			PprofEnabled:     flagsConfig.EnablePprof,
		},
		ApiRoutesConfig: *configs.ApiRoutesConfig,
		AccountsState:   currentNode.stateComponents.AccountsAdapter(),
		PeerState:       currentNode.stateComponents.PeerAccounts(),
	}

	ef, err := facade.NewNodeFacade(argNodeFacade)
	if err != nil {
		return nil, fmt.Errorf("%w while creating NodeFacade", err)
	}

	ef.SetSyncer(currentNode.coreComponents.SyncTimer())
	ef.SetTpsBenchmark(currentNode.statusComponents.TpsBenchmark())

	err = upgradableHttpServer.GetHttpServer().Close()
	if err != nil {
		return nil, err
	}

	err = upgradableHttpServer.UpdateFacade(ef)
	if err != nil {
		return nil, err
	}

	go upgradableHttpServer.GetHttpServer().Start()

	log.Debug("updated node facade and restarted the API services")

	log.Trace("starting background services")

	return ef, nil
}

func (nr *nodeRunner) createInitialHttpServer() (shared.UpgradeableHttpServerHandler, error) {
	httpServerArgs := gin.ArgsNewWebServer{
		Facade:          disabled.NewDisabledNodeFacade(nr.configs.FlagsConfig.RestApiInterface),
		ApiConfig:       *nr.configs.ApiRoutesConfig,
		AntiFloodConfig: nr.configs.GeneralConfig.Antiflood.WebServer,
	}

	httpServerWrapper, err := gin.NewGinWebServerHandler(httpServerArgs)
	if err != nil {
		return nil, err
	}

	httpSever, err := httpServerWrapper.CreateHttpServer()
	if err != nil {
		return nil, err
	}

	err = httpServerWrapper.SetHttpServer(httpSever)
	if err != nil {
		return nil, err
	}

	go httpSever.Start()

	return httpServerWrapper, nil
}

func (nr *nodeRunner) createMetrics(
	managedCoreComponents mainFactory.CoreComponentsHandler,
	managedCryptoComponents mainFactory.CryptoComponentsHandler,
	managedBootstrapComponents mainFactory.BootstrapComponentsHandler,
) error {
	err := metrics.InitMetrics(
		managedCoreComponents.StatusHandlerUtils(),
		managedCryptoComponents.PublicKeyString(),
		managedBootstrapComponents.NodeType(),
		managedBootstrapComponents.ShardCoordinator(),
		managedCoreComponents.GenesisNodesSetup(),
		nr.configs.FlagsConfig.Version,
		nr.configs.EconomicsConfig,
		nr.configs.GeneralConfig.EpochStartConfig.RoundsPerEpoch,
		managedCoreComponents.MinTransactionVersion(),
		nr.configs.EpochConfig,
	)

	if err != nil {
		return err
	}

	metrics.SaveStringMetric(managedCoreComponents.StatusHandler(), core.MetricNodeDisplayName, nr.configs.PreferencesConfig.Preferences.NodeDisplayName)
	metrics.SaveStringMetric(managedCoreComponents.StatusHandler(), core.MetricChainId, managedCoreComponents.ChainID())
	metrics.SaveUint64Metric(managedCoreComponents.StatusHandler(), core.MetricGasPerDataByte, managedCoreComponents.EconomicsData().GasPerDataByte())
	metrics.SaveUint64Metric(managedCoreComponents.StatusHandler(), core.MetricMinGasPrice, managedCoreComponents.EconomicsData().MinGasPrice())
	metrics.SaveUint64Metric(managedCoreComponents.StatusHandler(), core.MetricMinGasLimit, managedCoreComponents.EconomicsData().MinGasLimit())
	metrics.SaveStringMetric(managedCoreComponents.StatusHandler(), core.MetricRewardsTopUpGradientPoint, managedCoreComponents.EconomicsData().RewardsTopUpGradientPoint().String())
	metrics.SaveStringMetric(managedCoreComponents.StatusHandler(), core.MetricTopUpFactor, fmt.Sprintf("%g", managedCoreComponents.EconomicsData().RewardsTopUpFactor()))
	metrics.SaveStringMetric(managedCoreComponents.StatusHandler(), core.MetricGasPriceModifier, fmt.Sprintf("%g", managedCoreComponents.EconomicsData().GasPriceModifier()))

	return nil
}

func (nr *nodeRunner) createHealthService(flagsConfig *config.ContextFlagsConfig, managedDataComponents mainFactory.DataComponentsHandler) closing.Closer {
	healthService := health.NewHealthService(nr.configs.GeneralConfig.Health, flagsConfig.WorkingDir)
	if flagsConfig.UseHealthService {
		healthService.Start()
	}

	healthService.RegisterComponent(managedDataComponents.Datapool().Transactions())
	healthService.RegisterComponent(managedDataComponents.Datapool().UnsignedTransactions())
	healthService.RegisterComponent(managedDataComponents.Datapool().RewardTransactions())
	return healthService
}

// CreateManagedConsensusComponents is the managed consensus components factory
func (nr *nodeRunner) CreateManagedConsensusComponents(
	managedCoreComponents mainFactory.CoreComponentsHandler,
	managedNetworkComponents mainFactory.NetworkComponentsHandler,
	managedCryptoComponents mainFactory.CryptoComponentsHandler,
	managedBootstrapComponents mainFactory.BootstrapComponentsHandler,
	managedDataComponents mainFactory.DataComponentsHandler,
	managedStateComponents mainFactory.StateComponentsHandler,
	managedStatusComponents mainFactory.StatusComponentsHandler,
	managedProcessComponents mainFactory.ProcessComponentsHandler,
	nodesCoordinator sharding.NodesCoordinator,
	nodesShuffledOut update.Closer,
) (mainFactory.ConsensusComponentsHandler, error) {
	hardForkTrigger, err := CreateHardForkTrigger(
		nr.configs.GeneralConfig,
		nr.configs.EpochConfig,
		managedBootstrapComponents.ShardCoordinator(),
		nodesCoordinator,
		nodesShuffledOut,
		managedCoreComponents,
		managedStateComponents,
		managedDataComponents,
		managedCryptoComponents,
		managedProcessComponents,
		managedNetworkComponents,
		managedCoreComponents.EpochStartNotifierWithConfirm(),
		managedProcessComponents.ImportStartHandler(),
		nr.configs.FlagsConfig.WorkingDir,
	)
	if err != nil {
		return nil, err
	}

	consensusArgs := mainFactory.ConsensusComponentsFactoryArgs{
		Config:              *nr.configs.GeneralConfig,
		BootstrapRoundIndex: nr.configs.FlagsConfig.BootstrapRoundIndex,
		HardforkTrigger:     hardForkTrigger,
		CoreComponents:      managedCoreComponents,
		NetworkComponents:   managedNetworkComponents,
		CryptoComponents:    managedCryptoComponents,
		DataComponents:      managedDataComponents,
		ProcessComponents:   managedProcessComponents,
		StateComponents:     managedStateComponents,
		StatusComponents:    managedStatusComponents,
		IsInImportMode:      nr.configs.ImportDbConfig.IsImportDBMode,
	}

	consensusFactory, err := mainFactory.NewConsensusComponentsFactory(consensusArgs)
	if err != nil {
		return nil, fmt.Errorf("NewConsensusComponentsFactory failed: %w", err)
	}

	managedConsensusComponents, err := mainFactory.NewManagedConsensusComponents(consensusFactory)
	if err != nil {
		return nil, err
	}

	err = managedConsensusComponents.Create()
	if err != nil {
		return nil, err
	}
	return managedConsensusComponents, nil
}

// CreateManagedHeartbeatComponents is the managed heartbeat components factory
func (nr *nodeRunner) CreateManagedHeartbeatComponents(
	managedCoreComponents mainFactory.CoreComponentsHandler,
	managedNetworkComponents mainFactory.NetworkComponentsHandler,
	managedCryptoComponents mainFactory.CryptoComponentsHandler,
	managedDataComponents mainFactory.DataComponentsHandler,
	managedProcessComponents mainFactory.ProcessComponentsHandler,
	hardforkTrigger HardforkTrigger,
	redundancyHandler consensus.NodeRedundancyHandler,
) (mainFactory.HeartbeatComponentsHandler, error) {
	genesisTime := time.Unix(managedCoreComponents.GenesisNodesSetup().GetStartTime(), 0)

	heartbeatArgs := mainFactory.HeartbeatComponentsFactoryArgs{
		Config:            *nr.configs.GeneralConfig,
		Prefs:             *nr.configs.PreferencesConfig,
		AppVersion:        nr.configs.FlagsConfig.Version,
		GenesisTime:       genesisTime,
		HardforkTrigger:   hardforkTrigger,
		RedundancyHandler: redundancyHandler,
		CoreComponents:    managedCoreComponents,
		DataComponents:    managedDataComponents,
		NetworkComponents: managedNetworkComponents,
		CryptoComponents:  managedCryptoComponents,
		ProcessComponents: managedProcessComponents,
	}

	heartbeatComponentsFactory, err := mainFactory.NewHeartbeatComponentsFactory(heartbeatArgs)
	if err != nil {
		return nil, fmt.Errorf("NewHeartbeatComponentsFactory failed: %w", err)
	}

	managedHeartbeatComponents, err := mainFactory.NewManagedHeartbeatComponents(heartbeatComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedHeartbeatComponents.Create()
	if err != nil {
		return nil, err
	}
	return managedHeartbeatComponents, nil
}

func waitForSignal(
	sigs chan os.Signal,
	chanStopNodeProcess chan endProcess.ArgEndProcess,
	healthService closing.Closer,
	ef closing.Closer,
	httpServer shared.UpgradeableHttpServerHandler,
	currentNode *Node,
	goRoutinesNumberStart int,
) error {
	var sig endProcess.ArgEndProcess
	reshuffled := false
	select {
	case <-sigs:
		log.Info("terminating at user's signal...")
	case sig = <-chanStopNodeProcess:
		log.Info("terminating at internal stop signal", "reason", sig.Reason, "description", sig.Description)
		if sig.Reason == core.ShuffledOut {
			reshuffled = true
		}
	}

	chanCloseComponents := make(chan struct{})
	go func() {
		closeAllComponents(healthService, ef, httpServer, currentNode, chanCloseComponents)
	}()

	select {
	case <-chanCloseComponents:
		log.Debug("Closed all components gracefully")
	case <-time.After(maxTimeToClose):
		log.Warn("force closing the node", "error", "closeAllComponents did not finished on time")
		return fmt.Errorf("did NOT close all components gracefully")
	}

	if reshuffled {
		log.Info("=============================Shuffled out - soft restart==================================")
		logGoroutinesNumber(goRoutinesNumberStart)
	} else {
		return fmt.Errorf("not reshuffled, closing")
	}

	return nil
}

func (nr *nodeRunner) logInformation(
	managedCoreComponents mainFactory.CoreComponentsHandler,
	managedCryptoComponents mainFactory.CryptoComponentsHandler,
	managedBootstrapComponents mainFactory.BootstrapComponentsHandler,
) {
	log.Info("Bootstrap", "epoch", managedBootstrapComponents.EpochBootstrapParams().Epoch())
	if managedBootstrapComponents.EpochBootstrapParams().NodesConfig() != nil {
		log.Info("the epoch from nodesConfig is",
			"epoch", managedBootstrapComponents.EpochBootstrapParams().NodesConfig().CurrentEpoch)
	}

	var shardIdString = core.GetShardIDString(managedBootstrapComponents.ShardCoordinator().SelfId())
	logger.SetCorrelationShard(shardIdString)

	sessionInfoFileOutput := fmt.Sprintf("%s:%s\n%s:%s\n%s:%v\n%s:%s\n%s:%v\n",
		"PkBlockSign", managedCryptoComponents.PublicKeyString(),
		"ShardId", shardIdString,
		"TotalShards", managedBootstrapComponents.ShardCoordinator().NumberOfShards(),
		"AppVersion", nr.configs.FlagsConfig.Version,
		"GenesisTimeStamp", managedCoreComponents.GenesisTime().Unix(),
	)

	sessionInfoFileOutput += "\nStarted with parameters:\n"
	sessionInfoFileOutput += nr.configs.FlagsConfig.SessionInfoFileOutput

	nr.logSessionInformation(nr.configs.FlagsConfig.WorkingDir, sessionInfoFileOutput, managedCoreComponents)
}

func (nr *nodeRunner) getNodesFileName() (string, error) {
	flagsConfig := nr.configs.FlagsConfig
	configurationPaths := nr.configs.ConfigurationPathsHolder
	nodesFileName := configurationPaths.Nodes

	exportFolder := filepath.Join(flagsConfig.WorkingDir, nr.configs.GeneralConfig.Hardfork.ImportFolder)
	if nr.configs.GeneralConfig.Hardfork.AfterHardFork {
		exportFolderNodesSetupPath := filepath.Join(exportFolder, core.NodesSetupJsonFileName)
		if !core.DoesFileExist(exportFolderNodesSetupPath) {
			return "", fmt.Errorf("cannot find %s in the export folder", core.NodesSetupJsonFileName)
		}

		nodesFileName = exportFolderNodesSetupPath
	}
	return nodesFileName, nil
}

func logGoroutinesNumber(goRoutinesNumberStart int) {
	buffer := new(bytes.Buffer)
	err := pprof.Lookup("goroutine").WriteTo(buffer, 2)
	if err != nil {
		log.Error("could not dump goroutines")
	}
	log.Debug("go routines number",
		"start", goRoutinesNumberStart,
		"end", runtime.NumGoroutine())

	log.Debug(buffer.String())
}

// CreateManagedStatusComponents is the managed status components factory
func (nr *nodeRunner) CreateManagedStatusComponents(
	managedCoreComponents mainFactory.CoreComponentsHandler,
	managedNetworkComponents mainFactory.NetworkComponentsHandler,
	managedBootstrapComponents mainFactory.BootstrapComponentsHandler,
	managedDataComponents mainFactory.DataComponentsHandler,
	managedStateComponents mainFactory.StateComponentsHandler,
	nodesCoordinator sharding.NodesCoordinator,
	elasticTemplatePath string,
	isInImportMode bool,
) (mainFactory.StatusComponentsHandler, error) {
	statArgs := mainFactory.StatusComponentsFactoryArgs{
		Config:               *nr.configs.GeneralConfig,
		ExternalConfig:       *nr.configs.ExternalConfig,
		EconomicsConfig:      *nr.configs.EconomicsConfig,
		ShardCoordinator:     managedBootstrapComponents.ShardCoordinator(),
		NodesCoordinator:     nodesCoordinator,
		EpochStartNotifier:   managedCoreComponents.EpochStartNotifierWithConfirm(),
		CoreComponents:       managedCoreComponents,
		DataComponents:       managedDataComponents,
		NetworkComponents:    managedNetworkComponents,
		StateComponents:      managedStateComponents,
		ElasticTemplatesPath: elasticTemplatePath,
		IsInImportMode:       isInImportMode,
	}

	statusComponentsFactory, err := mainFactory.NewStatusComponentsFactory(statArgs)
	if err != nil {
		return nil, fmt.Errorf("NewStatusComponentsFactory failed: %w", err)
	}

	managedStatusComponents, err := mainFactory.NewManagedStatusComponents(statusComponentsFactory)
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
	managedCoreComponents mainFactory.CoreComponentsHandler,
) {
	statsFolder := filepath.Join(workingDir, core.DefaultStatsPath)
	configurationPaths := nr.configs.ConfigurationPathsHolder
	copyConfigToStatsFolder(
		statsFolder,
		configurationPaths.GasScheduleDirectoryName,
		[]string{
			configurationPaths.MainConfig,
			configurationPaths.Economics,
			configurationPaths.Ratings,
			configurationPaths.Preferences,
			configurationPaths.P2p,
			configurationPaths.Genesis,
			configurationPaths.Nodes,
			configurationPaths.ApiRoutes,
			configurationPaths.External,
			configurationPaths.SystemSC,
		})

	statsFile := filepath.Join(statsFolder, "session.info")
	err := ioutil.WriteFile(statsFile, []byte(sessionInfoFileOutput), os.ModePerm)
	log.LogIfError(err)

	computedRatingsDataStr := createStringFromRatingsData(managedCoreComponents.RatingsData())
	log.Debug("rating data", "rating", computedRatingsDataStr)
}

// CreateManagedProcessComponents is the managed process components factory
func (nr *nodeRunner) CreateManagedProcessComponents(
	managedCoreComponents mainFactory.CoreComponentsHandler,
	managedCryptoComponents mainFactory.CryptoComponentsHandler,
	managedNetworkComponents mainFactory.NetworkComponentsHandler,
	managedBootstrapComponents mainFactory.BootstrapComponentsHandler,
	managedStateComponents mainFactory.StateComponentsHandler,
	managedDataComponents mainFactory.DataComponentsHandler,
	managedStatusComponents mainFactory.StatusComponentsHandler,
	gasScheduleNotifier core.GasScheduleNotifier,
	nodesCoordinator sharding.NodesCoordinator,
) (mainFactory.ProcessComponentsHandler, error) {
	configs := nr.configs
	configurationPaths := nr.configs.ConfigurationPathsHolder
	importStartHandler, err := trigger.NewImportStartHandler(filepath.Join(configs.FlagsConfig.WorkingDir, core.DefaultDBPath), configs.FlagsConfig.Version)
	if err != nil {
		return nil, err
	}

	totalSupply, ok := big.NewInt(0).SetString(configs.EconomicsConfig.GlobalSettings.GenesisTotalSupply, 10)
	if !ok {
		return nil, fmt.Errorf("can not parse total suply from economics.toml, %s is not a valid value",
			configs.EconomicsConfig.GlobalSettings.GenesisTotalSupply)
	}

	accountsParser, err := parsing.NewAccountsParser(
		configurationPaths.Genesis,
		totalSupply,
		managedCoreComponents.AddressPubKeyConverter(),
		managedCryptoComponents.TxSignKeyGen(),
	)
	if err != nil {
		return nil, err
	}

	smartContractParser, err := parsing.NewSmartContractsParser(
		configurationPaths.SmartContracts,
		managedCoreComponents.AddressPubKeyConverter(),
		managedCryptoComponents.TxSignKeyGen(),
	)
	if err != nil {
		return nil, err
	}

	historyRepoFactoryArgs := &dbLookupFactory.ArgsHistoryRepositoryFactory{
		SelfShardID: managedBootstrapComponents.ShardCoordinator().SelfId(),
		Config:      configs.GeneralConfig.DbLookupExtensions,
		Hasher:      managedCoreComponents.Hasher(),
		Marshalizer: managedCoreComponents.InternalMarshalizer(),
		Store:       managedDataComponents.StorageService(),
	}
	historyRepositoryFactory, err := dbLookupFactory.NewHistoryRepositoryFactory(historyRepoFactoryArgs)
	if err != nil {
		return nil, err
	}

	whiteListCache, err := storageUnit.NewCache(storageFactory.GetCacherFromConfig(configs.GeneralConfig.WhiteListPool))
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
	requestedItemsHandler := timecache.NewTimeCache(
		time.Duration(uint64(time.Millisecond) * managedCoreComponents.GenesisNodesSetup().GetRoundDuration()))

	processArgs := mainFactory.ProcessComponentsFactoryArgs{
		Config:                 *configs.GeneralConfig,
		EpochConfig:            *configs.EpochConfig,
		PrefConfigs:            configs.PreferencesConfig.Preferences,
		ImportDBConfig:         *configs.ImportDbConfig,
		AccountsParser:         accountsParser,
		SmartContractParser:    smartContractParser,
		GasSchedule:            gasScheduleNotifier,
		NodesCoordinator:       nodesCoordinator,
		Data:                   managedDataComponents,
		CoreData:               managedCoreComponents,
		Crypto:                 managedCryptoComponents,
		State:                  managedStateComponents,
		Network:                managedNetworkComponents,
		BootstrapComponents:    managedBootstrapComponents,
		StatusComponents:       managedStatusComponents,
		RequestedItemsHandler:  requestedItemsHandler,
		WhiteListHandler:       whiteListRequest,
		WhiteListerVerifiedTxs: whiteListerVerifiedTxs,
		MaxRating:              configs.RatingsConfig.General.MaxRating,
		SystemSCConfig:         configs.SystemSCConfig,
		Version:                configs.FlagsConfig.Version,
		ImportStartHandler:     importStartHandler,
		WorkingDir:             configs.FlagsConfig.WorkingDir,
		HistoryRepo:            historyRepository,
	}
	processComponentsFactory, err := mainFactory.NewProcessComponentsFactory(processArgs)
	if err != nil {
		return nil, fmt.Errorf("NewProcessComponentsFactory failed: %w", err)
	}

	managedProcessComponents, err := mainFactory.NewManagedProcessComponents(processComponentsFactory)
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
	managedCoreComponents mainFactory.CoreComponentsHandler,
	managedBootstrapComponents mainFactory.BootstrapComponentsHandler,
) (mainFactory.DataComponentsHandler, error) {
	configs := nr.configs
	storerEpoch := managedBootstrapComponents.EpochBootstrapParams().Epoch()
	if !configs.GeneralConfig.StoragePruning.Enabled {
		// TODO: refactor this as when the pruning storer is disabled, the default directory path is Epoch_0
		// and it should be Epoch_ALL or something similar
		storerEpoch = 0
	}

	dataArgs := mainFactory.DataComponentsFactoryArgs{
		Config:                        *configs.GeneralConfig,
		ShardCoordinator:              managedBootstrapComponents.ShardCoordinator(),
		Core:                          managedCoreComponents,
		EpochStartNotifier:            managedCoreComponents.EpochStartNotifierWithConfirm(),
		CurrentEpoch:                  storerEpoch,
		CreateTrieEpochRootHashStorer: configs.ImportDbConfig.ImportDbSaveTrieEpochRootHash,
	}

	dataComponentsFactory, err := mainFactory.NewDataComponentsFactory(dataArgs)
	if err != nil {
		return nil, fmt.Errorf("NewDataComponentsFactory failed: %w", err)
	}
	managedDataComponents, err := mainFactory.NewManagedDataComponents(dataComponentsFactory)
	if err != nil {
		return nil, err
	}
	err = managedDataComponents.Create()
	if err != nil {
		return nil, err
	}

	err = managedCoreComponents.StatusHandlerUtils().UpdateStorerAndMetricsForPersistentHandler(
		managedDataComponents.StorageService().GetStorer(dataRetriever.StatusMetricsUnit),
	)

	if err != nil {
		return nil, err
	}

	return managedDataComponents, nil
}

// CreateManagedStateComponents is the managed state components factory
func (nr *nodeRunner) CreateManagedStateComponents(
	managedCoreComponents mainFactory.CoreComponentsHandler,
	managedBootstrapComponents mainFactory.BootstrapComponentsHandler,
) (mainFactory.StateComponentsHandler, error) {
	triesComponents, trieStorageManagers := managedBootstrapComponents.EpochStartBootstrapper().GetTriesComponents()
	stateArgs := mainFactory.StateComponentsFactoryArgs{
		Config:              *nr.configs.GeneralConfig,
		ShardCoordinator:    managedBootstrapComponents.ShardCoordinator(),
		Core:                managedCoreComponents,
		TriesContainer:      triesComponents,
		TrieStorageManagers: trieStorageManagers,
	}

	stateComponentsFactory, err := mainFactory.NewStateComponentsFactory(stateArgs)
	if err != nil {
		return nil, fmt.Errorf("NewStateComponentsFactory failed: %w", err)
	}

	managedStateComponents, err := mainFactory.NewManagedStateComponents(stateComponentsFactory)
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
	managedCoreComponents mainFactory.CoreComponentsHandler,
	managedCryptoComponents mainFactory.CryptoComponentsHandler,
	managedNetworkComponents mainFactory.NetworkComponentsHandler,
) (mainFactory.BootstrapComponentsHandler, error) {

	bootstrapComponentsFactoryArgs := mainFactory.BootstrapComponentsFactoryArgs{
		Config:            *nr.configs.GeneralConfig,
		EpochConfig:       *nr.configs.EpochConfig,
		PrefConfig:        *nr.configs.PreferencesConfig,
		ImportDbConfig:    *nr.configs.ImportDbConfig,
		WorkingDir:        nr.configs.FlagsConfig.WorkingDir,
		CoreComponents:    managedCoreComponents,
		CryptoComponents:  managedCryptoComponents,
		NetworkComponents: managedNetworkComponents,
	}

	bootstrapComponentsFactory, err := mainFactory.NewBootstrapComponentsFactory(bootstrapComponentsFactoryArgs)
	if err != nil {
		return nil, fmt.Errorf("NewBootstrapComponentsFactory failed: %w", err)
	}

	managedBootstrapComponents, err := mainFactory.NewManagedBootstrapComponents(bootstrapComponentsFactory)
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
	managedCoreComponents mainFactory.CoreComponentsHandler,
) (mainFactory.NetworkComponentsHandler, error) {

	networkComponentsFactoryArgs := mainFactory.NetworkComponentsFactoryArgs{
		P2pConfig:            *nr.configs.P2pConfig,
		MainConfig:           *nr.configs.GeneralConfig,
		RatingsConfig:        *nr.configs.RatingsConfig,
		StatusHandler:        managedCoreComponents.StatusHandler(),
		Marshalizer:          managedCoreComponents.InternalMarshalizer(),
		Syncer:               managedCoreComponents.SyncTimer(),
		BootstrapWaitSeconds: core.SecondsToWaitForP2PBootstrap,
	}
	if nr.configs.ImportDbConfig.IsImportDBMode {
		networkComponentsFactoryArgs.BootstrapWaitSeconds = 0
	}

	networkComponentsFactory, err := mainFactory.NewNetworkComponentsFactory(networkComponentsFactoryArgs)
	if err != nil {
		return nil, fmt.Errorf("NewNetworkComponentsFactory failed: %w", err)
	}

	managedNetworkComponents, err := mainFactory.NewManagedNetworkComponents(networkComponentsFactory)
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

	statusHandlersFactoryArgs := &factory.StatusHandlersFactoryArgs{
		UseTermUI: !nr.configs.FlagsConfig.UseLogView,
	}

	statusHandlersFactory, err := factory.NewStatusHandlersFactory(statusHandlersFactoryArgs)
	if err != nil {
		return nil, err
	}

	coreArgs := mainFactory.CoreComponentsFactoryArgs{
		Config:                *nr.configs.GeneralConfig,
		EpochConfig:           *nr.configs.EpochConfig,
		ImportDbConfig:        *nr.configs.ImportDbConfig,
		RatingsConfig:         *nr.configs.RatingsConfig,
		EconomicsConfig:       *nr.configs.EconomicsConfig,
		NodesFilename:         nr.configs.ConfigurationPathsHolder.Nodes,
		WorkingDirectory:      nr.configs.FlagsConfig.WorkingDir,
		ChanStopNodeProcess:   chanStopNodeProcess,
		StatusHandlersFactory: statusHandlersFactory,
	}

	coreComponentsFactory, err := mainFactory.NewCoreComponentsFactory(coreArgs)
	if err != nil {
		return nil, fmt.Errorf("NewCoreComponentsFactory failed: %w", err)
	}

	managedCoreComponents, err := mainFactory.NewManagedCoreComponents(coreComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedCoreComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedCoreComponents, nil
}

// CreateManagedCryptoComponents is the managed crypto components factory
func (nr *nodeRunner) CreateManagedCryptoComponents(
	managedCoreComponents mainFactory.CoreComponentsHandler,
) (mainFactory.CryptoComponentsHandler, error) {
	configs := nr.configs
	validatorKeyPemFileName := configs.ConfigurationPathsHolder.ValidatorKey
	cryptoComponentsHandlerArgs := mainFactory.CryptoComponentsFactoryArgs{
		ValidatorKeyPemFileName:              validatorKeyPemFileName,
		SkIndex:                              configs.FlagsConfig.ValidatorKeyIndex,
		Config:                               *configs.GeneralConfig,
		CoreComponentsHolder:                 managedCoreComponents,
		ActivateBLSPubKeyMessageVerification: configs.SystemSCConfig.StakingSystemSCConfig.ActivateBLSPubKeyMessageVerification,
		KeyLoader:                            &core.KeyLoader{},
		ImportModeNoSigCheck:                 configs.ImportDbConfig.ImportDbNoSigCheckFlag,
		IsInImportMode:                       configs.ImportDbConfig.IsImportDBMode,
	}

	cryptoComponentsFactory, err := mainFactory.NewCryptoComponentsFactory(cryptoComponentsHandlerArgs)
	if err != nil {
		return nil, fmt.Errorf("NewCryptoComponentsFactory failed: %w", err)
	}

	managedCryptoComponents, err := mainFactory.NewManagedCryptoComponents(cryptoComponentsFactory)
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
	if cleanupStorage {
		dbPath := filepath.Join(
			workingDir,
			core.DefaultDBPath)
		log.Trace("cleaning storage", "path", dbPath)
		err := os.RemoveAll(dbPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func copyConfigToStatsFolder(statsFolder string, gasScheduleFolder string, configs []string) {
	err := os.MkdirAll(statsFolder, os.ModePerm)
	log.LogIfError(err)

	err = copyDirectory(gasScheduleFolder, statsFolder)
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
		dstFilePath := path.Join(destination, fd.Name())
		if fd.IsDir() {
			err = copyDirectory(srcFilePath, dstFilePath)
			log.LogIfError(err)
		} else {
			copySingleFile(dstFilePath, srcFilePath)
		}
	}
	return nil
}

func copySingleFile(folder string, configFile string) {
	fileName := filepath.Base(configFile)

	source, err := core.OpenFile(configFile)
	if err != nil {
		return
	}
	defer func() {
		err = source.Close()
		if err != nil {
			log.Warn("copySingleFile", "Could not close file", source.Name(), "error", err.Error())
		}
	}()

	destPath := filepath.Join(folder, fileName)
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
	elasticIndexer process.Indexer,
	coordinator sharding.NodesCoordinator,
	epoch uint32,
) {
	if check.IfNil(elasticIndexer) {
		return
	}

	validatorsPubKeys, err := coordinator.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		log.Warn("GetAllEligibleValidatorPublicKeys for epoch 0 failed", "error", err)
	}

	if len(validatorsPubKeys) > 0 {
		elasticIndexer.SaveValidatorsPubKeys(validatorsPubKeys, epoch)
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
	whiteListCacheVerified, err := storageUnit.NewCache(storageFactory.GetCacherFromConfig(generalConfig.WhiteListerVerifiedTxs))
	if err != nil {
		return nil, err
	}
	return interceptors.NewWhiteListDataVerifier(whiteListCacheVerified)
}

func (nr *nodeRunner) attachFileLogger() (factory.FileLoggingHandler, error) {
	configs := nr.configs
	flagsConfig := configs.FlagsConfig
	var fileLogging factory.FileLoggingHandler
	var err error
	if flagsConfig.SaveLogFile {
		fileLogging, err = logging.NewFileLogging(flagsConfig.WorkingDir, defaultLogsPath, logFilePrefix)
		if err != nil {
			return nil, fmt.Errorf("%w creating a log file", err)
		}
	}

	err = logger.SetDisplayByteSlice(logger.ToHex)
	log.LogIfError(err)
	logger.ToggleCorrelation(flagsConfig.EnableLogCorrelation)
	logger.ToggleLoggerName(flagsConfig.EnableLogName)
	logLevelFlagValue := flagsConfig.LogLevel
	err = logger.SetLogLevel(logLevelFlagValue)
	if err != nil {
		return nil, err
	}

	if flagsConfig.DisableAnsiColor {
		err = logger.RemoveLogObserver(os.Stdout)
		if err != nil {
			return nil, err
		}

		err = logger.AddLogObserver(os.Stdout, &logger.PlainFormatter{})
		if err != nil {
			return nil, err
		}
	}
	log.Trace("logger updated", "level", logLevelFlagValue, "disable ANSI color", flagsConfig.DisableAnsiColor)

	if !check.IfNil(fileLogging) {
		err = fileLogging.ChangeFileLifeSpan(time.Second * time.Duration(configs.GeneralConfig.Logs.LogFileLifeSpanInSec))
		if err != nil {
			return nil, err
		}
	}

	return fileLogging, nil
}
