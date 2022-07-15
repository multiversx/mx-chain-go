package heartbeat_test

import (
	"testing"

	heartbeatComp "github.com/ElrondNetwork/elrond-go/factory/heartbeat"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	componentsMock "github.com/ElrondNetwork/elrond-go/factory/mock/components"
	"github.com/stretchr/testify/require"
)

// ------------ Test ManagedHeartbeatComponents --------------------
func TestManagedHeartbeatComponents_CreateWithInvalidArgsShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	heartbeatArgs := componentsMock.GetDefaultHeartbeatComponents(shardCoordinator)
	heartbeatArgs.Config.Heartbeat.MaxTimeToWaitBetweenBroadcastsInSec = 0
	heartbeatComponentsFactory, _ := heartbeatComp.NewHeartbeatComponentsFactory(heartbeatArgs)
	managedHeartbeatComponents, err := heartbeatComp.NewManagedHeartbeatComponents(heartbeatComponentsFactory)
	require.NoError(t, err)
	err = managedHeartbeatComponents.Create()
	require.Error(t, err)
	require.Nil(t, managedHeartbeatComponents.Monitor())
}

func TestManagedHeartbeatComponents_CreateShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	heartbeatArgs := componentsMock.GetDefaultHeartbeatComponents(shardCoordinator)
	heartbeatComponentsFactory, _ := heartbeatComp.NewHeartbeatComponentsFactory(heartbeatArgs)
	managedHeartbeatComponents, err := heartbeatComp.NewManagedHeartbeatComponents(heartbeatComponentsFactory)
	require.NoError(t, err)
	require.Nil(t, managedHeartbeatComponents.Monitor())
	require.Nil(t, managedHeartbeatComponents.MessageHandler())
	require.Nil(t, managedHeartbeatComponents.Sender())
	require.Nil(t, managedHeartbeatComponents.Storer())

	err = managedHeartbeatComponents.Create()
	require.NoError(t, err)
	require.NotNil(t, managedHeartbeatComponents.Monitor())
	require.NotNil(t, managedHeartbeatComponents.MessageHandler())
	require.NotNil(t, managedHeartbeatComponents.Sender())
	require.NotNil(t, managedHeartbeatComponents.Storer())
}

func TestManagedHeartbeatComponents_Close(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	heartbeatArgs := componentsMock.GetDefaultHeartbeatComponents(shardCoordinator)
	heartbeatComponentsFactory, _ := heartbeatComp.NewHeartbeatComponentsFactory(heartbeatArgs)
	managedHeartbeatComponents, _ := heartbeatComp.NewManagedHeartbeatComponents(heartbeatComponentsFactory)
	err := managedHeartbeatComponents.Create()
	require.NoError(t, err)

	err = managedHeartbeatComponents.Close()
	require.NoError(t, err)
	require.Nil(t, managedHeartbeatComponents.Monitor())
}
