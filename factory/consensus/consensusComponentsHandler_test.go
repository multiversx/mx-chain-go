package consensus_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/errors"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	mxFactory "github.com/multiversx/mx-chain-go/factory"

	consensusComp "github.com/multiversx/mx-chain-go/factory/consensus"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	factoryMocks "github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ------------ Test ManagedConsensusComponentsFactory --------------------
func TestManagedConsensusComponents_CreateWithInvalidArgsShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	statusCoreComponents := &factoryMocks.StatusCoreComponentsStub{
		AppStatusHandlerField: &statusHandler.AppStatusHandlerStub{},
	}
	args.StatusCoreComponents = statusCoreComponents
	consensusComponentsFactory, _ := consensusComp.NewConsensusComponentsFactory(args)
	managedConsensusComponents, err := consensusComp.NewManagedConsensusComponents(consensusComponentsFactory)
	require.NoError(t, err)

	statusCoreComponents.AppStatusHandlerField = nil
	err = managedConsensusComponents.Create()
	require.Error(t, err)
	require.NotNil(t, managedConsensusComponents.CheckSubcomponents())
}

func TestManagedConsensusComponents_CreateShouldWorkForShard(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)

	consensusComponentsFactory, _ := consensusComp.NewConsensusComponentsFactory(args)
	managedConsensusComponents, err := consensusComp.NewManagedConsensusComponents(consensusComponentsFactory)

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
	assert.Equal(t, "*sync.ShardBootstrap", fmt.Sprintf("%T", managedConsensusComponents.Bootstrapper()))
}

func TestNewManagedConsensusComponents(t *testing.T) {
	t.Parallel()

	t.Run("nil factory should error", func(t *testing.T) {
		t.Parallel()

		managedConsensusComponents, err := consensusComp.NewManagedConsensusComponents(nil)
		require.Equal(t, errorsMx.ErrNilConsensusComponentsFactory, err)
		require.Nil(t, managedConsensusComponents)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		consensusComponentsFactory, _ := consensusComp.NewConsensusComponentsFactory(createMockConsensusComponentsFactoryArgs())
		managedConsensusComponents, err := consensusComp.NewManagedConsensusComponents(consensusComponentsFactory)
		require.NoError(t, err)
		require.NotNil(t, managedConsensusComponents)
	})
}

func TestManagedConsensusComponents_Create(t *testing.T) {
	t.Parallel()

	t.Run("invalid params should error", func(t *testing.T) {
		t.Parallel()

		args := createMockConsensusComponentsFactoryArgs()
		statusCoreCompStub, ok := args.StatusCoreComponents.(*factoryMocks.StatusCoreComponentsStub)
		require.True(t, ok)
		statusCoreCompStub.AppStatusHandlerField = nil
		consensusComponentsFactory, _ := consensusComp.NewConsensusComponentsFactory(args)
		managedConsensusComponents, _ := consensusComp.NewManagedConsensusComponents(consensusComponentsFactory)
		require.NotNil(t, managedConsensusComponents)

		err := managedConsensusComponents.Create()
		require.Error(t, err)
	})
	t.Run("should work with getters", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := testscommon.NewMultiShardsCoordinatorMock(2)
		args := componentsMock.GetConsensusArgs(shardCoordinator)
		consensusComponentsFactory, _ := consensusComp.NewConsensusComponentsFactory(args)
		managedConsensusComponents, _ := consensusComp.NewManagedConsensusComponents(consensusComponentsFactory)
		require.NotNil(t, managedConsensusComponents)

		require.Nil(t, managedConsensusComponents.BroadcastMessenger())
		require.Nil(t, managedConsensusComponents.Chronology())
		require.Nil(t, managedConsensusComponents.ConsensusWorker())
		require.Nil(t, managedConsensusComponents.Bootstrapper())

		err := managedConsensusComponents.Create()
		require.NoError(t, err)
		require.NotNil(t, managedConsensusComponents.BroadcastMessenger())
		require.NotNil(t, managedConsensusComponents.Chronology())
		require.NotNil(t, managedConsensusComponents.ConsensusWorker())
		require.NotNil(t, managedConsensusComponents.Bootstrapper())

		require.Equal(t, mxFactory.ConsensusComponentsName, managedConsensusComponents.String())
	})
}

