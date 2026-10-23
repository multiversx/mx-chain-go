package chainSimulator

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
)

// RequireSupernova advances at most maxBlocks and verifies both activation gates on every shard and meta.
// Tests must call it before the transactions whose Supernova behavior they qualify.
func RequireSupernova(t *testing.T, simulator ChainSimulator, numShards uint32, maxBlocks int) {
	t.Helper()
	shards := []uint32{core.MetachainShardId}
	for shard := uint32(0); shard < numShards; shard++ {
		shards = append(shards, shard)
	}
	active := func() bool {
		for _, shard := range shards {
			node := simulator.GetNodeHandler(shard)
			header := node.GetChainHandler().GetCurrentBlockHeader()
			if header == nil || !node.GetCoreComponents().EnableEpochsHandler().IsFlagEnabledInEpoch(common.SupernovaFlag, header.GetEpoch()) ||
				!node.GetCoreComponents().EnableRoundsHandler().IsFlagEnabledInRound(common.SupernovaRoundFlag, header.GetRound()) {
				return false
			}
		}
		return true
	}
	for blocks := 0; blocks < maxBlocks && !active(); blocks++ {
		require.NoError(t, simulator.GenerateBlocks(1))
	}
	require.True(t, active(), "Supernova must be active on every shard and meta before smoke assertions")
}
