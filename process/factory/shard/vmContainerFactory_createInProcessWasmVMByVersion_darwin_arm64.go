//go:build darwin && arm64

package shard

import (
	"github.com/multiversx/mx-chain-go/config"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

func (vmf *vmContainerFactory) createInProcessWasmVMByVersion(version config.WasmVMVersionByEpoch) (vmcommon.VMExecutionHandler, error) {
	logVMContainerFactory.Debug("createInProcessWasmVMByVersion (darwin && arm64)", "version", version)
	switch version.Version {
	default:
		return vmf.createInProcessWasmVMV15()
	}
}
