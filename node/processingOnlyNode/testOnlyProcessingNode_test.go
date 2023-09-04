package processingOnlyNode

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/assert"
)

const pathForMainConfig = "../../cmd/node/config/config.toml"
const pathForEconomicsConfig = "../../cmd/node/config/economics.toml"
const pathForGasSchedules = "../../cmd/node/config/gasSchedules"

func createMockArgsTestOnlyProcessingNode(t *testing.T) ArgsTestOnlyProcessingNode {
	mainConfig := config.Config{}
	err := LoadConfigFromFile(pathForMainConfig, &mainConfig)
	assert.Nil(t, err)

	economicsConfig := config.EconomicsConfig{}
	err = LoadConfigFromFile(pathForEconomicsConfig, &economicsConfig)
	assert.Nil(t, err)

	gasScheduleName, err := GetLatestGasScheduleFilename(pathForGasSchedules)
	assert.Nil(t, err)

	return ArgsTestOnlyProcessingNode{
		Config:              mainConfig,
		EnableEpochsConfig:  config.EnableEpochs{},
		EconomicsConfig:     economicsConfig,
		GasScheduleFilename: gasScheduleName,
		NumShards:           0,
		ShardID:             3,
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
