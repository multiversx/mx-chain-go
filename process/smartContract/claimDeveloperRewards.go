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
}

// ProcessBuiltinFunction processes the protocol built-in smart contract function
func (c *claimDeveloperRewards) ProcessBuiltinFunction(
	tx data.TransactionHandler,
	_, acntDst state.UserAccountHandler,
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

	return value, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (c *claimDeveloperRewards) IsInterfaceNil() bool {
	return c == nil
}
