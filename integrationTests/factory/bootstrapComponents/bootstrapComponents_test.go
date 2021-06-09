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

// ------------ Test BootstrapComponents --------------------
func TestBootstrapComponents_Create_Close_ShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	defer factory.CleanupWorkingDir()
	time.Sleep(time.Second * 4)

	nrBefore := runtime.NumGoroutine()
	factory.PrintStack()

	configs := factory.CreateDefaultConfig()
	chanStopNodeProcess := make(chan endProcess.ArgEndProcess)
	nr, err := node.NewNodeRunner(configs)
	require.Nil(t, err)
	managedCoreComponents, err := nr.CreateManagedCoreComponents(chanStopNodeProcess)
	require.Nil(t, err)
	managedCryptoComponents, err := nr.CreateManagedCryptoComponents(managedCoreComponents)
	require.Nil(t, err)
	managedNetworkComponents, err := nr.CreateManagedNetworkComponents(managedCoreComponents)
	require.Nil(t, err)
	managedBootstrapComponents, err := nr.CreateManagedBootstrapComponents(managedCoreComponents, managedCryptoComponents, managedNetworkComponents)
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
	// TODO: find a clean solution
	// On the tests using managed network components, depending on the NAT config, there
	// might be one go routine hanging for up to 3 minutes
	if !(nrBefore == nrAfter || nrBefore == nrAfter-1) {
		factory.PrintStack()
	}

	require.True(t, nrBefore == nrAfter || nrBefore == nrAfter-1)
}
