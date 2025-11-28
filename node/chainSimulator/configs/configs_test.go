package configs

import (
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
