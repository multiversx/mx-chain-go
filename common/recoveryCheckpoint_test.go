package common_test

import (
	"encoding/hex"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
)

func TestRecoveryCheckpoint_RequiresExactVectorAndExclusion(t *testing.T) {
	hash0 := make([]byte, 32)
	hash0[0] = 1
	hashMeta := make([]byte, 32)
	hashMeta[0] = 2
	cfg := &config.Config{
		HardforkRoundExclusions: []config.HardforkRoundExclusionConfig{{StartRound: 101, EndRound: 199}},
		HardforkRecoveryCheckpoint: config.HardforkRecoveryCheckpointConfig{
			Enabled: true,
			Round:   100,
			Headers: []config.HardforkRecoveryHeaderConfig{
				{ShardID: 0, Hash: hex.EncodeToString(hash0)},
				{ShardID: core.MetachainShardId, Hash: hex.EncodeToString(hashMeta)},
			},
		},
	}
	checkpoint, err := common.NewRecoveryCheckpoint(cfg)
	require.NoError(t, err)
	require.True(t, checkpoint.HasAllShards(1))
	require.False(t, checkpoint.HasAllShards(2))
	require.False(t, checkpoint.IsDiscarded(100, 0, hash0))
	require.True(t, checkpoint.IsDiscarded(100, 0, hashMeta))
	require.True(t, checkpoint.IsDiscarded(100, 1, hash0))
	require.True(t, checkpoint.IsDiscarded(101, 0, hash0))
	require.True(t, checkpoint.IsDiscarded(199, 0, hash0))
	require.False(t, checkpoint.IsDiscarded(200, 0, hash0))

	cfg.HardforkRoundExclusions[0].StartRound = 102
	_, err = common.NewRecoveryCheckpoint(cfg)
	require.ErrorIs(t, err, common.ErrInvalidRecoveryCheckpoint)

	cfg.HardforkRoundExclusions[0].StartRound = 101
	cfg.HardforkRecoveryCheckpoint.Headers[1].Hash = "not-a-hash"
	_, err = common.NewRecoveryCheckpoint(cfg)
	require.ErrorIs(t, err, common.ErrInvalidRecoveryCheckpoint)
}

func TestConfiguredRoundExclusionHandler_RejectsOnlyCompetingCheckpointHash(t *testing.T) {
	goodHash := make([]byte, 32)
	badHash := make([]byte, 32)
	badHash[0] = 1
	cfg := &config.Config{
		HardforkRoundExclusions: []config.HardforkRoundExclusionConfig{{StartRound: 11, EndRound: 19}},
		HardforkRecoveryCheckpoint: config.HardforkRecoveryCheckpointConfig{
			Enabled: true,
			Round:   10,
			Headers: []config.HardforkRecoveryHeaderConfig{
				{ShardID: 0, Hash: hex.EncodeToString(goodHash)},
				{ShardID: core.MetachainShardId, Hash: hex.EncodeToString(goodHash)},
			},
		},
	}
	handler, err := common.NewConfiguredRoundExclusionHandler(cfg)
	require.NoError(t, err)
	require.False(t, common.IsHeaderExcluded(handler, 10, 0, goodHash))
	require.True(t, common.IsHeaderExcluded(handler, 10, 0, badHash))
	require.True(t, common.IsHeaderExcluded(handler, 11, 0, goodHash))
	require.False(t, common.IsHeaderExcluded(handler, 20, 0, badHash))
}
