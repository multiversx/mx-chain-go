//go:build darwin && arm64

package node

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/assert"
)

func TestApplyArchCustomConfigs(t *testing.T) {
	t.Parallel()

	executionVMConfig := config.VirtualMachineConfig{
		WasmVMVersions: []config.WasmVMVersionByEpoch{
			{
				StartEpoch: 0,
				Version:    "v1.2",
			},
			{
				StartEpoch: 1,
				Version:    "v1.3",
			},
			{
				StartEpoch: 2,
				Version:    "v1.4",
			},
			{
				StartEpoch: 3,
				Version:    "v1.5",
			},
		},
		TimeOutForSCExecutionInMilliseconds: 1,
		WasmerSIGSEGVPassthrough:            true,
	}

	queryVMConfig := config.QueryVirtualMachineConfig{
		VirtualMachineConfig: executionVMConfig,
		NumConcurrentVMs:     15,
	}

	expectedVMWasmVersionsConfig := []config.WasmVMVersionByEpoch{
		{
			StartEpoch: 0,
			Version:    "v1.5",
		},
	}

	t.Run("providing a configuration should alter it", func(t *testing.T) {
		t.Parallel()

		providedConfigs := &config.Configs{
			GeneralConfig: &config.Config{
				VirtualMachine: config.VirtualMachineServicesConfig{
					Execution: executionVMConfig,
					Querying:  queryVMConfig,
				},
			},
		}

		expectedVMConfig := providedConfigs.GeneralConfig.VirtualMachine
		expectedVMConfig.Execution.WasmVMVersions = expectedVMWasmVersionsConfig
		expectedVMConfig.Querying.WasmVMVersions = expectedVMWasmVersionsConfig

		applyArchCustomConfigs(providedConfigs)

		assert.Equal(t, expectedVMConfig, providedConfigs.GeneralConfig.VirtualMachine)
	})
	t.Run("empty config should return an altered config", func(t *testing.T) {
		t.Parallel()

		providedConfigs := &config.Configs{
			GeneralConfig: &config.Config{},
		}

		expectedVMConfig := providedConfigs.GeneralConfig.VirtualMachine
		expectedVMConfig.Execution.WasmVMVersions = expectedVMWasmVersionsConfig
		expectedVMConfig.Querying.WasmVMVersions = expectedVMWasmVersionsConfig

		applyArchCustomConfigs(providedConfigs)

		expectedConfig := &config.Configs{
			GeneralConfig: &config.Config{
				VirtualMachine: expectedVMConfig,
			},
		}

		assert.Equal(t, expectedConfig, providedConfigs)
	})
}
