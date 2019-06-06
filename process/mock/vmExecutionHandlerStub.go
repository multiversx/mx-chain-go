package mock

import (
	"github.com/ElrondNetwork/elrond-vm-common"
	"math/big"
)

type VMExecutionHandlerStub struct {
	G0CreateCalled               func(input *vmcommon.ContractCreateInput) (*big.Int, error)
	G0CallCalled                 func(input *vmcommon.ContractCallInput) (*big.Int, error)
	RunSmartContractCreateCalled func(input *vmcommon.ContractCreateInput) (*vmcommon.VMOutput, error)
	RunSmartContractCallCalled   func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error)
}

// G0Create yields the initial gas cost of creating a new smart contract
func (vm *VMExecutionHandlerStub) G0Create(input *vmcommon.ContractCreateInput) (*big.Int, error) {
	if vm.G0CreateCalled == nil {
		return big.NewInt(0), nil
	}

	return vm.G0CreateCalled(input)
}

// G0Call yields the initial gas cost of calling an existing smart contract
func (vm *VMExecutionHandlerStub) G0Call(input *vmcommon.ContractCallInput) (*big.Int, error) {
	if vm.G0CallCalled == nil {
		return big.NewInt(0), nil
	}

	return vm.G0CallCalled(input)
}

// Computes how a smart contract creation should be performed
func (vm *VMExecutionHandlerStub) RunSmartContractCreate(input *vmcommon.ContractCreateInput) (*vmcommon.VMOutput, error) {
	if vm.RunSmartContractCreateCalled == nil {
		return &vmcommon.VMOutput{}, nil
	}

	return vm.RunSmartContractCreateCalled(input)
}

// Computes the result of a smart contract call and how the system must change after the execution
func (vm *VMExecutionHandlerStub) RunSmartContractCall(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	if vm.RunSmartContractCallCalled == nil {
		return &vmcommon.VMOutput{}, nil
	}

	return vm.RunSmartContractCallCalled(input)
}
