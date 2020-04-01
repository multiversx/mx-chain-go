package smartContract

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const minimumUserNameLength = 10
const maximumUserNameLength = 20

type setUserName struct {
	gasCost uint64
}

// ProcessBuiltinFunction sets the username to the account if it is allowed
func (s *setUserName) ProcessBuiltinFunction(
	tx data.TransactionHandler,
	acntSnd, acntDst state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*big.Int, error) {
	return nil, nil
}

// GasUsed returns the used gas from the built-in function
func (s *setUserName) GasUsed() uint64 {
	return s.gasCost
}

// IsInterfaceNil
func (s *setUserName) IsInterfaceNil() bool {
	return s == nil
}
