package state

import (
	"math/big"
)

// JournalizedAccountWrap is an account wrapper that can journalize fields changes
type JournalizedAccountWrap struct {
	TrackableDataAccountWrapper
	accounts AccountsAdapter
}

// NewJournalizedAccountWrap returns an account wrapper that can journalize fields changes
// It needs a trackableDataAccountWrapper object
func NewJournalizedAccountWrap(trackableDataAccountWrapper TrackableDataAccountWrapper,
	accounts AccountsAdapter) (*JournalizedAccountWrap, error) {

	if trackableDataAccountWrapper == nil {
		return nil, ErrNilTrackableAccountWrapper
	}

	if accounts == nil {
		return nil, ErrNilAccountsAdapter
	}

	return &JournalizedAccountWrap{
		TrackableDataAccountWrapper: trackableDataAccountWrapper,
		accounts:                    accounts,
	}, nil
}

// NewJournalizedAccountWrapFromAccountContainer returns an account wrapper that can journalize fields changes
// It needs an addressContainer, accountContainer and an accountsAdapter
func NewJournalizedAccountWrapFromAccountContainer(
	addressContainer AddressContainer,
	account *Account,
	accounts AccountsAdapter) (*JournalizedAccountWrap, error) {

	if account == nil {
		return nil, ErrNilAccount
	}

	accountWrap, err := NewAccountWrap(addressContainer, account)
	if err != nil {
		return nil, err
	}
	trackableDataAccountWrap, err := NewTrackableDataAccountWrap(accountWrap)
	if err != nil {
		return nil, err
	}

	if accounts == nil {
		return nil, ErrNilAccountsAdapter
	}

	return &JournalizedAccountWrap{
		TrackableDataAccountWrapper: trackableDataAccountWrap,
		accounts:                    accounts,
	}, nil
}

// SetNonceWithJournal sets the account's nonce, saving the old nonce before changing
func (jaw *JournalizedAccountWrap) SetNonceWithJournal(nonce uint64) error {
	jaw.accounts.AddJournalEntry(NewJournalEntryNonce(jaw, jaw.BaseAccount().Nonce))
	jaw.BaseAccount().Nonce = nonce

	return jaw.accounts.SaveJournalizedAccount(jaw)
}

// SetBalanceWithJournal sets the account's balance, saving the old balance before changing
func (jaw *JournalizedAccountWrap) SetBalanceWithJournal(balance big.Int) error {
	jaw.accounts.AddJournalEntry(NewJournalEntryBalance(jaw, jaw.BaseAccount().Balance))
	jaw.BaseAccount().Balance = balance

	return jaw.accounts.SaveJournalizedAccount(jaw)
}

// SetCodeHashWithJournal sets the account's code hash, saving the old code hash before changing
func (jaw *JournalizedAccountWrap) SetCodeHashWithJournal(codeHash []byte) error {
	jaw.accounts.AddJournalEntry(NewJournalEntryCodeHash(jaw, jaw.BaseAccount().CodeHash))
	jaw.BaseAccount().CodeHash = codeHash

	return jaw.accounts.SaveJournalizedAccount(jaw)
}

// SetRootHashWithJournal sets the account's root hash, saving the old root hash before changing
func (jaw *JournalizedAccountWrap) SetRootHashWithJournal(rootHash []byte) error {
	jaw.accounts.AddJournalEntry(NewJournalEntryRootHash(jaw, jaw.BaseAccount().RootHash))
	jaw.BaseAccount().RootHash = rootHash

	return jaw.accounts.SaveJournalizedAccount(jaw)
}

// AppendDataRegistrationWithJournal appends a new registration data, keeping track that a new
// item was added
func (jaw *JournalizedAccountWrap) AppendDataRegistrationWithJournal(data *RegistrationData) error {
	journalEntry := NewJournalEntryAppendRegistration(jaw)
	err := jaw.AppendRegistrationData(data)

	if err != nil {
		return err
	}

	jaw.accounts.AddJournalEntry(journalEntry)
	return jaw.accounts.SaveJournalizedAccount(jaw)
}
