package statusCore_test

import (
	"testing"

	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/statusCore"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/stretchr/testify/require"
)

func TestNewManagedStatusCoreComponents(t *testing.T) {
	t.Parallel()

	t.Run("nil factory should error", func(t *testing.T) {
		t.Parallel()

		managedStatusCoreComponents, err := statusCore.NewManagedStatusCoreComponents(nil)
		require.Equal(t, errorsMx.ErrNilStatusCoreComponentsFactory, err)
		require.Nil(t, managedStatusCoreComponents)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetStatusCoreArgs(componentsMock.GetDefaultCoreComponents())
		statusCoreComponentsFactory, err := statusCore.NewStatusCoreComponentsFactory(args)
		require.NoError(t, err)
		managedStatusCoreComponents, err := statusCore.NewManagedStatusCoreComponents(statusCoreComponentsFactory)
		require.NoError(t, err)
		require.NotNil(t, managedStatusCoreComponents)
	})
}

func TestManagedStatusCoreComponents_Create(t *testing.T) {
	t.Parallel()

	t.Run("invalid params should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetStatusCoreArgs(componentsMock.GetDefaultCoreComponents())
		args.Config.ResourceStats.RefreshIntervalInSec = 0

		statusCoreComponentsFactory, err := statusCore.NewStatusCoreComponentsFactory(args)
		require.NoError(t, err)
		managedStatusCoreComponents, err := statusCore.NewManagedStatusCoreComponents(statusCoreComponentsFactory)
		require.NoError(t, err)

		err = managedStatusCoreComponents.Create()
		require.Error(t, err)
	})
	t.Run("should work with getters", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetStatusCoreArgs(componentsMock.GetCoreComponents())
		statusCoreComponentsFactory, err := statusCore.NewStatusCoreComponentsFactory(args)
		require.NoError(t, err)
		managedStatusCoreComponents, err := statusCore.NewManagedStatusCoreComponents(statusCoreComponentsFactory)
		require.NoError(t, err)

		require.Nil(t, managedStatusCoreComponents.ResourceMonitor())
		require.Nil(t, managedStatusCoreComponents.NetworkStatistics())
		require.Nil(t, managedStatusCoreComponents.TrieSyncStatistics())
		require.Nil(t, managedStatusCoreComponents.AppStatusHandler())
		require.Nil(t, managedStatusCoreComponents.StatusMetrics())
		require.Nil(t, managedStatusCoreComponents.PersistentStatusHandler())

		err = managedStatusCoreComponents.Create()
		require.NoError(t, err)

		require.NotNil(t, managedStatusCoreComponents.ResourceMonitor())
		require.NotNil(t, managedStatusCoreComponents.NetworkStatistics())
		require.NotNil(t, managedStatusCoreComponents.TrieSyncStatistics())
		require.NotNil(t, managedStatusCoreComponents.AppStatusHandler())
		require.NotNil(t, managedStatusCoreComponents.StatusMetrics())
		require.NotNil(t, managedStatusCoreComponents.PersistentStatusHandler())

		require.Equal(t, factory.StatusCoreComponentsName, managedStatusCoreComponents.String())
	})
}

func TestManagedStatusCoreComponents_CheckSubcomponents(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetStatusCoreArgs(componentsMock.GetCoreComponents())
	statusCoreComponentsFactory, _ := statusCore.NewStatusCoreComponentsFactory(args)
	managedStatusCoreComponents, _ := statusCore.NewManagedStatusCoreComponents(statusCoreComponentsFactory)

	err := managedStatusCoreComponents.CheckSubcomponents()
	require.Equal(t, errorsMx.ErrNilStatusCoreComponents, err)

	err = managedStatusCoreComponents.Create()
	require.NoError(t, err)

	err = managedStatusCoreComponents.CheckSubcomponents()
	require.NoError(t, err)
}

func TestManagedStatusCoreComponents_Close(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetStatusCoreArgs(componentsMock.GetCoreComponents())
	statusCoreComponentsFactory, err := statusCore.NewStatusCoreComponentsFactory(args)
	require.NoError(t, err)
	managedStatusCoreComponents, err := statusCore.NewManagedStatusCoreComponents(statusCoreComponentsFactory)
	require.NoError(t, err)

	err = managedStatusCoreComponents.Close()
	require.NoError(t, err)

	err = managedStatusCoreComponents.Create()
	require.NoError(t, err)

	err = managedStatusCoreComponents.Close()
	require.NoError(t, err)
}

func TestManagedStatusCoreComponents_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	managedStatusCoreComponents, _ := statusCore.NewManagedStatusCoreComponents(nil)
	require.True(t, managedStatusCoreComponents.IsInterfaceNil())

	args := componentsMock.GetStatusCoreArgs(componentsMock.GetCoreComponents())
	statusCoreComponentsFactory, _ := statusCore.NewStatusCoreComponentsFactory(args)
	managedStatusCoreComponents, _ = statusCore.NewManagedStatusCoreComponents(statusCoreComponentsFactory)
	require.False(t, managedStatusCoreComponents.IsInterfaceNil())
}
