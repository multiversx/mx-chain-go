package factory

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// AccountCreator has method to create a new account
type AccountCreator struct {
	hasher              hashing.Hasher
	marshaller          marshal.Marshalizer
	enableEpochsHandler common.EnableEpochsHandler
}

// NewAccountCreator creates an account creator
func NewAccountCreator(hasher hashing.Hasher, marshaller marshal.Marshalizer, enableEpochsHandler common.EnableEpochsHandler) (state.AccountFactory, error) {
	if check.IfNil(hasher) {
		return nil, errors.ErrNilHasher
	}
	if check.IfNil(marshaller) {
		return nil, errors.ErrNilMarshalizer
	}
	if check.IfNil(enableEpochsHandler) {
		return nil, errors.ErrNilEnableEpochsHandler
	}

	return &AccountCreator{
		hasher:              hasher,
		marshaller:          marshaller,
		enableEpochsHandler: enableEpochsHandler,
	}, nil
}

// CreateAccount calls the new Account creator and returns the result
func (ac *AccountCreator) CreateAccount(address []byte) (vmcommon.AccountHandler, error) {
	return state.NewUserAccount(address, ac.hasher, ac.marshaller, ac.enableEpochsHandler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ac *AccountCreator) IsInterfaceNil() bool {
	return ac == nil
}
