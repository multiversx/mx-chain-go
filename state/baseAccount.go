package state

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
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
func (ba *baseAccount) DataTrie() common.DataTrieHandler {
	return ba.dataTrieTracker.DataTrie()
}

// SetDataTrie sets the trie that holds the current account's data
func (ba *baseAccount) SetDataTrie(trie common.Trie) {
	ba.dataTrieTracker.SetDataTrie(trie)
}

// RetrieveValue fetches the value from a particular key searching the account data store in the data trie tracker
func (ba *baseAccount) RetrieveValue(key []byte) ([]byte, uint32, error) {
	if check.IfNil(ba.dataTrieTracker) {
		return nil, 0, ErrNilTrackableDataTrie
	}

	return ba.dataTrieTracker.RetrieveValue(key)
}

// SaveKeyValue adds the given key and value to the underlying trackable data trie
func (ba *baseAccount) SaveKeyValue(key []byte, value []byte) error {
	if check.IfNil(ba.dataTrieTracker) {
		return ErrNilTrackableDataTrie
	}

	return ba.dataTrieTracker.SaveKeyValue(key, value)
}

// SaveDirtyData triggers SaveDirtyData form the underlying trackableDataTrie
func (ba *baseAccount) SaveDirtyData(trie common.Trie) ([]core.TrieData, error) {
	if check.IfNil(ba.dataTrieTracker) {
		return nil, ErrNilTrackableDataTrie
	}

	return ba.dataTrieTracker.SaveDirtyData(trie)
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
