package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// NodesSetupStub -
type NodesSetupStub struct {
	GetRoundDurationCalled                    func() uint64
	GetShardConsensusGroupSizeCalled          func() uint32
	GetMetaConsensusGroupSizeCalled           func() uint32
	NumberOfShardsCalled                      func() uint32
	MinNumberOfNodesCalled                    func() uint32
	GetAdaptivityCalled                       func() bool
	GetHysteresisCalled                       func() float32
	GetShardIDForPubKeyCalled                 func(pubkey []byte) (uint32, error)
	InitialEligibleNodesPubKeysForShardCalled func(shardId uint32) ([]string, error)
	InitialNodesPubKeysCalled                 func() map[uint32][]string
	MinNumberOfMetaNodesCalled                func() uint32
	MinNumberOfShardNodesCalled               func() uint32
	MinNumberOfNodesWithHysteresisCalled      func() uint32
}

// MinNumberOfNodes -
func (n *NodesSetupStub) MinNumberOfNodes() uint32 {
	if n.MinNumberOfNodesCalled != nil {
		return n.MinNumberOfNodesCalled()
	}
	return 1
}

// GetRoundDuration -
func (n *NodesSetupStub) GetRoundDuration() uint64 {
	if n.GetRoundDurationCalled != nil {
		return n.GetRoundDurationCalled()
	}
	return 0
}

// GetShardConsensusGroupSize -
func (n *NodesSetupStub) GetShardConsensusGroupSize() uint32 {
	if n.GetShardConsensusGroupSizeCalled != nil {
		return n.GetShardConsensusGroupSizeCalled()
	}
	return 0
}

// GetMetaConsensusGroupSize -
func (n *NodesSetupStub) GetMetaConsensusGroupSize() uint32 {
	if n.GetMetaConsensusGroupSizeCalled != nil {
		return n.GetMetaConsensusGroupSizeCalled()
	}
	return 0
}

// NumberOfShards -
func (n *NodesSetupStub) NumberOfShards() uint32 {
	if n.NumberOfShardsCalled != nil {
		return n.NumberOfShardsCalled()
	}
	return 0
}

// GetAdaptivity -
func (n *NodesSetupStub) GetAdaptivity() bool {
	if n.GetAdaptivityCalled != nil {
		return n.GetAdaptivityCalled()
	}

	return false
}

// GetHysteresis -
func (n *NodesSetupStub) GetHysteresis() float32 {
	if n.GetHysteresisCalled != nil {
		return n.GetHysteresisCalled()
	}

	return 0
}

// GetShardIDForPubKey -
func (n *NodesSetupStub) GetShardIDForPubKey(pubkey []byte) (uint32, error) {
	if n.GetShardIDForPubKeyCalled != nil {
		return n.GetShardIDForPubKeyCalled(pubkey)
	}
	return 0, nil
}

// InitialEligibleNodesPubKeysForShard -
func (n *NodesSetupStub) InitialEligibleNodesPubKeysForShard(shardId uint32) ([]string, error) {
	if n.InitialEligibleNodesPubKeysForShardCalled != nil {
		return n.InitialEligibleNodesPubKeysForShardCalled(shardId)
	}

	return []string{"val1", "val2"}, nil
}

// InitialNodesPubKeys -
func (n *NodesSetupStub) InitialNodesPubKeys() map[uint32][]string {
	if n.InitialNodesPubKeysCalled != nil {
		return n.InitialNodesPubKeysCalled()
	}

	return map[uint32][]string{0: {"val1", "val2"}}
}

// MinNumberOfMetaNodes -
func (n *NodesSetupStub) MinNumberOfMetaNodes() uint32 {
	if n.MinNumberOfMetaNodesCalled != nil {
		return n.MinNumberOfMetaNodesCalled()
	}

	return 1
}

// MinNumberOfShardNodes -
func (n *NodesSetupStub) MinNumberOfShardNodes() uint32 {
	if n.MinNumberOfShardNodesCalled != nil {
		return n.MinNumberOfShardNodesCalled()
	}

	return 1
}

// MinNumberOfNodesWithHysteresis -
func (n *NodesSetupStub) MinNumberOfNodesWithHysteresis() uint32 {
	if n.MinNumberOfNodesWithHysteresisCalled != nil {
		return n.MinNumberOfNodesWithHysteresisCalled()
	}
	return n.MinNumberOfNodes()
}

// IsInterfaceNil -
func (n *NodesSetupStub) IsInterfaceNil() bool {
	return n == nil
}

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

	t.Run("staking v4 steps not in cardinal order, should return error", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()

		cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 1
		cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
		cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 6
		err := SanityCheckEnableEpochsStakingV4(cfg)
		require.Equal(t, errStakingV4StepsNotInOrder, err)

		cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 1
		cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 2
		cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 6
		err = SanityCheckEnableEpochsStakingV4(cfg)
		require.Equal(t, errStakingV4StepsNotInOrder, err)

		cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 1
		cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 5
		cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 6
		err = SanityCheckEnableEpochsStakingV4(cfg)
		require.Equal(t, errStakingV4StepsNotInOrder, err)
	})

	t.Run("no previous config for max nodes change, should return error", func(t *testing.T) {
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
		require.Equal(t, errNotEnoughMaxNodesChanges, err)
	})

	t.Run("no max nodes config change for StakingV4Step3EnableEpoch, should return error", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()
		cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch = []MaxNodesChangeConfig{
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

		err := SanityCheckEnableEpochsStakingV4(cfg)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), errNoMaxNodesConfigChangeForStakingV4.Error()))
		require.True(t, strings.Contains(err.Error(), "6"))
	})

	t.Run("max nodes config change for StakingV4Step3EnableEpoch has no previous config change, should return error", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()
		cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch = []MaxNodesChangeConfig{
			{
				EpochEnable:            cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch,
				MaxNumNodes:            48,
				NodesToShufflePerShard: 2,
			},
			{
				EpochEnable:            444,
				MaxNumNodes:            56,
				NodesToShufflePerShard: 2,
			},
		}

		err := SanityCheckEnableEpochsStakingV4(cfg)
		require.NotNil(t, err)
		require.ErrorIs(t, err, errNoMaxNodesConfigBeforeStakingV4)
	})

	t.Run("stakingV4 config for max nodes changed with different nodes to shuffle, should return error", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()
		cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[1].NodesToShufflePerShard = 2
		cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].NodesToShufflePerShard = 4

		err := SanityCheckEnableEpochsStakingV4(cfg)
		require.ErrorIs(t, err, errMismatchNodesToShuffle)
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

func TestSanityCheckNodesConfig(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cfg := generateCorrectConfig()
		err := SanityCheckNodesConfig(&NodesSetupStub{

			NumberOfShardsCalled: func() uint32 {
				return 3
			},
			MinNumberOfMetaNodesCalled: func() uint32 {
				return 5
			},
			MinNumberOfShardNodesCalled: func() uint32 {
				return 5
			},
			GetHysteresisCalled: func() float32 {
				return 0.2
			},
			MinNumberOfNodesWithHysteresisCalled: func() uint32 {
				return 5*4 + uint32(float32(5)*0.2) + uint32(float32(5)*0.2*float32(3))
			},
		}, cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch)

		require.Nil(t, err)
	})
}
