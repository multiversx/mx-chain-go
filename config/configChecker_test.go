package config

import (
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon/nodesSetupMock"
	"github.com/stretchr/testify/require"
)

const numOfShards = 3

func generateCorrectConfig() EnableEpochs {
	return EnableEpochs{
		StakingV4Step1EnableEpoch: 4,
		StakingV4Step2EnableEpoch: 5,
		StakingV4Step3EnableEpoch: 6,
		MaxNodesChangeEnableEpoch: []MaxNodesChangeConfig{
			{
				EpochEnable:            0,
				MaxNumNodes:            36,
				NodesToShufflePerShard: 4,
			},
			{
				EpochEnable:            1,
				MaxNumNodes:            56,
				NodesToShufflePerShard: 2,
			},
			{
				EpochEnable:            6,
				MaxNumNodes:            48,
				NodesToShufflePerShard: 2,
			},
		},
	}
}

func TestSanityCheckEnableEpochsStakingV4(t *testing.T) {
	t.Parallel()

	t.Run("correct config, should work", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()
		err := sanityCheckEnableEpochsStakingV4(cfg, numOfShards)
		require.Nil(t, err)
	})

	t.Run("staking v4 steps not in ascending order, should return error", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()
		cfg.StakingV4Step1EnableEpoch = 5
		cfg.StakingV4Step2EnableEpoch = 5
		err := sanityCheckEnableEpochsStakingV4(cfg, numOfShards)
		require.Equal(t, errStakingV4StepsNotInOrder, err)

		cfg = generateCorrectConfig()
		cfg.StakingV4Step2EnableEpoch = 5
		cfg.StakingV4Step3EnableEpoch = 4
		err = sanityCheckEnableEpochsStakingV4(cfg, numOfShards)
		require.Equal(t, errStakingV4StepsNotInOrder, err)
	})

	t.Run("staking v4 steps not in cardinal order, should return error", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()

		cfg.StakingV4Step1EnableEpoch = 1
		cfg.StakingV4Step2EnableEpoch = 3
		cfg.StakingV4Step3EnableEpoch = 6
		err := sanityCheckEnableEpochsStakingV4(cfg, numOfShards)
		require.Equal(t, errStakingV4StepsNotInOrder, err)

		cfg.StakingV4Step1EnableEpoch = 1
		cfg.StakingV4Step2EnableEpoch = 2
		cfg.StakingV4Step3EnableEpoch = 6
		err = sanityCheckEnableEpochsStakingV4(cfg, numOfShards)
		require.Equal(t, errStakingV4StepsNotInOrder, err)

		cfg.StakingV4Step1EnableEpoch = 1
		cfg.StakingV4Step2EnableEpoch = 5
		cfg.StakingV4Step3EnableEpoch = 6
		err = sanityCheckEnableEpochsStakingV4(cfg, numOfShards)
		require.Equal(t, errStakingV4StepsNotInOrder, err)
	})

	t.Run("no previous config for max nodes change with one entry, should not return error", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()
		cfg.MaxNodesChangeEnableEpoch = []MaxNodesChangeConfig{
			{
				EpochEnable:            6,
				MaxNumNodes:            48,
				NodesToShufflePerShard: 2,
			},
		}

		err := sanityCheckEnableEpochsStakingV4(cfg, numOfShards)
		require.Nil(t, err)
	})

	t.Run("no max nodes config change for StakingV4Step3EnableEpoch, should return error", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()
		cfg.MaxNodesChangeEnableEpoch = []MaxNodesChangeConfig{
			{
				EpochEnable:            1,
				MaxNumNodes:            56,
				NodesToShufflePerShard: 2,
			},
			{
				EpochEnable:            444,
				MaxNumNodes:            48,
				NodesToShufflePerShard: 2,
			},
		}

		err := sanityCheckEnableEpochsStakingV4(cfg, numOfShards)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), errNoMaxNodesConfigChangeForStakingV4.Error()))
		require.True(t, strings.Contains(err.Error(), "6"))
	})

	t.Run("max nodes config change for StakingV4Step3EnableEpoch has no previous config change, should not error", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()
		cfg.MaxNodesChangeEnableEpoch = []MaxNodesChangeConfig{
			{
				EpochEnable:            cfg.StakingV4Step3EnableEpoch,
				MaxNumNodes:            48,
				NodesToShufflePerShard: 2,
			},
			{
				EpochEnable:            444,
				MaxNumNodes:            56,
				NodesToShufflePerShard: 2,
			},
		}

		err := sanityCheckEnableEpochsStakingV4(cfg, numOfShards)
		require.Nil(t, err)
	})

	t.Run("stakingV4 config for max nodes changed with different nodes to shuffle, should return error", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()
		cfg.MaxNodesChangeEnableEpoch[1].NodesToShufflePerShard = 2
		cfg.MaxNodesChangeEnableEpoch[2].NodesToShufflePerShard = 4

		err := sanityCheckEnableEpochsStakingV4(cfg, numOfShards)
		require.ErrorIs(t, err, errMismatchNodesToShuffle)
	})

	t.Run("stakingV4 config for max nodes changed with wrong max num nodes, should return error", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()
		cfg.MaxNodesChangeEnableEpoch[2].MaxNumNodes = 56

		err := sanityCheckEnableEpochsStakingV4(cfg, numOfShards)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), "expected"))
		require.True(t, strings.Contains(err.Error(), "48"))
		require.True(t, strings.Contains(err.Error(), "got"))
		require.True(t, strings.Contains(err.Error(), "56"))
	})
}

