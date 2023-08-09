package sharding

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func TestSovereignShardCoordinator_ComputeId(t *testing.T) {
	shardCoordinator := NewSovereignShardCoordinator(core.SovereignChainShardId)
	require.Equal(t, core.SovereignChainShardId, shardCoordinator.ComputeId([]byte("address")))

	require.Equal(t, uint32(1), shardCoordinator.NumberOfShards())
}
