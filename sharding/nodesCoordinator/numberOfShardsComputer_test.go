package nodesCoordinator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSovereignNumberOfShardsComputer_ComputeNumberOfShards(t *testing.T) {
	t.Parallel()

	t.Run("should fail with less than 2", func(t *testing.T) {
		t.Parallel()

		shardsComputer := NewNumberOfShardsWithMetaComputer()
		eligibleMap := make(map[uint32][]Validator)
		eligibleMap[0] = []Validator{&validator{}}
		nodesConfig := &epochNodesConfig{
			eligibleMap: eligibleMap,
		}

		nbShards, err := shardsComputer.ComputeNumberOfShards(nodesConfig)
		require.Equal(t, ErrInvalidNumberOfShards, err)
		require.Equal(t, uint32(0), nbShards)
	})

	t.Run("should work with 2", func(t *testing.T) {
		t.Parallel()

		shardsComputer := NewNumberOfShardsWithMetaComputer()
		eligibleMap := make(map[uint32][]Validator)
		eligibleMap[0] = []Validator{&validator{}}
		eligibleMap[1] = []Validator{&validator{}}
		nodesConfig := &epochNodesConfig{
			eligibleMap: eligibleMap,
		}
		expectedNbShardsWithoutMeta := uint32(1)

		nbShards, err := shardsComputer.ComputeNumberOfShards(nodesConfig)
		require.Nil(t, err)
		require.Equal(t, expectedNbShardsWithoutMeta, nbShards)
	})

	t.Run("should work with 10", func(t *testing.T) {
		t.Parallel()

		shardsComputer := NewNumberOfShardsWithMetaComputer()
		eligibleMap := make(map[uint32][]Validator)
		for i := uint32(0); i < 10; i++ {
			eligibleMap[i] = []Validator{&validator{}}
		}
		nodesConfig := &epochNodesConfig{
			eligibleMap: eligibleMap,
		}
		expectedNbShardsWithoutMeta := uint32(9)

		nbShards, err := shardsComputer.ComputeNumberOfShards(nodesConfig)
		require.Nil(t, err)
		require.Equal(t, expectedNbShardsWithoutMeta, nbShards)
	})
}
