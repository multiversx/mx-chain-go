package smartContract

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-vm-common"
)

type claimDeveloperRewards struct {
	gasCost uint64
}

// ProcessBuiltinFunction processes the protocol built-in smart contract function
func (c *claimDeveloperRewards) ProcessBuiltinFunction(
	tx data.TransactionHandler,
	acntSnd, acntDst state.UserAccountHandler,
	_ *vmcommon.ContractCallInput,
) (*big.Int, error) {
	if check.IfNil(tx) {
		return nil, process.ErrNilTransaction
	}
	if check.IfNil(acntDst) {
		return nil, process.ErrNilSCDestAccount
	}

	value, err := acntDst.ClaimDeveloperRewards(tx.GetSndAddress())
	if err != nil {
		return nil, err
	}

	if check.IfNil(acntSnd) {
		return value, nil
	}

	err = acntSnd.AddToBalance(value)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// GasUsed returns the gas used for processing the change
func (c *claimDeveloperRewards) GasUsed() uint64 {
	return c.gasCost
}

// IsInterfaceNil returns true if underlying object is nil
func (c *claimDeveloperRewards) IsInterfaceNil() bool {
	return c == nil
}