func TestSanityCheckNodesConfig(t *testing.T) {
	t.Parallel()

	numShards := uint32(3)
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()
		nodesSetup := &nodesSetupMock.NodesSetupMock{
			NumberOfShardsField:        numShards,
			HysteresisField:            0,
			MinNumberOfMetaNodesField:  5,
			MinNumberOfShardNodesField: 5,
		}
		err := SanityCheckNodesConfig(nodesSetup, cfg)
		require.Nil(t, err)

		cfg.MaxNodesChangeEnableEpoch = []MaxNodesChangeConfig{
			{
				EpochEnable:            1,
				MaxNumNodes:            3200,
				NodesToShufflePerShard: 80,
			},
			{
				EpochEnable:            2,
				MaxNumNodes:            2880,
				NodesToShufflePerShard: 80,
			},
			{
				EpochEnable:            3,
				MaxNumNodes:            2240,
				NodesToShufflePerShard: 80,
			},
			{
				EpochEnable:            4,
				MaxNumNodes:            2240,
				NodesToShufflePerShard: 40,
			},
			{
				EpochEnable:            6,
				MaxNumNodes:            2080,
				NodesToShufflePerShard: 40,
			},
		}
		nodesSetup = &nodesSetupMock.NodesSetupMock{
			NumberOfShardsField:        numShards,
			HysteresisField:            0.2,
			MinNumberOfMetaNodesField:  400,
			MinNumberOfShardNodesField: 400,
		}
		err = SanityCheckNodesConfig(nodesSetup, cfg)
		require.Nil(t, err)

		cfg.MaxNodesChangeEnableEpoch = []MaxNodesChangeConfig{
			{
				EpochEnable:            0,
				MaxNumNodes:            36,
				NodesToShufflePerShard: 4,
			},
			{
				EpochEnable:            1,
				MaxNumNodes:            56,
				NodesToShufflePerShard: 2,
			},
			{
				EpochEnable:            6,
				MaxNumNodes:            48,
				NodesToShufflePerShard: 2,
			},
		}
		nodesSetup = &nodesSetupMock.NodesSetupMock{
			NumberOfShardsField:        numShards,
			HysteresisField:            0,
			MinNumberOfMetaNodesField:  3,
			MinNumberOfShardNodesField: 3,
		}
		err = SanityCheckNodesConfig(nodesSetup, cfg)
		require.Nil(t, err)

		cfg.MaxNodesChangeEnableEpoch = []MaxNodesChangeConfig{
			{
				EpochEnable:            0,
				MaxNumNodes:            36,
				NodesToShufflePerShard: 4,
			},
			{
				EpochEnable:            1,
				MaxNumNodes:            56,
				NodesToShufflePerShard: 2,
			},
			{
				EpochEnable:            6,
				MaxNumNodes:            48,
				NodesToShufflePerShard: 2,
			},
		}
		nodesSetup = &nodesSetupMock.NodesSetupMock{
			NumberOfShardsField:        numShards,
			HysteresisField:            0.2,
			MinNumberOfMetaNodesField:  7,
			MinNumberOfShardNodesField: 7,
		}
		err = SanityCheckNodesConfig(nodesSetup, cfg)
		require.Nil(t, err)

		cfg.MaxNodesChangeEnableEpoch = []MaxNodesChangeConfig{
			{
				EpochEnable:            0,
				MaxNumNodes:            48,
				NodesToShufflePerShard: 4,
			},
			{
				EpochEnable:            1,
				MaxNumNodes:            56,
				NodesToShufflePerShard: 2,
			},
			{
				EpochEnable:            6,
				MaxNumNodes:            48,
				NodesToShufflePerShard: 2,
			},
		}
		nodesSetup = &nodesSetupMock.NodesSetupMock{
			NumberOfShardsField:        numShards,
			HysteresisField:            0.2,
			MinNumberOfMetaNodesField:  10,
			MinNumberOfShardNodesField: 10,
		}
		err = SanityCheckNodesConfig(nodesSetup, cfg)
		require.Nil(t, err)
	})

	t.Run("zero nodes to shuffle per shard, should not return error", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()
		cfg.MaxNodesChangeEnableEpoch = []MaxNodesChangeConfig{
			{
				EpochEnable:            4,
				MaxNumNodes:            3200,
				NodesToShufflePerShard: 0,
			},
			{
				EpochEnable:            6,
				MaxNumNodes:            3200,
				NodesToShufflePerShard: 0,
			},
		}
		nodesSetup := &nodesSetupMock.NodesSetupMock{
			NumberOfShardsField:        numShards,
			HysteresisField:            0.2,
			MinNumberOfMetaNodesField:  400,
			MinNumberOfShardNodesField: 400,
		}
		err := SanityCheckNodesConfig(nodesSetup, cfg)
		require.Nil(t, err)
	})

	t.Run("maxNumNodes < minNumNodesWithHysteresis, should return error ", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()
		cfg.MaxNodesChangeEnableEpoch = []MaxNodesChangeConfig{
			{
				EpochEnable:            4,
				MaxNumNodes:            1900,
				NodesToShufflePerShard: 80,
			},
		}
		nodesSetup := &nodesSetupMock.NodesSetupMock{
			NumberOfShardsField:        numShards,
			HysteresisField:            0.2,
			MinNumberOfMetaNodesField:  400,
			MinNumberOfShardNodesField: 400,
		}
		err := SanityCheckNodesConfig(nodesSetup, cfg)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), errInvalidMaxMinNodes.Error()))
		require.True(t, strings.Contains(err.Error(), "maxNumNodes: 1900"))
		require.True(t, strings.Contains(err.Error(), "minNumNodesWithHysteresis: 1920"))
	})
}
