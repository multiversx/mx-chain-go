package nodesCoordinator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSovereignIndexHashedNodesCoordinator_ComputeNumberOfShards(t *testing.T) {
	t.Parallel()

	t.Run("should not work with 0", func(t *testing.T) {
		t.Parallel()

		ssc := NewSovereignNumberOfShardsComputer()
		eligibleMap := make(map[uint32][]Validator)
		nodesConfig := &epochNodesConfig{
			eligibleMap: eligibleMap,
		}

		nbShards, err := ssc.ComputeNumberOfShards(nodesConfig)
		require.Equal(t, ErrInvalidNumberOfShards, err)
		require.Equal(t, uint32(0), nbShards)
	})

	t.Run("should work only with 1", func(t *testing.T) {
		t.Parallel()

		ssc := NewSovereignNumberOfShardsComputer()
		eligibleMap := make(map[uint32][]Validator)
		eligibleMap[0] = []Validator{&validator{}}
		nodesConfig := &epochNodesConfig{
			eligibleMap: eligibleMap,
		}

		nbShards, err := ssc.ComputeNumberOfShards(nodesConfig)
		require.Nil(t, err)
		require.Equal(t, uint32(1), nbShards)
	})

	t.Run("should not work with 2", func(t *testing.T) {
		t.Parallel()

		ssc := NewSovereignNumberOfShardsComputer()
		eligibleMap := make(map[uint32][]Validator)
		eligibleMap[0] = []Validator{&validator{}}
		eligibleMap[1] = []Validator{&validator{}}
		nodesConfig := &epochNodesConfig{
			eligibleMap: eligibleMap,
		}

		nbShards, err := ssc.ComputeNumberOfShards(nodesConfig)
		require.NotNil(t, err)
		require.Equal(t, uint32(0), nbShards)
	})
}
