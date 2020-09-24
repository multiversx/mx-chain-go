package main

import (
	"bytes"
	"errors"
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
	"github.com/ElrondNetwork/elrond-go/core/accumulator"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/closing"
	"github.com/ElrondNetwork/elrond-go/core/dblookupext"
	dbLookupFactory "github.com/ElrondNetwork/elrond-go/core/dblookupext/factory"
	"github.com/ElrondNetwork/elrond-go/core/forking"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/logging"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/facade"
	mainFactory "github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/genesis/parsing"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/health"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/nodeDebugFactory"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	processFactory "github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/headerCheck"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/rating"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/blackList"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/ElrondNetwork/elrond-go/update"
	exportFactory "github.com/ElrondNetwork/elrond-go/update/factory"
	"github.com/ElrondNetwork/elrond-go/update/trigger"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
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

	closableComponents := make([]mainFactory.Closer, 0)

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

	for {
		log.Trace("creating core components")

		chanCreateViews := make(chan struct{}, 1)
		chanLogRewrite := make(chan struct{}, 1)

		useTermui := !ctx.GlobalBool(useLogView.Name)
		statusHandlersFactoryArgs := &factory.StatusHandlersFactoryArgs{
			UseTermUI:      useTermui,
			ChanStartViews: chanCreateViews,
			ChanLogRewrite: chanLogRewrite,
		}

		statusHandlersFactory, err := factory.NewStatusHandlersFactory(statusHandlersFactoryArgs)
		if err != nil {
			return err
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
			return fmt.Errorf("NewCoreComponentsFactory failed: %w", err)
		}

		managedCoreComponents, err := mainFactory.NewManagedCoreComponents(coreComponentsFactory)
		if err != nil {
			return err
		}

		err = managedCoreComponents.Create()
		if err != nil {
			return err
		}

		closableComponents = append(closableComponents, managedCoreComponents)

		log.Trace("creating crypto components")
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
			return fmt.Errorf("NewCryptoComponentsFactory failed: %w", err)
		}

		managedCryptoComponents, err := mainFactory.NewManagedCryptoComponents(cryptoComponentsFactory)
		if err != nil {
			return err
		}

		err = managedCryptoComponents.Create()
		if err != nil {
			return err
		}
		closableComponents = append(closableComponents, managedCryptoComponents)

		log.Debug("block sign pubkey", "value", managedCryptoComponents.PublicKeyString())

		if ctx.IsSet(destinationShardAsObserver.Name) {
			cfgs.preferencesConfig.Preferences.DestinationShardAsObserver = ctx.GlobalString(destinationShardAsObserver.Name)
		}

		if ctx.IsSet(nodeDisplayName.Name) {
			cfgs.preferencesConfig.Preferences.NodeDisplayName = ctx.GlobalString(nodeDisplayName.Name)
		}

		if ctx.IsSet(identityFlagName.Name) {
			cfgs.preferencesConfig.Preferences.Identity = ctx.GlobalString(identityFlagName.Name)
		}

		err = cleanupStorageIfNecessary(workingDir, ctx, log)
		if err != nil {
			return err
		}

		healthService := health.NewHealthService(cfgs.generalConfig.Health, workingDir)
		if ctx.IsSet(useHealthService.Name) {
			healthService.Start()
		}

		log.Trace("creating network components")
		networkComponentsFactoryArgs := mainFactory.NetworkComponentsFactoryArgs{
			P2pConfig:     *cfgs.p2pConfig,
			MainConfig:    *cfgs.generalConfig,
			RatingsConfig: *cfgs.ratingsConfig,
			StatusHandler: managedCoreComponents.StatusHandler(),
			Marshalizer:   managedCoreComponents.InternalMarshalizer(),
			Syncer:        managedCoreComponents.SyncTimer(),
		}

		networkComponentsFactory, err := mainFactory.NewNetworkComponentsFactory(networkComponentsFactoryArgs)
		if err != nil {
			return fmt.Errorf("NewNetworkComponentsFactory failed: %w", err)
		}

		managedNetworkComponents, err := mainFactory.NewManagedNetworkComponents(networkComponentsFactory)
		if err != nil {
			return err
		}
		err = managedNetworkComponents.Create()
		if err != nil {
			return err
		}
		closableComponents = append(closableComponents, managedNetworkComponents)

		err = managedNetworkComponents.NetworkMessenger().Bootstrap()
		if err != nil {
			return err
		}
		log.Info(fmt.Sprintf("waiting %d seconds for network discovery...", core.SecondsToWaitForP2PBootstrap))
		time.Sleep(core.SecondsToWaitForP2PBootstrap * time.Second)

		log.Trace("creating economics data components")
		economicsData, err := economics.NewEconomicsData(cfgs.economicsConfig)
		if err != nil {
			return err
		}

		log.Trace("creating ratings data components")

		nodesSetup := managedCoreComponents.GenesisNodesSetup()

		nodesShuffler := sharding.NewHashValidatorsShuffler(
			nodesSetup.MinNumberOfShardNodes(),
			nodesSetup.MinNumberOfMetaNodes(),
			nodesSetup.GetHysteresis(),
			nodesSetup.GetAdaptivity(),
			true,
		)

		destShardIdAsObserver, err := core.ProcessDestinationShardAsObserver(cfgs.preferencesConfig.Preferences.DestinationShardAsObserver)
		if err != nil {
			return err
		}

		versionsCache, err := storageUnit.NewCache(storageFactory.GetCacherFromConfig(cfgs.generalConfig.Versions.Cache))
		if err != nil {
			return err
		}

		headerIntegrityVerifier, err := headerCheck.NewHeaderIntegrityVerifier(
			[]byte(managedCoreComponents.ChainID()),
			cfgs.generalConfig.Versions.VersionsByEpochs,
			cfgs.generalConfig.Versions.DefaultVersion,
			versionsCache,
		)
		if err != nil {
			return err
		}
		genesisShardCoordinator, nodeType, err := mainFactory.CreateShardCoordinator(
			managedCoreComponents.GenesisNodesSetup(),
			managedCryptoComponents.PublicKey(),
			cfgs.preferencesConfig.Preferences,
			log,
		)
		if err != nil {
			return err
		}
		bootstrapComponentsFactoryArgs := mainFactory.BootstrapComponentsFactoryArgs{
			Config:                  *cfgs.generalConfig,
			WorkingDir:              workingDir,
			DestinationAsObserver:   destShardIdAsObserver,
			GenesisNodesSetup:       nodesSetup,
			NodeShuffler:            nodesShuffler,
			ShardCoordinator:        genesisShardCoordinator,
			CoreComponents:          managedCoreComponents,
			CryptoComponents:        managedCryptoComponents,
			NetworkComponents:       managedNetworkComponents,
			HeaderIntegrityVerifier: headerIntegrityVerifier,
		}

		bootstrapComponentsFactory, err := mainFactory.NewBootstrapComponentsFactory(bootstrapComponentsFactoryArgs)
		if err != nil {
			return fmt.Errorf("NewBootstrapComponentsFactory failed: %w", err)
		}

		managedBootstrapComponents, err := mainFactory.NewManagedBootstrapComponents(bootstrapComponentsFactory)
		if err != nil {
			return err
		}

		err = managedBootstrapComponents.Create()
		if err != nil {
			return err
		}
		closableComponents = append(closableComponents, managedBootstrapComponents)

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

		log.Trace("initializing stats file")
		var shardId = core.GetShardIDString(genesisShardCoordinator.SelfId())
		err = initStatsFileMonitor(
			cfgs.generalConfig,
			managedCoreComponents.PathHandler(),
			shardId)
		if err != nil {
			return err
		}

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

		closableComponents = append(closableComponents, managedStateComponents)

		log.Trace("creating data components")
		epochStartNotifier := notifier.NewEpochStartSubscriptionHandler()

		dataArgs := mainFactory.DataComponentsFactoryArgs{
			Config:             *cfgs.generalConfig,
			EconomicsData:      economicsData,
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
		closableComponents = append(closableComponents, managedDataComponents)

		healthService.RegisterComponent(managedDataComponents.Datapool().Transactions())
		healthService.RegisterComponent(managedDataComponents.Datapool().UnsignedTransactions())
		healthService.RegisterComponent(managedDataComponents.Datapool().RewardTransactions())

		log.Trace("initializing metrics")
		err = metrics.InitMetrics(
			managedCoreComponents.StatusHandler(),
			managedCryptoComponents.PublicKeyString(),
			nodeType,
			shardCoordinator,
			nodesSetup,
			version,
			cfgs.economicsConfig,
			cfgs.generalConfig.EpochStartConfig.RoundsPerEpoch,
			managedCoreComponents.MinTransactionVersion(),
		)
		if err != nil {
			return err
		}

		chanLogRewrite <- struct{}{}
		chanCreateViews <- struct{}{}

		ratingDataArgs := rating.RatingsDataArg{
			Config:                   *cfgs.ratingsConfig,
			ShardConsensusSize:       nodesSetup.GetShardConsensusGroupSize(),
			MetaConsensusSize:        nodesSetup.GetMetaConsensusGroupSize(),
			ShardMinNodes:            nodesSetup.MinNumberOfShardNodes(),
			MetaMinNodes:             nodesSetup.MinNumberOfMetaNodes(),
			RoundDurationMiliseconds: nodesSetup.GetRoundDuration(),
		}
		ratingsData, err := rating.NewRatingsData(ratingDataArgs)
		if err != nil {
			return err
		}

		rater, err := rating.NewBlockSigningRater(ratingsData)
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

		nodesCoordinator, nodeShufflerOut, err := createNodesCoordinator(
			log,
			nodesSetup,
			cfgs.preferencesConfig.Preferences,
			epochStartNotifier,
			managedCryptoComponents.PublicKey(),
			managedCoreComponents.InternalMarshalizer(),
			managedCoreComponents.Hasher(),
			rater,
			managedDataComponents.StorageService().GetStorer(dataRetriever.BootstrapUnit),
			nodesShuffler,
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
		metrics.SaveUint64Metric(managedCoreComponents.StatusHandler(), core.MetricGasPerDataByte, economicsData.GasPerDataByte())
		metrics.SaveUint64Metric(managedCoreComponents.StatusHandler(), core.MetricMinGasPrice, economicsData.MinGasPrice())
		metrics.SaveUint64Metric(managedCoreComponents.StatusHandler(), core.MetricMinGasLimit, economicsData.MinGasLimit())

		sessionInfoFileOutput := fmt.Sprintf("%s:%s\n%s:%s\n%s:%v\n%s:%s\n%s:%v\n",
			"PkBlockSign", managedCryptoComponents.PublicKeyString(),
			"ShardId", shardId,
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
		computedRatingsDataStr := createStringFromRatingsData(ratingsData)
		err = ioutil.WriteFile(computedRatingsData, []byte(computedRatingsDataStr), os.ModePerm)
		log.LogIfError(err)

		gasScheduleConfigurationFileName := ctx.GlobalString(gasScheduleConfigurationFile.Name)
		gasSchedule, err := core.LoadGasScheduleConfig(gasScheduleConfigurationFileName)
		if err != nil {
			return err
		}

		log.Trace("creating time cache for requested items components")
		requestedItemsHandler := timecache.NewTimeCache(time.Duration(uint64(time.Millisecond) * nodesSetup.GetRoundDuration()))

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
			RoundDurationSec:   nodesSetup.GetRoundDuration() / 1000,
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
		closableComponents = append(closableComponents, managedStateComponents)
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
			Config:                    coreArgs.Config,
			AccountsParser:            accountsParser,
			SmartContractParser:       smartContractParser,
			EconomicsData:             economicsData,
			NodesConfig:               nodesSetup,
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
			Rater:                     rater,
			RatingsData:               ratingsData,
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
			HeaderIntegrityVerifier:   headerIntegrityVerifier,
			ChanGracefullyClose:       chanStopNodeProcess,
		}
		processComponentsFactory, err := mainFactory.NewProcessComponentsFactory(processArgs)
		if err != nil {
			return fmt.Errorf("NewDataComponentsFactory failed: %w", err)
		}

		managedProcessComponents, err := mainFactory.NewManagedProcessComponents(processComponentsFactory)
		if err != nil {
			return err
		}
		err = managedProcessComponents.Create()
		if err != nil {
			return err
		}
		closableComponents = append(closableComponents, managedProcessComponents)

		historyRepository.RegisterToBlockTracker(managedProcessComponents.BlockTracker())
		managedStatusComponents.SetForkDetector(managedProcessComponents.ForkDetector())
		err = managedStatusComponents.StartPolling()

		hardForkTrigger, err := createHardForkTrigger(
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
			nodesSetup,
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

		log.Trace("creating node structure")
		currentNode, err := createNode(
			cfgs.generalConfig,
			cfgs.preferencesConfig,
			nodesSetup,
			managedBootstrapComponents,
			managedCoreComponents,
			managedCryptoComponents,
			managedDataComponents,
			managedNetworkComponents,
			managedProcessComponents,
			managedStateComponents,
			managedStatusComponents,
			ctx.GlobalUint64(bootstrapRoundIndex.Name),
			version,
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

		log.Trace("creating software checker structure")
		softwareVersionChecker, err := factory.CreateSoftwareVersionChecker(
			managedCoreComponents.StatusHandler(),
			cfgs.generalConfig.SoftwareVersionConfig,
		)
		if err != nil {
			log.Debug("nil software version checker", "error", err.Error())
		} else {
			softwareVersionChecker.StartCheckSoftwareVersion()
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
		apiResolver, err := createApiResolver(
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
			economicsData,
			managedCryptoComponents.MessageSignVerifier(),
			nodesSetup,
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

		log.Debug("starting node...")

		consensusArgs := mainFactory.ConsensusComponentsFactoryArgs{
			Config:              *cfgs.generalConfig,
			ConsensusGroupSize:  int(nodesSetup.GetShardConsensusGroupSize()),
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
		closableComponents = append(closableComponents, managedConsensusComponents)

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
			closeAllComponents(log, healthService, closableComponents, chanCloseComponents)
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
			err := pprof.Lookup("goroutine").WriteTo(buffer, 1)
			if err != nil {
				log.Error("could not dump goroutines")
			}

			log.Error("Remaining goroutines", "num", runtime.NumGoroutine)
			log.Warn(buffer.String())

			panic("test")
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
	closableComponents []mainFactory.Closer,
	chanCloseComponents chan struct{},
) {
	log.Debug("closing health service...")
	err := healthService.Close()
	log.LogIfError(err)

	log.Debug("closing all components")
	for i := len(closableComponents) - 1; i >= 0; i-- {
		managedComponent := closableComponents[i]
		log.Debug("closing", fmt.Sprintf("managedComponent %t", managedComponent))
		err = managedComponent.Close()
		log.LogIfError(err)
	}

	chanCloseComponents <- struct{}{}
}

func createStringFromRatingsData(ratingsData *rating.RatingsData) string {
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

func createNodesCoordinator(
	log logger.Logger,
	nodesConfig mainFactory.NodesSetupHandler,
	prefsConfig config.PreferencesConfig,
	epochStartNotifier epochStart.RegistrationHandler,
	pubKey crypto.PublicKey,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	ratingAndListIndexHandler sharding.PeerAccountListAndRatingHandler,
	bootStorer storage.Storer,
	nodeShuffler sharding.NodesShuffler,
	epochConfig config.EpochStartConfig,
	currentShardID uint32,
	chanStopNodeProcess chan endProcess.ArgEndProcess,
	bootstrapParameters mainFactory.BootstrapParamsHandler,
	startEpoch uint32,
) (sharding.NodesCoordinator, update.Closer, error) {
	shardIDAsObserver, err := core.ProcessDestinationShardAsObserver(prefsConfig.DestinationShardAsObserver)
	if err != nil {
		return nil, nil, err
	}
	if shardIDAsObserver == core.DisabledShardIDAsObserver {
		shardIDAsObserver = uint32(0)
	}

	nbShards := nodesConfig.NumberOfShards()
	shardConsensusGroupSize := int(nodesConfig.GetShardConsensusGroupSize())
	metaConsensusGroupSize := int(nodesConfig.GetMetaConsensusGroupSize())
	eligibleNodesInfo, waitingNodesInfo := nodesConfig.InitialNodesInfo()

	eligibleValidators, errEligibleValidators := sharding.NodesInfoToValidators(eligibleNodesInfo)
	if errEligibleValidators != nil {
		return nil, nil, errEligibleValidators
	}

	waitingValidators, errWaitingValidators := sharding.NodesInfoToValidators(waitingNodesInfo)
	if errWaitingValidators != nil {
		return nil, nil, errWaitingValidators
	}

	currentEpoch := startEpoch
	if bootstrapParameters.NodesConfig() != nil {
		nodeRegistry := bootstrapParameters.NodesConfig()
		currentEpoch = bootstrapParameters.Epoch()
		eligibles := nodeRegistry.EpochsConfig[fmt.Sprintf("%d", currentEpoch)].EligibleValidators
		eligibleValidators, err = sharding.SerializableValidatorsToValidators(eligibles)
		if err != nil {
			return nil, nil, err
		}

		waitings := nodeRegistry.EpochsConfig[fmt.Sprintf("%d", currentEpoch)].WaitingValidators
		waitingValidators, err = sharding.SerializableValidatorsToValidators(waitings)
		if err != nil {
			return nil, nil, err
		}
	}

	pubKeyBytes, err := pubKey.ToByteArray()
	if err != nil {
		return nil, nil, err
	}

	consensusGroupCache, err := lrucache.NewCache(25000)
	if err != nil {
		return nil, nil, err
	}

	maxThresholdEpochDuration := epochConfig.MaxShuffledOutRestartThreshold
	if !(maxThresholdEpochDuration >= 0.0 && maxThresholdEpochDuration <= 1.0) {
		return nil, nil, fmt.Errorf("invalid max threshold for shuffled out handler")
	}
	minThresholdEpochDuration := epochConfig.MinShuffledOutRestartThreshold
	if !(minThresholdEpochDuration >= 0.0 && minThresholdEpochDuration <= 1.0) {
		return nil, nil, fmt.Errorf("invalid min threshold for shuffled out handler")
	}

	epochDuration := int64(nodesConfig.GetRoundDuration()) * epochConfig.RoundsPerEpoch
	minDurationBeforeStopProcess := int64(minThresholdEpochDuration * float64(epochDuration))
	maxDurationBeforeStopProcess := int64(maxThresholdEpochDuration * float64(epochDuration))

	minDurationInterval := time.Millisecond * time.Duration(minDurationBeforeStopProcess)
	maxDurationInterval := time.Millisecond * time.Duration(maxDurationBeforeStopProcess)

	log.Debug("closing.NewShuffleOutCloser",
		"minDurationInterval", minDurationInterval,
		"maxDurationInterval", maxDurationInterval,
	)

	nodeShufflerOut, err := closing.NewShuffleOutCloser(
		minDurationInterval,
		maxDurationInterval,
		chanStopNodeProcess,
	)
	if err != nil {
		return nil, nil, err
	}
	shuffledOutHandler, err := sharding.NewShuffledOutTrigger(pubKeyBytes, currentShardID, nodeShufflerOut.EndOfProcessingHandler)
	if err != nil {
		return nil, nil, err
	}

	argumentsNodesCoordinator := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: shardConsensusGroupSize,
		MetaConsensusGroupSize:  metaConsensusGroupSize,
		Marshalizer:             marshalizer,
		Hasher:                  hasher,
		Shuffler:                nodeShuffler,
		EpochStartNotifier:      epochStartNotifier,
		BootStorer:              bootStorer,
		ShardIDAsObserver:       shardIDAsObserver,
		NbShards:                nbShards,
		EligibleNodes:           eligibleValidators,
		WaitingNodes:            waitingValidators,
		SelfPublicKey:           pubKeyBytes,
		ConsensusGroupCache:     consensusGroupCache,
		ShuffledOutHandler:      shuffledOutHandler,
		Epoch:                   currentEpoch,
		StartEpoch:              startEpoch,
	}

	baseNodesCoordinator, err := sharding.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
	if err != nil {
		return nil, nil, err
	}

	nodesCoordinator, err := sharding.NewIndexHashedNodesCoordinatorWithRater(baseNodesCoordinator, ratingAndListIndexHandler)
	if err != nil {
		return nil, nil, err
	}

	return nodesCoordinator, nodeShufflerOut, nil
}

func getConsensusGroupSize(nodesConfig mainFactory.NodesSetupHandler, shardCoordinator sharding.Coordinator) (uint32, error) {
	if shardCoordinator.SelfId() == core.MetachainShardId {
		return nodesConfig.GetMetaConsensusGroupSize(), nil
	}
	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return nodesConfig.GetShardConsensusGroupSize(), nil
	}

	return 0, state.ErrUnknownShardId
}

func createHardForkTrigger(
	config *config.Config,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	coreData mainFactory.CoreComponentsHolder,
	stateComponents mainFactory.StateComponentsHolder,
	data mainFactory.DataComponentsHolder,
	crypto mainFactory.CryptoComponentsHolder,
	process mainFactory.ProcessComponentsHolder,
	network mainFactory.NetworkComponentsHolder,
	whiteListRequest process.WhiteListHandler,
	whiteListerVerifiedTxs process.WhiteListHandler,
	chanStopNodeProcess chan endProcess.ArgEndProcess,
	epochNotifier factory.EpochStartNotifier,
	importStartHandler update.ImportStartHandler,
	nodesSetup update.GenesisNodesSetupHandler,
	workingDir string,
) (node.HardforkTrigger, error) {

	selfPubKeyBytes := crypto.PublicKeyBytes()
	triggerPubKeyBytes, err := coreData.ValidatorPubKeyConverter().Decode(config.Hardfork.PublicKeyToListenFrom)
	if err != nil {
		return nil, fmt.Errorf("%w while decoding HardforkConfig.PublicKeyToListenFrom", err)
	}

	accountsDBs := make(map[state.AccountsDbIdentifier]state.AccountsAdapter)
	accountsDBs[state.UserAccountsState] = stateComponents.AccountsAdapter()
	accountsDBs[state.PeerAccountsState] = stateComponents.PeerAccounts()
	hardForkConfig := config.Hardfork
	exportFolder := filepath.Join(workingDir, hardForkConfig.ImportFolder)
	argsExporter := exportFactory.ArgsExporter{
		CoreComponents:           coreData,
		CryptoComponents:         crypto,
		HeaderValidator:          process.HeaderConstructionValidator(),
		DataPool:                 data.Datapool(),
		StorageService:           data.StorageService(),
		RequestHandler:           process.RequestHandler(),
		ShardCoordinator:         shardCoordinator,
		Messenger:                network.NetworkMessenger(),
		ActiveAccountsDBs:        accountsDBs,
		ExistingResolvers:        process.ResolversFinder(),
		ExportFolder:             exportFolder,
		ExportTriesStorageConfig: hardForkConfig.ExportTriesStorageConfig,
		ExportStateStorageConfig: hardForkConfig.ExportStateStorageConfig,
		ExportStateKeysConfig:    hardForkConfig.ExportKeysStorageConfig,
		WhiteListHandler:         whiteListRequest,
		WhiteListerVerifiedTxs:   whiteListerVerifiedTxs,
		InterceptorsContainer:    process.InterceptorsContainer(),
		NodesCoordinator:         nodesCoordinator,
		HeaderSigVerifier:        process.HeaderSigVerifier(),
		HeaderIntegrityVerifier:  process.HeaderIntegrityVerifier(),
		MaxTrieLevelInMemory:     config.StateTriesConfig.MaxStateTrieLevelInMemory,
		InputAntifloodHandler:    network.InputAntiFloodHandler(),
		OutputAntifloodHandler:   network.OutputAntiFloodHandler(),
		ValidityAttester:         process.BlockTracker(),
		Rounder:                  process.Rounder(),
		GenesisNodesSetupHandler: nodesSetup,
	}
	hardForkExportFactory, err := exportFactory.NewExportHandlerFactory(argsExporter)
	if err != nil {
		return nil, err
	}

	atArgumentParser := smartContract.NewArgumentParser()
	argTrigger := trigger.ArgHardforkTrigger{
		TriggerPubKeyBytes:        triggerPubKeyBytes,
		SelfPubKeyBytes:           selfPubKeyBytes,
		Enabled:                   config.Hardfork.EnableTrigger,
		EnabledAuthenticated:      config.Hardfork.EnableTriggerFromP2P,
		ArgumentParser:            atArgumentParser,
		EpochProvider:             process.EpochStartTrigger(),
		ExportFactoryHandler:      hardForkExportFactory,
		ChanStopNodeProcess:       chanStopNodeProcess,
		EpochConfirmedNotifier:    epochNotifier,
		CloseAfterExportInMinutes: config.Hardfork.CloseAfterExportInMinutes,
		ImportStartHandler:        importStartHandler,
	}
	hardforkTrigger, err := trigger.NewTrigger(argTrigger)
	if err != nil {
		return nil, err
	}

	return hardforkTrigger, nil
}

func createNode(
	config *config.Config,
	preferencesConfig *config.Preferences,
	nodesConfig mainFactory.NodesSetupHandler,
	bootstrapComponents mainFactory.BootstrapComponentsHandler,
	coreComponents mainFactory.CoreComponentsHandler,
	cryptoComponents mainFactory.CryptoComponentsHandler,
	dataComponents mainFactory.DataComponentsHandler,
	networkComponents mainFactory.NetworkComponentsHandler,
	processComponents mainFactory.ProcessComponentsHandler,
	stateComponents mainFactory.StateComponentsHandler,
	statusComponents mainFactory.StatusComponentsHandler,
	bootstrapRoundIndex uint64,
	version string,
	requestedItemsHandler dataRetriever.RequestedItemsHandler,
	whiteListRequest process.WhiteListHandler,
	whiteListerVerifiedTxs process.WhiteListHandler,
	chanStopNodeProcess chan endProcess.ArgEndProcess,
	hardForkTrigger node.HardforkTrigger,
	historyRepository dblookupext.HistoryRepository,
) (*node.Node, error) {
	var err error
	var consensusGroupSize uint32
	consensusGroupSize, err = getConsensusGroupSize(nodesConfig, processComponents.ShardCoordinator())
	if err != nil {
		return nil, err
	}

	var txAccumulator node.Accumulator
	txAccumulatorConfig := config.Antiflood.TxAccumulator
	txAccumulator, err = accumulator.NewTimeAccumulator(
		time.Duration(txAccumulatorConfig.MaxAllowedTimeInMilliseconds)*time.Millisecond,
		time.Duration(txAccumulatorConfig.MaxDeviationTimeInMilliseconds)*time.Millisecond,
	)
	if err != nil {
		return nil, err
	}

	prepareOpenTopics(networkComponents.InputAntiFloodHandler(), processComponents.ShardCoordinator())

	peerDenialEvaluator, err := blackList.NewPeerDenialEvaluator(
		networkComponents.PeerBlackListHandler(),
		networkComponents.PubKeyCacher(),
		processComponents.PeerShardMapper(),
	)
	if err != nil {
		return nil, err
	}

	err = networkComponents.NetworkMessenger().SetPeerDenialEvaluator(peerDenialEvaluator)
	if err != nil {
		return nil, err
	}

	genesisTime := time.Unix(nodesConfig.GetStartTime(), 0)
	heartbeatArgs := mainFactory.HeartbeatComponentsFactoryArgs{
		Config:            *config,
		Prefs:             *preferencesConfig,
		AppVersion:        version,
		GenesisTime:       genesisTime,
		HardforkTrigger:   hardForkTrigger,
		CoreComponents:    coreComponents,
		DataComponents:    dataComponents,
		NetworkComponents: networkComponents,
		CryptoComponents:  cryptoComponents,
		ProcessComponents: processComponents,
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

	var nd *node.Node
	nd, err = node.NewNode(
		node.WithBootstrapComponents(bootstrapComponents),
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithNetworkComponents(networkComponents),
		node.WithProcessComponents(processComponents),
		node.WithCryptoComponents(cryptoComponents),
		node.WithStateComponents(stateComponents),
		node.WithStatusComponents(statusComponents),
		node.WithInitialNodesPubKeys(nodesConfig.InitialNodesPubKeys()),
		node.WithRoundDuration(nodesConfig.GetRoundDuration()),
		node.WithConsensusGroupSize(int(consensusGroupSize)),
		node.WithGenesisTime(genesisTime),
		node.WithConsensusType(config.Consensus.Type),
		node.WithBootstrapRoundIndex(bootstrapRoundIndex),
		node.WithPeerDenialEvaluator(peerDenialEvaluator),
		node.WithRequestedItemsHandler(requestedItemsHandler),
		node.WithTxAccumulator(txAccumulator),
		node.WithHardforkTrigger(hardForkTrigger),
		node.WithWhiteListHandler(whiteListRequest),
		node.WithWhiteListHandlerVerified(whiteListerVerifiedTxs),
		node.WithSignatureSize(config.ValidatorPubkeyConverter.SignatureLength),
		node.WithPublicKeySize(config.ValidatorPubkeyConverter.Length),
		node.WithNodeStopChannel(chanStopNodeProcess),
		node.WithHistoryRepository(historyRepository),
	)
	if err != nil {
		return nil, errors.New("error creating node: " + err.Error())
	}

	if processComponents.ShardCoordinator().SelfId() < processComponents.ShardCoordinator().NumberOfShards() {
		err = nd.CreateShardedStores()
		if err != nil {
			return nil, err
		}
	}

	err = nodeDebugFactory.CreateInterceptedDebugHandler(
		nd,
		processComponents.InterceptorsContainer(),
		processComponents.ResolversFinder(),
		config.Debug.InterceptorResolver,
	)
	if err != nil {
		return nil, err
	}

	return nd, nil
}

func initStatsFileMonitor(
	config *config.Config,
	pathManager storage.PathManagerHandler,
	shardId string,
) error {
	err := startStatisticsMonitor(config, pathManager, shardId)
	if err != nil {
		return err
	}

	return nil
}

func startStatisticsMonitor(
	generalConfig *config.Config,
	pathManager storage.PathManagerHandler,
	shardId string,
) error {
	if !generalConfig.ResourceStats.Enabled {
		return nil
	}

	if generalConfig.ResourceStats.RefreshIntervalInSec < 1 {
		return errors.New("invalid RefreshIntervalInSec in section [ResourceStats]. Should be an integer higher than 1")
	}

	resMon := statistics.NewResourceMonitor()

	go func() {
		for {
			resMon.SaveStatistics(generalConfig, pathManager, shardId)
			time.Sleep(time.Second * time.Duration(generalConfig.ResourceStats.RefreshIntervalInSec))
		}
	}()

	return nil
}

func createApiResolver(
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
	economics *economics.EconomicsData,
	messageSigVerifier vm.MessageSignVerifier,
	nodesSetup sharding.GenesisNodesSetupHandler,
	systemSCConfig *config.SystemSmartContractsConfig,
) (facade.ApiResolver, error) {
	var vmFactory process.VirtualMachinesContainerFactory
	var err error

	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasMap:          gasSchedule,
		MapDNSAddresses: make(map[string]struct{}),
		Marshalizer:     marshalizer,
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
		)
		if err != nil {
			return nil, err
		}
	} else {
		vmFactory, err = shard.NewVMContainerFactory(
			config.VirtualMachineConfig,
			economics.MaxGasLimitPerBlock(shardCoordinator.SelfId()),
			gasSchedule,
			argsHook)
		if err != nil {
			return nil, err
		}
	}

	vmContainer, err := vmFactory.Create()
	if err != nil {
		return nil, err
	}

	scQueryService, err := smartContract.NewSCQueryService(vmContainer, economics)
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

	return external.NewNodeApiResolver(scQueryService, statusMetrics, txCostHandler)
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

// prepareOpenTopics will set to the anti flood handler the topics for which
// the node can receive messages from others than validators
func prepareOpenTopics(
	antiflood mainFactory.P2PAntifloodHandler,
	shardCoordinator sharding.Coordinator,
) {
	selfID := shardCoordinator.SelfId()
	if selfID == core.MetachainShardId {
		antiflood.SetTopicsForAll(core.HeartbeatTopic)
		return
	}

	selfShardTxTopic := processFactory.TransactionTopic + core.CommunicationIdentifierBetweenShards(selfID, selfID)
	antiflood.SetTopicsForAll(core.HeartbeatTopic, selfShardTxTopic)
}
