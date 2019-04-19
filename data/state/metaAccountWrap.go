package state

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
)

// MetaAccountWrap wraps the AccountContainer adding data trie, code slice and address
type MetaAccountWrap struct {
	*MetaAccount

	addressContainer AddressContainer
	code             []byte
	dataTrie         trie.PatriciaMerkelTree
}

// NewMetaAccountWrap creates new simple account wrapper for an AccountContainer (that has just been initialized)
func NewMetaAccountWrap(addressContainer AddressContainer, account *MetaAccount) (*MetaAccountWrap, error) {
	if addressContainer == nil {
		return nil, ErrNilAddressContainer
	}
	if account == nil {
		return nil, ErrNilAccount
	}

	return &MetaAccountWrap{
		MetaAccount:      account,
		addressContainer: addressContainer,
	}, nil
}

// AddressContainer returns the address associated with the account
func (aw *MetaAccountWrap) AddressContainer() AddressContainer {
	return aw.addressContainer
}

// Code gets the actual code that needs to be run in the VM
func (aw *MetaAccountWrap) Code() []byte {
	return aw.code
}

// SetCode sets the actual code that needs to be run in the VM
func (aw *MetaAccountWrap) SetCode(code []byte) {
	aw.code = code
}

// DataTrie returns the trie that holds the current account's data
func (aw *MetaAccountWrap) DataTrie() trie.PatriciaMerkelTree {
	return aw.dataTrie
}

// SetDataTrie sets the trie that holds the current account's data
func (aw *MetaAccountWrap) SetDataTrie(trie trie.PatriciaMerkelTree) {
	aw.dataTrie = trie
}

// BaseAccount returns the account held by this wrapper
func (aw *MetaAccountWrap) BaseAccount() AccountInterface {
	return aw.MetaAccount
}

// AppendRegistrationData will add a new RegistrationData to the slice
func (aw *MetaAccountWrap) AppendRegistrationData(data *RegistrationData) error {
	if !bytes.Equal(aw.addressContainer.Bytes(), RegistrationAddress.Bytes()) {
		return ErrNotSupportedAccountsRegistration
	}

	aw.RegistrationData = append(aw.RegistrationData, *data)
	return nil
}

// TrimLastRegistrationData will remove the last item added in the slice
func (aw *MetaAccountWrap) TrimLastRegistrationData() error {
	if !bytes.Equal(aw.addressContainer.Bytes(), RegistrationAddress.Bytes()) {
		return ErrNotSupportedAccountsRegistration
	}

	if len(aw.RegistrationData) == 0 {
		return ErrTrimOperationNotSupported
	}

	aw.RegistrationData = aw.RegistrationData[:len(aw.RegistrationData)-1]
	return nil
}

// CleanRegistrationData will clean-up the RegistrationData slice
func (aw *MetaAccountWrap) CleanRegistrationData() error {
	//TODO when it has been establied how and when to call this method
	return nil
}
