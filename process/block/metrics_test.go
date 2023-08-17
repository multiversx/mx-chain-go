package block

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/testscommon"
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

func TestMetrics_IncrementMetricCountConsensusAcceptedBlocks(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("nodes coordinator returns error should early exit", func(t *testing.T) {
		t.Parallel()

		nodesCoord := &shardingMocks.NodesCoordinatorMock{
			GetValidatorsPublicKeysCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]string, error) {
				return nil, expectedErr
			},
		}
		statusHandler := &statusHandlerMock.AppStatusHandlerStub{
			IncrementHandler: func(_ string) {
				assert.Fail(t, "should have not been called")
			},
		}

		incrementMetricCountConsensusAcceptedBlocks(&block.HeaderV2{}, 0, nodesCoord, statusHandler, &testscommon.ManagedPeersHolderStub{})
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cntIncrement := 0
		mainKey := "main key"
		managedKeyInConsensus := "managed key in consensus"
		nodesCoord := &shardingMocks.NodesCoordinatorStub{
			GetOwnPublicKeyCalled: func() []byte {
				return []byte(mainKey)
			},
			GetValidatorsPublicKeysCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]string, error) {
				return []string{
					"some leader",
					mainKey,
					managedKeyInConsensus,
					"some other key",
				}, nil
			},
		}
		statusHandler := &statusHandlerMock.AppStatusHandlerStub{
			IncrementHandler: func(_ string) {
				cntIncrement++
			},
		}
		managedPeersHolder := &testscommon.ManagedPeersHolderStub{
			IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
				return string(pkBytes) == managedKeyInConsensus
			},
		}

		incrementMetricCountConsensusAcceptedBlocks(&block.MetaBlock{}, 0, nodesCoord, statusHandler, managedPeersHolder)
		assert.Equal(t, 2, cntIncrement) // main key + managed key
	})
}
