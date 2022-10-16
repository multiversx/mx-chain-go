package consensus_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus/chronology"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/sposFactory"
	commonErrors "github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory/consensus"
	factoryMock "github.com/ElrondNetwork/elrond-go/factory/mock"
	componentsMock "github.com/ElrondNetwork/elrond-go/testscommon/components"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConsensusComponentsFactoryV2_ShouldErrNilConsensusComponentFactory(t *testing.T) {
	t.Parallel()

	ccf, err := consensus.NewConsensusComponentsFactoryV2(nil)

	assert.Nil(t, ccf)
	assert.Equal(t, sposFactory.ErrNilConsensusComponentFactory, err)
}

func TestNewConsensusComponentsFactoryV2_ShouldWork(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)

	ccf, _ := consensus.NewConsensusComponentsFactory(args)

	ccfV2, err := consensus.NewConsensusComponentsFactoryV2(ccf)

	assert.NotNil(t, ccfV2)
	assert.Nil(t, err)
}

func TestConsensusComponentsFactoryV2_CreateForShard(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)

	ccf, _ := consensus.NewConsensusComponentsFactory(args)
	require.NotNil(t, ccf)

	ccfV2, _ := consensus.NewConsensusComponentsFactoryV2(ccf)
	require.NotNil(t, ccfV2)

	cc, err := ccfV2.Create()

	require.NoError(t, err)
	require.NotNil(t, cc)
}

func TestConsensusComponentsFactoryV2_CreateNilShardCoordinator(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	consensusArgs := componentsMock.GetConsensusArgs(shardCoordinator)
	processComponents := &factoryMock.ProcessComponentsMock{}
	consensusArgs.ProcessComponents = processComponents

	ccf, _ := consensus.NewConsensusComponentsFactory(consensusArgs)
	require.NotNil(t, ccf)

	ccfV2, _ := consensus.NewConsensusComponentsFactoryV2(ccf)
	require.NotNil(t, ccfV2)

	cc, err := ccfV2.Create()

	require.Nil(t, cc)
	require.Equal(t, commonErrors.ErrNilShardCoordinator, err)
}

func TestConsensusComponentsFactoryV2_CreateConsensusTopicCreateTopicError(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	localError := errors.New("error")
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	networkComponents := componentsMock.GetDefaultNetworkComponents()
	networkComponents.Messenger = &p2pmocks.MessengerStub{
		HasTopicValidatorCalled: func(name string) bool {
			return false
		},
		HasTopicCalled: func(name string) bool {
			return false
		},
		CreateTopicCalled: func(name string, createChannelForTopic bool) error {
			return localError
		},
	}
	args.NetworkComponents = networkComponents

	ccf, _ := consensus.NewConsensusComponentsFactory(args)
	require.NotNil(t, ccf)

	ccfV2, _ := consensus.NewConsensusComponentsFactoryV2(ccf)
	require.NotNil(t, ccfV2)

	cc, err := ccfV2.Create()

	require.Nil(t, cc)
	require.Equal(t, localError, err)
}

func TestConsensusComponentsFactoryV2_CreateConsensusTopicNilMessageProcessor(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	networkComponents := componentsMock.GetDefaultNetworkComponents()
	networkComponents.Messenger = nil
	args.NetworkComponents = networkComponents

	ccf, _ := consensus.NewConsensusComponentsFactory(args)
	require.NotNil(t, ccf)

	ccfV2, _ := consensus.NewConsensusComponentsFactoryV2(ccf)
	require.NotNil(t, ccfV2)

	cc, err := ccfV2.Create()

	require.Nil(t, cc)
	require.Equal(t, commonErrors.ErrNilMessenger, err)
}

func TestConsensusComponentsFactoryV2_CreateNilSyncTimer(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	coreComponents := componentsMock.GetDefaultCoreComponents()
	coreComponents.NtpSyncTimer = nil
	args.CoreComponents = coreComponents

	ccf, _ := consensus.NewConsensusComponentsFactory(args)
	require.NotNil(t, ccf)

	ccfV2, _ := consensus.NewConsensusComponentsFactoryV2(ccf)
	require.NotNil(t, ccfV2)

	cc, err := ccfV2.Create()

	require.Nil(t, cc)
	require.Equal(t, chronology.ErrNilSyncTimer, err)
}
