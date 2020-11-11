package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/stretchr/testify/require"
)

// ------------ Test ManagedHeartbeatComponents --------------------
func TestManagedHeartbeatComponents_CreateWithInvalidArgs_ShouldErr(t *testing.T) {
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	heartbeatArgs := getDefaultHeartbeatComponents(shardCoordinator)
	heartbeatArgs.Config.Heartbeat.MaxTimeToWaitBetweenBroadcastsInSec = 0
	heartbeatComponentsFactory, _ := factory.NewHeartbeatComponentsFactory(heartbeatArgs)
	managedHeartbeatComponents, err := factory.NewManagedHeartbeatComponents(heartbeatComponentsFactory)
	require.NoError(t, err)
	err = managedHeartbeatComponents.Create()
	require.Error(t, err)
	require.Nil(t, managedHeartbeatComponents.Monitor())
}

func TestManagedHeartbeatComponents_Create_ShouldWork(t *testing.T) {
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	heartbeatArgs := getDefaultHeartbeatComponents(shardCoordinator)
	heartbeatComponentsFactory, _ := factory.NewHeartbeatComponentsFactory(heartbeatArgs)
	managedHeartbeatComponents, err := factory.NewManagedHeartbeatComponents(heartbeatComponentsFactory)
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
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	heartbeatArgs := getDefaultHeartbeatComponents(shardCoordinator)
	heartbeatComponentsFactory, _ := factory.NewHeartbeatComponentsFactory(heartbeatArgs)
	managedHeartbeatComponents, _ := factory.NewManagedHeartbeatComponents(heartbeatComponentsFactory)
	err := managedHeartbeatComponents.Create()
	require.NoError(t, err)

	err = managedHeartbeatComponents.Close()
	require.NoError(t, err)
	require.Nil(t, managedHeartbeatComponents.Monitor())
}
