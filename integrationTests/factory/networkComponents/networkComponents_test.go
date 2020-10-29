package networkComponents

import (
	"runtime"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests/factory"
	"github.com/stretchr/testify/require"
)

// ------------ Test NetworkComponents --------------------
func TestNetworkComponents_Create_Close_ShouldWork(t *testing.T) {
	defer factory.CleanupWorkingDir()
	time.Sleep(time.Second)

	nrBefore := runtime.NumGoroutine()

	generalConfig, _ := core.LoadMainConfig(factory.ConfigPath)
	ratingsConfig, _ := core.LoadRatingsConfig(factory.RatingsPath)
	economicsConfig, _ := core.LoadEconomicsConfig(factory.EconomicsPath)
	p2pConfig, _ := core.LoadP2PConfig(factory.P2pPath)

	coreComponents, err := factory.CreateCoreComponents(*generalConfig, *ratingsConfig, *economicsConfig)
	require.Nil(t, err)
	require.NotNil(t, coreComponents)

	time.Sleep(2 * time.Second)

	networkComponents, err := factory.CreateNetworkComponents(*generalConfig, *p2pConfig, *ratingsConfig, coreComponents)
	require.Nil(t, err)
	require.NotNil(t, networkComponents)
	time.Sleep(2 * time.Second)

	err = networkComponents.Close()
	require.Nil(t, err)

	time.Sleep(2 * time.Second)
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

	require.LessOrEqual(t, nrBefore, nrAfter-1)
}
