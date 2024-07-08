package headerCheck

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/assert"
)

func TestComputeConsensusGroup(t *testing.T) {
	t.Parallel()

	t.Run("nil header should error", func(t *testing.T) {
		nodesCoordinatorInstance := shardingMocks.NewNodesCoordinatorMock()
		nodesCoordinatorInstance.ComputeValidatorsGroupCalled = func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
			assert.Fail(t, "should have not called ComputeValidatorsGroupCalled")
			return nil, nil
		}

		vGroup, err := ComputeConsensusGroup(nil, nodesCoordinatorInstance)
		assert.Equal(t, process.ErrNilHeaderHandler, err)
		assert.Nil(t, vGroup)
	})
	t.Run("nil nodes coordinator should error", func(t *testing.T) {
		header := &block.Header{
			Epoch:        1123,
			Round:        37373,
			Nonce:        38383,
			ShardID:      2,
			PrevRandSeed: []byte("prev rand seed"),
		}

		vGroup, err := ComputeConsensusGroup(header, nil)
		assert.Equal(t, process.ErrNilNodesCoordinator, err)
		assert.Nil(t, vGroup)
	})
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

func generatePubKeys(num int) []string {
	consensusGroup := make([]string, 0, num)
	for i := 0; i < num; i++ {
		consensusGroup = append(consensusGroup, fmt.Sprintf("pub key %d", i))
	}

	return consensusGroup
}

func TestComputeSignersPublicKeys(t *testing.T) {
	t.Parallel()

	t.Run("should compute with 16 validators", func(t *testing.T) {
		t.Parallel()

		consensusGroup := generatePubKeys(16)
		mask0 := byte(0b00110101)
		mask1 := byte(0b01001101)

		result := ComputeSignersPublicKeys(consensusGroup, []byte{mask0, mask1})
		expected := []string{
			"pub key 0",
			"pub key 2",
			"pub key 4",
			"pub key 5",

			"pub key 8",
			"pub key 10",
			"pub key 11",
			"pub key 14",
		}

		assert.Equal(t, expected, result)
	})
	t.Run("should compute with 14 validators", func(t *testing.T) {
		t.Parallel()

		consensusGroup := generatePubKeys(14)
		mask0 := byte(0b00110101)
		mask1 := byte(0b00001101)

		result := ComputeSignersPublicKeys(consensusGroup, []byte{mask0, mask1})
		expected := []string{
			"pub key 0",
			"pub key 2",
			"pub key 4",
			"pub key 5",

			"pub key 8",
			"pub key 10",
			"pub key 11",
		}

		assert.Equal(t, expected, result)
	})
	t.Run("should compute with 14 validators, mask is 0", func(t *testing.T) {
		t.Parallel()

		consensusGroup := generatePubKeys(14)
		mask0 := byte(0b00000000)
		mask1 := byte(0b00000000)

		result := ComputeSignersPublicKeys(consensusGroup, []byte{mask0, mask1})
		expected := make([]string, 0)

		assert.Equal(t, expected, result)
	})
	t.Run("should compute with 14 validators, mask contains all bits set", func(t *testing.T) {
		t.Parallel()

		consensusGroup := generatePubKeys(14)
		mask0 := byte(0b11111111)
		mask1 := byte(0b00111111)

		result := ComputeSignersPublicKeys(consensusGroup, []byte{mask0, mask1})

		assert.Equal(t, consensusGroup, result)
	})
	t.Run("should compute with 17 validators, mask contains 2 bytes", func(t *testing.T) {
		t.Parallel()

		consensusGroup := generatePubKeys(17)
		mask0 := byte(0b11111111)
		mask1 := byte(0b11111111)

		result := ComputeSignersPublicKeys(consensusGroup, []byte{mask0, mask1})
		expected := generatePubKeys(16)
		assert.Equal(t, expected, result)
	})
}
