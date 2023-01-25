package block

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_CalculateRoundDuration(t *testing.T) {
	t.Parallel()

	lastBlockTimestamp := uint64(80)
	currentBlockTimestamp := uint64(100)
	lastBlockRound := uint64(5)
	currentBlockRound := uint64(10)
	expectedRoundDuration := uint64(4)

	roundDuration := calculateRoundDuration(lastBlockTimestamp, currentBlockTimestamp, lastBlockRound, currentBlockRound)
	assert.Equal(t, expectedRoundDuration, roundDuration)
}

func TestMetrics_IncrementCountAcceptedBlocks_KeyNotFoundShouldNotIncrement(t *testing.T) {
	t.Parallel()

	incrementWasCalled := false

	nodesCoord := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]nodesCoordinator.Validator, error) {
			return []nodesCoordinator.Validator{
				shardingMocks.NewValidatorMock([]byte("wrong-key1"), 1, 1), // nodes coordinator default return for OwnPubKey()
				shardingMocks.NewValidatorMock([]byte("wrong-key2"), 1, 1),
			}, nil
		},
	}
	statusHandler := &statusHandlerMock.AppStatusHandlerStub{
		IncrementHandler: func(_ string) {
			incrementWasCalled = true
		},
	}

	incrementCountAcceptedBlocks(nodesCoord, statusHandler, &block.Header{PubKeysBitmap: []byte{1, 0}})
	assert.False(t, incrementWasCalled)
}

func TestMetrics_IncrementCountAcceptedBlocks_IndexOutOfBoundsShouldNotIncrement(t *testing.T) {
	t.Parallel()

	incrementWasCalled := false

	nodesCoord := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]nodesCoordinator.Validator, error) {
			return []nodesCoordinator.Validator{
				shardingMocks.NewValidatorMock([]byte("key"), 1, 1), // nodes coordinator default return for OwnPubKey()
				shardingMocks.NewValidatorMock([]byte("wrong-key2"), 1, 1),
			}, nil
		},
	}
	statusHandler := &statusHandlerMock.AppStatusHandlerStub{
		IncrementHandler: func(_ string) {
			incrementWasCalled = true
		},
	}

	incrementCountAcceptedBlocks(nodesCoord, statusHandler, &block.Header{PubKeysBitmap: []byte{}})
	assert.False(t, incrementWasCalled)
}

func TestMetrics_IncrementCountAcceptedBlocks_ShouldWork(t *testing.T) {
	t.Parallel()

	incrementWasCalled := false

	nodesCoord := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]nodesCoordinator.Validator, error) {
			return []nodesCoordinator.Validator{
				shardingMocks.NewValidatorMock([]byte("another-key"), 1, 1),
				shardingMocks.NewValidatorMock([]byte("key"), 1, 1), // nodes coordinator default return for OwnPubKey()
			}, nil
		},
	}
	statusHandler := &statusHandlerMock.AppStatusHandlerStub{
		IncrementHandler: func(_ string) {
			incrementWasCalled = true
		},
	}

	incrementCountAcceptedBlocks(nodesCoord, statusHandler, &block.Header{PubKeysBitmap: []byte{2, 0}})
	assert.True(t, incrementWasCalled)
}
