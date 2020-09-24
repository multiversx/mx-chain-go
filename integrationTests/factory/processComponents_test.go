package factory

import (
	"math/big"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	dbLookupFactory "github.com/ElrondNetwork/elrond-go/core/dblookupext/factory"
	"github.com/ElrondNetwork/elrond-go/core/forking"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
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
	t.Skip()

	generalConfig, _ := core.LoadMainConfig(configPath)
	ratingsConfig, _ := core.LoadRatingsConfig(ratingsPath)
	economicsConfig, _ := core.LoadEconomicsConfig(economicsPath)
	prefsConfig, _ := core.LoadPreferencesConfig(prefsPath)
	p2pConfig, _ := core.LoadP2PConfig(p2pPath)
	externalConfig, _ := core.LoadExternalConfig(externalPath)
	systemSCConfig, _ := core.LoadSystemSmartContractsConfig(systemSCConfigPath)

	nrBefore := runtime.NumGoroutine()
	printStack()

	log := logger.GetOrCreate("test")

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
	genesisShardCoordinator, _, _ := factory.CreateShardCoordinator(
		nodesSetup,
		managedCryptoComponents.PublicKey(),
		prefsConfig.Preferences,
		log,
	)
	ratingDataArgs := rating.RatingsDataArg{
		Config:                   *ratingsConfig,
		ShardConsensusSize:       nodesSetup.GetShardConsensusGroupSize(),
		MetaConsensusSize:        nodesSetup.GetMetaConsensusGroupSize(),
		ShardMinNodes:            nodesSetup.MinNumberOfShardNodes(),
		MetaMinNodes:             nodesSetup.MinNumberOfMetaNodes(),
		RoundDurationMiliseconds: nodesSetup.GetRoundDuration(),
	}

	ratingsData, _ := rating.NewRatingsData(ratingDataArgs)
	rater, _ := rating.NewBlockSigningRater(ratingsData)

	nodesShuffler := sharding.NewHashValidatorsShuffler(
		nodesSetup.MinNumberOfShardNodes(),
		nodesSetup.MinNumberOfMetaNodes(),
		nodesSetup.GetHysteresis(),
		nodesSetup.GetAdaptivity(),
		true,
	)
	chanStopNodeProcess := make(chan endProcess.ArgEndProcess, 1)
	nodesCoordinator, _, _ := factory.CreateNodesCoordinator(
		log,
		nodesSetup,
		prefsConfig.Preferences,
		epochStartNotifier,
		managedCryptoComponents.PublicKey(),
		managedCoreComponents.InternalMarshalizer(),
		managedCoreComponents.Hasher(),
		rater,
		managedDataComponents.StorageService().GetStorer(dataRetriever.BootstrapUnit),
		nodesShuffler,
		generalConfig.EpochStartConfig,
		genesisShardCoordinator.SelfId(),
		chanStopNodeProcess,
		managedBootstrapComponents.EpochBootstrapParams(),
		managedBootstrapComponents.EpochBootstrapParams().Epoch(),
	)

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

	time.Sleep(5 * time.Second)

	economicsData, err := economics.NewEconomicsData(economicsConfig)
	totalSupply, ok := big.NewInt(0).SetString(economicsConfig.GlobalSettings.GenesisTotalSupply, 10)
	gasSchedule, err := core.LoadGasScheduleConfig(gasSchedule)
	require.True(t, ok)

	accountsParser, err := parsing.NewAccountsParser(
		genesisPath,
		totalSupply,
		managedCoreComponents.AddressPubKeyConverter(),
		managedCryptoComponents.TxSignKeyGen(),
	)
	require.Nil(t, err)
	require.NotNil(t, accountsParser)

	smartContractParser, err := parsing.NewSmartContractsParser(
		genesisSmartContracts,
		managedCoreComponents.AddressPubKeyConverter(),
		managedCryptoComponents.TxSignKeyGen(),
	)
	require.Nil(t, err)
	require.NotNil(t, smartContractParser)

	log.Trace("creating time cache for requested items components")
	requestedItemsHandler := timecache.NewTimeCache(time.Duration(uint64(time.Millisecond) * nodesSetup.GetRoundDuration()))

	whiteListCache, err := storageUnit.NewCache(storageFactory.GetCacherFromConfig(generalConfig.WhiteListPool))
	require.Nil(t, err)
	require.NotNil(t, whiteListCache)

	whiteListRequest, err := interceptors.NewWhiteListDataVerifier(whiteListCache)
	require.Nil(t, err)
	require.NotNil(t, whiteListRequest)

	whiteListerVerifiedTxs, err := createWhiteListerVerifiedTxs(generalConfig)
	require.Nil(t, err)
	require.NotNil(t, whiteListerVerifiedTxs)

	historyRepoFactoryArgs := &dbLookupFactory.ArgsHistoryRepositoryFactory{
		SelfShardID: genesisShardCoordinator.SelfId(),
		Config:      generalConfig.DbLookupExtensions,
		Hasher:      managedCoreComponents.Hasher(),
		Marshalizer: managedCoreComponents.InternalMarshalizer(),
		Store:       managedDataComponents.StorageService(),
	}
	historyRepositoryFactory, err := dbLookupFactory.NewHistoryRepositoryFactory(historyRepoFactoryArgs)
	require.Nil(t, err)
	require.NotNil(t, historyRepositoryFactory)

	historyRepository, err := historyRepositoryFactory.Create()
	require.Nil(t, err)
	require.NotNil(t, historyRepository)

	versionsCache, _ := storageUnit.NewCache(storageFactory.GetCacherFromConfig(generalConfig.Versions.Cache))
	headerIntegrityVerifier, err := headerCheck.NewHeaderIntegrityVerifier(
		[]byte(managedCoreComponents.ChainID()),
		generalConfig.Versions.VersionsByEpochs,
		generalConfig.Versions.DefaultVersion,
		versionsCache,
	)
	require.Nil(t, err)
	require.NotNil(t, headerIntegrityVerifier)
	epochNotifier := forking.NewGenericEpochNotifier()

	importStartHandler, err := trigger.NewImportStartHandler(filepath.Join("workingDir", core.DefaultDBPath), "appVersion")
	require.Nil(t, err)
	require.NotNil(t, importStartHandler)

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
	require.Nil(t, err)
	require.NotNil(t, processComponentsFactory)

	managedProcessComponents, err := factory.NewManagedProcessComponents(processComponentsFactory)
	require.Nil(t, err)
	require.NotNil(t, managedProcessComponents)

	err = managedProcessComponents.Create()
	require.Nil(t, err)

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

func createWhiteListerVerifiedTxs(generalConfig *config.Config) (process.WhiteListHandler, error) {
	whiteListCacheVerified, err := storageUnit.NewCache(storageFactory.GetCacherFromConfig(generalConfig.WhiteListerVerifiedTxs))
	if err != nil {
		return nil, err
	}
	return interceptors.NewWhiteListDataVerifier(whiteListCacheVerified)
}
