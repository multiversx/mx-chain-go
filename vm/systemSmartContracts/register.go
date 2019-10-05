package systemSmartContracts

import (
	"errors"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"math/big"
)

type registerSC struct {
	eei        vm.SystemEI
	stakeValue *big.Int
}

// NewRegisterSmartContract creates a register smart contract
func NewRegisterSmartContract(stakeValue *big.Int, eei vm.SystemEI) (*registerSC, error) {
	if stakeValue == nil {
		return nil, errors.New("initial stake value is nil")
	}
	if eei == nil || eei.IsInterfaceNil() {
		return nil, errors.New("environment interface is nil")
	}

	reg := &registerSC{stakeValue: big.NewInt(0).Set(stakeValue), eei: eei}
	return reg, nil
}

// Execute calls one of the functions from the register smart contract and runs the code according to the input
func (r *registerSC) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if CheckIfNil(args) != nil {
		return vmcommon.UserError
	}

	if args.CallValue.Cmp(r.stakeValue) != 0 {
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

// ValueOf returns the value of a selected key
func (r *registerSC) ValueOf(key interface{}) interface{} {
	return nil
}

// IsInterfaceNil verifies if the underlying object is nil or not
func (r *registerSC) IsInterfaceNil() bool {
	if r == nil {
		return true
	}
	return false
}
