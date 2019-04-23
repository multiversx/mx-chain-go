package state

import (
	"math/big"
)

// JournalEntryNonce is used to revert a nonce change
type JournalEntryNonce struct {
	account  *Account
	oldNonce uint64
}

// JournalEntryBalance is used to revert a balance change
type JournalEntryBalance struct {
	account    *Account
	oldBalance *big.Int
}

//------- JournalEntryNonce

// NewJournalEntryNonce outputs a new JournalEntry implementation used to revert a nonce change
func NewJournalEntryNonce(account *Account, oldNonce uint64) (*JournalEntryNonce, error) {
	if account == nil {
		return nil, ErrNilAccount
	}

	return &JournalEntryNonce{
		account:  account,
		oldNonce: oldNonce,
	}, nil
}

// Revert applies undo operation
func (jen *JournalEntryNonce) Revert() (AccountWrapper, error) {
	jen.account.Nonce = jen.oldNonce

	return jen.account, nil
}

//------- JournalEntryBalance

// NewJournalEntryBalance outputs a new JournalEntry implementation used to revert a balance change
func NewJournalEntryBalance(account *Account, oldBalance *big.Int) (*JournalEntryBalance, error) {
	if account == nil {
		return nil, ErrNilAccount
	}

	return &JournalEntryBalance{
		account:    account,
		oldBalance: oldBalance,
	}, nil
}

// Revert applies undo operation
func (jeb *JournalEntryBalance) Revert() (AccountWrapper, error) {
	jeb.account.Balance = jeb.oldBalance

	return jeb.account, nil
}
