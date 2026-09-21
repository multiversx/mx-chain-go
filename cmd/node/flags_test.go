package main

import (
	"fmt"
	"math"
	"testing"

	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"

	"github.com/multiversx/mx-chain-go/config"
)

func TestStartInEpochOffsetFlag(t *testing.T) {
	t.Parallel()

	t.Run("default is zero", func(t *testing.T) {
		configs, err := applyNodeFlagsForTest(t)

		require.NoError(t, err)
		require.Zero(t, configs.FlagsConfig.StartInEpochOffset)
		require.False(t, configs.GeneralConfig.GeneralSettings.StartInEpochEnabled)
	})

	t.Run("one enables start in epoch", func(t *testing.T) {
		configs, err := applyNodeFlagsForTest(t, "--start-in-epoch=false", "--start-in-epoch-offset=1")

		require.NoError(t, err)
		require.Equal(t, uint32(1), configs.FlagsConfig.StartInEpochOffset)
		require.True(t, configs.GeneralConfig.GeneralSettings.StartInEpochEnabled)
	})

	for _, offset := range []uint32{2, 25, math.MaxUint32} {
		t.Run(fmt.Sprintf("offset %d enables network bootstrap", offset), func(t *testing.T) {
			configs, err := applyNodeFlagsForTest(t, "--start-in-epoch=false", fmt.Sprintf("--start-in-epoch-offset=%d", offset))
			require.NoError(t, err)
			require.Equal(t, offset, configs.FlagsConfig.StartInEpochOffset)
			require.True(t, configs.GeneralConfig.GeneralSettings.StartInEpochEnabled)
		})
	}

	t.Run("values above uint32 are rejected before truncation", func(t *testing.T) {
		configs, err := applyNodeFlagsForTest(t, "--start-in-epoch-offset=4294967296")
		require.ErrorContains(t, err, "start-in-epoch-offset must fit in uint32")
		require.Nil(t, configs.FlagsConfig)
	})

	t.Run("negative values are rejected by the unsigned flag", func(t *testing.T) {
		_, err := applyNodeFlagsForTest(t, "--start-in-epoch-offset=-1")

		require.Error(t, err)
	})

	t.Run("import db is rejected", func(t *testing.T) {
		_, err := applyNodeFlagsForTest(t, "--start-in-epoch-offset=25", "--import-db=import-dir")

		require.ErrorContains(t, err, "start-in-epoch-offset cannot be used with import-db")
	})

}

func TestStartInEpochOffsetWithFullArchive(t *testing.T) {
	t.Parallel()

	for _, source := range []struct {
		name       string
		args       []string
		fromConfig bool
	}{
		{name: "flag", args: []string{"--full-archive"}},
		{name: "operation mode", args: []string{"--operation-mode=full-archive"}},
		{name: "preferences config", fromConfig: true},
	} {
		for _, offset := range []uint32{0, 2, 25} {
			t.Run(fmt.Sprintf("%s offset %d", source.name, offset), func(t *testing.T) {
				configs := &config.Configs{
					GeneralConfig:            &config.Config{},
					PreferencesConfig:        &config.Preferences{},
					ConfigurationPathsHolder: &config.ConfigurationPathsHolder{},
				}
				configs.PreferencesConfig.Preferences.FullArchive = source.fromConfig
				configs.GeneralConfig.StoragePruning.ValidatorCleanOldEpochsData = true
				configs.GeneralConfig.StoragePruning.ObserverCleanOldEpochsData = true
				args := append([]string{"--start-in-epoch", fmt.Sprintf("--start-in-epoch-offset=%d", offset)}, source.args...)

				err := applyNodeFlagsToConfigForTest(configs, args...)

				require.NoError(t, err)
				require.True(t, configs.PreferencesConfig.Preferences.FullArchive)
				require.Equal(t, offset, configs.FlagsConfig.StartInEpochOffset)
				require.Equal(t, offset > 0, configs.GeneralConfig.GeneralSettings.StartInEpochEnabled)
				require.False(t, configs.GeneralConfig.StoragePruning.ValidatorCleanOldEpochsData)
				require.False(t, configs.GeneralConfig.StoragePruning.ObserverCleanOldEpochsData)
				require.True(t, configs.GeneralConfig.StoragePruning.Enabled)
			})
		}
	}
}

func applyNodeFlagsForTest(t *testing.T, args ...string) (*config.Configs, error) {
	t.Helper()

	configs := &config.Configs{
		GeneralConfig:            &config.Config{},
		PreferencesConfig:        &config.Preferences{},
		ConfigurationPathsHolder: &config.ConfigurationPathsHolder{},
	}
	return configs, applyNodeFlagsToConfigForTest(configs, args...)
}

func applyNodeFlagsToConfigForTest(configs *config.Configs, args ...string) error {
	app := cli.NewApp()
	app.Flags = getFlags()
	app.Action = func(ctx *cli.Context) error {
		flagsConfig := getFlagsConfig(ctx, logger.GetOrCreate("flags-test"))
		return applyFlags(ctx, configs, flagsConfig, logger.GetOrCreate("flags-test"))
	}

	return app.Run(append([]string{"node"}, args...))
}
