package state

import (
	"math/big"
)

// JournalizedAccountWrap is an account wrapper that can journalize fields changes
type JournalizedAccountWrap struct {
	ModifyingDataAccountWrapper
	accounts AccountsAdapter
}

// NewJournalizedAccountWrap returns an account wrapper that can journalize fields changes
// It needs a modifyingDataAccountWrapper object
func NewJournalizedAccountWrap(modifyingDataAccountWrapper ModifyingDataAccountWrapper,
	accounts AccountsAdapter) (*JournalizedAccountWrap, error) {

	if modifyingDataAccountWrapper == nil {
		return nil, ErrNilModifyingAccountWrapper
	}

	if accounts == nil {
		return nil, ErrNilAccountsAdapter
	}

	return &JournalizedAccountWrap{
		ModifyingDataAccountWrapper: modifyingDataAccountWrapper,
		accounts:                    accounts,
	}, nil
}

// NewJournalizedAccountWrapFromAccountContainer returns an account wrapper that can journalize fields changes
// It needs an addressContainer, accountContainer and an accountsAdapter
func NewJournalizedAccountWrapFromAccountContainer(
	addressContainer AddressContainer,
	accountContainer AccountContainer,
	accounts AccountsAdapter) (*JournalizedAccountWrap, error) {

	if accountContainer == nil {
		return nil, ErrNilAccountContainer
	}

	simpleAccountWrap, err := NewSimpleAccountWrap(addressContainer, accountContainer)
	if err != nil {
		return nil, err
	}
	modifyingDataAccountWrapper, err := NewModifyingDataAccountWrap(simpleAccountWrap)
	if err != nil {
		return nil, err
	}

	return &JournalizedAccountWrap{
		ModifyingDataAccountWrapper: modifyingDataAccountWrapper,
		accounts:                    accounts,
	}, nil
}

// SetNonceWithJournal sets the account's nonce, saving the old nonce before changing
func (jaw *JournalizedAccountWrap) SetNonceWithJournal(nonce uint64) error {
	jaw.accounts.AddJournalEntry(NewJournalEntryNonce(jaw, jaw.Nonce()))
	jaw.SetNonce(nonce)

	return jaw.accounts.SaveJournalizedAccount(jaw)
}

// SetBalanceWithJournal sets the account's balance, saving the old balance before changing
func (jaw *JournalizedAccountWrap) SetBalanceWithJournal(balance big.Int) error {
	jaw.accounts.AddJournalEntry(NewJournalEntryBalance(jaw, jaw.Balance()))
	jaw.SetBalance(balance)

	return jaw.accounts.SaveJournalizedAccount(jaw)
}

// SetCodeHashWithJournal sets the account's code hash, saving the old code hash before changing
func (jaw *JournalizedAccountWrap) SetCodeHashWithJournal(codeHash []byte) error {
	jaw.accounts.AddJournalEntry(NewJournalEntryCodeHash(jaw, jaw.CodeHash()))
	jaw.SetCodeHash(codeHash)

	return jaw.accounts.SaveJournalizedAccount(jaw)
}

// SetRootHashWithJournal sets the account's root hash, saving the old root hash before changing
func (jaw *JournalizedAccountWrap) SetRootHashWithJournal(rootHash []byte) error {
	jaw.accounts.AddJournalEntry(NewJournalEntryRootHash(jaw, jaw.RootHash()))
	jaw.SetRootHash(rootHash)

	return jaw.accounts.SaveJournalizedAccount(jaw)
}
