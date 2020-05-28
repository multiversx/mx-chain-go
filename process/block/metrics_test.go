package block

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
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

	nodesCoord := &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{
				mock.NewValidatorMock([]byte("wrong-key1")), // nodes coordinator default return for OwnPubKey()
				mock.NewValidatorMock([]byte("wrong-key2")),
			}, nil
		},
	}
	statusHandler := &mock.AppStatusHandlerStub{
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

	nodesCoord := &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{
				mock.NewValidatorMock([]byte("key")), // nodes coordinator default return for OwnPubKey()
				mock.NewValidatorMock([]byte("wrong-key2")),
			}, nil
		},
	}
	statusHandler := &mock.AppStatusHandlerStub{
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

	nodesCoord := &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{
				mock.NewValidatorMock([]byte("another-key")),
				mock.NewValidatorMock([]byte("key")), // nodes coordinator default return for OwnPubKey()
			}, nil
		},
	}
	statusHandler := &mock.AppStatusHandlerStub{
		IncrementHandler: func(_ string) {
			incrementWasCalled = true
		},
	}

	incrementCountAcceptedBlocks(nodesCoord, statusHandler, &block.Header{PubKeysBitmap: []byte{2, 0}})
	assert.True(t, incrementWasCalled)
}
