package overridableConfig

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	p2pConfig "github.com/multiversx/mx-chain-go/p2p/config"

	"github.com/stretchr/testify/require"
)

func TestOverrideConfigValues(t *testing.T) {
	t.Parallel()

	t.Run("empty overridable config, should not error", func(t *testing.T) {
		t.Parallel()

		err := OverrideConfigValues([]config.OverridableConfig{}, &config.Configs{})
		require.NoError(t, err)
	})

	t.Run("invalid config file, should error", func(t *testing.T) {
		t.Parallel()

		err := OverrideConfigValues([]config.OverridableConfig{{File: "invalid.toml"}}, &config.Configs{})
		availableOptionsString := "api.toml,config.toml,economics.toml,enableEpochs.toml,enableRounds.toml,external.toml,fullArchiveP2P.toml,p2p.toml,ratings.toml,systemSmartContractsConfig.toml"
		require.Equal(t, "invalid config file <invalid.toml>. Available options are "+availableOptionsString, err.Error())
	})

	t.Run("nil config, should error", func(t *testing.T) {
		t.Parallel()

		err := OverrideConfigValues([]config.OverridableConfig{{Path: "test", File: "config.toml"}}, &config.Configs{GeneralConfig: nil})
		require.Equal(t, "nil structure to update", err.Error())
	})

	t.Run("should work for config.toml", func(t *testing.T) {
		t.Parallel()

		configs := &config.Configs{GeneralConfig: &config.Config{}}

		err := OverrideConfigValues([]config.OverridableConfig{{Path: "BootstrapStorage.DB.FilePath", Value: "new path", File: "config.toml"}}, configs)
		require.NoError(t, err)
		require.Equal(t, "new path", configs.GeneralConfig.BootstrapStorage.DB.FilePath)
	})

	t.Run("should work for p2p.toml", func(t *testing.T) {
		t.Parallel()

		configs := &config.Configs{MainP2pConfig: &p2pConfig.P2PConfig{Sharding: p2pConfig.ShardingConfig{TargetPeerCount: 5}}}

		err := OverrideConfigValues([]config.OverridableConfig{{Path: "Sharding.TargetPeerCount", Value: uint32(37), File: "p2p.toml"}}, configs)
		require.NoError(t, err)
		require.Equal(t, uint32(37), configs.MainP2pConfig.Sharding.TargetPeerCount)
	})

	t.Run("should work for fullArchiveP2P.toml", func(t *testing.T) {
		t.Parallel()

		configs := &config.Configs{FullArchiveP2pConfig: &p2pConfig.P2PConfig{Sharding: p2pConfig.ShardingConfig{TargetPeerCount: 5}}}

		err := OverrideConfigValues([]config.OverridableConfig{{Path: "Sharding.TargetPeerCount", Value: uint32(37), File: "fullArchiveP2P.toml"}}, configs)
		require.NoError(t, err)
		require.Equal(t, uint32(37), configs.FullArchiveP2pConfig.Sharding.TargetPeerCount)
	})

	t.Run("should work for external.toml", func(t *testing.T) {
		t.Parallel()

		configs := &config.Configs{ExternalConfig: &config.ExternalConfig{ElasticSearchConnector: config.ElasticSearchConfig{Password: "old pass"}}}

		err := OverrideConfigValues([]config.OverridableConfig{{Path: "ElasticSearchConnector.Password", Value: "new pass", File: "external.toml"}}, configs)
		require.NoError(t, err)
		require.Equal(t, "new pass", configs.ExternalConfig.ElasticSearchConnector.Password)
	})

	t.Run("should work for enableEpochs.toml", func(t *testing.T) {
		t.Parallel()

		configs := &config.Configs{EpochConfig: &config.EpochConfig{EnableEpochs: config.EnableEpochs{ESDTMetadataContinuousCleanupEnableEpoch: 5}}}

		err := OverrideConfigValues([]config.OverridableConfig{{Path: "EnableEpochs.ESDTMetadataContinuousCleanupEnableEpoch", Value: uint32(37), File: "enableEpochs.toml"}}, configs)
		require.NoError(t, err)
		require.Equal(t, uint32(37), configs.EpochConfig.EnableEpochs.ESDTMetadataContinuousCleanupEnableEpoch)
	})

	t.Run("should work for api.toml", func(t *testing.T) {
		t.Parallel()

		configs := &config.Configs{ApiRoutesConfig: &config.ApiRoutesConfig{}}

		err := OverrideConfigValues([]config.OverridableConfig{{Path: "Logging.LoggingEnabled", Value: true, File: "api.toml"}}, configs)
		require.NoError(t, err)
		require.True(t, configs.ApiRoutesConfig.Logging.LoggingEnabled)
	})

	t.Run("should work for economics.toml", func(t *testing.T) {
		t.Parallel()

		configs := &config.Configs{EconomicsConfig: &config.EconomicsConfig{}}

		err := OverrideConfigValues([]config.OverridableConfig{{Path: "GlobalSettings.GenesisTotalSupply", Value: "37", File: "economics.toml"}}, configs)
		require.NoError(t, err)
		require.Equal(t, "37", configs.EconomicsConfig.GlobalSettings.GenesisTotalSupply)
	})

	t.Run("should work for enableRounds.toml", func(t *testing.T) {
		t.Parallel()

		configs := &config.Configs{RoundConfig: &config.RoundConfig{}}
		value := make(map[string]config.ActivationRoundByName)
		value["DisableAsyncCallV1"] = config.ActivationRoundByName{Round: "37"}

		err := OverrideConfigValues([]config.OverridableConfig{{Path: "RoundActivations", Value: value, File: "enableRounds.toml"}}, configs)
		require.NoError(t, err)
		require.Equal(t, "37", configs.RoundConfig.RoundActivations["DisableAsyncCallV1"].Round)
	})

	t.Run("should work for ratings.toml", func(t *testing.T) {
		t.Parallel()

		configs := &config.Configs{RatingsConfig: &config.RatingsConfig{}}

		err := OverrideConfigValues([]config.OverridableConfig{{Path: "General.StartRating", Value: 37, File: "ratings.toml"}}, configs)
		require.NoError(t, err)
		require.Equal(t, uint32(37), configs.RatingsConfig.General.StartRating)
	})

	t.Run("should work for systemSmartContractsConfig.toml", func(t *testing.T) {
		t.Parallel()

		configs := &config.Configs{SystemSCConfig: &config.SystemSmartContractsConfig{}}

		err := OverrideConfigValues([]config.OverridableConfig{{Path: "StakingSystemSCConfig.UnBondPeriod", Value: 37, File: "systemSmartContractsConfig.toml"}}, configs)
		require.NoError(t, err)
		require.Equal(t, uint64(37), configs.SystemSCConfig.StakingSystemSCConfig.UnBondPeriod)
	})

	t.Run("should work for go struct overwrite", func(t *testing.T) {
		t.Parallel()

		configs := &config.Configs{
			GeneralConfig: &config.Config{
				VirtualMachine: config.VirtualMachineServicesConfig{
					Execution: config.VirtualMachineConfig{
						WasmVMVersions: []config.WasmVMVersionByEpoch{
							{StartEpoch: 0, Version: "1.3"},
							{StartEpoch: 1, Version: "1.4"},
							{StartEpoch: 2, Version: "1.5"},
						},
					},
				},
			},
		}
		require.Equal(t, len(configs.GeneralConfig.VirtualMachine.Execution.WasmVMVersions), 3)

		newWasmVMVersion := []config.WasmVMVersionByEpoch{
			{StartEpoch: 0, Version: "1.5"},
		}

		err := OverrideConfigValues([]config.OverridableConfig{{Path: "VirtualMachine.Execution.WasmVMVersions", Value: newWasmVMVersion, File: "config.toml"}}, configs)
		require.NoError(t, err)
		require.Equal(t, len(configs.GeneralConfig.VirtualMachine.Execution.WasmVMVersions), 1)
		require.Equal(t, newWasmVMVersion, configs.GeneralConfig.VirtualMachine.Execution.WasmVMVersions)
	})
}
