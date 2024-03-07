//go:build arm64

package node

import (
	"runtime"

	"github.com/multiversx/mx-chain-go/config"
)

func applyArchCustomConfigs(configs *config.Configs) {
	log.Debug("applyArchCustomConfigs", "architecture", runtime.GOARCH)

	firstSupportedWasmer2VMVersion := "v1.5"
	log.Debug("applyArchCustomConfigs - hardcoding the initial VM to " + firstSupportedWasmer2VMVersion)
	configs.GeneralConfig.VirtualMachine.Execution.WasmVMVersions = []config.WasmVMVersionByEpoch{
		{
			StartEpoch: 0,
			Version:    firstSupportedWasmer2VMVersion,
		},
	}
	configs.GeneralConfig.VirtualMachine.Querying.WasmVMVersions = []config.WasmVMVersionByEpoch{
		{
			StartEpoch: 0,
			Version:    firstSupportedWasmer2VMVersion,
		},
	}
}
