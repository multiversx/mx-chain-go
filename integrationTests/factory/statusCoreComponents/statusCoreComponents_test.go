package coreComponents

import (
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/integrationTests/factory"
	"github.com/multiversx/mx-chain-go/node"
	"github.com/multiversx/mx-chain-go/testscommon/goroutines"
	"github.com/stretchr/testify/require"
)

// ------------ Test StatusCoreComponents --------------------
func TestStatusCoreComponents_CreateCloseShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	time.Sleep(time.Second * 4)

	gc := goroutines.NewGoCounter(goroutines.TestsRelevantGoRoutines)
	idxInitial, _ := gc.Snapshot()
	factory.PrintStack()

	configs := factory.CreateDefaultConfig(t)
	chanStopNodeProcess := make(chan endProcess.ArgEndProcess)

	nr, err := node.NewNodeRunner(configs)
	require.Nil(t, err)

	managedCoreComponents, err := nr.CreateManagedCoreComponents(chanStopNodeProcess)
	require.Nil(t, err)
	statusCoreComponents, err := nr.CreateManagedStatusCoreComponents(managedCoreComponents)
	require.Nil(t, err)
	require.NotNil(t, statusCoreComponents)

	time.Sleep(2 * time.Second)

	err = statusCoreComponents.Close()
	require.Nil(t, err)
	err = managedCoreComponents.Close()
	require.Nil(t, err)

	time.Sleep(2 * time.Second)

	idx, _ := gc.Snapshot()
	diff := gc.DiffGoRoutines(idxInitial, idx)
	require.Equal(t, 0, len(diff), fmt.Sprintf("%v", diff))
}
