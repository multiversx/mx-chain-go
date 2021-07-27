package state

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
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
func (ba *baseAccount) DataTrie() temporary.Trie {
	return ba.dataTrieTracker.DataTrie()
}

// SetDataTrie sets the trie that holds the current account's data
func (ba *baseAccount) SetDataTrie(trie temporary.Trie) {
	ba.dataTrieTracker.SetDataTrie(trie)
}

// DataTrieTracker returns the trie wrapper used in managing the SC data
func (ba *baseAccount) DataTrieTracker() DataTrieTracker {
	return ba.dataTrieTracker
}

// RetrieveValueFromDataTrieTracker  fetches the value from a particular key searching the account data store in the data trie tracker
func (ba *baseAccount) RetrieveValueFromDataTrieTracker(key []byte) ([]byte, error) {
	if check.IfNil(ba.dataTrieTracker) {
		return nil, ErrNilTrackableDataTrie
	}

	return ba.dataTrieTracker.RetrieveValue(key)
}

// AccountDataHandler returns the account data handler
func (ba *baseAccount) AccountDataHandler() vmcommon.AccountDataHandler {
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
