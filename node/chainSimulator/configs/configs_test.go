package configs

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/integrationTests/realcomponents"
	"github.com/multiversx/mx-chain-go/testscommon"

	"github.com/stretchr/testify/require"
)

func TestNewProcessorRunnerChainArguments(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	outputConfig, err := CreateChainSimulatorConfigs(ArgsChainSimulatorConfigs{
		NumOfShards:                    3,
		OriginalConfigsPath:            "../../../cmd/node/config",
		RoundDurationInMillis:          6000,
		SupernovaRoundDurationInMillis: 600,
		TempDir:                        t.TempDir(),
		MetaChainMinNodes:              1,
		MinNodesPerShard:               1,
		ConsensusGroupSize:             1,
		MetaChainConsensusGroupSize:    1,
	})
	require.Nil(t, err)

	pr := realcomponents.NewProcessorRunner(t, outputConfig.Configs)
	pr.Close(t)
}

func TestUpdateSupernovaConfigs(t *testing.T) {
	t.Parallel()

	configs, err := testscommon.CreateTestConfigs(t.TempDir(), "../../../cmd/node/config")
	require.Nil(t, err)

	chainSimulatorCfg := ArgsChainSimulatorConfigs{
		RoundsPerEpoch: core.OptionalUint64{
			Value:    20,
			HasValue: true,
		},
		SupernovaRoundsPerEpoch: core.OptionalUint64{
			Value:    200,
			HasValue: true,
		},
		SupernovaRoundDurationInMillis: 600,
	}

	updateSupernovaConfigs(configs, chainSimulatorCfg)
	require.Equal(t, uint64(600), configs.GeneralConfig.GeneralSettings.ChainParametersByEpoch[2].RoundDuration)
	require.Equal(t, configs.EpochConfig.EnableEpochs.SupernovaEnableEpoch, configs.GeneralConfig.GeneralSettings.ChainParametersByEpoch[2].EnableEpoch)
	require.Equal(t, "45", configs.RoundConfig.RoundActivations[string(common.SupernovaRoundFlag)].Round)
}

func TestUpdateSupernovaConfigs_NonZeroInitialRoundAndEpoch(t *testing.T) {
	t.Parallel()

	configs, err := testscommon.CreateTestConfigs(t.TempDir(), "../../../cmd/node/config")
	require.Nil(t, err)

	// resuming far into a real chain's history, two epochs before Supernova, with a much
	// lower RoundsPerEpoch than the real historical rate
	chainSimulatorCfg := ArgsChainSimulatorConfigs{
		InitialEpoch: 0,
		InitialRound: 31219386,
		RoundsPerEpoch: core.OptionalUint64{
			Value:    20,
			HasValue: true,
		},
		SupernovaRoundsPerEpoch: core.OptionalUint64{
			Value:    200,
			HasValue: true,
		},
		SupernovaRoundDurationInMillis: 600,
	}
	configs.EpochConfig.EnableEpochs.SupernovaEnableEpoch = 2

	updateSupernovaConfigs(configs, chainSimulatorCfg)

	// expected = InitialRound + (SupernovaEnableEpoch-InitialEpoch)*RoundsPerEpoch + numRoundsAfterSupernovaEnableEpoch
	expectedRound := chainSimulatorCfg.InitialRound + 2*20 + numRoundsAfterSupernovaEnableEpoch
	require.Equal(t, fmt.Sprintf("%d", expectedRound), configs.RoundConfig.RoundActivations[string(common.SupernovaRoundFlag)].Round)
	require.Equal(t, uint64(expectedRound), configs.GeneralConfig.Versions.VersionsByEpochs[2].StartRound)

	// must stay ahead of InitialRound, or round-based and epoch-based Supernova checks desync
	require.Greater(t, uint64(expectedRound), uint64(chainSimulatorCfg.InitialRound))
}

func TestUpdateSupernovaConfigs_UnsetSupernovaRoundsPerEpochDoesNotPreserveStaleManualRound(t *testing.T) {
	t.Parallel()

	configs, err := testscommon.CreateTestConfigs(t.TempDir(), "../../../cmd/node/config")
	require.Nil(t, err)

	// default config ships SupernovaEnableRound = "440"; unsigned subtraction must not wrap
	// around and mistake it for a valid manually pinned activation round
	configs.EpochConfig.EnableEpochs.SupernovaEnableEpoch = 2
	chainSimulatorCfg := ArgsChainSimulatorConfigs{
		InitialEpoch: 0,
		InitialRound: 0,
		RoundsPerEpoch: core.OptionalUint64{
			Value:    1,
			HasValue: true,
		},
	}

	updateSupernovaConfigs(configs, chainSimulatorCfg)

	expectedRound := chainSimulatorCfg.InitialRound + 2*1 + numRoundsAfterSupernovaEnableEpoch
	require.Equal(t, fmt.Sprintf("%d", expectedRound), configs.RoundConfig.RoundActivations[string(common.SupernovaRoundFlag)].Round)
}
