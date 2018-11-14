package state

import (
	"math/big"

	"github.com/pkg/errors"
)

// ErrNilAccountsHandler defines the error when trying to revert on nil accounts
var ErrNilAccountsHandler = errors.New("nil AccountsHandler")

// ErrNilAddress defines the error when trying to work with a nil address
var ErrNilAddress = errors.New("nil Address")

// ErrNilAccountState defines the error when trying to work with a nil account
var ErrNilAccountState = errors.New("nil AccountState")

type JurnalEntryBalance struct {
	address    *Address
	acnt       *AccountState
	oldBalance *big.Int
}

func NewJurnalEntryBalance(address *Address, acnt *AccountState, oldBalance *big.Int) *JurnalEntryBalance {
	return &JurnalEntryBalance{address: address, oldBalance: oldBalance, acnt: acnt}
}

func (jeb *JurnalEntryBalance) Revert(accounts AccountsHandler) error {
	if accounts == nil {
		return ErrNilAccountsHandler
	}

	if jeb.address == nil {
		return ErrNilAddress
	}

	if jeb.acnt == nil {
		return ErrNilAccountState
	}

	jeb.acnt.Balance = jeb.oldBalance
	return accounts.SaveAccountState(jeb.acnt)
}

func (jeb *JurnalEntryBalance) DirtyAddress() *Address {
	return jeb.address
}

type JurnalEntryNonce struct {
	address  *Address
	acnt     *AccountState
	oldNonce uint64
}

func NewJurnalEntryNonce(address *Address, acnt *AccountState, oldNonce uint64) *JurnalEntryNonce {
	return &JurnalEntryNonce{address: address, oldNonce: oldNonce, acnt: acnt}
}

func (jen *JurnalEntryNonce) Revert(accounts AccountsHandler) error {
	if accounts == nil {
		return ErrNilAccountsHandler
	}

	if jen.address == nil {
		return ErrNilAddress
	}

	if jen.acnt == nil {
		return ErrNilAccountState
	}

	jen.acnt.Nonce = jen.oldNonce
	return accounts.SaveAccountState(jen.acnt)
}

func (jen *JurnalEntryNonce) DirtyAddress() *Address {
	return jen.address
}

type JurnalEntryCreation struct {
	address *Address
	acnt    *AccountState
}

func NewJurnalEntryCreation(address *Address, acnt *AccountState) *JurnalEntryCreation {
	return &JurnalEntryCreation{address: address, acnt: acnt}
}

func (jec *JurnalEntryCreation) Revert(accounts AccountsHandler) error {
	if accounts == nil {
		return ErrNilAccountsHandler
	}

	if jec.address == nil {
		return ErrNilAddress
	}

	if jec.acnt == nil {
		return ErrNilAccountState
	}

	//TODO remove SC code
	return accounts.RemoveAccount(*jec.address)
}

func (jec *JurnalEntryCreation) DirtyAddress() *Address {
	return jec.address
}
