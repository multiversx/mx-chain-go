package chainSimulator

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func TestConsensusKeyIndexesIncludeEligibleValidators(t *testing.T) {
	args := ArgsBaseChainSimulator{
		ArgsChainSimulator: ArgsChainSimulator{
			NumOfShards:              3,
			MinNodesPerShard:         2,
			MetaChainMinNodes:        2,
			NumNodesWaitingListShard: 6,
			NumNodesWaitingListMeta:  6,
			ConsensusMode:            ConsensusModeBLS,
		},
	}

	require.Equal(t, []int{0, 1}, consensusKeyIndexes(args, -1))
	require.Equal(t, []int{2, 3}, consensusKeyIndexes(args, 0))
	require.Equal(t, []int{4, 5}, consensusKeyIndexes(args, 1))
	require.Equal(t, []int{6, 7}, consensusKeyIndexes(args, 2))
}

func TestConsensusKeyIndexesDisabled(t *testing.T) {
	require.Nil(t, consensusKeyIndexes(ArgsBaseChainSimulator{}, -1))
}

func TestWaitingShardAssignment(t *testing.T) {
	require.Equal(t,
		[]uint32{1, 2, core.MetachainShardId, 0, 1, 2, core.MetachainShardId, 0},
		waitingShardAssignment(8, 3),
	)
}
