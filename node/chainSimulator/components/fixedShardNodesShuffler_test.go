package components

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

func TestPinGenesisValidators(t *testing.T) {
	metaValidator, err := nodesCoordinator.NewValidator([]byte("meta"), 1, 0)
	require.NoError(t, err)
	shardValidator, err := nodesCoordinator.NewValidator([]byte("shard"), 1, 1)
	require.NoError(t, err)
	dynamicValidator, err := nodesCoordinator.NewValidator([]byte("dynamic"), 1, 2)
	require.NoError(t, err)
	secondShardValidator, err := nodesCoordinator.NewValidator([]byte("second-shard"), 1, 3)
	require.NoError(t, err)
	secondMetaValidator, err := nodesCoordinator.NewValidator([]byte("second-meta"), 1, 4)
	require.NoError(t, err)

	pinned := pinGenesisValidators(map[uint32][]nodesCoordinator.Validator{
		0:                     {metaValidator},
		1:                     {shardValidator, dynamicValidator, secondMetaValidator},
		core.MetachainShardId: {secondShardValidator},
	}, map[string]uint32{
		"meta":         core.MetachainShardId,
		"second-meta":  core.MetachainShardId,
		"shard":        0,
		"second-shard": 0,
	})

	// Sources are visited deterministically as shard 0, shard 1, then metachain.
	require.Equal(t, []nodesCoordinator.Validator{shardValidator, secondShardValidator}, pinned[0])
	require.Equal(t, []nodesCoordinator.Validator{dynamicValidator}, pinned[1])
	require.Equal(t, []nodesCoordinator.Validator{metaValidator, secondMetaValidator}, pinned[core.MetachainShardId])
}
