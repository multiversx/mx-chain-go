package processComponents

import (
	"runtime"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	factory "github.com/ElrondNetwork/elrond-go/integrationTests/factory"
	"github.com/stretchr/testify/require"
)

// ------------ Test TestProcessComponents --------------------
func TestProcessComponents_Close_ShouldWork(t *testing.T) {
	defer factory.CleanupWorkingDir()
	time.Sleep(time.Second)

	nrBefore := runtime.NumGoroutine()

	generalConfig, _ := core.LoadMainConfig(factory.ConfigPath)
	ratingsConfig, _ := core.LoadRatingsConfig(factory.RatingsPath)
	economicsConfig, _ := core.LoadEconomicsConfig(factory.EconomicsPath)
	prefsConfig, _ := core.LoadPreferencesConfig(factory.PrefsPath)
	p2pConfig, _ := core.LoadP2PConfig(factory.P2pPath)
	externalConfig, _ := core.LoadExternalConfig(factory.ExternalPath)
	systemSCConfig, _ := core.LoadSystemSmartContractsConfig(factory.SystemSCConfigPath)

	managedCoreComponents, err := factory.CreateCoreComponents(*generalConfig, *ratingsConfig, *economicsConfig)
	require.Nil(t, err)
	require.NotNil(t, managedCoreComponents)

	managedCryptoComponents, err := factory.CreateCryptoComponents(*generalConfig, *systemSCConfig, managedCoreComponents)
	require.Nil(t, err)
	require.NotNil(t, managedCryptoComponents)

	managedNetworkComponents, err := factory.CreateNetworkComponents(*generalConfig, *p2pConfig, *ratingsConfig, managedCoreComponents)
	require.Nil(t, err)
	require.NotNil(t, managedNetworkComponents)

	managedBootstrapComponents, err := factory.CreateBootstrapComponents(
		*generalConfig,
		*prefsConfig,
		managedCoreComponents,
		managedCryptoComponents,
		managedNetworkComponents)
	require.Nil(t, err)
	require.NotNil(t, managedBootstrapComponents)

	epochStartNotifier := notifier.NewEpochStartSubscriptionHandler()
	managedDataComponents, err := factory.CreateDataComponents(*generalConfig, epochStartNotifier, managedCoreComponents)
	require.Nil(t, err)
	require.NotNil(t, managedDataComponents)

	managedStateComponents, err := factory.CreateStateComponents(*generalConfig, managedCoreComponents, managedBootstrapComponents)
	require.Nil(t, err)
	require.NotNil(t, managedStateComponents)

	nodesSetup := managedCoreComponents.GenesisNodesSetup()
	chanStopNodeProcess := make(chan endProcess.ArgEndProcess, 1)
	genesisShardCoordinator, nodesCoordinator, _, ratingsData, rater := factory.CreateCoordinators(generalConfig, prefsConfig, ratingsConfig, nodesSetup, epochStartNotifier, chanStopNodeProcess, managedCoreComponents, managedCryptoComponents, managedDataComponents, managedBootstrapComponents)

	managedStatusComponents, err := factory.CreateStatusComponents(
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

	managedProcessComponents, err := factory.CreateProcessComponents(
		generalConfig,
		economicsConfig,
		ratingsConfig,
		systemSCConfig,
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
	require.NotNil(t, managedProcessComponents)

	time.Sleep(5 * time.Second)

	managedStatusComponents.SetForkDetector(managedProcessComponents.ForkDetector())
	err = managedStatusComponents.StartPolling()
	require.Nil(t, err)
	time.Sleep(5 * time.Second)

	err = managedProcessComponents.Close()
	require.Nil(t, err)
	time.Sleep(5 * time.Second)

	err = managedStatusComponents.Close()
	require.Nil(t, err)

	err = managedStateComponents.Close()
	require.Nil(t, err)

	err = managedDataComponents.Close()
	require.Nil(t, err)

	err = managedBootstrapComponents.Close()
	require.Nil(t, err)

	err = managedNetworkComponents.Close()
	require.Nil(t, err)

	err = managedCryptoComponents.Close()
	require.Nil(t, err)

	err = managedCoreComponents.Close()
	require.Nil(t, err)

	time.Sleep(5 * time.Second)

	nrAfter := runtime.NumGoroutine()
	if nrBefore != nrAfter {
		factory.PrintStack()
	}

	require.Equal(t, nrBefore, nrAfter)
}
