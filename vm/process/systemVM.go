package process

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type systemVM struct {
	systemEI        vm.ContextHandler
	vmType          []byte
	systemContracts vm.SystemSCContainer
}

// NewSystemVM instantiates the system VM which is capable of running in protocol smart contracts
func NewSystemVM(
	systemEI vm.ContextHandler,
	systemContracts vm.SystemSCContainer,
	vmType []byte,
) (*systemVM, error) {
	if check.IfNil(systemEI) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if check.IfNil(systemContracts) {
		return nil, vm.ErrNilSystemContractsContainer
	}
	if len(vmType) == 0 { // no need for nil check, len() for nil returns 0
		return nil, vm.ErrNilVMType
	}

	sVm := &systemVM{
		systemEI:        systemEI,
		systemContracts: systemContracts,
		vmType:          make([]byte, len(vmType)),
	}
	copy(sVm.vmType, vmType)

	return sVm, nil
}

// RunSmartContractCreate creates and saves a new smart contract to the trie
func (s *systemVM) RunSmartContractCreate(input *vmcommon.ContractCreateInput) (*vmcommon.VMOutput, error) {
	if input == nil {
		return nil, vm.ErrInputArgsIsNil
	}
	if input.CallerAddr == nil {
		return nil, vm.ErrInputCallerAddrIsNil
	}

	s.systemEI.CleanCache()
	s.systemEI.SetSCAddress(input.CallerAddr)

	contract, err := s.systemContracts.Get(input.CallerAddr)
	if err != nil {
		return nil, vm.ErrUnknownSystemSmartContract
	}
	if !contract.CanUseContract() {
		// backward compatibility
		return nil, vm.ErrUnknownSystemSmartContract
	}

	deployInput := &vmcommon.ContractCallInput{
		VMInput:       input.VMInput,
		RecipientAddr: input.CallerAddr,
		Function:      core.SCDeployInitFunctionName,
	}

	returnCode := contract.Execute(deployInput)

	s.systemEI.AddCode(input.CallerAddr, input.ContractCode)

	vmOutput := s.systemEI.CreateVMOutput()
	vmOutput.ReturnCode = returnCode

	return vmOutput, nil
}

// RunSmartContractCall executes a smart contract according to the input
func (s *systemVM) RunSmartContractCall(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	s.systemEI.CleanCache()
	s.systemEI.SetSCAddress(input.RecipientAddr)
	s.systemEI.AddTxValueToSmartContract(input.CallValue, input.RecipientAddr)
	s.systemEI.SetGasProvided(input.GasProvided)

	contract, err := s.systemContracts.Get(input.RecipientAddr)
	if err != nil {
		return nil, vm.ErrUnknownSystemSmartContract
	}

	if input.Function == core.SCDeployInitFunctionName {
		return &vmcommon.VMOutput{
			ReturnCode:    vmcommon.UserError,
			ReturnMessage: "cannot call smart contract init function",
		}, nil
	}

	returnCode := contract.Execute(input)

	vmOutput := s.systemEI.CreateVMOutput()
	vmOutput.ReturnCode = returnCode

	return vmOutput, nil
}
