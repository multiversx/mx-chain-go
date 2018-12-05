package state

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
)

// SimpleAccountWrap wraps the AccountContainer adding data trie, code slice and address
type SimpleAccountWrap struct {
	AccountContainer

	addressContainer AddressContainer
	code             []byte
	dataTrie         trie.PatriciaMerkelTree
}

// NewSimpleAccountWrap creates new simple account wrapper for an AccountContainer (that has just been initialized)
func NewSimpleAccountWrap(addressContainer AddressContainer, accountContainer AccountContainer) (*SimpleAccountWrap, error) {
	if addressContainer == nil {
		return nil, ErrNilAddressContainer
	}

	if accountContainer == nil {
		return nil, ErrNilAccountContainer
	}

	return &SimpleAccountWrap{
		AccountContainer: accountContainer,
		addressContainer: addressContainer,
	}, nil
}

// AddressContainer returns the address associated with the account
func (saw *SimpleAccountWrap) AddressContainer() AddressContainer {
	return saw.addressContainer
}

// Code gets the actual code that needs to be run in the VM
func (saw *SimpleAccountWrap) Code() []byte {
	return saw.code
}

// SetCode sets the actual code that needs to be run in the VM
func (saw *SimpleAccountWrap) SetCode(code []byte) {
	saw.code = code
}

// DataTrie returns the trie that holds the current account's data
func (saw *SimpleAccountWrap) DataTrie() trie.PatriciaMerkelTree {
	return saw.dataTrie
}

// SetDataTrie sets the trie that holds the current account's data
func (saw *SimpleAccountWrap) SetDataTrie(trie trie.PatriciaMerkelTree) {
	saw.dataTrie = trie
}
