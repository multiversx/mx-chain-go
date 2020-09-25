package statusComponents

import (
	"runtime"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/integrationTests/factory"
	"github.com/stretchr/testify/require"
)

// ------------ Test StatusComponents --------------------
func TestStatusComponents_Create_Close_ShouldWork(t *testing.T) {
	nrBefore := runtime.NumGoroutine()

	generalConfig, _ := core.LoadMainConfig(factory.ConfigPath)
	ratingsConfig, _ := core.LoadRatingsConfig(factory.RatingsPath)
	economicsConfig, _ := core.LoadEconomicsConfig(factory.EconomicsPath)
	prefsConfig, _ := core.LoadPreferencesConfig(factory.PrefsPath)
	p2pConfig, _ := core.LoadP2PConfig(factory.P2pPath)
	externalConfig, _ := core.LoadExternalConfig(factory.ExternalPath)
	systemSCConfig, _ := core.LoadSystemSmartContractsConfig(factory.SystemSCConfigPath)

	coreComponents, err := factory.CreateCoreComponents(*generalConfig, *ratingsConfig, *economicsConfig)
	require.Nil(t, err)
	require.NotNil(t, coreComponents)

	cryptoComponents, err := factory.CreateCryptoComponents(*generalConfig, *systemSCConfig, coreComponents)
	require.Nil(t, err)
	require.NotNil(t, cryptoComponents)

	networkComponents, err := factory.CreateNetworkComponents(*generalConfig, *p2pConfig, *ratingsConfig, coreComponents)
	require.Nil(t, err)
	require.NotNil(t, networkComponents)

	bootstrapComponents, err := factory.CreateBootstrapComponents(
		*generalConfig,
		prefsConfig.Preferences,
		coreComponents,
		cryptoComponents,
		networkComponents)
	require.Nil(t, err)
	require.NotNil(t, bootstrapComponents)

	epochStartNotifier := notifier.NewEpochStartSubscriptionHandler()
	dataComponents, err := factory.CreateDataComponents(*generalConfig, *economicsConfig, epochStartNotifier, coreComponents)
	require.Nil(t, err)
	require.NotNil(t, dataComponents)
	time.Sleep(2 * time.Second)

	nodesSetup := coreComponents.GenesisNodesSetup()
	chanStopNodeProcess := make(chan endProcess.ArgEndProcess, 1)
	genesisShardCoordinator, nodesCoordinator, _, _, _ := factory.CreateCoordinators(
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

	statusComponents, err := factory.CreateStatusComponents(
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
	require.Nil(t, err)
	time.Sleep(5 * time.Second)

	err = dataComponents.Close()
	require.Nil(t, err)

	err = bootstrapComponents.Close()
	require.Nil(t, err)

	err = networkComponents.Close()
	require.Nil(t, err)

	err = cryptoComponents.Close()
	require.Nil(t, err)

	err = coreComponents.Close()
	require.Nil(t, err)

	time.Sleep(5 * time.Second)

	nrAfter := runtime.NumGoroutine()
	if nrBefore != nrAfter {
		factory.PrintStack()
	}

	require.Equal(t, nrBefore, nrAfter)

	factory.CleanupWorkingDir()
}
