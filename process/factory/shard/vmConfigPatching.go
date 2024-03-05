//go:build !(darwin && arm64)

package shard

import "github.com/multiversx/mx-chain-go/config"

func patchVirtualMachineConfigGivenArchitecture(config *config.VirtualMachineConfig) {
}