func TestManagedConsensusComponents_ConsensusGroupSize(t *testing.T) {
	t.Parallel()

	consensusComponentsFactory, _ := consensusComp.NewConsensusComponentsFactory(createMockConsensusComponentsFactoryArgs())
	managedConsensusComponents, _ := consensusComp.NewManagedConsensusComponents(consensusComponentsFactory)
	require.NotNil(t, managedConsensusComponents)

	size, err := managedConsensusComponents.ConsensusGroupSize()
	require.Equal(t, errorsMx.ErrNilConsensusComponentsHolder, err)
	require.Zero(t, size)

	err = managedConsensusComponents.Create()
	require.NoError(t, err)
	size, err = managedConsensusComponents.ConsensusGroupSize()
	require.NoError(t, err)
	require.Equal(t, 2, size)
	assert.Equal(t, "*sync.ShardBootstrap", fmt.Sprintf("%T", managedConsensusComponents.Bootstrapper()))
}

func TestManagedConsensusComponents_CreateShouldWorkForSovereign(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	args.ChainRunType = common.ChainRunTypeSovereign
	args.ConsensusModel = consensus.ConsensusModelV2

	consensusComponentsFactory, _ := consensusComp.NewConsensusComponentsFactory(args)
	managedConsensusComponents, err := consensusComp.NewManagedConsensusComponents(consensusComponentsFactory)

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
	assert.Equal(t, "*sync.SovereignChainShardBootstrap", fmt.Sprintf("%T", managedConsensusComponents.Bootstrapper()))
}

func TestManagedConsensusComponents_CreateShouldFailForInvalidChainType(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	args.ChainRunType = "invalid"

	consensusComponentsFactory, _ := consensusComp.NewConsensusComponentsFactory(args)
	managedConsensusComponents, err := consensusComp.NewManagedConsensusComponents(consensusComponentsFactory)
	require.NoError(t, err)

	err = managedConsensusComponents.Create()
	assert.Contains(t, err.Error(), errors.ErrUnimplementedChainRunType.Error())
}

func TestManagedConsensusComponents_CheckSubcomponents(t *testing.T) {
	t.Parallel()

	consensusComponentsFactory, _ := consensusComp.NewConsensusComponentsFactory(createMockConsensusComponentsFactoryArgs())
	managedConsensusComponents, _ := consensusComp.NewManagedConsensusComponents(consensusComponentsFactory)
	require.NotNil(t, managedConsensusComponents)

	require.Equal(t, errorsMx.ErrNilConsensusComponentsHolder, managedConsensusComponents.CheckSubcomponents())

	err := managedConsensusComponents.Create()
	require.NoError(t, err)

	require.Nil(t, managedConsensusComponents.CheckSubcomponents())
}

func TestManagedConsensusComponents_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	managedConsensusComponents, _ := consensusComp.NewManagedConsensusComponents(nil)
	require.True(t, managedConsensusComponents.IsInterfaceNil())

	consensusComponentsFactory, _ := consensusComp.NewConsensusComponentsFactory(createMockConsensusComponentsFactoryArgs())
	managedConsensusComponents, _ = consensusComp.NewManagedConsensusComponents(consensusComponentsFactory)
	require.False(t, managedConsensusComponents.IsInterfaceNil())
}

func TestManagedConsensusComponents_Close(t *testing.T) {
	t.Parallel()

	consensusComponentsFactory, _ := consensusComp.NewConsensusComponentsFactory(createMockConsensusComponentsFactoryArgs())
	managedConsensusComponents, _ := consensusComp.NewManagedConsensusComponents(consensusComponentsFactory)

	err := managedConsensusComponents.Close()
	require.NoError(t, err)

	err = managedConsensusComponents.Create()
	require.NoError(t, err)

	err = managedConsensusComponents.Close()
	require.NoError(t, err)
	require.Nil(t, managedConsensusComponents.BroadcastMessenger())
}
