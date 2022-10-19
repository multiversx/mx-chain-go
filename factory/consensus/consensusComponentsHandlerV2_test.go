package consensus_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory/consensus"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	componentsMock "github.com/ElrondNetwork/elrond-go/testscommon/components"
	"github.com/stretchr/testify/assert"
)

func TestNewConsensusComponentsHandlerV2_ShouldErrNilManagedConsensusComponents(t *testing.T) {
	t.Parallel()

	managedConsensusComponentsV2, err := consensus.NewManagedConsensusComponentsV2(nil, nil)

	assert.Nil(t, managedConsensusComponentsV2)
	assert.Equal(t, errors.ErrNilManagedConsensusComponents, err)
}

func TestNewConsensusComponentsHandlerV2_ShouldErrNilConsensusComponentsFactory(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	coreComponents := componentsMock.GetDefaultCoreComponents()
	args.CoreComponents = coreComponents
	consensusComponentsFactory, _ := consensus.NewConsensusComponentsFactory(args)
	managedConsensusComponents, _ := consensus.NewManagedConsensusComponents(consensusComponentsFactory)

	managedConsensusComponentsV2, err := consensus.NewManagedConsensusComponentsV2(managedConsensusComponents, nil)

	assert.Nil(t, managedConsensusComponentsV2)
	assert.Equal(t, errors.ErrNilConsensusComponentsFactory, err)
}

func TestNewConsensusComponentsHandlerV2_ShouldWork(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	coreComponents := componentsMock.GetDefaultCoreComponents()
	args.CoreComponents = coreComponents
	consensusComponentsFactory, _ := consensus.NewConsensusComponentsFactory(args)
	managedConsensusComponents, _ := consensus.NewManagedConsensusComponents(consensusComponentsFactory)
	consensusComponentsFactoryV2, _ := consensus.NewConsensusComponentsFactoryV2(consensusComponentsFactory)

	managedConsensusComponentsV2, err := consensus.NewManagedConsensusComponentsV2(managedConsensusComponents, consensusComponentsFactoryV2)

	assert.NotNil(t, managedConsensusComponentsV2)
	assert.Nil(t, err)
}

func TestConsensusComponentsHandlerV2_CreateWithInvalidArgsShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	coreComponents := componentsMock.GetDefaultCoreComponents()
	args.CoreComponents = coreComponents
	consensusComponentsFactory, _ := consensus.NewConsensusComponentsFactory(args)
	managedConsensusComponents, _ := consensus.NewManagedConsensusComponents(consensusComponentsFactory)
	consensusComponentsFactoryV2, _ := consensus.NewConsensusComponentsFactoryV2(consensusComponentsFactory)
	managedConsensusComponentsV2, _ := consensus.NewManagedConsensusComponentsV2(managedConsensusComponents, consensusComponentsFactoryV2)
	coreComponents.AppStatusHdl = nil

	err := managedConsensusComponentsV2.Create()

	assert.Error(t, err)
	assert.NotNil(t, managedConsensusComponentsV2.CheckSubcomponents())
}

func TestConsensusComponentsHandlerV2_CreateShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	consensusComponentsFactory, _ := consensus.NewConsensusComponentsFactory(args)
	managedConsensusComponents, _ := consensus.NewManagedConsensusComponents(consensusComponentsFactory)
	consensusComponentsFactoryV2, _ := consensus.NewConsensusComponentsFactoryV2(consensusComponentsFactory)
	managedConsensusComponentsV2, _ := consensus.NewManagedConsensusComponentsV2(managedConsensusComponents, consensusComponentsFactoryV2)

	assert.Nil(t, managedConsensusComponentsV2.BroadcastMessenger())
	assert.Nil(t, managedConsensusComponentsV2.Chronology())
	assert.Nil(t, managedConsensusComponentsV2.ConsensusWorker())
	assert.Error(t, managedConsensusComponentsV2.CheckSubcomponents())

	err := managedConsensusComponentsV2.Create()

	assert.Nil(t, err)
	assert.NotNil(t, managedConsensusComponentsV2.BroadcastMessenger())
	assert.NotNil(t, managedConsensusComponentsV2.Chronology())
	assert.NotNil(t, managedConsensusComponentsV2.ConsensusWorker())
	assert.NoError(t, managedConsensusComponentsV2.CheckSubcomponents())
}
