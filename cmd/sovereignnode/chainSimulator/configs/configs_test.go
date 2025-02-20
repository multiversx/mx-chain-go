package configs

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
)

func TestCreateChainSimulatorConfigsWithSovereignGenesisFile(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	outputConfig, err := configs.CreateChainSimulatorConfigs(configs.ArgsChainSimulatorConfigs{
		NumOfShards:           1,
		OriginalConfigsPath:   "../../../../cmd/node/config",
		RoundDurationInMillis: 6000,
		GenesisTimeStamp:      0,
		TempDir:               t.TempDir(),
		MinNodesPerShard:      1,
		GenerateGenesisFile:   GenerateSovereignGenesisFile,
	})
	require.Nil(t, err)
	require.NotNil(t, outputConfig)
}
