package track_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/stretchr/testify/require"
)

func TestShardBlockTrack_ComputeCrossInfo(t *testing.T) {
	t.Parallel()

	t.Run("legacy meta block reads pending miniblocks from shard info", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateShardTrackerMockArguments()
		sbt, err := track.NewShardBlockTrack(shardArguments)
		require.Nil(t, err)

		metaBlock := &block.MetaBlock{
			ShardInfo: []block.ShardData{
				{ShardID: 0, NumPendingMiniBlocks: 3, LastIncludedMetaNonce: 11},
				{ShardID: 1, NumPendingMiniBlocks: 5, LastIncludedMetaNonce: 22},
			},
		}

		sbt.ComputeCrossInfo([]data.HeaderHandler{metaBlock})

		require.Equal(t, uint32(3), sbt.GetNumPendingMiniBlocks(0))
		require.Equal(t, uint64(11), sbt.GetLastShardProcessedMetaNonce(0))
		require.Equal(t, uint32(5), sbt.GetNumPendingMiniBlocks(1))
		require.Equal(t, uint64(22), sbt.GetLastShardProcessedMetaNonce(1))
	})

	t.Run("V3 meta block reads pending miniblocks from shard info proposal", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateShardTrackerMockArguments()
		sbt, err := track.NewShardBlockTrack(shardArguments)
		require.Nil(t, err)

		metaBlock := &block.MetaBlockV3{
			ShardInfo: []block.ShardData{
				{ShardID: 0, NumPendingMiniBlocks: 0, LastIncludedMetaNonce: 11},
				{ShardID: 1, NumPendingMiniBlocks: 0, LastIncludedMetaNonce: 22},
			},
			ShardInfoProposal: []block.ShardDataProposal{
				{ShardID: 0, NumPendingMiniBlocks: 3},
				{ShardID: 1, NumPendingMiniBlocks: 5},
			},
		}

		sbt.ComputeCrossInfo([]data.HeaderHandler{metaBlock})

		require.Equal(t, uint32(3), sbt.GetNumPendingMiniBlocks(0))
		require.Equal(t, uint64(11), sbt.GetLastShardProcessedMetaNonce(0))
		require.Equal(t, uint32(5), sbt.GetNumPendingMiniBlocks(1))
		require.Equal(t, uint64(22), sbt.GetLastShardProcessedMetaNonce(1))
	})
}
