package state

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type baseAccount struct {
	address         []byte
	code            []byte
	dataTrieTracker DataTrieTracker
	hasNewCode      bool
}

// AddressBytes returns the address associated with the account as byte slice
func (ba *baseAccount) AddressBytes() []byte {
	return ba.address
}

// SetCode sets the actual code that needs to be run in the VM
func (ba *baseAccount) SetCode(code []byte) {
	ba.hasNewCode = true
	ba.code = code
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
func (ba *baseAccount) DataTrieTracker() DataTrieTracker {
	return ba.dataTrieTracker
}

// HasNewCode returns true if there was a code change for the account
func (ba *baseAccount) HasNewCode() bool {
	return ba.hasNewCode
}

// IsInterfaceNil returns true if there is no value under the interface
func (ba *baseAccount) IsInterfaceNil() bool {
	return ba == nil
}
