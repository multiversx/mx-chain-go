package integrationTests

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/sharding"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// TestBuiltinFunction wraps a builtin function defined ad-hoc, for testing
type TestBuiltinFunction struct {
	Function func(acntSnd, acntDst vmcommon.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error)
}

// ProcessBuiltinFunction is a method implementation required by the BuiltinFunction interface
func (bf *TestBuiltinFunction) ProcessBuiltinFunction(acntSnd, acntDst vmcommon.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	return bf.Function(acntSnd, acntDst, vmInput)
}

// SetNewGasConfig -
func (bf *TestBuiltinFunction) SetNewGasConfig(_ *vmcommon.GasCost) {
}

// IsActive -
func (bf *TestBuiltinFunction) IsActive() bool {
	return true
}

// IsInterfaceNil -
func (bf *TestBuiltinFunction) IsInterfaceNil() bool {
	return bf == nil
}

// GenerateOneAddressPerShard -
func GenerateOneAddressPerShard(shardCoordinator sharding.Coordinator) [][]byte {
	addresses := make([][]byte, shardCoordinator.NumberOfShards())
	for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
		for j := byte(1); j < 255; j++ {
			generatedAddress := bytes.Repeat([]byte{j}, 32)
			computedShardId := shardCoordinator.ComputeId(generatedAddress)
			if computedShardId == i {
				addresses[i] = generatedAddress
				break
			}
		}
	}
	return addresses
}
