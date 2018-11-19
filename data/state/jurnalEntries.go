package state

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
)

// JurnalEntryCreation is used to revert an account creation
type JurnalEntryCreation struct {
	address *Address
	acnt    *AccountState
}

// JurnalEntryNonce is used to revert a nonce change
type JurnalEntryNonce struct {
	address  *Address
	acnt     *AccountState
	oldNonce uint64
}

// JurnalEntryBalance is used to revert a balance change
type JurnalEntryBalance struct {
	address    *Address
	acnt       *AccountState
	oldBalance *big.Int
}

// JurnalEntryCodeHash is used to revert a code hash change
type JurnalEntryCodeHash struct {
	address     *Address
	acnt        *AccountState
	oldCodeHash []byte
}

// JurnalEntryCode is used to revert a code addition to the trie
type JurnalEntryCode struct {
	codeHash []byte
}

// JurnalEntryRoot is used to revert an account's root change
type JurnalEntryRoot struct {
	address *Address
	acnt    *AccountState
	oldRoot []byte
}

// JurnalEntryData is used to mark an account's data change
type JurnalEntryData struct {
	trie trie.PatriciaMerkelTree
	as   *AccountState
}

//------- JurnalEntryCreation

// NewJurnalEntryCreation outputs a new JurnalEntry implementation used to revert an account creation
func NewJurnalEntryCreation(address *Address, acnt *AccountState) *JurnalEntryCreation {
	return &JurnalEntryCreation{address: address, acnt: acnt}
}

// Revert apply undo operation
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

	return accounts.RemoveAccount(*jec.address)
}

// DirtyAddress returns the address on which this JurnalEntry will apply
func (jec *JurnalEntryCreation) DirtyAddress() *Address {
	return jec.address
}

//------- JurnalEntryNonce

// NewJurnalEntryNonce outputs a new JurnalEntry implementation used to revert a nonce change
func NewJurnalEntryNonce(address *Address, acnt *AccountState, oldNonce uint64) *JurnalEntryNonce {
	return &JurnalEntryNonce{address: address, oldNonce: oldNonce, acnt: acnt}
}

// Revert apply undo operation
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

	//access nonce through account pointer
	//as to not register a new entry of nonce change
	jen.acnt.Account().Nonce = jen.oldNonce
	return accounts.SaveAccountState(jen.acnt)
}

// DirtyAddress returns the address on which this JurnalEntry will apply
func (jen *JurnalEntryNonce) DirtyAddress() *Address {
	return jen.address
}

//------- JurnalEntryBalance

// NewJurnalEntryBalance outputs a new JurnalEntry implementation used to revert a balance change
func NewJurnalEntryBalance(address *Address, acnt *AccountState, oldBalance *big.Int) *JurnalEntryBalance {
	return &JurnalEntryBalance{address: address, oldBalance: oldBalance, acnt: acnt}
}

// Revert apply undo operation
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

	//access balance through account pointer
	//as to not register a new entry of balance change
	jeb.acnt.Account().Balance = jeb.oldBalance
	return accounts.SaveAccountState(jeb.acnt)
}

// DirtyAddress returns the address on which this JurnalEntry will apply
func (jeb *JurnalEntryBalance) DirtyAddress() *Address {
	return jeb.address
}

//------- JurnalEntryCodeHash

// NewJurnalEntryCodeHash outputs a new JurnalEntry implementation used to revert a code hash change
func NewJurnalEntryCodeHash(address *Address, acnt *AccountState, oldCodeHash []byte) *JurnalEntryCodeHash {
	return &JurnalEntryCodeHash{address: address, acnt: acnt, oldCodeHash: oldCodeHash}
}

// Revert apply undo operation
func (jech *JurnalEntryCodeHash) Revert(accounts AccountsHandler) error {
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

// DirtyAddress returns the address on which this JurnalEntry will apply
func (jech *JurnalEntryCodeHash) DirtyAddress() *Address {
	return jech.address
}

//------- JurnalEntryCode

// NewJurnalEntryCode outputs a new JurnalEntry implementation used to revert a code addition to the trie
func NewJurnalEntryCode(codeHash []byte) *JurnalEntryCode {
	return &JurnalEntryCode{codeHash: codeHash}
}

// Revert apply undo operation
func (jec *JurnalEntryCode) Revert(accounts AccountsHandler) error {
	if accounts == nil {
		return ErrNilAccountsHandler
	}

	return accounts.RemoveCode(jec.codeHash)
}

// DirtyAddress will return nil as there is no address involved in code saving in a trie
func (jec *JurnalEntryCode) DirtyAddress() *Address {
	return nil
}

//------- JurnalEntryRoot

// NewJurnalEntryRoot outputs a new JurnalEntry implementation used to revert an account's root change
func NewJurnalEntryRoot(address *Address, acnt *AccountState, oldRoot []byte) *JurnalEntryRoot {
	return &JurnalEntryRoot{address: address, acnt: acnt, oldRoot: oldRoot}
}

// Revert apply undo operation
func (jer *JurnalEntryRoot) Revert(accounts AccountsHandler) error {
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

// DirtyAddress returns the address on which this JurnalEntry will apply
func (jer *JurnalEntryRoot) DirtyAddress() *Address {
	return jer.address
}

//------- JurnalEntryData

// NewJurnalEntryData outputs a new JurnalEntry implementation used to keep track of data change.
// The revert will practically empty the dirty data map
func NewJurnalEntryData(trie trie.PatriciaMerkelTree, as *AccountState) *JurnalEntryData {
	return &JurnalEntryData{trie: trie, as: as}
}

// Revert will empty the dirtyData map from AccountState
func (jed *JurnalEntryData) Revert(accounts AccountsHandler) error {
	jed.as.ClearDirty()
	return nil
}

// DirtyAddress will return nil as there is no address involved in saving data
func (jed *JurnalEntryData) DirtyAddress() *Address {
	return nil
}

// Trie returns the referenced PatriciaMerkelTree for committing the changes
func (jed *JurnalEntryData) Trie() trie.PatriciaMerkelTree {
	return jed.trie
}
