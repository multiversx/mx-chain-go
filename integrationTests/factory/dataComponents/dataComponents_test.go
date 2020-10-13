package dataComponents

import (
	"runtime"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/integrationTests/factory"
	"github.com/stretchr/testify/require"
)

func TestDataComponents_Create_Close_ShouldWork(t *testing.T) {
	defer factory.CleanupWorkingDir()
	time.Sleep(time.Second)

	nrBefore := runtime.NumGoroutine()

	factory.PrintStack()

	generalConfig, _ := core.LoadMainConfig(factory.ConfigPath)
	ratingsConfig, _ := core.LoadRatingsConfig(factory.RatingsPath)
	economicsConfig, _ := core.LoadEconomicsConfig(factory.EconomicsPath)

	time.Sleep(2 * time.Second)

	coreComponents, err := factory.CreateCoreComponents(*generalConfig, *ratingsConfig, *economicsConfig)
	require.Nil(t, err)
	require.NotNil(t, coreComponents)

	epochStartNotifier := notifier.NewEpochStartSubscriptionHandler()
	dataComponents, err := factory.CreateDataComponents(*generalConfig, epochStartNotifier, coreComponents)
	require.Nil(t, err)
	require.NotNil(t, dataComponents)
	time.Sleep(2 * time.Second)

	err = dataComponents.Close()
	require.Nil(t, err)

	err = coreComponents.Close()
	require.Nil(t, err)

	time.Sleep(2 * time.Second)

	nrAfter := runtime.NumGoroutine()
	if nrBefore != nrAfter {
		factory.PrintStack()
	}

	require.Equal(t, nrBefore, nrAfter)
}
