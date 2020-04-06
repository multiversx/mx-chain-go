package builtInFunctions

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const minimumUserNameLength = 10
const maximumUserNameLength = 20

type userName struct {
	gasCost         uint64
	mapDnsAddresses map[string]struct{}
}

// NewUserNameFunc returns a username built in function implementation
func (u *userName) NewUserNameFunc(
	gasCost uint64,
	mapDnsAddresses map[string]struct{},
) (*userName, error) {
	if mapDnsAddresses == nil {
		return nil, process.ErrNilDnsAddresses
	}

	return &userName{
		gasCost:         gasCost,
		mapDnsAddresses: mapDnsAddresses,
	}, nil
}

// ProcessBuiltinFunction sets the username to the account if it is allowed
func (u *userName) ProcessBuiltinFunction(acntSnd, acntDst state.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*big.Int, error) {
	if vmInput == nil {
		return nil, process.ErrNilVmInput
	}
	if check.IfNil(acntDst) {
		return nil, process.ErrNilSCDestAccount
	}

	return nil, nil
}

// GasUsed returns the used gas from the built-in function
func (u *userName) GasUsed() uint64 {
	return u.gasCost
}

// IsInterfaceNil
func (u *userName) IsInterfaceNil() bool {
	return u == nil
}
