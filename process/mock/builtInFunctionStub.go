package mock

import (
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// BuiltInFunctionStub -
type BuiltInFunctionStub struct {
	ProcessBuiltinFunctionCalled func(acntSnd, acntDst vmcommon.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error)
	SetNewGasConfigCalled        func(gasCost *vmcommon.GasCost)
	IsActiveCalled               func() bool
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

// IsActive -
func (b *BuiltInFunctionStub) IsActive() bool {
	if b.IsActiveCalled != nil {
		return b.IsActiveCalled()
	}
	return true
}

// IsInterfaceNil -
func (b *BuiltInFunctionStub) IsInterfaceNil() bool {
	return b == nil
}
