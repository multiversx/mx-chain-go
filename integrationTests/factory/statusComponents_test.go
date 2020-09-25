package factory

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	mainFactory "github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/process/rating"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/stretchr/testify/require"
)

// ------------ Test StatusComponents --------------------
func TestStatusComponents_Create_Close_ShouldWork(t *testing.T) {
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

	coreComponents, _ := createCoreComponents(*generalConfig, *ratingsConfig, *economicsConfig)
	cryptoComponents, _ := createCryptoComponents(*generalConfig, *systemSCConfig, coreComponents)
	networkComponents, _ := createNetworkComponents(*generalConfig, *p2pConfig, *ratingsConfig, coreComponents)
	bootstrapComponents, _ := createBootstrapComponents(
		*generalConfig,
		prefsConfig.Preferences,
		coreComponents,
		cryptoComponents,
		networkComponents)
	epochStartNotifier := notifier.NewEpochStartSubscriptionHandler()
	dataComponents, _ := createDataComponents(*generalConfig, *economicsConfig, epochStartNotifier, coreComponents)
	time.Sleep(2 * time.Second)

	nodesSetup := coreComponents.GenesisNodesSetup()
	chanStopNodeProcess := make(chan endProcess.ArgEndProcess, 1)
	genesisShardCoordinator, nodesCoordinator, _, _, _ := createCoordinators(
		generalConfig,
		prefsConfig,
		ratingsConfig,
		nodesSetup,
		epochStartNotifier,
		chanStopNodeProcess,
		coreComponents,
		cryptoComponents,
		dataComponents,
		bootstrapComponents)

	statusComponents, err := createStatusComponents(
		*generalConfig,
		*externalConfig,
		genesisShardCoordinator,
		nodesCoordinator,
		epochStartNotifier,
		coreComponents,
		dataComponents,
		networkComponents)
	require.Nil(t, err)
	require.NotNil(t, statusComponents)

	time.Sleep(5 * time.Second)

	err = statusComponents.Close()
	time.Sleep(5 * time.Second)

	_ = dataComponents.Close()
	_ = bootstrapComponents.Close()
	_ = networkComponents.Close()
	_ = cryptoComponents.Close()
	_ = coreComponents.Close()

	time.Sleep(5 * time.Second)

	nrAfter := runtime.NumGoroutine()
	if nrBefore != nrAfter {
		printStack()
	}

	require.Equal(t, nrBefore, nrAfter)
}

func createCoordinators(
	generalConfig *config.Config,
	prefsConfig *config.Preferences,
	ratingsConfig *config.RatingsConfig,
	nodesSetup mainFactory.NodesSetupHandler,
	epochStartNotifier factory.EpochStartNotifier,
	chanStopNodeProcess chan endProcess.ArgEndProcess,
	coreComponents mainFactory.CoreComponentsHandler,
	cryptoComponents mainFactory.CryptoComponentsHandler,
	dataComponents mainFactory.DataComponentsHandler,
	bootstrapComponents mainFactory.BootstrapComponentsHandler,
) (sharding.Coordinator, sharding.NodesCoordinator, update.Closer, *rating.RatingsData, sharding.PeerAccountListAndRatingHandler) {
	log := logger.GetOrCreate("test")

	genesisShardCoordinator, _, _ := mainFactory.CreateShardCoordinator(
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

	nodesCoordinator, nodesShufflerOut, _ := mainFactory.CreateNodesCoordinator(
		log,
		nodesSetup,
		prefsConfig.Preferences,
		epochStartNotifier,
		cryptoComponents.PublicKey(),
		coreComponents.InternalMarshalizer(),
		coreComponents.Hasher(),
		rater,
		dataComponents.StorageService().GetStorer(dataRetriever.BootstrapUnit),
		nodesShuffler,
		generalConfig.EpochStartConfig,
		genesisShardCoordinator.SelfId(),
		chanStopNodeProcess,
		bootstrapComponents.EpochBootstrapParams(),
		bootstrapComponents.EpochBootstrapParams().Epoch(),
	)
	return genesisShardCoordinator, nodesCoordinator, nodesShufflerOut, ratingsData, rater
}

func createStatusComponents(
	generalConfig config.Config,
	externalConfig config.ExternalConfig,
	shardCoordinator storage.ShardCoordinator,
	nodesCoordinator sharding.NodesCoordinator,
	notifier mainFactory.EpochStartNotifier,
	managedCoreComponents mainFactory.CoreComponentsHandler,
	managedDataComponents mainFactory.DataComponentsHandler,
	managedNetworkComponents mainFactory.NetworkComponentsHandler,
) (mainFactory.StatusComponentsHandler, error) {

	statArgs := mainFactory.StatusComponentsFactoryArgs{
		Config:             generalConfig,
		ExternalConfig:     externalConfig,
		RoundDurationSec:   managedCoreComponents.GenesisNodesSetup().GetRoundDuration() / 1000,
		ElasticOptions:     &indexer.Options{TxIndexingEnabled: true},
		ShardCoordinator:   shardCoordinator,
		NodesCoordinator:   nodesCoordinator,
		EpochStartNotifier: notifier,
		CoreComponents:     managedCoreComponents,
		DataComponents:     managedDataComponents,
		NetworkComponents:  managedNetworkComponents,
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
