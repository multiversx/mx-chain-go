package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func generateCorrectConfig() *Configs {
	return &Configs{
		EpochConfig: &EpochConfig{
			EnableEpochs: EnableEpochs{
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
			},
		},
		GeneralConfig: &Config{
			GeneralSettings: GeneralSettingsConfig{
				GenesisMaxNumberOfShards: 3,
			},
		},
	}
}

func TestSanityCheckEnableEpochsStakingV4(t *testing.T) {
	t.Parallel()

	t.Run("correct config, should work", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()
		err := SanityCheckEnableEpochsStakingV4(cfg)
		require.Nil(t, err)
	})

	t.Run("staking v4 steps not in ascending order, should return error", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()
		cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 5
		cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 5
		err := SanityCheckEnableEpochsStakingV4(cfg)
		require.Equal(t, errStakingV4StepsNotInOrder, err)

		cfg = generateCorrectConfig()
		cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 5
		cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4
		err = SanityCheckEnableEpochsStakingV4(cfg)
		require.Equal(t, errStakingV4StepsNotInOrder, err)
	})

	t.Run("staking v4 steps not in cardinal order, should work", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()

		cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 1
		cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
		cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 6
		err := SanityCheckEnableEpochsStakingV4(cfg)
		require.Nil(t, err)

		cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 1
		cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 2
		cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 6
		err = SanityCheckEnableEpochsStakingV4(cfg)
		require.Nil(t, err)

		cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 1
		cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 5
		cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 6
		err = SanityCheckEnableEpochsStakingV4(cfg)
		require.Nil(t, err)
	})

	t.Run("no previous config for max nodes change, should work", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()
		cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch = []MaxNodesChangeConfig{
			{
				EpochEnable:            6,
				MaxNumNodes:            48,
				NodesToShufflePerShard: 2,
			},
		}

		err := SanityCheckEnableEpochsStakingV4(cfg)
		require.Nil(t, err)
	})

	t.Run("no max nodes config change for StakingV4Step3EnableEpoch, should return error", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()
		cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch = []MaxNodesChangeConfig{
			{
				EpochEnable:            444,
				MaxNumNodes:            48,
				NodesToShufflePerShard: 2,
			},
		}

		err := SanityCheckEnableEpochsStakingV4(cfg)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), errNoMaxNodesConfigChangeForStakingV4.Error()))
		require.True(t, strings.Contains(err.Error(), "6"))
	})

	t.Run("stakingV4 config for max nodes changed with different nodes to shuffle, should work", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()
		cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[1].NodesToShufflePerShard = 2
		cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].NodesToShufflePerShard = 4

		err := SanityCheckEnableEpochsStakingV4(cfg)
		require.Nil(t, err)
	})

	t.Run("stakingV4 config for max nodes changed with wrong max num nodes, should return error", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()
		cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].MaxNumNodes = 56

		err := SanityCheckEnableEpochsStakingV4(cfg)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), "expected"))
		require.True(t, strings.Contains(err.Error(), "48"))
		require.True(t, strings.Contains(err.Error(), "got"))
		require.True(t, strings.Contains(err.Error(), "56"))
	})
}
