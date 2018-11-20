package state

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
)

// JournalEntryCreation is used to revert an account creation
type JournalEntryCreation struct {
	address *Address
	acnt    *AccountState
}

// JournalEntryNonce is used to revert a nonce change
type JournalEntryNonce struct {
	address  *Address
	acnt     *AccountState
	oldNonce uint64
}

// JournalEntryBalance is used to revert a balance change
type JournalEntryBalance struct {
	address    *Address
	acnt       *AccountState
	oldBalance *big.Int
}

// JournalEntryCodeHash is used to revert a code hash change
type JournalEntryCodeHash struct {
	address     *Address
	acnt        *AccountState
	oldCodeHash []byte
}

// JournalEntryCode is used to revert a code addition to the trie
type JournalEntryCode struct {
	codeHash []byte
}

// JournalEntryRoot is used to revert an account's root change
type JournalEntryRoot struct {
	address *Address
	acnt    *AccountState
	oldRoot []byte
}

// JournalEntryData is used to mark an account's data change
type JournalEntryData struct {
	trie trie.PatriciaMerkelTree
	as   *AccountState
}

//------- JournalEntryCreation

// NewJournalEntryCreation outputs a new JournalEntry implementation used to revert an account creation
func NewJournalEntryCreation(address *Address, acnt *AccountState) *JournalEntryCreation {
	return &JournalEntryCreation{address: address, acnt: acnt}
}

// Revert apply undo operation
func (jec *JournalEntryCreation) Revert(accounts AccountsHandler) error {
	if accounts == nil {
		return ErrNilAccountsHandler
	}

	if jec.address == nil {
		return ErrNilAddress
	}

	if jec.acnt == nil {
		return ErrNilAccountState
	}

	return accounts.RemoveAccount(*jec.address)
}

// DirtyAddress returns the address on which this JournalEntry will apply
func (jec *JournalEntryCreation) DirtyAddress() *Address {
	return jec.address
}

//------- JournalEntryNonce

// NewJournalEntryNonce outputs a new JournalEntry implementation used to revert a nonce change
func NewJournalEntryNonce(address *Address, acnt *AccountState, oldNonce uint64) *JournalEntryNonce {
	return &JournalEntryNonce{address: address, oldNonce: oldNonce, acnt: acnt}
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

	//access nonce through account pointer
	//as to not register a new entry of nonce change
	jen.acnt.Account().Nonce = jen.oldNonce
	return accounts.SaveAccountState(jen.acnt)
}

// DirtyAddress returns the address on which this JournalEntry will apply
func (jen *JournalEntryNonce) DirtyAddress() *Address {
	return jen.address
}

//------- JournalEntryBalance

// NewJournalEntryBalance outputs a new JournalEntry implementation used to revert a balance change
func NewJournalEntryBalance(address *Address, acnt *AccountState, oldBalance *big.Int) *JournalEntryBalance {
	return &JournalEntryBalance{address: address, oldBalance: oldBalance, acnt: acnt}
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

	//access balance through account pointer
	//as to not register a new entry of balance change
	jeb.acnt.Account().Balance = jeb.oldBalance
	return accounts.SaveAccountState(jeb.acnt)
}

// DirtyAddress returns the address on which this JournalEntry will apply
func (jeb *JournalEntryBalance) DirtyAddress() *Address {
	return jeb.address
}

//------- JournalEntryCodeHash

// NewJournalEntryCodeHash outputs a new JournalEntry implementation used to revert a code hash change
func NewJournalEntryCodeHash(address *Address, acnt *AccountState, oldCodeHash []byte) *JournalEntryCodeHash {
	return &JournalEntryCodeHash{address: address, acnt: acnt, oldCodeHash: oldCodeHash}
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

	//access code hash through account pointer
	//as to not register a new entry of code hash change
	jech.acnt.Account().CodeHash = jech.oldCodeHash
	return accounts.SaveAccountState(jech.acnt)
}

// DirtyAddress returns the address on which this JournalEntry will apply
func (jech *JournalEntryCodeHash) DirtyAddress() *Address {
	return jech.address
}

//------- JournalEntryCode

// NewJournalEntryCode outputs a new JournalEntry implementation used to revert a code addition to the trie
func NewJournalEntryCode(codeHash []byte) *JournalEntryCode {
	return &JournalEntryCode{codeHash: codeHash}
}

// Revert apply undo operation
func (jec *JournalEntryCode) Revert(accounts AccountsHandler) error {
	if accounts == nil {
		return ErrNilAccountsHandler
	}

	return accounts.RemoveCode(jec.codeHash)
}

// DirtyAddress will return nil as there is no address involved in code saving in a trie
func (jec *JournalEntryCode) DirtyAddress() *Address {
	return nil
}

//------- JournalEntryRoot

// NewJournalEntryRoot outputs a new JournalEntry implementation used to revert an account's root change
func NewJournalEntryRoot(address *Address, acnt *AccountState, oldRoot []byte) *JournalEntryRoot {
	return &JournalEntryRoot{address: address, acnt: acnt, oldRoot: oldRoot}
}

// Revert apply undo operation
func (jer *JournalEntryRoot) Revert(accounts AccountsHandler) error {
	if accounts == nil {
		return ErrNilAccountsHandler
	}

	if jer.address == nil {
		return ErrNilAddress
	}

	if jer.acnt == nil {
		return ErrNilAccountState
	}

	//access code hash through account pointer
	//as to not register a new entry of root change
	jer.acnt.Account().Root = jer.oldRoot
	err := accounts.RetrieveDataTrie(jer.acnt)
	if err != nil {
		return err
	}
	return accounts.SaveAccountState(jer.acnt)
}

// DirtyAddress returns the address on which this JournalEntry will apply
func (jer *JournalEntryRoot) DirtyAddress() *Address {
	return jer.address
}

//------- JournalEntryData

// NewJournalEntryData outputs a new JournalEntry implementation used to keep track of data change.
// The revert will practically empty the dirty data map
func NewJournalEntryData(trie trie.PatriciaMerkelTree, as *AccountState) *JournalEntryData {
	return &JournalEntryData{trie: trie, as: as}
}

// Revert will empty the dirtyData map from AccountState
func (jed *JournalEntryData) Revert(accounts AccountsHandler) error {
	jed.as.ClearDirty()
	return nil
}

// DirtyAddress will return nil as there is no address involved in saving data
func (jed *JournalEntryData) DirtyAddress() *Address {
	return nil
}

// Trie returns the referenced PatriciaMerkelTree for committing the changes
func (jed *JournalEntryData) Trie() trie.PatriciaMerkelTree {
	return jed.trie
}
