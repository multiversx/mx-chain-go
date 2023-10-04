package chainSimulator

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/assert"
)

const (
	pathTestData           = "./testdata/"
	pathToConfigFolder     = "../../cmd/node/config/"
	pathForMainConfig      = "../../cmd/node/config/config.toml"
	pathForEconomicsConfig = "../../cmd/node/config/economics.toml"
	pathForGasSchedules    = "../../cmd/node/config/gasSchedules"
	nodesSetupConfig       = "../../cmd/node/config/nodesSetup.json"
	pathForPrefsConfig     = "../../cmd/node/config/prefs.toml"
	validatorPemFile       = "../../cmd/node/config/testKeys/validatorKey.pem"
	pathSystemSCConfig     = "../../cmd/node/config/systemSmartContractsConfig.toml"
)

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

	systemSCConfig := config.SystemSmartContractsConfig{}
	err = LoadConfigFromFile(pathSystemSCConfig, &systemSCConfig)
	assert.Nil(t, err)

	workingDir := t.TempDir()

	epochConfig := config.EpochConfig{}
	err = LoadConfigFromFile(pathToConfigFolder+"enableEpochs.toml", &epochConfig)
	assert.Nil(t, err)

	return ArgsTestOnlyProcessingNode{
		Config:      mainConfig,
		WorkingDir:  workingDir,
		EpochConfig: epochConfig,
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
			Version:    "1",
		},
		ConfigurationPathsHolder: config.ConfigurationPathsHolder{
			GasScheduleDirectoryName: pathToConfigFolder + "gasSchedules",
			Genesis:                  pathToConfigFolder + "genesis.json",
			SmartContracts:           pathTestData + "genesisSmartContracts.json",
		},
		SystemSCConfig:      systemSCConfig,
		ChanStopNodeProcess: make(chan endProcess.ArgEndProcess),
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
