package mock

import (
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// BuiltInFunctionStub -
type BuiltInFunctionStub struct {
	ProcessBuiltinFunctionCalled func(acntSnd, acntDst vmcommon.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error)
	SetNewGasConfigCalled        func(gasCost *vmcommon.GasCost)
}

// ProcessBuiltinFunction -
func (b *BuiltInFunctionStub) ProcessBuiltinFunction(acntSnd, acntDst vmcommon.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	if b.ProcessBuiltinFunctionCalled != nil {
		return b.ProcessBuiltinFunctionCalled(acntSnd, acntDst, vmInput)
	}
	return &vmcommon.VMOutput{}, nil
}

// SetNewGasConfig -
func (b *BuiltInFunctionStub) SetNewGasConfig(gasCost *vmcommon.GasCost) {
	if b.SetNewGasConfigCalled != nil {
		b.SetNewGasConfigCalled(gasCost)
	}
}

// IsInterfaceNil -
func (b *BuiltInFunctionStub) IsInterfaceNil() bool {
	return b == nil
}
