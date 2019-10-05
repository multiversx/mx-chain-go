package systemSmartContracts

import (
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

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
	if args.GasProvided == nil {
		return vm.ErrInputGasProvidedIsNil
	}
	if args.GasPrice == nil {
		return vm.ErrInputGasPriceIsNil
	}
	return nil
}
