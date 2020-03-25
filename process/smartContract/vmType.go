package smartContract

import (
	"github.com/ElrondNetwork/elrond-go/core"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

func parseVMTypeFromContractAddress(contractAddress []byte) ([]byte, error) {
	if len(contractAddress) < core.NumInitCharactersForScAddress {
		return nil, ErrInvalidVMType
	}

	startIndex := core.NumInitCharactersForScAddress - core.VMTypeLen
	endIndex := core.NumInitCharactersForScAddress
	return contractAddress[startIndex:endIndex], nil
}

func parseVMTypeFromVMInput(vmInput *vmcommon.VMInput) ([]byte, error) {
	vmType := vmInput.Arguments[indexOfVMTypeInArguments]

	if len(vmType) != core.VMTypeLen {
		return nil, ErrInvalidVMType
	}

	return vmType, nil
}
