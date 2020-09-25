package factory

import (
	"math/big"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	dbLookupFactory "github.com/ElrondNetwork/elrond-go/core/dblookupext/factory"
	"github.com/ElrondNetwork/elrond-go/core/forking"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/genesis/parsing"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/headerCheck"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/rating"
	"github.com/ElrondNetwork/elrond-go/sharding"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/ElrondNetwork/elrond-go/update/trigger"
	"github.com/stretchr/testify/require"
)

// ------------ Test TestProcessComponents --------------------
func TestProcessComponents_Close_ShouldWork(t *testing.T) {
	//t.Skip()

	generalConfig, _ := core.LoadMainConfig(configPath)
	ratingsConfig, _ := core.LoadRatingsConfig(ratingsPath)
	economicsConfig, _ := core.LoadEconomicsConfig(economicsPath)
	prefsConfig, _ := core.LoadPreferencesConfig(prefsPath)
	p2pConfig, _ := core.LoadP2PConfig(p2pPath)
	externalConfig, _ := core.LoadExternalConfig(externalPath)
	systemSCConfig, _ := core.LoadSystemSmartContractsConfig(systemSCConfigPath)

	nrBefore := runtime.NumGoroutine()
	printStack()

	managedCoreComponents, _ := createCoreComponents(*generalConfig, *ratingsConfig, *economicsConfig)
	managedCryptoComponents, _ := createCryptoComponents(*generalConfig, *systemSCConfig, managedCoreComponents)
	managedNetworkComponents, _ := createNetworkComponents(*generalConfig, *p2pConfig, *ratingsConfig, managedCoreComponents)
	managedBootstrapComponents, _ := createBootstrapComponents(
		*generalConfig,
		prefsConfig.Preferences,
		managedCoreComponents,
		managedCryptoComponents,
		managedNetworkComponents)
	epochStartNotifier := notifier.NewEpochStartSubscriptionHandler()
	managedDataComponents, _ := createDataComponents(*generalConfig, *economicsConfig, epochStartNotifier, managedCoreComponents)
	managedStateComponents, _ := createStateComponents(*generalConfig, managedCoreComponents, managedBootstrapComponents)
	nodesSetup := managedCoreComponents.GenesisNodesSetup()
	chanStopNodeProcess := make(chan endProcess.ArgEndProcess, 1)
	genesisShardCoordinator, nodesCoordinator, _, ratingsData, rater := createCoordinators(generalConfig, prefsConfig, ratingsConfig, nodesSetup, epochStartNotifier, chanStopNodeProcess, managedCoreComponents, managedCryptoComponents, managedDataComponents, managedBootstrapComponents)

	managedStatusComponents, err := createStatusComponents(
		*generalConfig,
		*externalConfig,
		genesisShardCoordinator,
		nodesCoordinator,
		epochStartNotifier,
		managedCoreComponents,
		managedDataComponents,
		managedNetworkComponents)
	require.Nil(t, err)
	require.NotNil(t, managedStatusComponents)
	require.Nil(t, err)
	require.NotNil(t, managedStatusComponents)

	time.Sleep(5 * time.Second)

	managedProcessComponents, err := createProcessComponents(
		generalConfig,
		economicsConfig,
		ratingsConfig,
		systemSCConfig,
		nodesSetup,
		nodesCoordinator,
		epochStartNotifier,
		genesisShardCoordinator,
		ratingsData,
		rater,
		managedCoreComponents,
		managedCryptoComponents,
		managedDataComponents,
		managedStateComponents,
		managedNetworkComponents,
		managedBootstrapComponents,
		managedStatusComponents,
		chanStopNodeProcess)
	require.Nil(t, err)

	time.Sleep(5 * time.Second)

	managedStatusComponents.SetForkDetector(managedProcessComponents.ForkDetector())
	err = managedStatusComponents.StartPolling()

	time.Sleep(5 * time.Second)

	err = managedProcessComponents.Close()

	time.Sleep(5 * time.Second)

	_ = managedStatusComponents.Close()
	_ = managedStateComponents.Close()
	_ = managedDataComponents.Close()
	_ = managedBootstrapComponents.Close()
	_ = managedNetworkComponents.Close()
	_ = managedCryptoComponents.Close()
	_ = managedCoreComponents.Close()

	time.Sleep(5 * time.Second)

	nrAfter := runtime.NumGoroutine()
	if nrBefore != nrAfter {
		printStack()
	}

	require.Equal(t, nrBefore, nrAfter)
}

