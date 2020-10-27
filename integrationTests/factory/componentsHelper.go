package factory

import (
	"bytes"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"

	"github.com/ElrondNetwork/elrond-go-logger"
	nodeFactory "github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	factory2 "github.com/ElrondNetwork/elrond-go/core/dblookupext/factory"
	"github.com/ElrondNetwork/elrond-go/core/forking"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/genesis/parsing"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/headerCheck"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/rating"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	factory3 "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/trigger"
)

// PrintStack -
func PrintStack() {

	buffer := new(bytes.Buffer)
	err := pprof.Lookup("goroutine").WriteTo(buffer, 2)
	if err != nil {
		fmt.Println("could not dump goroutines")
	}

	fmt.Printf("\n%s", buffer.String())
}

// CreateWhiteListerVerifiedTxs -
func CreateWhiteListerVerifiedTxs(generalConfig *config.Config) (process.WhiteListHandler, error) {
	whiteListCacheVerified, err := storageUnit.NewCache(factory3.GetCacherFromConfig(generalConfig.WhiteListerVerifiedTxs))
	if err != nil {
		return nil, err
	}
	return interceptors.NewWhiteListDataVerifier(whiteListCacheVerified)
}

// CreateCoreComponents -
func CreateCoreComponents(
	generalConfig config.Config,
	ratingsConfig config.RatingsConfig,
	economicsConfig config.EconomicsConfig) (factory.CoreComponentsHandler, error) {

	statusHandlersFactoryArgs := &nodeFactory.StatusHandlersFactoryArgs{
		UseTermUI: false,
	}

	statusHandlersFactory, err := nodeFactory.NewStatusHandlersFactory(statusHandlersFactoryArgs)
	if err != nil {
		return nil, err
	}

	chanStopNodeProcess := make(chan endProcess.ArgEndProcess, 1)

	coreArgs := factory.CoreComponentsFactoryArgs{
		Config:                generalConfig,
		RatingsConfig:         ratingsConfig,
		EconomicsConfig:       economicsConfig,
		NodesFilename:         NodesSetupPath,
		WorkingDirectory:      "workingDir",
		ChanStopNodeProcess:   chanStopNodeProcess,
		StatusHandlersFactory: statusHandlersFactory,
	}

	coreComponentsFactory, err := factory.NewCoreComponentsFactory(coreArgs)
	if err != nil {
		return nil, fmt.Errorf("NewCoreComponentsFactory failed: %w", err)
	}

	managedCoreComponents, err := factory.NewManagedCoreComponents(coreComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedCoreComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedCoreComponents, nil
}

// CreateCryptoComponents -
func CreateCryptoComponents(
	generalConfig config.Config,
	systemSCConfig config.SystemSmartContractsConfig,
	managedCoreComponents factory.CoreComponentsHandler) (factory.CryptoComponentsHandler, error) {
	cryptoComponentsHandlerArgs := factory.CryptoComponentsFactoryArgs{
		ValidatorKeyPemFileName:              "../validatorKey.pem",
		SkIndex:                              0,
		Config:                               generalConfig,
		CoreComponentsHolder:                 managedCoreComponents,
		ActivateBLSPubKeyMessageVerification: systemSCConfig.StakingSystemSCConfig.ActivateBLSPubKeyMessageVerification,
		KeyLoader:                            &core.KeyLoader{},
	}

	cryptoComponentsFactory, err := factory.NewCryptoComponentsFactory(cryptoComponentsHandlerArgs)
	if err != nil {
		return nil, fmt.Errorf("NewCryptoComponentsFactory failed: %w", err)
	}

	managedCryptoComponents, err := factory.NewManagedCryptoComponents(cryptoComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedCryptoComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedCryptoComponents, nil
}

// CreateNetworkComponents -
func CreateNetworkComponents(
	config config.Config,
	p2pConfig config.P2PConfig,
	ratingsConfig config.RatingsConfig,
	managedCoreComponents factory.CoreComponentsHandler) (factory.NetworkComponentsHandler, error) {
	networkComponentsFactoryArgs := factory.NetworkComponentsFactoryArgs{
		P2pConfig:     p2pConfig,
		MainConfig:    config,
		RatingsConfig: ratingsConfig,
		StatusHandler: managedCoreComponents.StatusHandler(),
		Marshalizer:   managedCoreComponents.InternalMarshalizer(),
		Syncer:        managedCoreComponents.SyncTimer(),
	}

	networkComponentsFactory, err := factory.NewNetworkComponentsFactory(networkComponentsFactoryArgs)
	if err != nil {
		return nil, fmt.Errorf("NewNetworkComponentsFactory failed: %w", err)
	}

	managedNetworkComponents, err := factory.NewManagedNetworkComponents(networkComponentsFactory)
	if err != nil {
		return nil, err
	}
	err = managedNetworkComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedNetworkComponents, nil
}

// CreateBootstrapComponents -
func CreateBootstrapComponents(
	config config.Config,
	preferences config.Preferences,
	managedCoreComponents factory.CoreComponentsHandler,
	managedCryptoComponents factory.CryptoComponentsHandler,
	managedNetworkComponents factory.NetworkComponentsHandler,
) (factory.BootstrapComponentsHandler, error) {
	bootstrapComponentsFactoryArgs := factory.BootstrapComponentsFactoryArgs{
		Config:            config,
		PrefConfig:        preferences,
		WorkingDir:        "workingDir",
		CoreComponents:    managedCoreComponents,
		CryptoComponents:  managedCryptoComponents,
		NetworkComponents: managedNetworkComponents,
	}

	bootstrapComponentsFactory, err := factory.NewBootstrapComponentsFactory(bootstrapComponentsFactoryArgs)
	if err != nil {
		return nil, fmt.Errorf("NewBootstrapComponentsFactory failed: %w", err)
	}

	managedBootstrapComponents, err := factory.NewManagedBootstrapComponents(bootstrapComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedBootstrapComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedBootstrapComponents, nil
}

// CreateDataComponents -
func CreateDataComponents(
	genConfig config.Config,
	epochStartNotifier factory.EpochStartNotifier,
	coreComponents factory.CoreComponentsHolder,
) (factory.DataComponentsHandler, error) {
	currentEpoch := uint32(0)
	nbShards := uint32(3)
	selfShardID := uint32(0)
	shardCoordinator, err := sharding.NewMultiShardCoordinator(nbShards, selfShardID)
	if err != nil {
		return nil, err
	}

	dataArgs := factory.DataComponentsFactoryArgs{
		Config:             genConfig,
		ShardCoordinator:   shardCoordinator,
		Core:               coreComponents,
		EpochStartNotifier: epochStartNotifier,
		CurrentEpoch:       currentEpoch,
	}

	dataComponentsFactory, err := factory.NewDataComponentsFactory(dataArgs)
	if err != nil {
		return nil, fmt.Errorf("NewDataComponentsFactory failed: %w", err)
	}
	managedDataComponents, err := factory.NewManagedDataComponents(dataComponentsFactory)
	if err != nil {
		return nil, err
	}
	err = managedDataComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedDataComponents, nil
}

// CreateStatusComponents -
func CreateStatusComponents(
	generalConfig config.Config,
	externalConfig config.ExternalConfig,
	shardCoordinator storage.ShardCoordinator,
	nodesCoordinator sharding.NodesCoordinator,
	notifier factory.EpochStartNotifier,
	managedCoreComponents factory.CoreComponentsHandler,
	managedDataComponents factory.DataComponentsHandler,
	managedNetworkComponents factory.NetworkComponentsHandler,
) (factory.StatusComponentsHandler, error) {

	statArgs := factory.StatusComponentsFactoryArgs{
		Config:             generalConfig,
		ExternalConfig:     externalConfig,
		ElasticOptions:     &indexer.Options{TxIndexingEnabled: true},
		ShardCoordinator:   shardCoordinator,
		NodesCoordinator:   nodesCoordinator,
		EpochStartNotifier: notifier,
		CoreComponents:     managedCoreComponents,
		DataComponents:     managedDataComponents,
		NetworkComponents:  managedNetworkComponents,
	}

	statusComponentsFactory, err := factory.NewStatusComponentsFactory(statArgs)
	if err != nil {
		return nil, fmt.Errorf("NewStatusComponentsFactory failed: %w", err)
	}

	managedStatusComponents, err := factory.NewManagedStatusComponents(statusComponentsFactory)
	if err != nil {
		return nil, err
	}
	err = managedStatusComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedStatusComponents, nil
}

// CreateStateComponents -
func CreateStateComponents(
	genConfig config.Config,
	coreComponents factory.CoreComponentsHolder,
	bootstrapComponents factory.BootstrapComponentsHolder,
) (factory.StateComponentsHandler, error) {
	nbShards := uint32(3)
	selfShardID := uint32(0)
	shardCoordinator, err := sharding.NewMultiShardCoordinator(nbShards, selfShardID)
	if err != nil {
		return nil, err
	}

	triesComponents, trieStorageManagers := bootstrapComponents.EpochStartBootstrapper().GetTriesComponents()
	stateArgs := factory.StateComponentsFactoryArgs{
		Config:              genConfig,
		ShardCoordinator:    shardCoordinator,
		Core:                coreComponents,
		TriesContainer:      triesComponents,
		TrieStorageManagers: trieStorageManagers,
	}

	stateComponentsFactory, err := factory.NewStateComponentsFactory(stateArgs)
	if err != nil {
		return nil, fmt.Errorf("NewStateComponentsFactory failed: %w", err)
	}

	managedStateComponents, err := factory.NewManagedStateComponents(stateComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedStateComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedStateComponents, nil
}

// CreateCoordinators -
func CreateCoordinators(
	generalConfig *config.Config,
	prefsConfig *config.Preferences,
	ratingsConfig *config.RatingsConfig,
	nodesSetup sharding.GenesisNodesSetupHandler,
	epochStartNotifier factory.EpochStartNotifier,
	chanStopNodeProcess chan endProcess.ArgEndProcess,
	coreComponents factory.CoreComponentsHandler,
	cryptoComponents factory.CryptoComponentsHandler,
	dataComponents factory.DataComponentsHandler,
	bootstrapComponents factory.BootstrapComponentsHandler,
) (sharding.Coordinator, sharding.NodesCoordinator, update.Closer, *rating.RatingsData, sharding.PeerAccountListAndRatingHandler) {
	log := logger.GetOrCreate("test")

	genesisShardCoordinator, _, _ := factory.CreateShardCoordinator(
		nodesSetup,
		cryptoComponents.PublicKey(),
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

	nodesShufflerOut, _ := factory.CreateNodesShuffleOut(nodesSetup, generalConfig.EpochStartConfig, chanStopNodeProcess)

	nodesCoordinator, _ := factory.CreateNodesCoordinator(
		nodesShufflerOut,
		nodesSetup,
		prefsConfig.Preferences,
		epochStartNotifier,
		cryptoComponents.PublicKey(),
		coreComponents.InternalMarshalizer(),
		coreComponents.Hasher(),
		rater,
		dataComponents.StorageService().GetStorer(dataRetriever.BootstrapUnit),
		nodesShuffler,
		genesisShardCoordinator.SelfId(),
		bootstrapComponents.EpochBootstrapParams(),
		bootstrapComponents.EpochBootstrapParams().Epoch(),
	)
	return genesisShardCoordinator, nodesCoordinator, nodesShufflerOut, ratingsData, rater
}

// CreateProcessComponents -
func CreateProcessComponents(
	generalConfig *config.Config,
	economicsConfig *config.EconomicsConfig,
	ratingsConfig *config.RatingsConfig,
	systemSCConfig *config.SystemSmartContractsConfig,
	nodesCoordinator sharding.NodesCoordinator,
	epochStartNotifier factory.EpochStartNotifier,
	genesisShardCoordinator sharding.Coordinator,
	ratingsData *rating.RatingsData,
	rater sharding.PeerAccountListAndRatingHandler,
	managedCoreComponents factory.CoreComponentsHandler,
	managedCryptoComponents factory.CryptoComponentsHandler,
	managedDataComponents factory.DataComponentsHandler,
	managedStateComponents factory.StateComponentsHandler,
	managedNetworkComponents factory.NetworkComponentsHandler,
	managedBootstrapComponents factory.BootstrapComponentsHandler,
	managedStatusComponents factory.StatusComponentsHandler,
	chanStopNodeProcess chan endProcess.ArgEndProcess,
) (factory.ProcessComponentsHandler, error) {
	economicsData, err := economics.NewEconomicsData(economicsConfig)
	totalSupply, _ := big.NewInt(0).SetString(economicsConfig.GlobalSettings.GenesisTotalSupply, 10)
	fmt.Println(os.Getwd())
	gasSchedule, err := core.LoadGasScheduleConfig(GasSchedule)
	if err != nil {
		return nil, err
	}

	accountsParser, err := parsing.NewAccountsParser(
		GenesisPath,
		totalSupply,
		managedCoreComponents.AddressPubKeyConverter(),
		managedCryptoComponents.TxSignKeyGen(),
	)
	if err != nil {
		return nil, err
	}

	smartContractParser, err := parsing.NewSmartContractsParser(
		GenesisSmartContracts,
		managedCoreComponents.AddressPubKeyConverter(),
		managedCryptoComponents.TxSignKeyGen(),
	)
	if err != nil {
		return nil, err
	}

	requestedItemsHandler := timecache.NewTimeCache(time.Duration(uint64(time.Millisecond) * managedCoreComponents.GenesisNodesSetup().GetRoundDuration()))

	whiteListCache, err := storageUnit.NewCache(factory3.GetCacherFromConfig(generalConfig.WhiteListPool))
	if err != nil {
		return nil, err
	}

	whiteListRequest, err := interceptors.NewWhiteListDataVerifier(whiteListCache)
	if err != nil {
		return nil, err
	}

	whiteListerVerifiedTxs, err := CreateWhiteListerVerifiedTxs(generalConfig)
	if err != nil {
		return nil, err
	}

	importStartHandler, err := trigger.NewImportStartHandler(filepath.Join("workingDir", core.DefaultDBPath), "appVersion")
	if err != nil {
		return nil, err
	}

	historyRepoFactoryArgs := &factory2.ArgsHistoryRepositoryFactory{
		SelfShardID: genesisShardCoordinator.SelfId(),
		Config:      generalConfig.DbLookupExtensions,
		Hasher:      managedCoreComponents.Hasher(),
		Marshalizer: managedCoreComponents.InternalMarshalizer(),
		Store:       managedDataComponents.StorageService(),
	}
	historyRepositoryFactory, err := factory2.NewHistoryRepositoryFactory(historyRepoFactoryArgs)
	if err != nil {
		return nil, err
	}

	historyRepository, err := historyRepositoryFactory.Create()
	if err != nil {
		return nil, err
	}

	versionsCache, _ := storageUnit.NewCache(factory3.GetCacherFromConfig(generalConfig.Versions.Cache))
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

// CleanupWorkingDir -
func CleanupWorkingDir() {
	workingDir := "workingDir"
	if _, err := os.Stat(workingDir); !os.IsNotExist(err) {
		err = os.RemoveAll(workingDir)
		if err != nil {
			fmt.Println("CleanupWorkingDir failed:" + err.Error())
		}
	}
}
