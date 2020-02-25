package accounts

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

type baseAccount struct {
	Nonce    uint64
	CodeHash []byte
	RootHash []byte

	addressContainer state.AddressContainer
	code             []byte
	dataTrieTracker  state.DataTrieTracker
}

// AddressContainer returns the address associated with the account
func (ba *baseAccount) AddressContainer() state.AddressContainer {
	return ba.addressContainer
}

//SetNonce saves the nonce to the account
func (ba *baseAccount) SetNonce(nonce uint64) {
	ba.Nonce = nonce
}

// GetNonce gets the nonce of the account
func (ba *baseAccount) GetNonce() uint64 {
	return ba.Nonce
}

// GetCode gets the actual code that needs to be run in the VM
func (ba *baseAccount) GetCode() []byte {
	return ba.code
}

// SetCode sets the actual code that needs to be run in the VM
func (ba *baseAccount) SetCode(code []byte) {
	ba.code = code
}

// GetCodeHash returns the code hash associated with this account
func (ba *baseAccount) GetCodeHash() []byte {
	return ba.CodeHash
}

// SetCodeHash sets the code hash associated with the account
func (ba *baseAccount) SetCodeHash(codeHash []byte) {
	ba.CodeHash = codeHash
}

// GetRootHash returns the root hash associated with this account
func (ba *baseAccount) GetRootHash() []byte {
	return ba.RootHash
}

// SetRootHash sets the root hash associated with the account
func (ba *baseAccount) SetRootHash(roothash []byte) {
	ba.RootHash = roothash
}

// DataTrie returns the trie that holds the current account's data
func (ba *baseAccount) DataTrie() data.Trie {
	return ba.dataTrieTracker.DataTrie()
}

// SetDataTrie sets the trie that holds the current account's data
func (ba *baseAccount) SetDataTrie(trie data.Trie) {
	ba.dataTrieTracker.SetDataTrie(trie)
}

// DataTrieTracker returns the trie wrapper used in managing the SC data
func (ba *baseAccount) DataTrieTracker() state.DataTrieTracker {
	return ba.dataTrieTracker
}

// IsInterfaceNil returns true if there is no value under the interface
func (ba *baseAccount) IsInterfaceNil() bool {
	return ba == nil
}
