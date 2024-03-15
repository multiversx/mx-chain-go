//go:build arm64

package node

import (
	"runtime"

	"github.com/multiversx/mx-chain-go/config"
)

// ApplyArchCustomConfigs will apply configuration tweaks based on the architecture the node is running on
func ApplyArchCustomConfigs(configs *config.Configs) {
	log.Debug("ApplyArchCustomConfigs", "architecture", runtime.GOARCH)

	firstSupportedWasmer2VMVersion := "v1.5"
	log.Debug("ApplyArchCustomConfigs - hardcoding the initial VM to " + firstSupportedWasmer2VMVersion)
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
