//go:build !(darwin && arm64)

package shard

import (
	"github.com/multiversx/mx-chain-go/config"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

func (vmf *vmContainerFactory) createInProcessWasmVMByVersion(version config.WasmVMVersionByEpoch) (vmcommon.VMExecutionHandler, error) {
	logVMContainerFactory.Debug("createInProcessWasmVMByVersion !(darwin && arm64)", "version", version)
	switch version.Version {
	case "v1.2":
		return vmf.createInProcessWasmVMV12()
	case "v1.3":
		return vmf.createInProcessWasmVMV13()
	case "v1.4":
		return vmf.createInProcessWasmVMV14()
	default:
		return vmf.createInProcessWasmVMV15()
	}
}
