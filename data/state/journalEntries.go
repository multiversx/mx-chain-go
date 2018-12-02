package state

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
)

// JournalEntryCreation is used to revert an account creation
type JournalEntryCreation struct {
	address AddressHandler
}

// JournalEntryNonce is used to revert a nonce change
type JournalEntryNonce struct {
	address  AddressHandler
	acnt     AccountStateHandler
	oldNonce uint64
}

// JournalEntryBalance is used to revert a balance change
type JournalEntryBalance struct {
	address    AddressHandler
	acnt       AccountStateHandler
	oldBalance *big.Int
}

// JournalEntryCodeHash is used to revert a code hash change
type JournalEntryCodeHash struct {
	address     AddressHandler
	acnt        AccountStateHandler
	oldCodeHash []byte
}

// JournalEntryCode is used to revert a code addition to the trie
type JournalEntryCode struct {
	codeHash []byte
}

// JournalEntryRootHash is used to revert an account's root hash change
type JournalEntryRootHash struct {
	address     AddressHandler
	acnt        AccountStateHandler
	oldRootHash []byte
}

// JournalEntryData is used to mark an account's data change
type JournalEntryData struct {
	trie trie.PatriciaMerkelTree
	as   AccountStateHandler
}

//------- JournalEntryCreation

// NewJournalEntryCreation outputs a new JournalEntry implementation used to revert an account creation
func NewJournalEntryCreation(address AddressHandler) *JournalEntryCreation {
	return &JournalEntryCreation{
		address: address,
	}
}

// Revert apply undo operation
func (jec *JournalEntryCreation) Revert(accounts AccountsHandler) error {
	if accounts == nil {
		return ErrNilAccountsHandler
	}

	if jec.address == nil {
		return ErrNilAddress
	}

	return accounts.RemoveAccount(jec.address)
}

// DirtiedAddress returns the address on which this JournalEntry will apply
func (jec *JournalEntryCreation) DirtiedAddress() AddressHandler {
	return jec.address
}

//------- JournalEntryNonce

// NewJournalEntryNonce outputs a new JournalEntry implementation used to revert a nonce change
func NewJournalEntryNonce(address AddressHandler, acnt AccountStateHandler, oldNonce uint64) *JournalEntryNonce {
	return &JournalEntryNonce{
		address:  address,
		oldNonce: oldNonce,
		acnt:     acnt,
	}
}

// Revert apply undo operation
func (jen *JournalEntryNonce) Revert(accounts AccountsHandler) error {
	if accounts == nil {
		return ErrNilAccountsHandler
	}

	if jen.address == nil {
		return ErrNilAddress
	}

	if jen.acnt == nil {
		return ErrNilAccountState
	}

	//access nonce through dedicated func
	//as to not register a new entry of nonce change
	jen.acnt.SetNonceNoJournal(jen.oldNonce)
	return accounts.SaveAccountState(jen.acnt)
}

// DirtiedAddress returns the address on which this JournalEntry will apply
func (jen *JournalEntryNonce) DirtiedAddress() AddressHandler {
	return jen.address
}

//------- JournalEntryBalance

// NewJournalEntryBalance outputs a new JournalEntry implementation used to revert a balance change
func NewJournalEntryBalance(address AddressHandler, acnt AccountStateHandler, oldBalance *big.Int) *JournalEntryBalance {
	return &JournalEntryBalance{
		address:    address,
		oldBalance: oldBalance,
		acnt:       acnt,
	}
}

// Revert apply undo operation
func (jeb *JournalEntryBalance) Revert(accounts AccountsHandler) error {
	if accounts == nil {
		return ErrNilAccountsHandler
	}

	if jeb.address == nil {
		return ErrNilAddress
	}

	if jeb.acnt == nil {
		return ErrNilAccountState
	}

	//save balance through dedicated func
	//as to not register a new entry of balance change
	jeb.acnt.SetBalanceNoJournal(jeb.oldBalance)
	return accounts.SaveAccountState(jeb.acnt)
}

