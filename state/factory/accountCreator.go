package factory

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// AccountCreator has method to create a new account
type AccountCreator struct {
	accountArgs state.ArgsAccountCreation
}

// NewAccountCreator creates a new instance of AccountCreator
func NewAccountCreator(args state.ArgsAccountCreation) (state.AccountFactory, error) {
	if check.IfNil(args.Hasher) {
		return nil, errors.ErrNilHasher
	}
	if check.IfNil(args.Marshaller) {
		return nil, errors.ErrNilMarshalizer
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, errors.ErrNilEnableEpochsHandler
	}

	return &AccountCreator{
		accountArgs: args,
	}, nil
}

// CreateAccount calls the new Account creator and returns the result
func (ac *AccountCreator) CreateAccount(address []byte) (vmcommon.AccountHandler, error) {
	return state.NewUserAccount(address, ac.accountArgs)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ac *AccountCreator) IsInterfaceNil() bool {
	return ac == nil
}
