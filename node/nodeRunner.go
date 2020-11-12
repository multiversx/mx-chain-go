package node

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/cmd/node/metrics"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/closing"
	dbLookupFactory "github.com/ElrondNetwork/elrond-go/core/dblookupext/factory"
	"github.com/ElrondNetwork/elrond-go/core/forking"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/logging"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/facade"
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

type NodeRunner struct {
	configs *config.Configs
	log     logger.Logger
}

func NewNodeRunner(cfgs *config.Configs, log logger.Logger) (*NodeRunner, error) {
	return &NodeRunner{
		configs: cfgs,
		log:     log,
	}, nil
}

func (nr *NodeRunner) Start() error {
	log := nr.log
	configs := nr.configs
	flagsConfig := configs.FlagsConfig
	chanStopNodeProcess1 := make(chan endProcess.ArgEndProcess, 1)

	log.Info("starting node", "version", flagsConfig.Version, "pid", os.Getpid())
	log.Debug("config", "file", flagsConfig.GenesisFileName)

	enableGopsIfNeeded(flagsConfig.EnableGops, log)

	fileLogging, err := attachFileLogger(configs, log)
	if err != nil {
		return err
	}

	flagsConfig.NodesFileName, err = getNodesFileName(configs)
	if err != nil {
		return err
	}

	log.Debug("config", "file", flagsConfig.NodesFileName)

	err = cleanupStorageIfNecessary(flagsConfig.WorkingDir, flagsConfig.CleanupStorage, log)
	if err != nil {
		return err
	}

	logGoroutinesNumber(log, 0)

	for {
		goRoutinesNumberStart := runtime.NumGoroutine()

		log.Debug("\n\n====================Starting managedComponents creation================================")

		log.Debug("creating core components")
		managedCoreComponents, err := CreateManagedCoreComponents(
			configs,
			chanStopNodeProcess1,
		)
		if err != nil {
			return err
		}

		log.Debug("creating crypto components")
		managedCryptoComponents, err := CreateManagedCryptoComponents(configs, managedCoreComponents)
		if err != nil {
			return err
		}

		log.Debug("creating network components")
		managedNetworkComponents, err := CreateManagedNetworkComponents(configs, managedCoreComponents)
		if err != nil {
			return err
		}

		log.Debug("creating bootstrap components")
		managedBootstrapComponents, err := CreateManagedBootstrapComponents(configs, managedCoreComponents, managedCryptoComponents, managedNetworkComponents)
		if err != nil {
			return err
		}

		logInformation(log, configs, managedCoreComponents, managedCryptoComponents, managedBootstrapComponents)

		log.Debug("creating state components")
		managedStateComponents, err := CreateManagedStateComponents(configs, managedCoreComponents, managedBootstrapComponents)
		if err != nil {
			return err
		}

		log.Debug("creating data components")
		managedDataComponents, err := CreateManagedDataComponents(configs, managedCoreComponents, managedBootstrapComponents)
		if err != nil {
			return err
		}

		log.Debug("creating healthService")
		healthService := createHealthService(configs, flagsConfig, managedDataComponents)

		log.Trace("creating metrics")
		err = createMetrics(configs, managedCoreComponents, managedCryptoComponents, managedBootstrapComponents)
		if err != nil {
			return err
		}

		nodesShufflerOut, err := mainFactory.CreateNodesShuffleOut(managedCoreComponents.GenesisNodesSetup(), configs.GeneralConfig.EpochStartConfig, managedCoreComponents.ChanStopNodeProcess())
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
		managedStatusComponents, err := CreateManagedStatusComponents(
			configs,
			managedCoreComponents,
			managedNetworkComponents,
			managedBootstrapComponents,
			managedDataComponents,
			managedStateComponents,
			nodesCoordinator,
			flagsConfig.ElasticSearchTemplatesPath,
			flagsConfig.IsInImportMode,
		)
		if err != nil {
			return err
		}

		gasSchedule, err := core.LoadGasScheduleConfig(flagsConfig.GasScheduleConfigurationFileName)
		if err != nil {
			return err
		}

		log.Trace("creating process components")
		managedProcessComponents, err := CreateManagedProcessComponents(
			configs,
			managedCoreComponents,
			managedCryptoComponents,
			managedNetworkComponents,
			managedBootstrapComponents,
			managedStateComponents,
			managedDataComponents,
			managedStatusComponents,
			gasSchedule,
			nodesCoordinator,
		)
		if err != nil {
			return err
		}

		managedStatusComponents.SetForkDetector(managedProcessComponents.ForkDetector())
		err = managedStatusComponents.StartPolling()

		elasticIndexer := managedStatusComponents.ElasticIndexer()
		if !elasticIndexer.IsNilIndexer() {
			elasticIndexer.SetTxLogsProcessor(managedProcessComponents.TxLogsProcessor())
			managedProcessComponents.TxLogsProcessor().EnableLogToBeSavedInCache()
		}

		log.Debug("starting node...")

		managedConsensusComponents, err := CreateManagedConsensusComponents(configs, managedCoreComponents, managedNetworkComponents, managedCryptoComponents, managedBootstrapComponents, managedDataComponents, managedStateComponents, managedStatusComponents, managedProcessComponents, nodesCoordinator, nodesShufflerOut)
		if err != nil {
			return err
		}

		managedHeartbeatComponents, err := CreateManagedHeartbeatComponents(configs, managedCoreComponents, managedNetworkComponents, managedCryptoComponents, managedDataComponents, managedProcessComponents, managedConsensusComponents.HardforkTrigger())
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
				log,
			)
		}

		ef, err := createApiFacade(configs, currentNode, gasSchedule, log)
		if err != nil {
			return err
		}

		log.Info("application is now running")
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		err = waitForSignal(sigs, log, managedCoreComponents.ChanStopNodeProcess(), healthService, ef, currentNode, goRoutinesNumberStart)
		if err != nil {
			break
		}
	}

	log.Debug("closing node")
	if !check.IfNil(fileLogging) {
		err = fileLogging.Close()
		log.LogIfError(err)
	}

	return nil
}

