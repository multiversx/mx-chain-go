package state

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
)

// AccountWrap wraps the AccountContainer adding data trie, code slice and address
type AccountWrap struct {
	*Account

	addressContainer AddressContainer
	code             []byte
	dataTrie         trie.PatriciaMerkelTree
}

// NewAccountWrap creates new simple account wrapper for an AccountContainer (that has just been initialized)
func NewAccountWrap(addressContainer AddressContainer, account *Account) (*AccountWrap, error) {
	if addressContainer == nil {
		return nil, ErrNilAddressContainer
	}

	if account == nil {
		return nil, ErrNilAccount
	}

	return &AccountWrap{
		Account:          account,
		addressContainer: addressContainer,
	}, nil
}

// AddressContainer returns the address associated with the account
func (aw *AccountWrap) AddressContainer() AddressContainer {
	return aw.addressContainer
}

// Code gets the actual code that needs to be run in the VM
func (aw *AccountWrap) Code() []byte {
	return aw.code
}

// SetCode sets the actual code that needs to be run in the VM
func (aw *AccountWrap) SetCode(code []byte) {
	aw.code = code
}

// DataTrie returns the trie that holds the current account's data
func (aw *AccountWrap) DataTrie() trie.PatriciaMerkelTree {
	return aw.dataTrie
}

// SetDataTrie sets the trie that holds the current account's data
func (aw *AccountWrap) SetDataTrie(trie trie.PatriciaMerkelTree) {
	aw.dataTrie = trie
}

// BaseAccount returns the account held by this wrapper
func (aw *AccountWrap) BaseAccount() *Account {
	return aw.Account
}

func (aw *AccountWrap) AppendRegistrationData(data *RegistrationData) error {
	if !bytes.Equal(aw.addressContainer.Bytes(), RegistrationAddress.Bytes()) {
		return ErrNotSupportedAccountsRegistration
	}

	aw.RegistrationData = append(aw.RegistrationData, *data)
	return nil
}

func (aw *AccountWrap) TrimLastRegistrationData() error {
	if !bytes.Equal(aw.addressContainer.Bytes(), RegistrationAddress.Bytes()) {
		return ErrNotSupportedAccountsRegistration
	}

	if len(aw.RegistrationData) == 0 {
		return ErrTrimOperationNotSupported
	}

	aw.RegistrationData = aw.RegistrationData[:len(aw.RegistrationData)-1]
	return nil
}

func (aw *AccountWrap) CleanRegistrationData() error {
	//TODO when it has been establied how and when to call this method
	return nil
}
