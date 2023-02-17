package consensus_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/errors"
	consensusComp "github.com/multiversx/mx-chain-go/factory/consensus"
	"github.com/multiversx/mx-chain-go/factory/mock"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
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
	statusCoreComponents := &factory.StatusCoreComponentsStub{
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

func TestManagedConsensusComponents_CreateShouldWorkForSovereign(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	args.ChainRunType = common.ChainRunTypeSovereign
	args.SubroundType = consensus.SubroundTypeV2

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

func TestManagedConsensusComponents_Close(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	consensusArgs := componentsMock.GetConsensusArgs(shardCoordinator)
	consensusComponentsFactory, _ := consensusComp.NewConsensusComponentsFactory(consensusArgs)
	managedConsensusComponents, _ := consensusComp.NewManagedConsensusComponents(consensusComponentsFactory)
	err := managedConsensusComponents.Create()
	require.NoError(t, err)

	err = managedConsensusComponents.Close()
	require.NoError(t, err)
	require.Nil(t, managedConsensusComponents.BroadcastMessenger())
}