func createApiFacade(configs *config.Configs, currentNode *Node, gasSchedule map[string]map[string]uint64, log logger.Logger) (closing.Closer, error) {
	log.Trace("creating api resolver structure")
	apiResolver, err := mainFactory.CreateApiResolver(
		configs.GeneralConfig,
		currentNode.stateComponents.AccountsAdapter(),
		currentNode.stateComponents.PeerAccounts(),
		currentNode.coreComponents.AddressPubKeyConverter(),
		currentNode.dataComponents.StorageService(),
		currentNode.dataComponents.Blockchain(),
		currentNode.coreComponents.InternalMarshalizer(),
		currentNode.coreComponents.Hasher(),
		currentNode.coreComponents.Uint64ByteSliceConverter(),
		currentNode.bootstrapComponents.ShardCoordinator(),
		currentNode.coreComponents.StatusHandlerUtils().Metrics(),
		gasSchedule,
		currentNode.coreComponents.EconomicsData(),
		currentNode.cryptoComponents.MessageSignVerifier(),
		currentNode.coreComponents.GenesisNodesSetup(),
		configs.SystemSCConfig,
		currentNode.coreComponents.Rater(),
		currentNode.coreComponents.EpochNotifier(),
	)
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

	log.Trace("starting background services")
	ef.StartBackgroundServices()
	return ef, nil
}

func createMetrics(configs *config.Configs, managedCoreComponents mainFactory.CoreComponentsHandler, managedCryptoComponents mainFactory.CryptoComponentsHandler, managedBootstrapComponents mainFactory.BootstrapComponentsHandler) error {
	err := metrics.InitMetrics(
		managedCoreComponents.StatusHandlerUtils(),
		managedCryptoComponents.PublicKeyString(),
		managedBootstrapComponents.NodeType(),
		managedBootstrapComponents.ShardCoordinator(),
		managedCoreComponents.GenesisNodesSetup(),
		configs.FlagsConfig.Version,
		configs.EconomicsConfig,
		configs.GeneralConfig.EpochStartConfig.RoundsPerEpoch,
		managedCoreComponents.MinTransactionVersion(),
	)

	if err != nil {
		return err
	}

	metrics.SaveStringMetric(managedCoreComponents.StatusHandler(), core.MetricNodeDisplayName, configs.PreferencesConfig.Preferences.NodeDisplayName)
	metrics.SaveStringMetric(managedCoreComponents.StatusHandler(), core.MetricChainId, managedCoreComponents.ChainID())
	metrics.SaveUint64Metric(managedCoreComponents.StatusHandler(), core.MetricGasPerDataByte, managedCoreComponents.EconomicsData().GasPerDataByte())
	metrics.SaveUint64Metric(managedCoreComponents.StatusHandler(), core.MetricMinGasPrice, managedCoreComponents.EconomicsData().MinGasPrice())
	metrics.SaveUint64Metric(managedCoreComponents.StatusHandler(), core.MetricMinGasLimit, managedCoreComponents.EconomicsData().MinGasLimit())

	return nil
}

