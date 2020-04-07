package builtInFunctions

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type changeOwnerAddress struct {
	gasCost uint64
}

// NewChangeOwnerAddressFunc create a new change owner built in function
func NewChangeOwnerAddressFunc(gasCost uint64) *changeOwnerAddress {
	return &changeOwnerAddress{gasCost: gasCost}
}

// ProcessBuiltinFunction processes simple protocol built-in function
func (c *changeOwnerAddress) ProcessBuiltinFunction(_, acntDst state.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*big.Int, uint64, error) {
	if vmInput == nil {
		return nil, 0, process.ErrNilVmInput
	}
	if len(vmInput.Arguments) == 0 {
		return nil, vmInput.GasProvided, process.ErrInvalidArguments
	}
	if check.IfNil(acntDst) {
		return nil, vmInput.GasProvided, process.ErrNilSCDestAccount
	}

	if !bytes.Equal(vmInput.CallerAddr, acntDst.GetOwnerAddress()) {
		return nil, 0, fmt.Errorf("%w not the owner of the account", process.ErrOperationNotPermitted)
	}
	if len(vmInput.Arguments[0]) != len(acntDst.AddressContainer().Bytes()) {
		return nil, 0, process.ErrInvalidAddressLength
	}
	if vmInput.GasProvided < c.gasCost {
		return nil, vmInput.GasProvided, process.ErrNotEnoughGas
	}

	err := acntDst.ChangeOwnerAddress(vmInput.CallerAddr, vmInput.Arguments[0])
	if err != nil {
		return nil, vmInput.GasProvided, err
	}

	return big.NewInt(0), c.gasCost, nil
}

// GasUsed returns the gas used for processing the change
func (c *changeOwnerAddress) GasUsed() uint64 {
	return c.gasCost
}

// IsInterfaceNil returns true if underlying object in nil
func (c *changeOwnerAddress) IsInterfaceNil() bool {
	return c == nil
}
