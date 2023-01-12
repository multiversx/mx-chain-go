package integrationTests

import (
	"bytes"

	"github.com/multiversx/mx-chain-go/sharding"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
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
		generatedAddress := bytes.Repeat([]byte{1}, 32)
		generatedAddress[len(generatedAddress)-1] = byte(i)
		addresses[i] = generatedAddress
	}
	return addresses
}
