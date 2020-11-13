package integrationTests

import (
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
)

// TestBuiltinFunction wraps a builtin function defined ad-hoc, for testing
type TestBuiltinFunction struct {
	Function func(acntSnd, acntDst state.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error)
}

// ProcessBuiltinFunction is a method implementation required by the BuiltinFunction interface
func (bf *TestBuiltinFunction) ProcessBuiltinFunction(acntSnd, acntDst state.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	return bf.Function(acntSnd, acntDst, vmInput)
}

// SetNewGasConfig -
func (bf *TestBuiltinFunction) SetNewGasConfig(_ *process.GasCost) {
}

// IsInterfaceNil --
func (bf *TestBuiltinFunction) IsInterfaceNil() bool {
	return bf == nil
}