// DirtiedAddress returns the address on which this JournalEntry will apply
func (jeb *JournalEntryBalance) DirtiedAddress() AddressHandler {
	return jeb.address
}

//------- JournalEntryCodeHash

// NewJournalEntryCodeHash outputs a new JournalEntry implementation used to revert a code hash change
func NewJournalEntryCodeHash(address AddressHandler, acnt AccountStateHandler, oldCodeHash []byte) *JournalEntryCodeHash {
	return &JournalEntryCodeHash{
		address:     address,
		acnt:        acnt,
		oldCodeHash: oldCodeHash,
	}
}

// Revert apply undo operation
func (jech *JournalEntryCodeHash) Revert(accounts AccountsHandler) error {
	if accounts == nil {
		return ErrNilAccountsHandler
	}

	if jech.address == nil {
		return ErrNilAddress
	}

	if jech.acnt == nil {
		return ErrNilAccountState
	}

	//access code hash through dedicated func
	//as to not register a new entry of code hash change
	jech.acnt.SetCodeHashNoJournal(jech.oldCodeHash)
	return accounts.SaveAccountState(jech.acnt)
}

// DirtiedAddress returns the address on which this JournalEntry will apply
func (jech *JournalEntryCodeHash) DirtiedAddress() AddressHandler {
	return jech.address
}

//------- JournalEntryCode

// NewJournalEntryCode outputs a new JournalEntry implementation used to revert a code addition to the trie
func NewJournalEntryCode(codeHash []byte) *JournalEntryCode {
	return &JournalEntryCode{
		codeHash: codeHash,
	}
}

// Revert apply undo operation
func (jec *JournalEntryCode) Revert(accounts AccountsHandler) error {
	if accounts == nil {
		return ErrNilAccountsHandler
	}

	return accounts.RemoveCode(jec.codeHash)
}

// DirtiedAddress will return nil as there is no address involved in code saving in a trie
func (jec *JournalEntryCode) DirtiedAddress() AddressHandler {
	return nil
}

//------- JournalEntryRoot

// NewJournalEntryRootHash outputs a new JournalEntry implementation used to revert an account's root hash change
func NewJournalEntryRootHash(address AddressHandler, acnt AccountStateHandler, oldRootHash []byte) *JournalEntryRootHash {
	return &JournalEntryRootHash{
		address:     address,
		acnt:        acnt,
		oldRootHash: oldRootHash,
	}
}

// Revert apply undo operation
func (jer *JournalEntryRootHash) Revert(accounts AccountsHandler) error {
	if accounts == nil {
		return ErrNilAccountsHandler
	}

	if jer.address == nil {
		return ErrNilAddress
	}

	if jer.acnt == nil {
		return ErrNilAccountState
	}

	//access code hash through dedicated func
	//as to not register a new entry of root hash change
	jer.acnt.SetRootHashNoJournal(jer.oldRootHash)
	err := accounts.RetrieveDataTrie(jer.acnt)
	if err != nil {
		return err
	}
	return accounts.SaveAccountState(jer.acnt)
}

// DirtiedAddress returns the address on which this JournalEntry will apply
func (jer *JournalEntryRootHash) DirtiedAddress() AddressHandler {
	return jer.address
}

//------- JournalEntryData

// NewJournalEntryData outputs a new JournalEntry implementation used to keep track of data change.
// The revert will practically empty the dirty data map
func NewJournalEntryData(trie trie.PatriciaMerkelTree, as AccountStateHandler) *JournalEntryData {
	return &JournalEntryData{
		trie: trie,
		as:   as,
	}
}

// Revert will empty the dirtyData map from AccountState
func (jed *JournalEntryData) Revert(accounts AccountsHandler) error {
	jed.as.ClearDataCaches()
	return nil
}

// DirtiedAddress will return nil as there is no address involved in saving data
func (jed *JournalEntryData) DirtiedAddress() AddressHandler {
	return nil
}

// Trie returns the referenced PatriciaMerkelTree for committing the changes
func (jed *JournalEntryData) Trie() trie.PatriciaMerkelTree {
	return jed.trie
}
