package systemSmartContracts

import (
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/vm"
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

	return nil
}
