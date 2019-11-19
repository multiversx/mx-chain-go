package systemSmartContracts

import (
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// CheckIfNil verifies if contract call input is not nil
func CheckIfNil(args *vmcommon.ContractCallInput) error {
	if args == nil {
		return vm.ErrInputArgsIsNil
	}
	if args.CallValue == nil {
		return vm.ErrInputCallValueIsNil
	}
	if args.Function == "" {
		return vm.ErrInputFunctionIsNil
	}
	if args.CallerAddr == nil {
		return vm.ErrInputCallerAddrIsNil
	}
	if args.RecipientAddr == nil {
		return vm.ErrInputRecipientAddrIsNil
	}
	if args.GasProvided < 0 {
		return vm.ErrInputGasProvidedIsNil
	}
	if args.GasPrice < 0 {
		return vm.ErrInputGasPriceIsNil
	}
	if args.Header == nil {
		return vm.ErrInputHeaderIsNil
	}

	return nil
}
