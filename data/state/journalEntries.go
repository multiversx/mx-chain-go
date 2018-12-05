package state

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
)

// JournalEntryCreation is used to revert an account creation
type JournalEntryCreation struct {
	addressContainer AddressContainer
}

// JournalEntryNonce is used to revert a nonce change
type JournalEntryNonce struct {
	jurnalizedAccount JournalizedAccountWrapper
	oldNonce          uint64
}

// JournalEntryBalance is used to revert a balance change
type JournalEntryBalance struct {
	jurnalizedAccount JournalizedAccountWrapper
	oldBalance        big.Int
}

// JournalEntryCodeHash is used to revert a code hash change
type JournalEntryCodeHash struct {
	jurnalizedAccount JournalizedAccountWrapper
	oldCodeHash       []byte
}

// JournalEntryCode is used to revert a code addition to the trie
type JournalEntryCode struct {
	codeHash []byte
}

// JournalEntryRootHash is used to revert an account's root hash change
type JournalEntryRootHash struct {
	jurnalizedAccount JournalizedAccountWrapper
	oldRootHash       []byte
}

// JournalEntryData is used to mark an account's data change
type JournalEntryData struct {
	trie              trie.PatriciaMerkelTree
	jurnalizedAccount JournalizedAccountWrapper
}

//------- JournalEntryCreation

// NewJournalEntryCreation outputs a new JournalEntry implementation used to revert an account creation
func NewJournalEntryCreation(addressContainer AddressContainer) *JournalEntryCreation {
	return &JournalEntryCreation{
		addressContainer: addressContainer,
	}
}

// Revert apply undo operation
func (jec *JournalEntryCreation) Revert(accountsAdapter AccountsAdapter) error {
	if accountsAdapter == nil {
		return ErrNilAccountsAdapter
	}

	if jec.addressContainer == nil {
		return ErrNilAddressContainer
	}

	return accountsAdapter.RemoveAccount(jec.addressContainer)
}

// DirtiedAddress returns the address on which this JournalEntry will apply
func (jec *JournalEntryCreation) DirtiedAddress() AddressContainer {
	return jec.addressContainer
}

//------- JournalEntryNonce

// NewJournalEntryNonce outputs a new JournalEntry implementation used to revert a nonce change
func NewJournalEntryNonce(jurnalizedAccount JournalizedAccountWrapper, oldNonce uint64) *JournalEntryNonce {
	return &JournalEntryNonce{
		jurnalizedAccount: jurnalizedAccount,
		oldNonce:          oldNonce,
	}
}

// Revert apply undo operation
func (jen *JournalEntryNonce) Revert(accountsAdapter AccountsAdapter) error {
	if accountsAdapter == nil {
		return ErrNilAccountsAdapter
	}

	if jen.jurnalizedAccount == nil {
		return ErrNilJurnalizingAccountWrapper
	}

	if jen.jurnalizedAccount.AddressContainer() == nil {
		return ErrNilAddressContainer
	}

	//access nonce through implicit func as to not re-register the modification
	jen.jurnalizedAccount.SetNonce(jen.oldNonce)
	return accountsAdapter.SaveJournalizedAccount(jen.jurnalizedAccount)
}

// DirtiedAddress returns the address on which this JournalEntry will apply
func (jen *JournalEntryNonce) DirtiedAddress() AddressContainer {
	return jen.jurnalizedAccount.AddressContainer()
}

//------- JournalEntryBalance

// NewJournalEntryBalance outputs a new JournalEntry implementation used to revert a balance change
func NewJournalEntryBalance(jurnalizedAccount JournalizedAccountWrapper, oldBalance big.Int) *JournalEntryBalance {
	return &JournalEntryBalance{
		jurnalizedAccount: jurnalizedAccount,
		oldBalance:        oldBalance,
	}
}

// Revert apply undo operation
func (jeb *JournalEntryBalance) Revert(accountsAdapter AccountsAdapter) error {
	if accountsAdapter == nil {
		return ErrNilAccountsAdapter
	}

	if jeb.jurnalizedAccount == nil {
		return ErrNilJurnalizingAccountWrapper
	}

	if jeb.jurnalizedAccount.AddressContainer() == nil {
		return ErrNilAddressContainer
	}

	//save balance through implicit func as to not re-register the modification
	jeb.jurnalizedAccount.SetBalance(jeb.oldBalance)
	return accountsAdapter.SaveJournalizedAccount(jeb.jurnalizedAccount)
}

// DirtiedAddress returns the address on which this JournalEntry will apply
func (jeb *JournalEntryBalance) DirtiedAddress() AddressContainer {
	return jeb.jurnalizedAccount.AddressContainer()
}

//------- JournalEntryCodeHash

// NewJournalEntryCodeHash outputs a new JournalEntry implementation used to revert a code hash change
func NewJournalEntryCodeHash(jurnalizedAccount JournalizedAccountWrapper, oldCodeHash []byte) *JournalEntryCodeHash {
	return &JournalEntryCodeHash{
		jurnalizedAccount: jurnalizedAccount,
		oldCodeHash:       oldCodeHash,
	}
}

