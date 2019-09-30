package systemSmartContracts

import (
	"errors"
	"github.com/ElrondNetwork/elrond-go/vm"
)

func CheckIfNil(args *vm.ExecuteArguments) error {
	if args == nil {
		return errors.New("input system smart contract arguments are nil")
	}
	if args.Value == nil {
		return errors.New("input value for system smart contract is nil")
	}
	if args.Function == "" {
		return errors.New("input function for system samrt contract is nil")
	}
	return nil
}
