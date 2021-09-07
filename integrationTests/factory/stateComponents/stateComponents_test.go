package stateComponents

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	factory "github.com/ElrondNetwork/elrond-go/integrationTests/factory"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/testscommon/goroutines"
	"github.com/stretchr/testify/require"
)

func TestStateComponents_Create_Close_ShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	defer factory.CleanupWorkingDir()
	time.Sleep(time.Second * 4)

	gc := goroutines.NewGoCounter(goroutines.TestsRelevantGoRoutines)
	idxInitial, _ := gc.Snapshot()
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
	managedDataComponents, err := nr.CreateManagedDataComponents(managedCoreComponents, managedBootstrapComponents)
	require.Nil(t, err)
	managedStateComponents, err := nr.CreateManagedStateComponents(managedCoreComponents, managedBootstrapComponents)
	require.Nil(t, err)
	require.NotNil(t, managedStateComponents)

	time.Sleep(5 * time.Second)

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

	idx, _ := gc.Snapshot()
	diff := gc.DiffGoRoutines(idxInitial, idx)
	require.Equal(t, 0, len(diff), fmt.Sprintf("%v", diff))
}
