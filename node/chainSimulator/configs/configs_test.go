package configs

import (
	"testing"

	"github.com/multiversx/mx-chain-go/integrationTests/realcomponents"
	"github.com/stretchr/testify/require"
)

func TestNewProcessorRunnerChainArguments(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	outputConfig, err := CreateChainSimulatorConfigs(ArgsChainSimulatorConfigs{
		NumOfShards:           3,
		OriginalConfigsPath:   "../../../cmd/node/config",
		RoundDurationInMillis: 6000,
		GenesisTimeStamp:      0,
		TempDir:               t.TempDir(),
	})
	require.Nil(t, err)

	pr := realcomponents.NewProcessorRunner(t, *outputConfig.Configs)
	pr.Close(t)
}
