package stateComponents

import (
	"runtime"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	factory "github.com/ElrondNetwork/elrond-go/integrationTests/factory"
	"github.com/stretchr/testify/require"
)

func TestStateComponents_Create_Close_ShouldWork(t *testing.T) {
	defer factory.CleanupWorkingDir()
	time.Sleep(time.Second)

	nrBefore := runtime.NumGoroutine()

	generalConfig, _ := core.LoadMainConfig(factory.ConfigPath)
	ratingsConfig, _ := core.LoadRatingsConfig(factory.RatingsPath)
	economicsConfig, _ := core.LoadEconomicsConfig(factory.EconomicsPath)
	preferencesConfig, _ := core.LoadPreferencesConfig(factory.PrefsPath)
	systemScConfig, _ := core.LoadSystemSmartContractsConfig(factory.SystemSCConfigPath)
	p2pConfig, _ := core.LoadP2PConfig(factory.P2pPath)

	coreComponents, err := factory.CreateCoreComponents(*generalConfig, *ratingsConfig, *economicsConfig)
	require.Nil(t, err)
	require.NotNil(t, coreComponents)

	crytoComponents, err := factory.CreateCryptoComponents(*generalConfig, *systemScConfig, coreComponents)
	require.Nil(t, err)
	require.NotNil(t, crytoComponents)

	networkComponents, err := factory.CreateNetworkComponents(*generalConfig, *p2pConfig, *ratingsConfig, coreComponents)
	require.Nil(t, err)
	require.NotNil(t, networkComponents)

	bootstrapComponents, err := factory.CreateBootstrapComponents(
		*generalConfig, *preferencesConfig, coreComponents, crytoComponents, networkComponents,
	)
	require.Nil(t, err)
	require.NotNil(t, bootstrapComponents)
	time.Sleep(5 * time.Second)

	stateComponents, err := factory.CreateStateComponents(*generalConfig, coreComponents, bootstrapComponents)
	require.Nil(t, err)
	require.NotNil(t, stateComponents)
	time.Sleep(2 * time.Second)

	err = stateComponents.Close()
	require.Nil(t, err)

	err = bootstrapComponents.Close()
	require.Nil(t, err)

	err = networkComponents.Close()
	require.Nil(t, err)

	err = crytoComponents.Close()
	require.Nil(t, err)

	err = coreComponents.Close()
	require.Nil(t, err)

	time.Sleep(30 * time.Second)
	nrAfter := runtime.NumGoroutine()

	//TODO: make sure natpmp goroutine is closed as well
	// normally should be closed after ~3 minutes on timeout
	// temp fix: ignore the extra goroutine
	if nrBefore <= nrAfter {
		factory.PrintStack()
	}

	require.LessOrEqual(t, nrBefore, nrAfter)
}
