package headerCheck

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/assert"
)

func TestComputeConsensusGroup(t *testing.T) {
	t.Parallel()

	t.Run("should work for a random block", func(t *testing.T) {
		header := &block.Header{
			Epoch:        1123,
			Round:        37373,
			Nonce:        38383,
			ShardID:      2,
			PrevRandSeed: []byte("prev rand seed"),
		}

		validator1, _ := nodesCoordinator.NewValidator([]byte("pk1"), 1, 1)
		validator2, _ := nodesCoordinator.NewValidator([]byte("pk2"), 1, 2)

		validatorGroup := []nodesCoordinator.Validator{validator1, validator2}
		nodesCoordinatorInstance := shardingMocks.NewNodesCoordinatorMock()
		nodesCoordinatorInstance.ComputeValidatorsGroupCalled = func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
			assert.Equal(t, header.PrevRandSeed, randomness)
			assert.Equal(t, header.Round, round)
			assert.Equal(t, header.ShardID, shardId)
			assert.Equal(t, header.Epoch, epoch)

			return validatorGroup, nil
		}

		vGroup, err := ComputeConsensusGroup(header, nodesCoordinatorInstance)
		assert.Nil(t, err)
		assert.Equal(t, validatorGroup, vGroup)
	})
	t.Run("should work for a start of epoch block", func(t *testing.T) {
		header := &block.Header{
			Epoch:              1123,
			Round:              37373,
			Nonce:              38383,
			ShardID:            2,
			PrevRandSeed:       []byte("prev rand seed"),
			EpochStartMetaHash: []byte("epoch start metahash"),
		}

		validator1, _ := nodesCoordinator.NewValidator([]byte("pk1"), 1, 1)
		validator2, _ := nodesCoordinator.NewValidator([]byte("pk2"), 1, 2)

		validatorGroup := []nodesCoordinator.Validator{validator1, validator2}
		nodesCoordinatorInstance := shardingMocks.NewNodesCoordinatorMock()
		nodesCoordinatorInstance.ComputeValidatorsGroupCalled = func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
			assert.Equal(t, header.PrevRandSeed, randomness)
			assert.Equal(t, header.Round, round)
			assert.Equal(t, header.ShardID, shardId)
			assert.Equal(t, header.Epoch-1, epoch)

			return validatorGroup, nil
		}

		vGroup, err := ComputeConsensusGroup(header, nodesCoordinatorInstance)
		assert.Nil(t, err)
		assert.Equal(t, validatorGroup, vGroup)
	})
}
