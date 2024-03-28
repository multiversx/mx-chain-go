package notifier

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	epochNotifierMock "github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/stretchr/testify/require"
)

func getEnableEpochCfg() config.EnableEpochs {
	return config.EnableEpochs{
		StakingV4Step1EnableEpoch: 2,
		StakingV4Step2EnableEpoch: 3,
		StakingV4Step3EnableEpoch: 4,
		MaxNodesChangeEnableEpoch: []config.MaxNodesChangeConfig{
			{
				EpochEnable:            0,
				MaxNumNodes:            36,
				NodesToShufflePerShard: 4,
			},
			{
				EpochEnable:            1,
				MaxNumNodes:            64,
				NodesToShufflePerShard: 2,
			},
			{
				EpochEnable:            4,
				MaxNumNodes:            56,
				NodesToShufflePerShard: 2,
			},
		},
	}
}

func TestNewNodesConfigProviderAPI(t *testing.T) {
	t.Parallel()

	t.Run("nil epoch notifier, should return error", func(t *testing.T) {
		ncp, err := NewNodesConfigProviderAPI(nil, config.EnableEpochs{})
		require.Equal(t, process.ErrNilEpochNotifier, err)
		require.Nil(t, ncp)
	})

	t.Run("no nodes config for staking v4 step 3, should return error", func(t *testing.T) {
		ncp, err := NewNodesConfigProviderAPI(&epochNotifierMock.EpochNotifierStub{}, config.EnableEpochs{})
		require.ErrorIs(t, err, errNoMaxNodesConfigChangeForStakingV4)
		require.Nil(t, ncp)
	})

	t.Run("should work", func(t *testing.T) {
		ncp, err := NewNodesConfigProviderAPI(&epochNotifierMock.EpochNotifierStub{}, getEnableEpochCfg())
		require.Nil(t, err)
		require.False(t, ncp.IsInterfaceNil())
	})
}

func TestNodesConfigProviderAPI_GetCurrentNodesConfig(t *testing.T) {
	t.Parallel()

	epochNotifier := forking.NewGenericEpochNotifier()
	enableEpochCfg := getEnableEpochCfg()
	ncp, _ := NewNodesConfigProviderAPI(epochNotifier, enableEpochCfg)

	maxNodesConfig1 := enableEpochCfg.MaxNodesChangeEnableEpoch[0]
	maxNodesConfig2 := enableEpochCfg.MaxNodesChangeEnableEpoch[1]
	maxNodesConfigStakingV4Step3 := enableEpochCfg.MaxNodesChangeEnableEpoch[2]

	require.Equal(t, maxNodesConfig1, ncp.GetCurrentNodesConfig())

	epochNotifier.CheckEpoch(&block.Header{Epoch: enableEpochCfg.StakingV4Step1EnableEpoch})
	require.Equal(t, maxNodesConfig2, ncp.GetCurrentNodesConfig())

	epochNotifier.CheckEpoch(&block.Header{Epoch: enableEpochCfg.StakingV4Step2EnableEpoch})
	require.Equal(t, maxNodesConfigStakingV4Step3, ncp.GetCurrentNodesConfig())

	epochNotifier.CheckEpoch(&block.Header{Epoch: enableEpochCfg.StakingV4Step3EnableEpoch})
	require.Equal(t, maxNodesConfigStakingV4Step3, ncp.GetCurrentNodesConfig())

	epochNotifier.CheckEpoch(&block.Header{Epoch: enableEpochCfg.StakingV4Step3EnableEpoch + 1})
	require.Equal(t, maxNodesConfigStakingV4Step3, ncp.GetCurrentNodesConfig())

	// simulate restart
	epochNotifier.CheckEpoch(&block.Header{Epoch: 0})
	epochNotifier.CheckEpoch(&block.Header{Epoch: enableEpochCfg.StakingV4Step2EnableEpoch})
	require.Equal(t, maxNodesConfigStakingV4Step3, ncp.GetCurrentNodesConfig())

	// simulate restart
	epochNotifier.CheckEpoch(&block.Header{Epoch: 0})
	epochNotifier.CheckEpoch(&block.Header{Epoch: enableEpochCfg.StakingV4Step3EnableEpoch})
	require.Equal(t, maxNodesConfigStakingV4Step3, ncp.GetCurrentNodesConfig())
}
