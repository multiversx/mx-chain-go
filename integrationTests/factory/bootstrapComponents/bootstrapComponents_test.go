package bootstrapComponents

import (
	"runtime"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/integrationTests/factory"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/stretchr/testify/require"
)

// ------------ Test CryptoComponents --------------------
func TestBootstrapComponents_Create_Close_ShouldWork(t *testing.T) {
	defer factory.CleanupWorkingDir()
	time.Sleep(time.Second)

	nrBefore := runtime.NumGoroutine()
	factory.PrintStack()

	configs := factory.CreateDefaultConfig()
	chanStopNodeProcess := make(chan endProcess.ArgEndProcess)
	managedCoreComponents, err := node.CreateManagedCoreComponents(configs, chanStopNodeProcess)
	require.Nil(t, err)
	managedCryptoComponents, err := node.CreateManagedCryptoComponents(configs, managedCoreComponents)
	require.Nil(t, err)
	managedNetworkComponents, err := node.CreateManagedNetworkComponents(configs, managedCoreComponents)
	require.Nil(t, err)
	managedBootstrapComponents, err := node.CreateManagedBootstrapComponents(configs, managedCoreComponents, managedCryptoComponents, managedNetworkComponents)
	require.Nil(t, err)
	require.NotNil(t, managedBootstrapComponents)

	time.Sleep(5 * time.Second)

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