func createProcessComponents(generalConfig *config.Config, economicsConfig *config.EconomicsConfig, ratingsConfig *config.RatingsConfig, systemSCConfig *config.SystemSmartContractsConfig, nodesSetup factory.NodesSetupHandler, nodesCoordinator sharding.NodesCoordinator, epochStartNotifier factory.EpochStartNotifier, genesisShardCoordinator sharding.Coordinator, ratingsData *rating.RatingsData, rater sharding.PeerAccountListAndRatingHandler, managedCoreComponents factory.CoreComponentsHandler, managedCryptoComponents factory.CryptoComponentsHandler, managedDataComponents factory.DataComponentsHandler, managedStateComponents factory.StateComponentsHandler, managedNetworkComponents factory.NetworkComponentsHandler, managedBootstrapComponents factory.BootstrapComponentsHandler, managedStatusComponents factory.StatusComponentsHandler, chanStopNodeProcess chan endProcess.ArgEndProcess) (factory.ProcessComponentsHandler, error) {
	economicsData, err := economics.NewEconomicsData(economicsConfig)
	totalSupply, _ := big.NewInt(0).SetString(economicsConfig.GlobalSettings.GenesisTotalSupply, 10)
	gasSchedule, err := core.LoadGasScheduleConfig(gasSchedule)
	if err != nil {
		return nil, err
	}

	accountsParser, err := parsing.NewAccountsParser(
		genesisPath,
		totalSupply,
		managedCoreComponents.AddressPubKeyConverter(),
		managedCryptoComponents.TxSignKeyGen(),
	)
	if err != nil {
		return nil, err
	}

	smartContractParser, err := parsing.NewSmartContractsParser(
		genesisSmartContracts,
		managedCoreComponents.AddressPubKeyConverter(),
		managedCryptoComponents.TxSignKeyGen(),
	)
	if err != nil {
		return nil, err
	}

	requestedItemsHandler := timecache.NewTimeCache(time.Duration(uint64(time.Millisecond) * nodesSetup.GetRoundDuration()))

	whiteListCache, err := storageUnit.NewCache(storageFactory.GetCacherFromConfig(generalConfig.WhiteListPool))
	if err != nil {
		return nil, err
	}

	whiteListRequest, err := interceptors.NewWhiteListDataVerifier(whiteListCache)
	if err != nil {
		return nil, err
	}

	whiteListerVerifiedTxs, err := createWhiteListerVerifiedTxs(generalConfig)
	if err != nil {
		return nil, err
	}

	importStartHandler, err := trigger.NewImportStartHandler(filepath.Join("workingDir", core.DefaultDBPath), "appVersion")
	if err != nil {
		return nil, err
	}

	historyRepoFactoryArgs := &dbLookupFactory.ArgsHistoryRepositoryFactory{
		SelfShardID: genesisShardCoordinator.SelfId(),
		Config:      generalConfig.DbLookupExtensions,
		Hasher:      managedCoreComponents.Hasher(),
		Marshalizer: managedCoreComponents.InternalMarshalizer(),
		Store:       managedDataComponents.StorageService(),
	}
	historyRepositoryFactory, err := dbLookupFactory.NewHistoryRepositoryFactory(historyRepoFactoryArgs)
	if err != nil {
		return nil, err
	}

	historyRepository, err := historyRepositoryFactory.Create()
	if err != nil {
		return nil, err
	}

	versionsCache, _ := storageUnit.NewCache(storageFactory.GetCacherFromConfig(generalConfig.Versions.Cache))
	headerIntegrityVerifier, err := headerCheck.NewHeaderIntegrityVerifier(
		[]byte(managedCoreComponents.ChainID()),
		generalConfig.Versions.VersionsByEpochs,
		generalConfig.Versions.DefaultVersion,
		versionsCache,
	)
	if err != nil {
		return nil, err
	}
	epochNotifier := forking.NewGenericEpochNotifier()

	processArgs := factory.ProcessComponentsFactoryArgs{
		Config:                    *generalConfig,
		AccountsParser:            accountsParser,
		SmartContractParser:       smartContractParser,
		EconomicsData:             economicsData,
		NodesConfig:               nodesSetup,
		GasSchedule:               gasSchedule,
		Rounder:                   managedCoreComponents.Rounder(),
		ShardCoordinator:          genesisShardCoordinator,
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
		EpochStart:                &generalConfig.EpochStartConfig,
		Rater:                     rater,
		RatingsData:               ratingsData,
		StartEpochNum:             managedBootstrapComponents.EpochBootstrapParams().Epoch(),
		SizeCheckDelta:            generalConfig.Marshalizer.SizeCheckDelta,
		StateCheckpointModulus:    generalConfig.StateTriesConfig.CheckpointRoundsModulus,
		MaxComputableRounds:       generalConfig.GeneralSettings.MaxComputableRounds,
		NumConcurrentResolverJobs: generalConfig.Antiflood.NumConcurrentResolverJobs,
		MinSizeInBytes:            generalConfig.BlockSizeThrottleConfig.MinSizeInBytes,
		MaxSizeInBytes:            generalConfig.BlockSizeThrottleConfig.MaxSizeInBytes,
		MaxRating:                 ratingsConfig.General.MaxRating,
		ValidatorPubkeyConverter:  managedCoreComponents.ValidatorPubKeyConverter(),
		SystemSCConfig:            systemSCConfig,
		Version:                   "version",
		ImportStartHandler:        importStartHandler,
		WorkingDir:                "workingDir",
		Indexer:                   managedStatusComponents.ElasticIndexer(),
		TpsBenchmark:              managedStatusComponents.TpsBenchmark(),
		HistoryRepo:               historyRepository,
		EpochNotifier:             epochNotifier,
		HeaderIntegrityVerifier:   headerIntegrityVerifier,
		ChanGracefullyClose:       chanStopNodeProcess,
	}
	processComponentsFactory, err := factory.NewProcessComponentsFactory(processArgs)
	if err != nil {
		return nil, err
	}

	managedProcessComponents, err := factory.NewManagedProcessComponents(processComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedProcessComponents.Create()
	return managedProcessComponents, err
}

func createWhiteListerVerifiedTxs(generalConfig *config.Config) (process.WhiteListHandler, error) {
	whiteListCacheVerified, err := storageUnit.NewCache(storageFactory.GetCacherFromConfig(generalConfig.WhiteListerVerifiedTxs))
	if err != nil {
		return nil, err
	}
	return interceptors.NewWhiteListDataVerifier(whiteListCacheVerified)
}
