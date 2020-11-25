package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/stretchr/testify/require"
)

// ------------ Test ManagedConsensusComponentsFactory --------------------
func TestManagedConsensusComponents_CreateWithInvalidArgs_ShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := getConsensusArgs(shardCoordinator)
	coreComponents := getDefaultCoreComponents()
	args.CoreComponents = coreComponents
	consensusComponentsFactory, _ := factory.NewConsensusComponentsFactory(args)
	managedConsensusComponents, err := factory.NewManagedConsensusComponents(consensusComponentsFactory)
	require.NoError(t, err)

	coreComponents.AppStatusHdl = nil
	err = managedConsensusComponents.Create()
	require.Error(t, err)
	require.NotNil(t, managedConsensusComponents.CheckSubcomponents())
}

func TestManagedConsensusComponents_Create_ShouldWork(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := getConsensusArgs(shardCoordinator)

	consensusComponentsFactory, _ := factory.NewConsensusComponentsFactory(args)
	managedConsensusComponents, err := factory.NewManagedConsensusComponents(consensusComponentsFactory)

	require.NoError(t, err)
	require.Nil(t, managedConsensusComponents.BroadcastMessenger())
	require.Nil(t, managedConsensusComponents.Chronology())
	require.Nil(t, managedConsensusComponents.ConsensusWorker())
	require.Error(t, managedConsensusComponents.CheckSubcomponents())

	err = managedConsensusComponents.Create()
	require.NoError(t, err)
	require.NotNil(t, managedConsensusComponents.BroadcastMessenger())
	require.NotNil(t, managedConsensusComponents.Chronology())
	require.NotNil(t, managedConsensusComponents.ConsensusWorker())
	require.NoError(t, managedConsensusComponents.CheckSubcomponents())
}

func TestManagedConsensusComponents_Close(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	consensusArgs := getConsensusArgs(shardCoordinator)
	consensusComponentsFactory, _ := factory.NewConsensusComponentsFactory(consensusArgs)
	managedConsensusComponents, _ := factory.NewManagedConsensusComponents(consensusComponentsFactory)
	err := managedConsensusComponents.Create()
	require.NoError(t, err)

	err = managedConsensusComponents.Close()
	require.NoError(t, err)
	require.Nil(t, managedConsensusComponents.BroadcastMessenger())
}
