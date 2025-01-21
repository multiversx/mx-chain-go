package nodesCoordinator

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func TestNumberOfShardsComputer_ComputeNumberOfShards(t *testing.T) {
	t.Parallel()

	t.Run("should fail with less than 2", func(t *testing.T) {
		t.Parallel()

		shardsComputer := newNumberOfShardsWithMetaComputer()
		eligibleMap := make(map[uint32][]Validator)
		eligibleMap[0] = []Validator{&validator{}}
		nodesConfig := &epochNodesConfig{
			eligibleMap: eligibleMap,
		}

		nbShards, err := shardsComputer.ComputeNumberOfShards(nodesConfig)
		require.Equal(t, ErrInvalidNumberOfShards, err)
		require.Equal(t, uint32(0), nbShards)
	})

	t.Run("should fail with just meta", func(t *testing.T) {
		t.Parallel()

		shardsComputer := newNumberOfShardsWithMetaComputer()
		eligibleMap := make(map[uint32][]Validator)
		eligibleMap[core.MetachainShardId] = []Validator{&validator{}}
		nodesConfig := &epochNodesConfig{
			eligibleMap: eligibleMap,
		}

		nbShards, err := shardsComputer.ComputeNumberOfShards(nodesConfig)
		require.Equal(t, ErrInvalidNumberOfShards, err)
		require.Equal(t, uint32(0), nbShards)
	})

	t.Run("should work with 1 shard and meta", func(t *testing.T) {
		t.Parallel()

		shardsComputer := newNumberOfShardsWithMetaComputer()
		eligibleMap := make(map[uint32][]Validator)
		eligibleMap[0] = []Validator{&validator{}}
		eligibleMap[core.MetachainShardId] = []Validator{&validator{}}
		nodesConfig := &epochNodesConfig{
			eligibleMap: eligibleMap,
		}
		expectedNbShardsWithoutMeta := uint32(1)

		nbShards, err := shardsComputer.ComputeNumberOfShards(nodesConfig)
		require.Nil(t, err)
		require.Equal(t, expectedNbShardsWithoutMeta, nbShards)
	})

	t.Run("should not work with 2 shards", func(t *testing.T) {
		t.Parallel()

		shardsComputer := newNumberOfShardsWithMetaComputer()
		eligibleMap := make(map[uint32][]Validator)
		eligibleMap[0] = []Validator{&validator{}}
		eligibleMap[1] = []Validator{&validator{}}
		nodesConfig := &epochNodesConfig{
			eligibleMap: eligibleMap,
		}

		nbShards, err := shardsComputer.ComputeNumberOfShards(nodesConfig)
		require.Equal(t, ErrMetachainShardIdNotFound, err)
		require.Equal(t, uint32(0), nbShards)
	})

	t.Run("should work with 10 shards and meta", func(t *testing.T) {
		t.Parallel()

		shardsComputer := newNumberOfShardsWithMetaComputer()
		eligibleMap := make(map[uint32][]Validator)
		for i := uint32(0); i < 10; i++ {
			eligibleMap[i] = []Validator{&validator{}}
		}
		eligibleMap[core.MetachainShardId] = []Validator{&validator{}}
		nodesConfig := &epochNodesConfig{
			eligibleMap: eligibleMap,
		}
		expectedNbShardsWithoutMeta := uint32(10)

		nbShards, err := shardsComputer.ComputeNumberOfShards(nodesConfig)
		require.Nil(t, err)
		require.Equal(t, expectedNbShardsWithoutMeta, nbShards)
	})
}

func TestNumberOfShardsComputer_ShardIdFromNodesConfig(t *testing.T) {
	t.Parallel()

	ssc := newNumberOfShardsWithMetaComputer()
	require.Equal(t, uint32(3), ssc.ShardIdFromNodesConfig(&epochNodesConfig{shardID: 3}))
	require.Equal(t, core.MetachainShardId, ssc.ShardIdFromNodesConfig(&epochNodesConfig{shardID: core.MetachainShardId}))
}
