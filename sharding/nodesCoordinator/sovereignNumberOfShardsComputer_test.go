package nodesCoordinator

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func TestSovereignNumberOfShardsComputer_ComputeNumberOfShards(t *testing.T) {
	t.Parallel()

	t.Run("should not work with 0", func(t *testing.T) {
		t.Parallel()

		ssc := newSovereignNumberOfShardsComputer()
		eligibleMap := make(map[uint32][]Validator)
		nodesConfig := &epochNodesConfig{
			eligibleMap: eligibleMap,
		}

		nbShards, err := ssc.ComputeNumberOfShards(nodesConfig)
		require.Equal(t, ErrInvalidNumberOfShards, err)
		require.Equal(t, uint32(0), nbShards)
	})

	t.Run("should work with core.SovereignChainShardId", func(t *testing.T) {
		t.Parallel()

		ssc := newSovereignNumberOfShardsComputer()
		eligibleMap := make(map[uint32][]Validator)
		eligibleMap[core.SovereignChainShardId] = []Validator{&validator{}}
		nodesConfig := &epochNodesConfig{
			eligibleMap: eligibleMap,
		}

		nbShards, err := ssc.ComputeNumberOfShards(nodesConfig)
		require.Nil(t, err)
		require.Equal(t, uint32(1), nbShards)
	})

	t.Run("should not work with other than core.SovereignChainShardId", func(t *testing.T) {
		t.Parallel()

		ssc := newSovereignNumberOfShardsComputer()
		eligibleMap := make(map[uint32][]Validator)
		eligibleMap[1] = []Validator{&validator{}}
		nodesConfig := &epochNodesConfig{
			eligibleMap: eligibleMap,
		}

		nbShards, err := ssc.ComputeNumberOfShards(nodesConfig)
		require.Equal(t, ErrInvalidSovereignChainShardId, err)
		require.Equal(t, uint32(0), nbShards)
	})

	t.Run("should not work with 2", func(t *testing.T) {
		t.Parallel()

		ssc := newSovereignNumberOfShardsComputer()
		eligibleMap := make(map[uint32][]Validator)
		eligibleMap[core.SovereignChainShardId] = []Validator{&validator{}}
		eligibleMap[1] = []Validator{&validator{}}
		nodesConfig := &epochNodesConfig{
			eligibleMap: eligibleMap,
		}

		nbShards, err := ssc.ComputeNumberOfShards(nodesConfig)
		require.Equal(t, ErrInvalidNumberOfShards, err)
		require.Equal(t, uint32(0), nbShards)
	})
}

func TestSovereignNumberOfShardsComputer_ShardIdFromNodesConfig(t *testing.T) {
	t.Parallel()

	ssc := newSovereignNumberOfShardsComputer()
	require.Equal(t, core.SovereignChainShardId, ssc.ShardIdFromNodesConfig(nil))
	require.Equal(t, core.SovereignChainShardId, ssc.ShardIdFromNodesConfig(&epochNodesConfig{shardID: core.MetachainShardId}))
}
