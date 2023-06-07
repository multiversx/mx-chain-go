package vmcommonMocks

import (
	"math/big"

	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// BuiltInFunctionExecutableStub -
type BuiltInFunctionExecutableStub struct {
	ProcessBuiltinFunctionCalled func(acntSnd, acntDst vmcommon.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error)
	SetNewGasConfigCalled        func(gasCost *vmcommon.GasCost)
	IsActiveCalled               func() bool
	CheckIsExecutableCalled      func(senderAddr []byte, value *big.Int, receiverAddr []byte, gasProvidedForCall uint64, arguments [][]byte) error
}

// ProcessBuiltinFunction -
func (b *BuiltInFunctionExecutableStub) ProcessBuiltinFunction(acntSnd, acntDst vmcommon.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	if b.ProcessBuiltinFunctionCalled != nil {
		return b.ProcessBuiltinFunctionCalled(acntSnd, acntDst, vmInput)
	}
	return &vmcommon.VMOutput{}, nil
}

// SetNewGasConfig -
func (b *BuiltInFunctionExecutableStub) SetNewGasConfig(gasCost *vmcommon.GasCost) {
	if b.SetNewGasConfigCalled != nil {
		b.SetNewGasConfigCalled(gasCost)
	}
}

// IsActive -
func (b *BuiltInFunctionExecutableStub) IsActive() bool {
	if b.IsActiveCalled != nil {
		return b.IsActiveCalled()
	}
	return true
}

// CheckIsExecutable -
func (b *BuiltInFunctionExecutableStub) CheckIsExecutable(senderAddr []byte, value *big.Int, receiverAddr []byte, gasProvidedForCall uint64, arguments [][]byte) error {
	if b.CheckIsExecutableCalled != nil {
		return b.CheckIsExecutableCalled(senderAddr, value, receiverAddr, gasProvidedForCall, arguments)
	}
	return nil
}

// IsInterfaceNil -
func (b *BuiltInFunctionExecutableStub) IsInterfaceNil() bool {
	return b == nil
}
