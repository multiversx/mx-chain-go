package components

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	pathTestData           = "../testdata/"
	pathToConfigFolder     = "../../../cmd/node/config/"
	pathForMainConfig      = "../../../cmd/node/config/config.toml"
	pathForEconomicsConfig = "../../../cmd/node/config/economics.toml"
	pathForGasSchedules    = "../../../cmd/node/config/gasSchedules"
	nodesSetupConfig       = "../../../cmd/node/config/nodesSetup.json"
	pathForPrefsConfig     = "../../../cmd/node/config/prefs.toml"
	validatorPemFile       = "../../../cmd/node/config/testKeys/validatorKey.pem"
	pathSystemSCConfig     = "../../../cmd/node/config/systemSmartContractsConfig.toml"
)

func createMockArgsTestOnlyProcessingNode(t *testing.T) ArgsTestOnlyProcessingNode {
	mainConfig := config.Config{}
	err := LoadConfigFromFile(pathForMainConfig, &mainConfig)
	assert.Nil(t, err)

	economicsConfig := config.EconomicsConfig{}
	err = LoadConfigFromFile(pathForEconomicsConfig, &economicsConfig)
	assert.Nil(t, err)

	gasScheduleName, err := configs.GetLatestGasScheduleFilename(pathForGasSchedules)
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
		NumShards:              3,
		ShardID:                0,
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
			Nodes:                    nodesSetupConfig,
			ValidatorKey:             validatorPemFile,
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
		if testing.Short() {
			t.Skip("cannot run with -race -short; requires Wasm VM fix")
		}

		t.Parallel()

		args := createMockArgsTestOnlyProcessingNode(t)
		node, err := NewTestOnlyProcessingNode(args)
		assert.Nil(t, err)
		assert.NotNil(t, node)
	})

	t.Run("try commit a block", func(t *testing.T) {
		if testing.Short() {
			t.Skip("cannot run with -race -short; requires Wasm VM fix")
		}

		t.Parallel()

		args := createMockArgsTestOnlyProcessingNode(t)
		node, err := NewTestOnlyProcessingNode(args)
		assert.Nil(t, err)
		assert.NotNil(t, node)

		newHeader, err := node.ProcessComponentsHolder.BlockProcessor().CreateNewHeader(1, 1)
		assert.Nil(t, err)

		err = newHeader.SetPrevHash(node.ChainHandler.GetGenesisHeaderHash())
		assert.Nil(t, err)

		header, block, err := node.ProcessComponentsHolder.BlockProcessor().CreateBlock(newHeader, func() bool {
			return true
		})
		assert.Nil(t, err)
		require.NotNil(t, header)
		require.NotNil(t, block)

		err = node.ProcessComponentsHolder.BlockProcessor().ProcessBlock(header, block, func() time.Duration {
			return 1000
		})
		assert.Nil(t, err)

		err = node.ProcessComponentsHolder.BlockProcessor().CommitBlock(header, block)
		assert.Nil(t, err)
	})
}
