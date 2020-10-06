package main

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
	dbLookupFactory "github.com/ElrondNetwork/elrond-go/core/dblookupext/factory"
	"github.com/ElrondNetwork/elrond-go/core/forking"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/logging"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/facade"
	mainFactory "github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/genesis/parsing"
	"github.com/ElrondNetwork/elrond-go/health"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	"github.com/ElrondNetwork/elrond-go/sharding"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/ElrondNetwork/elrond-go/update/trigger"
	"github.com/denisbrodbeck/machineid"
	"github.com/google/gops/agent"
	"github.com/urfave/cli"
)

const (
	defaultLogsPath = "logs"
	maxTimeToClose  = 10 * time.Second
	maxMachineIDLen = 10
)

var (
	nodeHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
VERSION:
   {{.Version}}
   {{end}}
`
)

// appVersion should be populated at build time using ldflags
// Usage examples:
// linux/mac:
//            go build -i -v -ldflags="-X main.appVersion=$(git describe --tags --long --dirty)"
// windows:
//            for /f %i in ('git describe --tags --long --dirty') do set VERS=%i
//            go build -i -v -ldflags="-X main.appVersion=%VERS%"
var appVersion = core.UnVersionedAppString

type configs struct {
	generalConfig                    *config.Config
	apiRoutesConfig                  *config.ApiRoutesConfig
	economicsConfig                  *config.EconomicsConfig
	systemSCConfig                   *config.SystemSmartContractsConfig
	ratingsConfig                    *config.RatingsConfig
	preferencesConfig                *config.Preferences
	externalConfig                   *config.ExternalConfig
	p2pConfig                        *config.P2PConfig
	configurationFileName            string
	configurationEconomicsFileName   string
	configurationRatingsFileName     string
	configurationPreferencesFileName string
	p2pConfigurationFileName         string
}

func main() {
	_ = logger.SetDisplayByteSlice(logger.ToHexShort)
	log := logger.GetOrCreate("main")

	app := cli.NewApp()
	cli.AppHelpTemplate = nodeHelpTemplate
	app.Name = "Elrond Node CLI App"
	machineID, err := machineid.ProtectedID(app.Name)
	if err != nil {
		log.Warn("error fetching machine id", "error", err)
		machineID = "unknown"
	}
	if len(machineID) > maxMachineIDLen {
		machineID = machineID[:maxMachineIDLen]
	}

	app.Version = fmt.Sprintf("%s/%s/%s-%s/%s", appVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH, machineID)
	app.Usage = "This is the entry point for starting a new Elrond node - the app will start after the genesis timestamp"
	app.Flags = getFlags()
	app.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		return startNode(c, log, app.Version)
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startNode(ctx *cli.Context, log logger.Logger, version string) error {
	log.Trace("startNode called")
	chanStopNodeProcess := make(chan endProcess.ArgEndProcess, 1)
	workingDir := getWorkingDir(ctx, log)
	fileLogging, err := updateLogger(workingDir, ctx, log)
	if err != nil {
		return err
	}

	enableGopsIfNeeded(ctx, log)

	log.Info("starting node", "version", version, "pid", os.Getpid())

	var cfgs *configs
	cfgs, err = readConfigs(log, ctx)
	if err != nil {
		return err
	}

	if !check.IfNil(fileLogging) {
		err = fileLogging.ChangeFileLifeSpan(time.Second * time.Duration(cfgs.generalConfig.Logs.LogFileLifeSpanInSec))
		if err != nil {
			return err
		}
	}

	nodesFileName := ctx.GlobalString(nodesFile.Name)

	//TODO when refactoring main, maybe initialize economics data before this line
	totalSupply, ok := big.NewInt(0).SetString(cfgs.economicsConfig.GlobalSettings.GenesisTotalSupply, 10)
	if !ok {
		return fmt.Errorf("can not parse total suply from economics.toml, %s is not a valid value",
			cfgs.economicsConfig.GlobalSettings.GenesisTotalSupply)
	}

	log.Debug("config", "file", ctx.GlobalString(genesisFile.Name))

	exportFolder := filepath.Join(workingDir, cfgs.generalConfig.Hardfork.ImportFolder)
	if cfgs.generalConfig.Hardfork.AfterHardFork {
		exportFolderNodesSetupPath := filepath.Join(exportFolder, core.NodesSetupJsonFileName)
		if !core.DoesFileExist(exportFolderNodesSetupPath) {
			return fmt.Errorf("cannot find %s in the export folder", core.NodesSetupJsonFileName)
		}

		nodesFileName = exportFolderNodesSetupPath
	}

	if err != nil {
		return err
	}
	log.Debug("config", "file", nodesFileName)

	if ctx.IsSet(startInEpoch.Name) {
		log.Debug("start in epoch is enabled")
		cfgs.generalConfig.GeneralSettings.StartInEpochEnabled = ctx.GlobalBool(startInEpoch.Name)
	}

	err = cleanupStorageIfNecessary(workingDir, ctx, log)
	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	err = pprof.Lookup("goroutine").WriteTo(buffer, 2)
	if err != nil {
		log.Error("could not dump goroutines")
	}
	log.Debug("go routines number",
		"start", runtime.NumGoroutine())

	log.Warn(buffer.String())

	for {
		goRoutinesNumberStart := runtime.NumGoroutine()

		log.Debug("\n\n====================Starting managedComponents creation================================")

		healthService := health.NewHealthService(cfgs.generalConfig.Health, workingDir)
		if ctx.IsSet(useHealthService.Name) {
			healthService.Start()
		}

		log.Trace("creating core components")
		managedCoreComponents, err := createManagedCoreComponents(
			ctx,
			cfgs,
			nodesFileName,
			workingDir,
			chanStopNodeProcess,
		)
		if err != nil {
			return err
		}

		log.Trace("creating crypto components")
		managedCryptoComponents, err := createManagedCryptoComponents(ctx, cfgs, managedCoreComponents)
		if err != nil {
			return err
		}

		log.Trace("creating network components")
		managedNetworkComponents, err := createManagedNetworkComponents(cfgs, managedCoreComponents)
		if err != nil {
			return err
		}

		managedBootstrapComponents, err := createManagedBootstrapComponents(cfgs, managedCoreComponents, managedCryptoComponents, managedNetworkComponents, workingDir)
		if err != nil {
			return err
		}

		shardCoordinator, err := sharding.NewMultiShardCoordinator(
			managedBootstrapComponents.EpochBootstrapParams().NumOfShards(),
			managedBootstrapComponents.EpochBootstrapParams().SelfShardID())
		if err != nil {
			return err
		}

		currentEpoch := managedBootstrapComponents.EpochBootstrapParams().Epoch()
		storerEpoch := currentEpoch
		if !cfgs.generalConfig.StoragePruning.Enabled {
			// TODO: refactor this as when the pruning storer is disabled, the default directory path is Epoch_0
			// and it should be Epoch_ALL or something similar
			storerEpoch = 0
		}

		log.Info("Bootstrap", "epoch", managedBootstrapComponents.EpochBootstrapParams().Epoch())
		if managedBootstrapComponents.EpochBootstrapParams().NodesConfig() != nil {
			log.Info("the epoch from nodesConfig is",
				"epoch", managedBootstrapComponents.EpochBootstrapParams().NodesConfig().CurrentEpoch)
		}

		var shardIdString = core.GetShardIDString(shardCoordinator.SelfId())
		logger.SetCorrelationShard(shardIdString)

		log.Trace("creating state components")
		triesComponents, trieStorageManagers := managedBootstrapComponents.EpochStartBootstrapper().GetTriesComponents()
		stateArgs := mainFactory.StateComponentsFactoryArgs{
			Config:              *cfgs.generalConfig,
			ShardCoordinator:    shardCoordinator,
			Core:                managedCoreComponents,
			TriesContainer:      triesComponents,
			TrieStorageManagers: trieStorageManagers,
		}

		stateComponentsFactory, err := mainFactory.NewStateComponentsFactory(stateArgs)
		if err != nil {
			return fmt.Errorf("NewStateComponentsFactory failed: %w", err)
		}

		managedStateComponents, err := mainFactory.NewManagedStateComponents(stateComponentsFactory)
		if err != nil {
			return err
		}

		err = managedStateComponents.Create()
		if err != nil {
			return err
		}

		log.Trace("creating data components")
		epochStartNotifier := notifier.NewEpochStartSubscriptionHandler()

		dataArgs := mainFactory.DataComponentsFactoryArgs{
			Config:             *cfgs.generalConfig,
			ShardCoordinator:   shardCoordinator,
			Core:               managedCoreComponents,
			EpochStartNotifier: epochStartNotifier,
			CurrentEpoch:       storerEpoch,
		}

		dataComponentsFactory, err := mainFactory.NewDataComponentsFactory(dataArgs)
		if err != nil {
			return fmt.Errorf("NewDataComponentsFactory failed: %w", err)
		}
		managedDataComponents, err := mainFactory.NewManagedDataComponents(dataComponentsFactory)
		if err != nil {
			return err
		}
		err = managedDataComponents.Create()
		if err != nil {
			return err
		}

		healthService.RegisterComponent(managedDataComponents.Datapool().Transactions())
		healthService.RegisterComponent(managedDataComponents.Datapool().UnsignedTransactions())
		healthService.RegisterComponent(managedDataComponents.Datapool().RewardTransactions())

		log.Trace("initializing metrics")
		err = metrics.InitMetrics(
			managedCoreComponents.StatusHandlerUtils(),
			managedCryptoComponents.PublicKeyString(),
			managedBootstrapComponents.NodeType(),
			shardCoordinator,
			managedCoreComponents.GenesisNodesSetup(),
			version,
			cfgs.economicsConfig,
			cfgs.generalConfig.EpochStartConfig.RoundsPerEpoch,
			managedCoreComponents.MinTransactionVersion(),
		)
		if err != nil {
			return err
		}

		err = managedCoreComponents.StatusHandlerUtils().UpdateStorerAndMetricsForPersistentHandler(
			managedDataComponents.StorageService().GetStorer(dataRetriever.StatusMetricsUnit),
		)
		if err != nil {
			return err
		}

		log.Trace("creating nodes coordinator")
		if ctx.IsSet(keepOldEpochsData.Name) {
			cfgs.generalConfig.StoragePruning.CleanOldEpochsData = !ctx.GlobalBool(keepOldEpochsData.Name)
		}
		if ctx.IsSet(numEpochsToSave.Name) {
			cfgs.generalConfig.StoragePruning.NumEpochsToKeep = ctx.GlobalUint64(numEpochsToSave.Name)
		}
		if ctx.IsSet(numActivePersisters.Name) {
			cfgs.generalConfig.StoragePruning.NumActivePersisters = ctx.GlobalUint64(numActivePersisters.Name)
		}

		nodesCoordinator, nodeShufflerOut, err := mainFactory.CreateNodesCoordinator(
			log,
			managedCoreComponents.GenesisNodesSetup(),
			cfgs.preferencesConfig.Preferences,
			epochStartNotifier,
			managedCryptoComponents.PublicKey(),
			managedCoreComponents.InternalMarshalizer(),
			managedCoreComponents.Hasher(),
			managedCoreComponents.Rater(),
			managedDataComponents.StorageService().GetStorer(dataRetriever.BootstrapUnit),
			managedCoreComponents.NodesShuffler(),
			cfgs.generalConfig.EpochStartConfig,
			shardCoordinator.SelfId(),
			chanStopNodeProcess,
			managedBootstrapComponents.EpochBootstrapParams(),
			currentEpoch,
		)
		if err != nil {
			return err
		}

		metrics.SaveStringMetric(managedCoreComponents.StatusHandler(), core.MetricNodeDisplayName, cfgs.preferencesConfig.Preferences.NodeDisplayName)
		metrics.SaveStringMetric(managedCoreComponents.StatusHandler(), core.MetricChainId, managedCoreComponents.ChainID())
		metrics.SaveUint64Metric(managedCoreComponents.StatusHandler(), core.MetricGasPerDataByte, managedCoreComponents.EconomicsData().GasPerDataByte())
		metrics.SaveUint64Metric(managedCoreComponents.StatusHandler(), core.MetricMinGasPrice, managedCoreComponents.EconomicsData().MinGasPrice())
		metrics.SaveUint64Metric(managedCoreComponents.StatusHandler(), core.MetricMinGasLimit, managedCoreComponents.EconomicsData().MinGasLimit())

		sessionInfoFileOutput := fmt.Sprintf("%s:%s\n%s:%s\n%s:%v\n%s:%s\n%s:%v\n",
			"PkBlockSign", managedCryptoComponents.PublicKeyString(),
			"ShardId", shardIdString,
			"TotalShards", shardCoordinator.NumberOfShards(),
			"AppVersion", version,
			"GenesisTimeStamp", managedCoreComponents.GenesisTime().Unix(),
		)

		sessionInfoFileOutput += fmt.Sprintf("\nStarted with parameters:\n")
		for _, flag := range ctx.App.Flags {
			flagValue := fmt.Sprintf("%v", ctx.GlobalGeneric(flag.GetName()))
			if flagValue != "" {
				sessionInfoFileOutput += fmt.Sprintf("%s = %v\n", flag.GetName(), flagValue)
			}
		}

		statsFolder := filepath.Join(workingDir, core.DefaultStatsPath)
		copyConfigToStatsFolder(
			log,
			statsFolder,
			[]string{
				cfgs.configurationFileName,
				cfgs.configurationEconomicsFileName,
				cfgs.configurationRatingsFileName,
				cfgs.configurationPreferencesFileName,
				cfgs.p2pConfigurationFileName,
				cfgs.configurationFileName,
				ctx.GlobalString(genesisFile.Name),
				ctx.GlobalString(nodesFile.Name),
			})

		statsFile := filepath.Join(statsFolder, "session.info")
		err = ioutil.WriteFile(statsFile, []byte(sessionInfoFileOutput), os.ModePerm)
		log.LogIfError(err)

		//TODO: remove this in the future and add just a log debug
		computedRatingsData := filepath.Join(statsFolder, "ratings.info")
		computedRatingsDataStr := createStringFromRatingsData(managedCoreComponents.RatingsData())
		err = ioutil.WriteFile(computedRatingsData, []byte(computedRatingsDataStr), os.ModePerm)
		log.LogIfError(err)

		gasScheduleConfigurationFileName := ctx.GlobalString(gasScheduleConfigurationFile.Name)
		gasSchedule, err := core.LoadGasScheduleConfig(gasScheduleConfigurationFileName)
		if err != nil {
			return err
		}

		log.Trace("creating time cache for requested items components")
		requestedItemsHandler := timecache.NewTimeCache(
			time.Duration(uint64(time.Millisecond) * managedCoreComponents.GenesisNodesSetup().GetRoundDuration()))

		whiteListCache, err := storageUnit.NewCache(storageFactory.GetCacherFromConfig(cfgs.generalConfig.WhiteListPool))
		if err != nil {
			return err
		}
		whiteListRequest, err := interceptors.NewWhiteListDataVerifier(whiteListCache)
		if err != nil {
			return err
		}

		whiteListerVerifiedTxs, err := createWhiteListerVerifiedTxs(cfgs.generalConfig)
		if err != nil {
			return err
		}

		historyRepoFactoryArgs := &dbLookupFactory.ArgsHistoryRepositoryFactory{
			SelfShardID: shardCoordinator.SelfId(),
			Config:      cfgs.generalConfig.DbLookupExtensions,
			Hasher:      managedCoreComponents.Hasher(),
			Marshalizer: managedCoreComponents.InternalMarshalizer(),
			Store:       managedDataComponents.StorageService(),
		}
		historyRepositoryFactory, err := dbLookupFactory.NewHistoryRepositoryFactory(historyRepoFactoryArgs)
		if err != nil {
			return err
		}

		historyRepository, err := historyRepositoryFactory.Create()
		if err != nil {
			return err
		}

		log.Trace("starting status pooling components")
		statArgs := mainFactory.StatusComponentsFactoryArgs{
			Config:             *cfgs.generalConfig,
			ExternalConfig:     *cfgs.externalConfig,
			ElasticOptions:     &indexer.Options{TxIndexingEnabled: ctx.GlobalBoolT(enableTxIndexing.Name)},
			ShardCoordinator:   shardCoordinator,
			NodesCoordinator:   nodesCoordinator,
			EpochStartNotifier: epochStartNotifier,
			CoreComponents:     managedCoreComponents,
			DataComponents:     managedDataComponents,
			NetworkComponents:  managedNetworkComponents,
		}

		statusComponentsFactory, err := mainFactory.NewStatusComponentsFactory(statArgs)
		if err != nil {
			return fmt.Errorf("NewStatusComponentsFactory failed: %w", err)
		}

		managedStatusComponents, err := mainFactory.NewManagedStatusComponents(statusComponentsFactory)
		if err != nil {
			return err
		}
		err = managedStatusComponents.Create()
		if err != nil {
			return err
		}
		epochNotifier := forking.NewGenericEpochNotifier()

		log.Trace("creating process components")

		importStartHandler, err := trigger.NewImportStartHandler(filepath.Join(workingDir, core.DefaultDBPath), appVersion)
		if err != nil {
			return err
		}

		accountsParser, err := parsing.NewAccountsParser(
			ctx.GlobalString(genesisFile.Name),
			totalSupply,
			managedCoreComponents.AddressPubKeyConverter(),
			managedCryptoComponents.TxSignKeyGen(),
		)
		if err != nil {
			return err
		}

		smartContractParser, err := parsing.NewSmartContractsParser(
			ctx.GlobalString(smartContractsFile.Name),
			managedCoreComponents.AddressPubKeyConverter(),
			managedCryptoComponents.TxSignKeyGen(),
		)
		if err != nil {
			return err
		}

		processArgs := mainFactory.ProcessComponentsFactoryArgs{
			Config:                    *cfgs.generalConfig,
			AccountsParser:            accountsParser,
			SmartContractParser:       smartContractParser,
			GasSchedule:               gasSchedule,
			Rounder:                   managedCoreComponents.Rounder(),
			ShardCoordinator:          shardCoordinator,
			NodesCoordinator:          nodesCoordinator,
			Data:                      managedDataComponents,
			CoreData:                  managedCoreComponents,
			Crypto:                    managedCryptoComponents,
			State:                     managedStateComponents,
			Network:                   managedNetworkComponents,
			RequestedItemsHandler:     requestedItemsHandler,
			WhiteListHandler:          whiteListRequest,
			WhiteListerVerifiedTxs:    whiteListerVerifiedTxs,
			EpochStartNotifier:        epochStartNotifier,
			EpochStart:                &cfgs.generalConfig.EpochStartConfig,
			Rater:                     managedCoreComponents.Rater(),
			RatingsData:               managedCoreComponents.RatingsData(),
			StartEpochNum:             currentEpoch,
			SizeCheckDelta:            cfgs.generalConfig.Marshalizer.SizeCheckDelta,
			StateCheckpointModulus:    cfgs.generalConfig.StateTriesConfig.CheckpointRoundsModulus,
			MaxComputableRounds:       cfgs.generalConfig.GeneralSettings.MaxComputableRounds,
			NumConcurrentResolverJobs: cfgs.generalConfig.Antiflood.NumConcurrentResolverJobs,
			MinSizeInBytes:            cfgs.generalConfig.BlockSizeThrottleConfig.MinSizeInBytes,
			MaxSizeInBytes:            cfgs.generalConfig.BlockSizeThrottleConfig.MaxSizeInBytes,
			MaxRating:                 cfgs.ratingsConfig.General.MaxRating,
			ValidatorPubkeyConverter:  managedCoreComponents.ValidatorPubKeyConverter(),
			SystemSCConfig:            cfgs.systemSCConfig,
			Version:                   version,
			ImportStartHandler:        importStartHandler,
			WorkingDir:                workingDir,
			Indexer:                   managedStatusComponents.ElasticIndexer(),
			TpsBenchmark:              managedStatusComponents.TpsBenchmark(),
			HistoryRepo:               historyRepository,
			EpochNotifier:             epochNotifier,
			HeaderIntegrityVerifier:   managedBootstrapComponents.HeaderIntegrityVerifier(),
			ChanGracefullyClose:       chanStopNodeProcess,
			EconomicsData:             managedCoreComponents.EconomicsData(),
		}
		processComponentsFactory, err := mainFactory.NewProcessComponentsFactory(processArgs)
		if err != nil {
			return fmt.Errorf("NewProcessComponentsFactory failed: %w", err)
		}

		managedProcessComponents, err := mainFactory.NewManagedProcessComponents(processComponentsFactory)
		if err != nil {
			return err
		}
		err = managedProcessComponents.Create()

		if err != nil {
			return err
		}

		historyRepository.RegisterToBlockTracker(managedProcessComponents.BlockTracker())
		managedStatusComponents.SetForkDetector(managedProcessComponents.ForkDetector())
		err = managedStatusComponents.StartPolling()

		hardForkTrigger, err := node.CreateHardForkTrigger(
			cfgs.generalConfig,
			shardCoordinator,
			nodesCoordinator,
			managedCoreComponents,
			managedStateComponents,
			managedDataComponents,
			managedCryptoComponents,
			managedProcessComponents,
			managedNetworkComponents,
			whiteListRequest,
			whiteListerVerifiedTxs,
			chanStopNodeProcess,
			epochStartNotifier,
			importStartHandler,
			workingDir,
		)
		if err != nil {
			return err
		}

		err = hardForkTrigger.AddCloser(nodeShufflerOut)
		if err != nil {
			return fmt.Errorf("%w when adding nodeShufflerOut in hardForkTrigger", err)
		}

		elasticIndexer := managedStatusComponents.ElasticIndexer()
		if !elasticIndexer.IsNilIndexer() {
			elasticIndexer.SetTxLogsProcessor(managedProcessComponents.TxLogsProcessor())
			managedProcessComponents.TxLogsProcessor().EnableLogToBeSavedInCache()
		}

		genesisTime := time.Unix(managedCoreComponents.GenesisNodesSetup().GetStartTime(), 0)
		heartbeatArgs := mainFactory.HeartbeatComponentsFactoryArgs{
			Config:            *cfgs.generalConfig,
			Prefs:             *cfgs.preferencesConfig,
			AppVersion:        version,
			GenesisTime:       genesisTime,
			HardforkTrigger:   hardForkTrigger,
			CoreComponents:    managedCoreComponents,
			DataComponents:    managedDataComponents,
			NetworkComponents: managedNetworkComponents,
			CryptoComponents:  managedCryptoComponents,
			ProcessComponents: managedProcessComponents,
		}

		heartbeatComponentsFactory, err := mainFactory.NewHeartbeatComponentsFactory(heartbeatArgs)
		if err != nil {
			return fmt.Errorf("NewHeartbeatComponentsFactory failed: %w", err)
		}

		managedHeartbeatComponents, err := mainFactory.NewManagedHeartbeatComponents(heartbeatComponentsFactory)
		if err != nil {
			return err
		}

		err = managedHeartbeatComponents.Create()
		if err != nil {
			return err
		}

		log.Debug("starting node...")

		consensusArgs := mainFactory.ConsensusComponentsFactoryArgs{
			Config:              *cfgs.generalConfig,
			ConsensusGroupSize:  int(managedCoreComponents.GenesisNodesSetup().GetShardConsensusGroupSize()),
			BootstrapRoundIndex: ctx.GlobalUint64(bootstrapRoundIndex.Name),
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
			return fmt.Errorf("NewConsensusComponentsFactory failed: %w", err)
		}

		managedConsensusComponents, err := mainFactory.NewManagedConsensusComponents(consensusFactory)
		if err != nil {
			return err
		}

		err = managedConsensusComponents.Create()
		if err != nil {
			log.Error("starting node failed", "epoch", currentEpoch, "error", err.Error())
			return err
		}

		log.Trace("creating node structure")
		currentNode, err := node.CreateNode(
			cfgs.generalConfig,
			cfgs.preferencesConfig,
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
			ctx.GlobalUint64(bootstrapRoundIndex.Name),
			requestedItemsHandler,
			whiteListRequest,
			whiteListerVerifiedTxs,
			chanStopNodeProcess,
			hardForkTrigger,
			historyRepository,
		)
		if err != nil {
			return err
		}

		if shardCoordinator.SelfId() == core.MetachainShardId {
			log.Trace("activating nodesCoordinator's validators indexing")
			indexValidatorsListIfNeeded(
				elasticIndexer,
				nodesCoordinator,
				managedProcessComponents.EpochStartTrigger().Epoch(),
				log,
			)
		}

		log.Trace("creating api resolver structure")
		apiResolver, err := mainFactory.CreateApiResolver(
			cfgs.generalConfig,
			managedStateComponents.AccountsAdapter(),
			managedStateComponents.PeerAccounts(),
			managedCoreComponents.AddressPubKeyConverter(),
			managedDataComponents.StorageService(),
			managedDataComponents.Blockchain(),
			managedCoreComponents.InternalMarshalizer(),
			managedCoreComponents.Hasher(),
			managedCoreComponents.Uint64ByteSliceConverter(),
			shardCoordinator,
			managedCoreComponents.StatusHandlerUtils().Metrics(),
			gasSchedule,
			managedCoreComponents.EconomicsData(),
			managedCryptoComponents.MessageSignVerifier(),
			managedCoreComponents.GenesisNodesSetup(),
			cfgs.systemSCConfig,
		)
		if err != nil {
			return err
		}

		log.Trace("creating elrond node facade")
		restAPIServerDebugMode := ctx.GlobalBool(restApiDebug.Name)

		argNodeFacade := facade.ArgNodeFacade{
			Node:                   currentNode,
			ApiResolver:            apiResolver,
			TxSimulatorProcessor:   managedProcessComponents.TransactionSimulatorProcessor(),
			RestAPIServerDebugMode: restAPIServerDebugMode,
			WsAntifloodConfig:      cfgs.generalConfig.Antiflood.WebServer,
			FacadeConfig: config.FacadeConfig{
				RestApiInterface: ctx.GlobalString(restApiInterface.Name),
				PprofEnabled:     ctx.GlobalBool(profileMode.Name),
			},
			ApiRoutesConfig: *cfgs.apiRoutesConfig,
			AccountsState:   managedStateComponents.AccountsAdapter(),
			PeerState:       managedStateComponents.PeerAccounts(),
		}

		ef, err := facade.NewNodeFacade(argNodeFacade)
		if err != nil {
			return fmt.Errorf("%w while creating NodeFacade", err)
		}

		ef.SetSyncer(managedCoreComponents.SyncTimer())
		ef.SetTpsBenchmark(managedStatusComponents.TpsBenchmark())

		log.Trace("starting background services")
		ef.StartBackgroundServices()

		log.Info("application is now running")
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
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
			break
		}

		if reshuffled {
			log.Info("=============================Shuffled out - soft restart==================================")

			buffer := new(bytes.Buffer)
			err := pprof.Lookup("goroutine").WriteTo(buffer, 2)
			if err != nil {
				log.Error("could not dump goroutines")
			}
			log.Debug("go routines number",
				"start", goRoutinesNumberStart,
				"end", runtime.NumGoroutine())

			log.Warn(buffer.String())
		} else {
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

func createManagedBootstrapComponents(
	cfgs *configs,
	managedCoreComponents mainFactory.CoreComponentsHandler,
	managedCryptoComponents mainFactory.CryptoComponentsHandler,
	managedNetworkComponents mainFactory.NetworkComponentsHandler,
	workingDir string,
) (mainFactory.BootstrapComponentsHandler, error) {

	bootstrapComponentsFactoryArgs := mainFactory.BootstrapComponentsFactoryArgs{
		Config:            *cfgs.generalConfig,
		PrefConfig:        *cfgs.preferencesConfig,
		WorkingDir:        workingDir,
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

func createManagedNetworkComponents(
	cfgs *configs,
	managedCoreComponents mainFactory.CoreComponentsHandler,
) (mainFactory.NetworkComponentsHandler, error) {

	networkComponentsFactoryArgs := mainFactory.NetworkComponentsFactoryArgs{
		P2pConfig:            *cfgs.p2pConfig,
		MainConfig:           *cfgs.generalConfig,
		RatingsConfig:        *cfgs.ratingsConfig,
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

func createManagedCoreComponents(
	ctx *cli.Context,
	cfgs *configs,
	nodesFileName string,
	workingDir string,
	chanStopNodeProcess chan endProcess.ArgEndProcess,
) (mainFactory.CoreComponentsHandler, error) {
	useTermui := !ctx.GlobalBool(useLogView.Name)
	statusHandlersFactoryArgs := &factory.StatusHandlersFactoryArgs{
		UseTermUI: useTermui,
	}

	statusHandlersFactory, err := factory.NewStatusHandlersFactory(statusHandlersFactoryArgs)
	if err != nil {
		return nil, err
	}

	coreArgs := mainFactory.CoreComponentsFactoryArgs{
		Config:                *cfgs.generalConfig,
		RatingsConfig:         *cfgs.ratingsConfig,
		EconomicsConfig:       *cfgs.economicsConfig,
		NodesFilename:         nodesFileName,
		WorkingDirectory:      workingDir,
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

func createManagedCryptoComponents(
	ctx *cli.Context,
	cfgs *configs,
	managedCoreComponents mainFactory.CoreComponentsHandler,
) (mainFactory.CryptoComponentsHandler, error) {
	validatorKeyPemFileName := ctx.GlobalString(validatorKeyPemFile.Name)
	cryptoComponentsHandlerArgs := mainFactory.CryptoComponentsFactoryArgs{
		ValidatorKeyPemFileName:              validatorKeyPemFileName,
		SkIndex:                              ctx.GlobalInt(validatorKeyIndex.Name),
		Config:                               *cfgs.generalConfig,
		CoreComponentsHolder:                 managedCoreComponents,
		ActivateBLSPubKeyMessageVerification: cfgs.systemSCConfig.StakingSystemSCConfig.ActivateBLSPubKeyMessageVerification,
		KeyLoader:                            &core.KeyLoader{},
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

func applyCompatibleConfigs(log logger.Logger, config *config.Config, ctx *cli.Context) {
	importDbDirectoryValue := ctx.GlobalString(importDbDirectory.Name)
	if len(importDbDirectoryValue) > 0 {
		importCheckpointRoundsModulus := uint(config.EpochStartConfig.RoundsPerEpoch)
		log.Info("import DB directory is set, altering config values!",
			"GeneralSettings.StartInEpochEnabled", "false",
			"StateTriesConfig.CheckpointRoundsModulus", importCheckpointRoundsModulus,
			"import DB path", importDbDirectoryValue,
		)
		config.GeneralSettings.StartInEpochEnabled = false
		config.StateTriesConfig.CheckpointRoundsModulus = importCheckpointRoundsModulus
	}
}

func closeAllComponents(
	log logger.Logger,
	healthService io.Closer,
	facade mainFactory.Closer,
	node *node.Node,
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

func cleanupStorageIfNecessary(workingDir string, ctx *cli.Context, log logger.Logger) error {
	storageCleanupFlagValue := ctx.GlobalBool(storageCleanup.Name)
	if storageCleanupFlagValue {
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

func getWorkingDir(ctx *cli.Context, log logger.Logger) string {
	var workingDir string
	var err error
	if ctx.IsSet(workingDirectory.Name) {
		workingDir = ctx.GlobalString(workingDirectory.Name)
	} else {
		workingDir, err = os.Getwd()
		if err != nil {
			log.LogIfError(err)
			workingDir = ""
		}
	}
	log.Trace("working directory", "path", workingDir)

	return workingDir
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
		go elasticIndexer.SaveValidatorsPubKeys(validatorsPubKeys, epoch)
	}
}

func enableGopsIfNeeded(ctx *cli.Context, log logger.Logger) {
	var gopsEnabled bool
	if ctx.IsSet(gopsEn.Name) {
		gopsEnabled = ctx.GlobalBool(gopsEn.Name)
	}

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

func updateLogger(workingDir string, ctx *cli.Context, log logger.Logger) (factory.FileLoggingHandler, error) {
	var fileLogging factory.FileLoggingHandler
	var err error
	withLogFile := ctx.GlobalBool(logSaveFile.Name)
	if withLogFile {
		fileLogging, err = logging.NewFileLogging(workingDir, defaultLogsPath)
		if err != nil {
			return nil, fmt.Errorf("%w creating a log file", err)
		}
	}

	err = logger.SetDisplayByteSlice(logger.ToHex)
	log.LogIfError(err)
	logger.ToggleCorrelation(ctx.GlobalBool(logWithCorrelation.Name))
	logger.ToggleLoggerName(ctx.GlobalBool(logWithLoggerName.Name))
	logLevelFlagValue := ctx.GlobalString(logLevel.Name)
	err = logger.SetLogLevel(logLevelFlagValue)
	if err != nil {
		return nil, err
	}
	noAnsiColor := ctx.GlobalBool(disableAnsiColor.Name)
	if noAnsiColor {
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
	log.Trace("logger updated", "level", logLevelFlagValue, "disable ANSI color", noAnsiColor)

	return fileLogging, nil
}

func readConfigs(log logger.Logger, ctx *cli.Context) (*configs, error) {
	log.Trace("reading configs")

	configurationFileName := ctx.GlobalString(configurationFile.Name)
	generalConfig, err := core.LoadMainConfig(configurationFileName)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationFileName)

	applyCompatibleConfigs(log, generalConfig, ctx)

	configurationApiFileName := ctx.GlobalString(configurationApiFile.Name)
	apiRoutesConfig, err := core.LoadApiConfig(configurationApiFileName)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationApiFileName)

	configurationEconomicsFileName := ctx.GlobalString(configurationEconomicsFile.Name)
	economicsConfig, err := core.LoadEconomicsConfig(configurationEconomicsFileName)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationEconomicsFileName)

	configurationSystemSCConfigFileName := ctx.GlobalString(configurationSystemSCFile.Name)
	systemSCConfig, err := core.LoadSystemSmartContractsConfig(configurationSystemSCConfigFileName)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationSystemSCConfigFileName)

	configurationRatingsFileName := ctx.GlobalString(configurationRatingsFile.Name)
	ratingsConfig, err := core.LoadRatingsConfig(configurationRatingsFileName)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationRatingsFileName)

	configurationPreferencesFileName := ctx.GlobalString(configurationPreferencesFile.Name)
	preferencesConfig, err := core.LoadPreferencesConfig(configurationPreferencesFileName)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPreferencesFileName)

	externalConfigurationFileName := ctx.GlobalString(externalConfigFile.Name)
	externalConfig, err := core.LoadExternalConfig(externalConfigurationFileName)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", externalConfigurationFileName)

	p2pConfigurationFileName := ctx.GlobalString(p2pConfigurationFile.Name)
	p2pConfig, err := core.LoadP2PConfig(p2pConfigurationFileName)
	if err != nil {
		return nil, err
	}

	log.Debug("config", "file", p2pConfigurationFileName)
	if ctx.IsSet(port.Name) {
		p2pConfig.Node.Port = ctx.GlobalString(port.Name)
	}
	if ctx.IsSet(destinationShardAsObserver.Name) {
		preferencesConfig.Preferences.DestinationShardAsObserver = ctx.GlobalString(destinationShardAsObserver.Name)
	}
	if ctx.IsSet(nodeDisplayName.Name) {
		preferencesConfig.Preferences.NodeDisplayName = ctx.GlobalString(nodeDisplayName.Name)
	}
	if ctx.IsSet(identityFlagName.Name) {
		preferencesConfig.Preferences.Identity = ctx.GlobalString(identityFlagName.Name)
	}

	return &configs{
		generalConfig:                    generalConfig,
		apiRoutesConfig:                  apiRoutesConfig,
		economicsConfig:                  economicsConfig,
		systemSCConfig:                   systemSCConfig,
		ratingsConfig:                    ratingsConfig,
		preferencesConfig:                preferencesConfig,
		externalConfig:                   externalConfig,
		p2pConfig:                        p2pConfig,
		configurationFileName:            configurationFileName,
		configurationEconomicsFileName:   configurationEconomicsFileName,
		configurationRatingsFileName:     configurationRatingsFileName,
		configurationPreferencesFileName: configurationPreferencesFileName,
		p2pConfigurationFileName:         p2pConfigurationFileName,
	}, nil
}