func createHealthService(configs *config.Configs, flagsConfig *config.ContextFlagsConfig, managedDataComponents mainFactory.DataComponentsHandler) closing.Closer {
	healthService := health.NewHealthService(configs.GeneralConfig.Health, flagsConfig.WorkingDir)
	if flagsConfig.UseHealthService {
		healthService.Start()
	}

	healthService.RegisterComponent(managedDataComponents.Datapool().Transactions())
	healthService.RegisterComponent(managedDataComponents.Datapool().UnsignedTransactions())
	healthService.RegisterComponent(managedDataComponents.Datapool().RewardTransactions())
	return healthService
}

func CreateManagedConsensusComponents(
	configs *config.Configs,
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
		configs.GeneralConfig,
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
		configs.FlagsConfig.WorkingDir,
	)
	if err != nil {
		return nil, err
	}

	consensusArgs := mainFactory.ConsensusComponentsFactoryArgs{
		Config:              *configs.GeneralConfig,
		BootstrapRoundIndex: configs.FlagsConfig.BootstrapRoundIndex,
		HardforkTrigger:     hardForkTrigger,
		CoreComponents:      managedCoreComponents,
		NetworkComponents:   managedNetworkComponents,
		CryptoComponents:    managedCryptoComponents,
		DataComponents:      managedDataComponents,
		ProcessComponents:   managedProcessComponents,
		StateComponents:     managedStateComponents,
		StatusComponents:    managedStatusComponents,
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

func CreateManagedHeartbeatComponents(configs *config.Configs,
	managedCoreComponents mainFactory.CoreComponentsHandler,
	managedNetworkComponents mainFactory.NetworkComponentsHandler,
	managedCryptoComponents mainFactory.CryptoComponentsHandler,
	managedDataComponents mainFactory.DataComponentsHandler,
	managedProcessComponents mainFactory.ProcessComponentsHandler,
	hardforkTrigger HardforkTrigger) (mainFactory.HeartbeatComponentsHandler, error) {
	genesisTime := time.Unix(managedCoreComponents.GenesisNodesSetup().GetStartTime(), 0)
	heartbeatArgs := mainFactory.HeartbeatComponentsFactoryArgs{
		Config:            *configs.GeneralConfig,
		Prefs:             *configs.PreferencesConfig,
		AppVersion:        configs.FlagsConfig.Version,
		GenesisTime:       genesisTime,
		HardforkTrigger:   hardforkTrigger,
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

func waitForSignal(sigs chan os.Signal, log logger.Logger, chanStopNodeProcess chan endProcess.ArgEndProcess, healthService closing.Closer, ef closing.Closer, currentNode *Node, goRoutinesNumberStart int) error {
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
		closeAllComponents(log, healthService, ef, currentNode, chanCloseComponents)
	}()

	select {
	case <-chanCloseComponents:
		log.Debug("Closed all components gracefully")
	case <-time.After(maxTimeToClose):
		log.Warn("force closing the node", "error", "closeAllComponents did not finished on time")
		return fmt.Errorf("Did NOT close all components gracefully")
	}

	if reshuffled {
		log.Info("=============================Shuffled out - soft restart==================================")
		logGoroutinesNumber(log, goRoutinesNumberStart)
	} else {
		return fmt.Errorf("not reshuffled, closing")
	}

	return nil
}

func logInformation(log logger.Logger, configs *config.Configs, managedCoreComponents mainFactory.CoreComponentsHandler, managedCryptoComponents mainFactory.CryptoComponentsHandler, managedBootstrapComponents mainFactory.BootstrapComponentsHandler) {
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
		"AppVersion", configs.FlagsConfig.Version,
		"GenesisTimeStamp", managedCoreComponents.GenesisTime().Unix(),
	)

	sessionInfoFileOutput += fmt.Sprintf("\nStarted with parameters:\n")
	sessionInfoFileOutput += configs.FlagsConfig.SessionInfoFileOutput

	logSessionInformation(configs.FlagsConfig.WorkingDir, log, configs, sessionInfoFileOutput, managedCoreComponents)
}

func getNodesFileName(configs *config.Configs) (string, error) {
	flagsConfig := configs.FlagsConfig
	nodesFileName := flagsConfig.NodesFileName

	exportFolder := filepath.Join(flagsConfig.WorkingDir, configs.GeneralConfig.Hardfork.ImportFolder)
	if configs.GeneralConfig.Hardfork.AfterHardFork {
		exportFolderNodesSetupPath := filepath.Join(exportFolder, core.NodesSetupJsonFileName)
		if !core.DoesFileExist(exportFolderNodesSetupPath) {
			return "", fmt.Errorf("cannot find %s in the export folder", core.NodesSetupJsonFileName)
		}

		nodesFileName = exportFolderNodesSetupPath
	}
	return nodesFileName, nil
}

func logGoroutinesNumber(log logger.Logger, goRoutinesNumberStart int) {
	buffer := new(bytes.Buffer)
	err := pprof.Lookup("goroutine").WriteTo(buffer, 2)
	if err != nil {
		log.Error("could not dump goroutines")
	}
	log.Debug("go routines number",
		"start", goRoutinesNumberStart,
		"end", runtime.NumGoroutine())

	log.Warn(buffer.String())
}

func CreateManagedStatusComponents(
	configs *config.Configs,
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
		Config:               *configs.GeneralConfig,
		ExternalConfig:       *configs.ExternalConfig,
		EconomicsConfig:      *configs.EconomicsConfig,
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

func logSessionInformation(workingDir string, log logger.Logger, configs *config.Configs, sessionInfoFileOutput string, managedCoreComponents mainFactory.CoreComponentsHandler) {
	statsFolder := filepath.Join(workingDir, core.DefaultStatsPath)
	copyConfigToStatsFolder(
		log,
		statsFolder,
		[]string{
			configs.ConfigurationFileName,
			configs.ConfigurationEconomicsFileName,
			configs.ConfigurationRatingsFileName,
			configs.ConfigurationPreferencesFileName,
			configs.P2pConfigurationFileName,
			configs.ConfigurationFileName,
			configs.FlagsConfig.GenesisFileName,
			configs.FlagsConfig.NodesFileName,
		})

	statsFile := filepath.Join(statsFolder, "session.info")
	err := ioutil.WriteFile(statsFile, []byte(sessionInfoFileOutput), os.ModePerm)
	log.LogIfError(err)

	//TODO: remove this in the future and add just a log debug
	computedRatingsData := filepath.Join(statsFolder, "ratings.info")
	computedRatingsDataStr := createStringFromRatingsData(managedCoreComponents.RatingsData())
	err = ioutil.WriteFile(computedRatingsData, []byte(computedRatingsDataStr), os.ModePerm)
	log.LogIfError(err)
}

func CreateManagedProcessComponents(configs *config.Configs, managedCoreComponents mainFactory.CoreComponentsHandler, managedCryptoComponents mainFactory.CryptoComponentsHandler, managedNetworkComponents mainFactory.NetworkComponentsHandler, managedBootstrapComponents mainFactory.BootstrapComponentsHandler, managedStateComponents mainFactory.StateComponentsHandler, managedDataComponents mainFactory.DataComponentsHandler, managedStatusComponents mainFactory.StatusComponentsHandler, gasSchedule map[string]map[string]uint64, nodesCoordinator sharding.NodesCoordinator) (mainFactory.ProcessComponentsHandler, error) {
	importStartHandler, err := trigger.NewImportStartHandler(filepath.Join(configs.FlagsConfig.WorkingDir, core.DefaultDBPath), configs.FlagsConfig.Version)
	if err != nil {
		return nil, err
	}

	//TODO when refactoring main, maybe initialize economics data before this line
	totalSupply, ok := big.NewInt(0).SetString(configs.EconomicsConfig.GlobalSettings.GenesisTotalSupply, 10)
	if !ok {
		return nil, fmt.Errorf("can not parse total suply from economics.toml, %s is not a valid value",
			configs.EconomicsConfig.GlobalSettings.GenesisTotalSupply)
	}

	accountsParser, err := parsing.NewAccountsParser(
		configs.FlagsConfig.GenesisFileName,
		totalSupply,
		managedCoreComponents.AddressPubKeyConverter(),
		managedCryptoComponents.TxSignKeyGen(),
	)
	if err != nil {
		return nil, err
	}

	smartContractParser, err := parsing.NewSmartContractsParser(
		configs.FlagsConfig.SmartContractsFileName,
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

	epochNotifier := forking.NewGenericEpochNotifier()

	log.Trace("creating time cache for requested items components")
	requestedItemsHandler := timecache.NewTimeCache(
		time.Duration(uint64(time.Millisecond) * managedCoreComponents.GenesisNodesSetup().GetRoundDuration()))

	processArgs := mainFactory.ProcessComponentsFactoryArgs{
		Config:                    *configs.GeneralConfig,
		AccountsParser:            accountsParser,
		SmartContractParser:       smartContractParser,
		GasSchedule:               gasSchedule,
		Rounder:                   managedCoreComponents.Rounder(),
		ShardCoordinator:          managedBootstrapComponents.ShardCoordinator(),
		NodesCoordinator:          nodesCoordinator,
		Data:                      managedDataComponents,
		CoreData:                  managedCoreComponents,
		Crypto:                    managedCryptoComponents,
		State:                     managedStateComponents,
		Network:                   managedNetworkComponents,
		RequestedItemsHandler:     requestedItemsHandler,
		WhiteListHandler:          whiteListRequest,
		WhiteListerVerifiedTxs:    whiteListerVerifiedTxs,
		EpochStartNotifier:        managedCoreComponents.EpochStartNotifierWithConfirm(),
		EpochStart:                &configs.GeneralConfig.EpochStartConfig,
		Rater:                     managedCoreComponents.Rater(),
		RatingsData:               managedCoreComponents.RatingsData(),
		StartEpochNum:             managedBootstrapComponents.EpochBootstrapParams().Epoch(),
		SizeCheckDelta:            configs.GeneralConfig.Marshalizer.SizeCheckDelta,
		StateCheckpointModulus:    configs.GeneralConfig.StateTriesConfig.CheckpointRoundsModulus,
		MaxComputableRounds:       configs.GeneralConfig.GeneralSettings.MaxComputableRounds,
		NumConcurrentResolverJobs: configs.GeneralConfig.Antiflood.NumConcurrentResolverJobs,
		MinSizeInBytes:            configs.GeneralConfig.BlockSizeThrottleConfig.MinSizeInBytes,
		MaxSizeInBytes:            configs.GeneralConfig.BlockSizeThrottleConfig.MaxSizeInBytes,
		MaxRating:                 configs.RatingsConfig.General.MaxRating,
		ValidatorPubkeyConverter:  managedCoreComponents.ValidatorPubKeyConverter(),
		SystemSCConfig:            configs.SystemSCConfig,
		Version:                   configs.FlagsConfig.Version,
		ImportStartHandler:        importStartHandler,
		WorkingDir:                configs.FlagsConfig.WorkingDir,
		Indexer:                   managedStatusComponents.ElasticIndexer(),
		TpsBenchmark:              managedStatusComponents.TpsBenchmark(),
		HistoryRepo:               historyRepository,
		EpochNotifier:             epochNotifier,
		HeaderIntegrityVerifier:   managedBootstrapComponents.HeaderIntegrityVerifier(),
		ChanGracefullyClose:       managedCoreComponents.ChanStopNodeProcess(),
		EconomicsData:             managedCoreComponents.EconomicsData(),
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

func CreateManagedDataComponents(configs *config.Configs, managedCoreComponents mainFactory.CoreComponentsHandler, managedBootstrapComponents mainFactory.BootstrapComponentsHandler) (mainFactory.DataComponentsHandler, error) {
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
		CreateTrieEpochRootHashStorer: configs.FlagsConfig.ImportDbSaveTrieEpochRootHash,
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

func CreateManagedStateComponents(configs *config.Configs, managedCoreComponents mainFactory.CoreComponentsHandler, managedBootstrapComponents mainFactory.BootstrapComponentsHandler) (mainFactory.StateComponentsHandler, error) {
	triesComponents, trieStorageManagers := managedBootstrapComponents.EpochStartBootstrapper().GetTriesComponents()
	stateArgs := mainFactory.StateComponentsFactoryArgs{
		Config:              *configs.GeneralConfig,
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

func CreateManagedBootstrapComponents(
	cfgs *config.Configs,
	managedCoreComponents mainFactory.CoreComponentsHandler,
	managedCryptoComponents mainFactory.CryptoComponentsHandler,
	managedNetworkComponents mainFactory.NetworkComponentsHandler,
) (mainFactory.BootstrapComponentsHandler, error) {

	bootstrapComponentsFactoryArgs := mainFactory.BootstrapComponentsFactoryArgs{
		Config:            *cfgs.GeneralConfig,
		PrefConfig:        *cfgs.PreferencesConfig,
		WorkingDir:        cfgs.FlagsConfig.WorkingDir,
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

func CreateManagedNetworkComponents(
	cfgs *config.Configs,
	managedCoreComponents mainFactory.CoreComponentsHandler,
) (mainFactory.NetworkComponentsHandler, error) {

	networkComponentsFactoryArgs := mainFactory.NetworkComponentsFactoryArgs{
		P2pConfig:            *cfgs.P2pConfig,
		MainConfig:           *cfgs.GeneralConfig,
		RatingsConfig:        *cfgs.RatingsConfig,
		StatusHandler:        managedCoreComponents.StatusHandler(),
		Marshalizer:          managedCoreComponents.InternalMarshalizer(),
		Syncer:               managedCoreComponents.SyncTimer(),
		BootstrapWaitSeconds: core.SecondsToWaitForP2PBootstrap,
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

func CreateManagedCoreComponents(
	cfgs *config.Configs,
	chanStopNodeProcess chan endProcess.ArgEndProcess,
) (mainFactory.CoreComponentsHandler, error) {

	statusHandlersFactoryArgs := &factory.StatusHandlersFactoryArgs{
		UseTermUI: !cfgs.FlagsConfig.UseLogView,
	}

	statusHandlersFactory, err := factory.NewStatusHandlersFactory(statusHandlersFactoryArgs)
	if err != nil {
		return nil, err
	}

	coreArgs := mainFactory.CoreComponentsFactoryArgs{
		Config:                *cfgs.GeneralConfig,
		RatingsConfig:         *cfgs.RatingsConfig,
		EconomicsConfig:       *cfgs.EconomicsConfig,
		NodesFilename:         cfgs.FlagsConfig.NodesFileName,
		WorkingDirectory:      cfgs.FlagsConfig.WorkingDir,
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

func CreateManagedCryptoComponents(
	cfgs *config.Configs,
	managedCoreComponents mainFactory.CoreComponentsHandler,
) (mainFactory.CryptoComponentsHandler, error) {
	validatorKeyPemFileName := cfgs.FlagsConfig.ValidatorKeyPemFileName
	cryptoComponentsHandlerArgs := mainFactory.CryptoComponentsFactoryArgs{
		ValidatorKeyPemFileName:              validatorKeyPemFileName,
		SkIndex:                              cfgs.FlagsConfig.ValidatorKeyIndex,
		Config:                               *cfgs.GeneralConfig,
		CoreComponentsHolder:                 managedCoreComponents,
		ActivateBLSPubKeyMessageVerification: cfgs.SystemSCConfig.StakingSystemSCConfig.ActivateBLSPubKeyMessageVerification,
		KeyLoader:                            &core.KeyLoader{},
		UseDisabledSigVerifier:               cfgs.FlagsConfig.ImportDbNoSigCheckFlag,
		IsInImportMode:                       cfgs.FlagsConfig.IsInImportMode,
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
	log logger.Logger,
	healthService io.Closer,
	facade mainFactory.Closer,
	node *Node,
	chanCloseComponents chan struct{},
) {
	log.Debug("closing health service...")
	err := healthService.Close()
	log.LogIfError(err)

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

func cleanupStorageIfNecessary(workingDir string, cleanupStorage bool, log logger.Logger) error {
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

func copyConfigToStatsFolder(log logger.Logger, statsFolder string, configs []string) {
	err := os.MkdirAll(statsFolder, os.ModePerm)
	log.LogIfError(err)

	for _, configFile := range configs {
		copySingleFile(statsFolder, configFile)
	}
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
			fmt.Println(fmt.Sprintf("Could not close %s", source.Name()))
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
			fmt.Println(fmt.Sprintf("Could not close %s", source.Name()))
		}
	}()

	_, err = io.Copy(destination, source)
	if err != nil {
		fmt.Println(fmt.Sprintf("Could not copy %s", source.Name()))
	}
}

func indexValidatorsListIfNeeded(
	elasticIndexer indexer.Indexer,
	coordinator sharding.NodesCoordinator,
	epoch uint32,
	log logger.Logger,

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

func enableGopsIfNeeded(gopsEnabled bool, log logger.Logger) {
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

func attachFileLogger(configs *config.Configs, log logger.Logger) (factory.FileLoggingHandler, error) {
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
			//we need to print this manually as we do not have console log observer
			fmt.Println("error removing log observer: " + err.Error())
			return nil, err
		}

		err = logger.AddLogObserver(os.Stdout, &logger.PlainFormatter{})
		if err != nil {
			//we need to print this manually as we do not have console log observer
			fmt.Println("error setting log observer: " + err.Error())
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
