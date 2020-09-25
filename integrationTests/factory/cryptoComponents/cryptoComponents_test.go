package cryptoComponents

import (
	"runtime"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests/factory"
	"github.com/stretchr/testify/require"
)

// ------------ Test CryptoComponents --------------------
func TestCryptoComponents_Create_Close_ShouldWork(t *testing.T) {
	nrBefore := runtime.NumGoroutine()

	generalConfig, _ := core.LoadMainConfig(factory.ConfigPath)
	systemSCConfig, _ := core.LoadSystemSmartContractsConfig(factory.SystemSCConfigPath)
	ratingsConfig, _ := core.LoadRatingsConfig(factory.RatingsPath)
	economicsConfig, _ := core.LoadEconomicsConfig(factory.EconomicsPath)

	coreComponents, err := factory.CreateCoreComponents(*generalConfig, *ratingsConfig, *economicsConfig)
	require.Nil(t, err)
	require.NotNil(t, coreComponents)

	time.Sleep(2 * time.Second)

	cryptoComponents, err := factory.CreateCryptoComponents(*generalConfig, *systemSCConfig, coreComponents)
	require.Nil(t, err)
	require.NotNil(t, cryptoComponents)

	err = cryptoComponents.Close()
	require.Nil(t, err)

	time.Sleep(2 * time.Second)

	err = coreComponents.Close()
	require.Nil(t, err)

	time.Sleep(2 * time.Second)

	nrAfter := runtime.NumGoroutine()
	if nrBefore != nrAfter {
		factory.PrintStack()
	}

	require.Equal(t, nrBefore, nrAfter)

	factory.CleanupWorkingDir()
}
