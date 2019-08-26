package state

import (
	"math/big"
)

//------- JournalEntryBalance

// JournalEntryBalance is used to revert a balance change
type JournalEntryBalance struct {
	account    *Account
	oldBalance *big.Int
}

// NewJournalEntryBalance outputs a new JournalEntry implementation used to revert a balance change
func NewJournalEntryBalance(account *Account, oldBalance *big.Int) (*JournalEntryBalance, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &JournalEntryBalance{
		account:    account,
		oldBalance: oldBalance,
	}, nil
}

// Revert applies undo operation
func (jeb *JournalEntryBalance) Revert() (AccountHandler, error) {
	jeb.account.Balance = jeb.oldBalance

	return jeb.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (jeb *JournalEntryBalance) IsInterfaceNil() bool {
	if jeb == nil {
		return true
	}
	return false
}
