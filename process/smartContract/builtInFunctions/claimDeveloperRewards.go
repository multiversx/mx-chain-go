package builtInFunctions

import (
	"bytes"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.BuiltinFunction = (*claimDeveloperRewards)(nil)

type claimDeveloperRewards struct {
	gasCost      uint64
	mutExecution sync.RWMutex
}

// NewClaimDeveloperRewardsFunc returns a new developer rewards implementation
func NewClaimDeveloperRewardsFunc(gasCost uint64) *claimDeveloperRewards {
	return &claimDeveloperRewards{gasCost: gasCost}
}

// SetNewGasConfig is called whenever gas cost is changed
func (c *claimDeveloperRewards) SetNewGasConfig(gasCost *process.GasCost) {
	c.mutExecution.Lock()
	c.gasCost = gasCost.BuiltInCost.ClaimDeveloperRewards
	c.mutExecution.Unlock()
}

// ProcessBuiltinFunction processes the protocol built-in smart contract function
func (c *claimDeveloperRewards) ProcessBuiltinFunction(
	acntSnd, acntDst state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {
	c.mutExecution.RLock()
	defer c.mutExecution.RUnlock()

	if vmInput == nil {
		return nil, process.ErrNilVmInput
	}
	if vmInput.CallValue.Cmp(zero) != 0 {
		return nil, process.ErrBuiltInFunctionCalledWithValue
	}
	gasRemaining := computeGasRemaining(acntSnd, vmInput.GasProvided, c.gasCost)
	if check.IfNil(acntDst) {
		// cross-shard call, in sender shard only the gas is taken out
		return &vmcommon.VMOutput{ReturnCode: vmcommon.Ok, GasRemaining: gasRemaining}, nil
	}

	if !bytes.Equal(vmInput.CallerAddr, acntDst.GetOwnerAddress()) {
		return nil, process.ErrOperationNotPermitted
	}
	if vmInput.GasProvided < c.gasCost {
		return nil, process.ErrNotEnoughGas
	}

	value, err := acntDst.ClaimDeveloperRewards(vmInput.CallerAddr)
	if err != nil {
		return nil, err
	}

	vmOutput := &vmcommon.VMOutput{GasRemaining: gasRemaining}
	outTransfer := vmcommon.OutputTransfer{
		Value:    big.NewInt(0).Set(value),
		GasLimit: 0,
		Data:     nil,
		CallType: vmcommon.DirectCall,
	}
	if vmInput.CallType == vmcommon.AsynchronousCall {
		outTransfer.GasLocked = vmInput.GasLocked
		outTransfer.GasLimit = vmOutput.GasRemaining
		outTransfer.CallType = vmcommon.AsynchronousCallBack
		vmOutput.GasRemaining = 0
	}
	outputAcc := &vmcommon.OutputAccount{
		Address:         vmInput.CallerAddr,
		BalanceDelta:    big.NewInt(0).Set(value),
		OutputTransfers: []vmcommon.OutputTransfer{outTransfer},
	}

	vmOutput.OutputAccounts = make(map[string]*vmcommon.OutputAccount)
	vmOutput.OutputAccounts[string(outputAcc.Address)] = outputAcc

	if check.IfNil(acntSnd) {
		return vmOutput, nil
	}

	err = acntSnd.AddToBalance(value)
	if err != nil {
		return nil, err
	}

	if core.IsSmartContractAddress(vmInput.CallerAddr) {
		vmOutput.OutputAccounts = make(map[string]*vmcommon.OutputAccount)
	}

	return vmOutput, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (c *claimDeveloperRewards) IsInterfaceNil() bool {
	return c == nil
}
