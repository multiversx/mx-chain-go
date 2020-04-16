package builtInFunctions

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var _ process.BuiltinFunction = (*saveUserName)(nil)

const userNameHashLength = 32

type saveUserName struct {
	gasCost         uint64
	mapDnsAddresses map[string]struct{}
	enableChange    bool
}

// NewSaveUserNameFunc returns a username built in function implementation
func NewSaveUserNameFunc(
	gasCost uint64,
	mapDnsAddresses map[string]struct{},
	enableChange bool,
) (*saveUserName, error) {
	if mapDnsAddresses == nil {
		return nil, process.ErrNilDnsAddresses
	}

	return &saveUserName{
		gasCost:         gasCost,
		mapDnsAddresses: mapDnsAddresses,
		enableChange:    enableChange,
	}, nil
}

// ProcessBuiltinFunction sets the username to the account if it is allowed
func (s *saveUserName) ProcessBuiltinFunction(
	_, acntDst state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*big.Int, uint64, error) {
	if vmInput == nil {
		return big.NewInt(0), 0, process.ErrNilVmInput
	}
	if check.IfNil(acntDst) {
		return big.NewInt(0), vmInput.GasProvided, process.ErrNilSCDestAccount
	}

	_, ok := s.mapDnsAddresses[string(vmInput.CallerAddr)]
	if !ok {
		return big.NewInt(0), vmInput.GasProvided, process.ErrCallerIsNotTheDNSAddress
	}

	if len(vmInput.Arguments) == 0 || len(vmInput.Arguments[0]) != userNameHashLength {
		return big.NewInt(0), vmInput.GasProvided, process.ErrInvalidArguments
	}
	if vmInput.GasProvided < s.gasCost {
		return big.NewInt(0), vmInput.GasProvided, process.ErrNotEnoughGas
	}

	currentUserName := acntDst.GetUserName()
	if !s.enableChange && len(currentUserName) > 0 {
		return big.NewInt(0), vmInput.GasProvided, process.ErrUserNameChangeIsDisabled
	}

	acntDst.SetUserName(vmInput.Arguments[0])

	return big.NewInt(0), s.gasCost, nil
}

// IsInterfaceNil returns true if underlying object in nil
func (s *saveUserName) IsInterfaceNil() bool {
	return s == nil
}
