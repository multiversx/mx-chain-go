package statusCore_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/factory/statusCore"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/stretchr/testify/require"
)

func TestManagedStatusCoreComponents_CreateWithInvalidArgsShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetStatusCoreArgs(componentsMock.GetDefaultCoreComponents())
	args.Config.ResourceStats.RefreshIntervalInSec = 0

	statusCoreComponentsFactory, err := statusCore.NewStatusCoreComponentsFactory(args)
	require.NoError(t, err)
	managedStatusCoreComponents, err := statusCore.NewManagedStatusCoreComponents(statusCoreComponentsFactory)
	require.NoError(t, err)

	err = managedStatusCoreComponents.Create()
	require.Error(t, err)
	require.Nil(t, managedStatusCoreComponents.ResourceMonitor())
}

func TestManagedStatusCoreComponents_CreateShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetStatusCoreArgs(componentsMock.GetCoreComponents())
	statusCoreComponentsFactory, err := statusCore.NewStatusCoreComponentsFactory(args)
	require.NoError(t, err)
	managedStatusCoreComponents, err := statusCore.NewManagedStatusCoreComponents(statusCoreComponentsFactory)
	require.NoError(t, err)

	require.Nil(t, managedStatusCoreComponents.ResourceMonitor())
	require.Nil(t, managedStatusCoreComponents.NetworkStatistics())

	err = managedStatusCoreComponents.Create()
	require.NoError(t, err)

	require.NotNil(t, managedStatusCoreComponents.ResourceMonitor())
	require.NotNil(t, managedStatusCoreComponents.NetworkStatistics())
}

func TestManagedCoreComponents_Close(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetStatusCoreArgs(componentsMock.GetCoreComponents())
	statusCoreComponentsFactory, err := statusCore.NewStatusCoreComponentsFactory(args)
	require.NoError(t, err)
	managedStatusCoreComponents, err := statusCore.NewManagedStatusCoreComponents(statusCoreComponentsFactory)
	require.NoError(t, err)

	err = managedStatusCoreComponents.Create()
	require.NoError(t, err)

	err = managedStatusCoreComponents.Close()
	require.NoError(t, err)
	require.Nil(t, managedStatusCoreComponents.ResourceMonitor())
}
