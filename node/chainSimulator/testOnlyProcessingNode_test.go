package chainSimulator

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/assert"
)

const pathForMainConfig = "../../cmd/node/config/config.toml"
const pathForEconomicsConfig = "../../cmd/node/config/economics.toml"
const pathForGasSchedules = "../../cmd/node/config/gasSchedules"
const nodesSetupConfig = "../../cmd/node/config/nodesSetup.json"
const pathForPrefsConfig = "../../cmd/node/config/prefs.toml"
const validatorPemFile = "../../cmd/node/config/testKeys/validatorKey.pem"

func createMockArgsTestOnlyProcessingNode(t *testing.T) ArgsTestOnlyProcessingNode {
	mainConfig := config.Config{}
	err := LoadConfigFromFile(pathForMainConfig, &mainConfig)
	assert.Nil(t, err)

	economicsConfig := config.EconomicsConfig{}
	err = LoadConfigFromFile(pathForEconomicsConfig, &economicsConfig)
	assert.Nil(t, err)

	gasScheduleName, err := GetLatestGasScheduleFilename(pathForGasSchedules)
	assert.Nil(t, err)

	prefsConfig := config.Preferences{}
	err = LoadConfigFromFile(pathForPrefsConfig, &prefsConfig)
	assert.Nil(t, err)

	workingDir := t.TempDir()

	return ArgsTestOnlyProcessingNode{
		Config:     mainConfig,
		WorkingDir: workingDir,
		EnableEpochsConfig: config.EnableEpochs{
			BLSMultiSignerEnableEpoch: []config.MultiSignerConfig{
				{EnableEpoch: 0, Type: "KOSK"},
			},
		},
		RoundsConfig: config.RoundConfig{
			RoundActivations: map[string]config.ActivationRoundByName{
				"DisableAsyncCallV1": {
					Round: "18446744073709551614",
				},
			},
		},
		EconomicsConfig:        economicsConfig,
		GasScheduleFilename:    gasScheduleName,
		NodesSetupPath:         nodesSetupConfig,
		NumShards:              3,
		ShardID:                0,
		ValidatorPemFile:       validatorPemFile,
		PreferencesConfig:      prefsConfig,
		SyncedBroadcastNetwork: NewSyncedBroadcastNetwork(),
		ImportDBConfig:         config.ImportDbConfig{},
		ContextFlagsConfig: config.ContextFlagsConfig{
			WorkingDir: workingDir,
		},
	}
}

func TestNewTestOnlyProcessingNode(t *testing.T) {
	t.Parallel()

	t.Run("invalid shard configuration should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTestOnlyProcessingNode(t)
		args.ShardID = args.NumShards
		node, err := NewTestOnlyProcessingNode(args)
		assert.NotNil(t, err)
		assert.Nil(t, node)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTestOnlyProcessingNode(t)
		node, err := NewTestOnlyProcessingNode(args)
		assert.Nil(t, err)
		assert.NotNil(t, node)
	})
}
