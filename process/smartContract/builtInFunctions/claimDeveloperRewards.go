package builtInFunctions

import (
	"bytes"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type claimDeveloperRewards struct {
	gasCost uint64
}

// NewClaimDeveloperRewardsFunc returns a new developer rewards implementation
func NewClaimDeveloperRewardsFunc(gasCost uint64) *claimDeveloperRewards {
	return &claimDeveloperRewards{gasCost: gasCost}
}

// ProcessBuiltinFunction processes the protocol built-in smart contract function
func (c *claimDeveloperRewards) ProcessBuiltinFunction(acntSnd, acntDst state.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*big.Int, uint64, error) {
	if vmInput == nil {
		return nil, 0, process.ErrNilVmInput
	}
	if check.IfNil(acntDst) {
		return nil, vmInput.GasProvided, process.ErrNilSCDestAccount
	}

	if !bytes.Equal(vmInput.CallerAddr, acntDst.GetOwnerAddress()) {
		return nil, vmInput.GasProvided, process.ErrOperationNotPermitted
	}
	if vmInput.GasProvided < c.gasCost {
		return nil, vmInput.GasProvided, process.ErrNotEnoughGas
	}

	value, err := acntDst.ClaimDeveloperRewards(vmInput.CallerAddr)
	if err != nil {
		return nil, vmInput.GasProvided, err
	}

	if check.IfNil(acntSnd) {
		return value, vmInput.GasProvided, nil
	}

	err = acntSnd.AddToBalance(value)
	if err != nil {
		return nil, vmInput.GasProvided, err
	}

	return value, c.gasCost, nil
}

// GasUsed returns the gas used for processing the change
func (c *claimDeveloperRewards) GasUsed() uint64 {
	return c.gasCost
}

// IsInterfaceNil returns true if underlying object is nil
func (c *claimDeveloperRewards) IsInterfaceNil() bool {
	return c == nil
}
