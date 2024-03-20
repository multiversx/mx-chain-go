//go:build !arm64

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

	t.Run("providing a configuration should not alter it", func(t *testing.T) {
		t.Parallel()

		providedConfigs := &config.Configs{
			GeneralConfig: &config.Config{
				VirtualMachine: config.VirtualMachineServicesConfig{
					Execution: executionVMConfig,
					Querying:  queryVMConfig,
				},
			},
		}

		ApplyArchCustomConfigs(providedConfigs)

		assert.Equal(t, executionVMConfig, providedConfigs.GeneralConfig.VirtualMachine.Execution)
		assert.Equal(t, queryVMConfig, providedConfigs.GeneralConfig.VirtualMachine.Querying)
	})
	t.Run("empty config should return an empty config", func(t *testing.T) {
		t.Parallel()

		// this test will prevent adding new config changes without handling them in this test
		providedConfigs := &config.Configs{
			GeneralConfig: &config.Config{},
		}
		emptyConfigs := &config.Configs{
			GeneralConfig: &config.Config{},
		}
		ApplyArchCustomConfigs(providedConfigs)

		assert.Equal(t, emptyConfigs, providedConfigs)
	})
}