// Revert apply undo operation
func (jech *JournalEntryCodeHash) Revert(accountsAdapter AccountsAdapter) error {
	if accountsAdapter == nil {
		return ErrNilAccountsAdapter
	}

	if jech.jurnalizedAccount == nil {
		return ErrNilJurnalizingAccountWrapper
	}

	if jech.jurnalizedAccount.AddressContainer() == nil {
		return ErrNilAddressContainer
	}

	//bare in mind that nil oldRootHash are permitted because this is the way a new SC with data account
	//is constructed
	if jech.oldCodeHash != nil && len(jech.oldCodeHash) != AdrLen {
		return NewErrorWrongSize(AdrLen, len(jech.oldCodeHash))
	}

	//access code hash through implicit func as to not re-register the modification
	jech.jurnalizedAccount.SetCodeHash(jech.oldCodeHash)
	return accountsAdapter.SaveJournalizedAccount(jech.jurnalizedAccount)
}

// DirtiedAddress returns the address on which this JournalEntry will apply
func (jech *JournalEntryCodeHash) DirtiedAddress() AddressContainer {
	return jech.jurnalizedAccount.AddressContainer()
}

//------- JournalEntryCode

// NewJournalEntryCode outputs a new JournalEntry implementation used to revert a code addition to the trie
func NewJournalEntryCode(codeHash []byte) *JournalEntryCode {
	return &JournalEntryCode{
		codeHash: codeHash,
	}
}

// Revert apply undo operation
func (jec *JournalEntryCode) Revert(accountsAdapter AccountsAdapter) error {
	if accountsAdapter == nil {
		return ErrNilAccountsAdapter
	}

	if jec.codeHash == nil {
		return ErrNilValue
	}

	if len(jec.codeHash) != AdrLen {
		return NewErrorWrongSize(AdrLen, len(jec.codeHash))
	}

	return accountsAdapter.RemoveCode(jec.codeHash)
}

// DirtiedAddress will return nil as there is no address involved in code saving in a trie
func (jec *JournalEntryCode) DirtiedAddress() AddressContainer {
	return nil
}

//------- JournalEntryRoot

// NewJournalEntryRootHash outputs a new JournalEntry implementation used to revert an account's root hash change
func NewJournalEntryRootHash(jurnalizedAccount JournalizedAccountWrapper, oldRootHash []byte) *JournalEntryRootHash {
	return &JournalEntryRootHash{
		jurnalizedAccount: jurnalizedAccount,
		oldRootHash:       oldRootHash,
	}
}

// Revert apply undo operation
func (jer *JournalEntryRootHash) Revert(accountsAdapter AccountsAdapter) error {
	if accountsAdapter == nil {
		return ErrNilAccountsAdapter
	}

	if jer.jurnalizedAccount == nil {
		return ErrNilJurnalizingAccountWrapper
	}

	if jer.jurnalizedAccount.AddressContainer() == nil {
		return ErrNilAddressContainer
	}

	//bare in mind that nil oldRootHash are permitted because this is the way a new SC with data account
	//is constructed
	if jer.oldRootHash != nil && len(jer.oldRootHash) != AdrLen {
		return NewErrorWrongSize(AdrLen, len(jer.oldRootHash))
	}

	//access code hash through implicit func as to not re-register the modification
	jer.jurnalizedAccount.SetRootHash(jer.oldRootHash)
	err := accountsAdapter.LoadDataTrie(jer.jurnalizedAccount)
	if err != nil {
		return err
	}
	return accountsAdapter.SaveJournalizedAccount(jer.jurnalizedAccount)
}

// DirtiedAddress returns the address on which this JournalEntry will apply
func (jer *JournalEntryRootHash) DirtiedAddress() AddressContainer {
	return jer.jurnalizedAccount.AddressContainer()
}

//------- JournalEntryData

// NewJournalEntryData outputs a new JournalEntry implementation used to keep track of data change.
// The revert will practically empty the dirty data map
func NewJournalEntryData(jurnalizedAccount JournalizedAccountWrapper, trie trie.PatriciaMerkelTree) *JournalEntryData {
	return &JournalEntryData{
		jurnalizedAccount: jurnalizedAccount,
		trie:              trie,
	}
}

// Revert will empty the dirtyData map from AccountState
func (jed *JournalEntryData) Revert(accountsAdapter AccountsAdapter) error {
	if jed.jurnalizedAccount == nil {
		return ErrNilJurnalizingAccountWrapper
	}

	jed.jurnalizedAccount.ClearDataCaches()
	return nil
}

// DirtiedAddress will return nil as there is no address involved in saving data
func (jed *JournalEntryData) DirtiedAddress() AddressContainer {
	return nil
}

// Trie returns the referenced PatriciaMerkelTree for committing the changes
func (jed *JournalEntryData) Trie() trie.PatriciaMerkelTree {
	return jed.trie
}
