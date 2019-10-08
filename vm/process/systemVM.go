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

	return sVm, nil
}

// RunSmartContractCreate creates and saves a new smart contract to the trie
func (s *systemVM) RunSmartContractCreate(input *vmcommon.ContractCreateInput) (*vmcommon.VMOutput, error) {
	// currently this function is not used, as all the contracts are deployed and created at startNode time only
	// register the system smart contract with a name into the map

	return nil, vm.ErrCannotCreateNewSystemSmartContract
}

// RunSmartContractCall executes a smart contract according to the input
func (s *systemVM) RunSmartContractCall(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	s.systemEI.CleanCache()

	contract, err := s.systemContracts.Get(input.RecipientAddr)
	if err != nil {
		return nil, vm.ErrUnknownSystemSmartContract
	}

	returnCode := contract.Execute(input)

	vmOutput := s.systemEI.CreateVMOutput()
	vmOutput.ReturnCode = returnCode

	return vmOutput, nil
}
