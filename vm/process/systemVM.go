package process

import (
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type systemVM struct {
	systemEI        vm.SystemEI
	vmType          []byte
	systemContracts vm.SystemSCContainer
}

// NewSystemVM instantiates the system VM which is capable of running in protocol smart contracts
func NewSystemVM(
	systemEI vm.SystemEI,
	systemContracts vm.SystemSCContainer,
	vmType []byte,
) (*systemVM, error) {
	if systemEI == nil || systemEI.IsInterfaceNil() {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if systemContracts == nil || systemContracts.IsInterfaceNil() {
		return nil, vm.ErrNilSystemContractsContainer
	}
	if vmType == nil || len(vmType) == 0 {
		return nil, vm.ErrNilVMType
	}

	sVm := &systemVM{
		systemEI:        systemEI,
		systemContracts: systemContracts,
		vmType:          make([]byte, len(vmType)),
	}
	copy(sVm.vmType, vmType)

	//TODO: run all system smart contracts INIT function at genesis, update roothash and block

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
	// currently this function is not used, as all the contracts are deployed and created at startNode time only
	// register the system smart contract with a name into the map
	s.systemEI.CleanCache()
	s.systemEI.SetSCAddress(input.CallerAddr)

	contract, err := s.systemContracts.Get(input.CallerAddr)
	if err != nil {
		return nil, vm.ErrUnknownSystemSmartContract
	}

	deployInput := &vmcommon.ContractCallInput{
		VMInput:       input.VMInput,
		RecipientAddr: input.CallerAddr,
		Function:      "_init",
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

	contract, err := s.systemContracts.Get(input.RecipientAddr)
	if err != nil {
		return nil, vm.ErrUnknownSystemSmartContract
	}

	if input.Function == "_init" {
		return &vmcommon.VMOutput{ReturnCode: vmcommon.UserError}, nil
	}

	returnCode := contract.Execute(input)

	vmOutput := s.systemEI.CreateVMOutput()
	vmOutput.ReturnCode = returnCode

	return vmOutput, nil
}
