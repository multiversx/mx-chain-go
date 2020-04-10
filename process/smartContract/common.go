package smartContract

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// GetBuiltinFunctions gets the list of built-in functions
func GetBuiltinFunctions() []string {
	return []string{claimDeveloperRewardsFunctionName, changeOwnerAddressFunctionName}
}

func findVMByTransaction(container process.VirtualMachinesContainer, tx data.TransactionHandler) (vmcommon.VMExecutionHandler, error) {
	scAddress := tx.GetRcvAddr()
	return findVMByScAddress(container, scAddress)
}

func findVMByScAddress(container process.VirtualMachinesContainer, scAddress []byte) (vmcommon.VMExecutionHandler, error) {
	vmType, err := parseVMTypeFromContractAddress(scAddress)
	if err != nil {
		return nil, err
	}

	vm, err := container.Get(vmType)
	if err != nil {
		return nil, err
	}

	return vm, nil
}

func parseVMTypeFromContractAddress(contractAddress []byte) ([]byte, error) {
	// TODO: Why not check against AddressLength (32)?
	if len(contractAddress) < core.NumInitCharactersForScAddress {
		return nil, vmcommon.ErrInvalidVMType
	}

	startIndex := core.NumInitCharactersForScAddress - core.VMTypeLen
	endIndex := core.NumInitCharactersForScAddress
	return contractAddress[startIndex:endIndex], nil
}
