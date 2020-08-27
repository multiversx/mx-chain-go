package integrationTests

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// TestBuiltinFunction wraps a builtin function defined ad-hoc, for testing
type TestBuiltinFunction struct {
	Function func(acntSnd, acntDst state.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error)
}

// ProcessBuiltinFunction is a method implementation required by the BuiltinFunction interface
func (bf *TestBuiltinFunction) ProcessBuiltinFunction(acntSnd, acntDst state.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	return bf.Function(acntSnd, acntDst, vmInput)
}

// IsInterfaceNil --
func (bf *TestBuiltinFunction) IsInterfaceNil() bool {
	return bf == nil
}
