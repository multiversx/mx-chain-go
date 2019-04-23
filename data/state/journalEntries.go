package state

import (
	"math/big"
)

//------- JournalEntryNonce

// JournalEntryNonce is used to revert a nonce change
type JournalEntryNonce struct {
	account  AccountWrapper
	oldNonce uint64
}

// NewJournalEntryNonce outputs a new JournalEntry implementation used to revert a nonce change
func NewJournalEntryNonce(account AccountWrapper, oldNonce uint64) (*JournalEntryNonce, error) {
	if account == nil {
		return nil, ErrNilAccountWrapper
	}

	return &JournalEntryNonce{
		account:  account,
		oldNonce: oldNonce,
	}, nil
}

// Revert applies undo operation
func (jen *JournalEntryNonce) Revert() (AccountWrapper, error) {
	acnt, ok := jen.account.(*Account)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	acnt.Nonce = jen.oldNonce

	return jen.account, nil
}

//------- JournalEntryBalance

// JournalEntryBalance is used to revert a balance change
type JournalEntryBalance struct {
	account    AccountWrapper
	oldBalance *big.Int
}

// NewJournalEntryBalance outputs a new JournalEntry implementation used to revert a balance change
func NewJournalEntryBalance(account AccountWrapper, oldBalance *big.Int) (*JournalEntryBalance, error) {
	if account == nil {
		return nil, ErrNilAccountWrapper
	}

	return &JournalEntryBalance{
		account:    account,
		oldBalance: oldBalance,
	}, nil
}

// Revert applies undo operation
func (jeb *JournalEntryBalance) Revert() (AccountWrapper, error) {
	acnt, ok := jeb.account.(*Account)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	acnt.Balance = jeb.oldBalance

	return jeb.account, nil
}
