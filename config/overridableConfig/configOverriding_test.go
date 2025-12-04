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

	t.Run("should work for HostDriversConfig partial override - smart merge", func(t *testing.T) {
		t.Parallel()

		// Set up existing config with all default values
		configs := &config.Configs{
			ExternalConfig: &config.ExternalConfig{
				HostDriversConfig: []config.HostDriversConfig{
					{
						Enabled:                    false,
						Mode:                       "client",
						URL:                        "ws://127.0.0.1:22111",
						WithAcknowledge:            true,
						AcknowledgeTimeoutInSec:    60,
						MarshallerType:             "json",
						RetryDurationInSec:         5,
						BlockingAckOnError:         true,
						DropMessagesIfNoConnection: false,
						Version:                    1,
					},
				},
			},
		}

		// Partial override with only the fields we want to change
		partialOverride := []interface{}{
			map[string]interface{}{
				"URL":            "ws://new-url:8080",
				"Enabled":        true,
				"MarshallerType": "gogo protobuf",
			},
		}

		err := OverrideConfigValues([]config.OverridableConfig{{
			Path:  "HostDriversConfig",
			Value: partialOverride,
			File:  "external.toml",
		}}, configs)

		require.NoError(t, err)
		require.Len(t, configs.ExternalConfig.HostDriversConfig, 1)

		result := configs.ExternalConfig.HostDriversConfig[0]
		// Check that overridden fields are set correctly
		require.Equal(t, "ws://new-url:8080", result.URL)
		require.True(t, result.Enabled)
		require.Equal(t, "gogo protobuf", result.MarshallerType)

		// Check that non-overridden fields are preserved
		require.Equal(t, "client", result.Mode)
		require.True(t, result.WithAcknowledge)
		require.Equal(t, 60, result.AcknowledgeTimeoutInSec)
		require.Equal(t, 5, result.RetryDurationInSec)
		require.True(t, result.BlockingAckOnError)
		require.False(t, result.DropMessagesIfNoConnection)
		require.Equal(t, uint32(1), result.Version)
	})

	t.Run("should work for HostDriversConfig multiple elements partial override", func(t *testing.T) {
		t.Parallel()

		// Set up existing config with multiple elements
		configs := &config.Configs{
			ExternalConfig: &config.ExternalConfig{
				HostDriversConfig: []config.HostDriversConfig{
					{
						Enabled:                    false,
						Mode:                       "client",
						URL:                        "ws://127.0.0.1:22111",
						WithAcknowledge:            true,
						AcknowledgeTimeoutInSec:    60,
						MarshallerType:             "json",
						RetryDurationInSec:         5,
						BlockingAckOnError:         true,
						DropMessagesIfNoConnection: false,
						Version:                    1,
					},
					{
						Enabled:                    false,
						Mode:                       "server",
						URL:                        "ws://127.0.0.1:22112",
						WithAcknowledge:            false,
						AcknowledgeTimeoutInSec:    30,
						MarshallerType:             "protobuf",
						RetryDurationInSec:         10,
						BlockingAckOnError:         false,
						DropMessagesIfNoConnection: true,
						Version:                    2,
					},
				},
			},
		}

		// Override only the first element partially
		partialOverride := []interface{}{
			map[string]interface{}{
				"Enabled": true,
				"URL":     "ws://new-url:8080",
			},
		}

		err := OverrideConfigValues([]config.OverridableConfig{{
			Path:  "HostDriversConfig",
			Value: partialOverride,
			File:  "external.toml",
		}}, configs)

		require.NoError(t, err)
		require.Len(t, configs.ExternalConfig.HostDriversConfig, 2)

		// First element should have overridden fields + preserved fields
		first := configs.ExternalConfig.HostDriversConfig[0]
		require.True(t, first.Enabled)                     // overridden
		require.Equal(t, "ws://new-url:8080", first.URL)   // overridden
		require.Equal(t, "client", first.Mode)             // preserved
		require.True(t, first.WithAcknowledge)             // preserved
		require.Equal(t, "json", first.MarshallerType)     // preserved

		// Second element should be completely unchanged
		second := configs.ExternalConfig.HostDriversConfig[1]
		require.False(t, second.Enabled)
		require.Equal(t, "server", second.Mode)
		require.Equal(t, "ws://127.0.0.1:22112", second.URL)
		require.False(t, second.WithAcknowledge)
		require.Equal(t, 30, second.AcknowledgeTimeoutInSec)
		require.Equal(t, "protobuf", second.MarshallerType)
		require.Equal(t, 10, second.RetryDurationInSec)
		require.False(t, second.BlockingAckOnError)
		require.True(t, second.DropMessagesIfNoConnection)
		require.Equal(t, uint32(2), second.Version)
	})

	t.Run("should use smart merge for longer override arrays", func(t *testing.T) {
		t.Parallel()

		// Set up existing config with 1 element
		configs := &config.Configs{
			ExternalConfig: &config.ExternalConfig{
				HostDriversConfig: []config.HostDriversConfig{
					{
						Enabled:                    false,
						Mode:                       "client",
						URL:                        "ws://127.0.0.1:22111",
						WithAcknowledge:            true,
						AcknowledgeTimeoutInSec:    60,
						MarshallerType:             "json",
						RetryDurationInSec:         5,
						BlockingAckOnError:         true,
						DropMessagesIfNoConnection: false,
						Version:                    1,
					},
				},
			},
		}

		// Override with 2 elements (longer than existing) - should still use smart merge
		smartOverride := []interface{}{
			map[string]interface{}{
				"Enabled": true,
				"URL":     "ws://new-url-1:8080",
			},
			map[string]interface{}{
				"Enabled": false,
				"URL":     "ws://new-url-2:8081",
			},
		}

		err := OverrideConfigValues([]config.OverridableConfig{{
			Path:  "HostDriversConfig",
			Value: smartOverride,
			File:  "external.toml",
		}}, configs)

		require.NoError(t, err)
		require.Len(t, configs.ExternalConfig.HostDriversConfig, 2)

		// First element should merge with existing values
		first := configs.ExternalConfig.HostDriversConfig[0]
		require.True(t, first.Enabled)                      // overridden
		require.Equal(t, "ws://new-url-1:8080", first.URL)  // overridden
		require.Equal(t, "client", first.Mode)              // preserved from existing
		require.True(t, first.WithAcknowledge)              // preserved from existing
		require.Equal(t, 60, first.AcknowledgeTimeoutInSec) // preserved from existing
		require.Equal(t, "json", first.MarshallerType)      // preserved from existing
		require.Equal(t, 5, first.RetryDurationInSec)       // preserved from existing
		require.True(t, first.BlockingAckOnError)           // preserved from existing
		require.False(t, first.DropMessagesIfNoConnection)  // preserved from existing
		require.Equal(t, uint32(1), first.Version)          // preserved from existing

		// Second element should inherit from existing element + apply its own overrides
		second := configs.ExternalConfig.HostDriversConfig[1]
		require.False(t, second.Enabled)                     // overridden
		require.Equal(t, "ws://new-url-2:8081", second.URL)  // overridden
		require.Equal(t, "client", second.Mode)              // inherited from existing element
		require.True(t, second.WithAcknowledge)              // inherited from existing element
		require.Equal(t, 60, second.AcknowledgeTimeoutInSec) // inherited from existing element
		require.Equal(t, "json", second.MarshallerType)      // inherited from existing element
		require.Equal(t, 5, second.RetryDurationInSec)       // inherited from existing element
		require.True(t, second.BlockingAckOnError)           // inherited from existing element
		require.False(t, second.DropMessagesIfNoConnection)  // inherited from existing element
		require.Equal(t, uint32(1), second.Version)          // inherited from existing element
	})

	t.Run("should use zero values when starting with empty array", func(t *testing.T) {
		t.Parallel()

		// Set up existing config with empty array
		configs := &config.Configs{
			ExternalConfig: &config.ExternalConfig{
				HostDriversConfig: []config.HostDriversConfig{},
			},
		}

		// Override with 2 elements on empty array
		override := []interface{}{
			map[string]interface{}{
				"Enabled": true,
				"URL":     "ws://first-url:8080",
			},
			map[string]interface{}{
				"Enabled": false,
				"URL":     "ws://second-url:8081",
			},
		}

		err := OverrideConfigValues([]config.OverridableConfig{{
			Path:  "HostDriversConfig",
			Value: override,
			File:  "external.toml",
		}}, configs)

		require.NoError(t, err)
		require.Len(t, configs.ExternalConfig.HostDriversConfig, 2)

		// Both elements should have zero values + overrides (no existing element to inherit from)
		first := configs.ExternalConfig.HostDriversConfig[0]
		require.True(t, first.Enabled)                      // overridden
		require.Equal(t, "ws://first-url:8080", first.URL)  // overridden
		require.Equal(t, "", first.Mode)                    // zero value
		require.False(t, first.WithAcknowledge)             // zero value
		require.Equal(t, 0, first.AcknowledgeTimeoutInSec)  // zero value

		second := configs.ExternalConfig.HostDriversConfig[1]
		require.False(t, second.Enabled)                     // overridden
		require.Equal(t, "ws://second-url:8081", second.URL) // overridden
		require.Equal(t, "", second.Mode)                    // zero value
		require.False(t, second.WithAcknowledge)             // zero value
		require.Equal(t, 0, second.AcknowledgeTimeoutInSec)  // zero value
	})
}
