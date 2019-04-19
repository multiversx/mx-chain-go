package state

import (
	"bytes"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
)

// ActionRequested defines the type used to refer to the action requested in the registration transaction
type ActionRequested int

const (
	// ArRegister defines a requested action of registration
	ArRegister ActionRequested = iota
	// ArUnregister defines a requested action of unregistration
	ArUnregister
)

// RegistrationData holds the data which is sent with a registration transaction
type RegistrationData struct {
	OriginatorPubKey []byte
	NodePubKey       []byte
	Stake            *big.Int
	Action           ActionRequested
	RoundIndex       int32
	EpochIndex       int32
}

// Account is the struct used in serialization/deserialization
type Account struct {
	Nonce            uint64
	Balance          *big.Int
	CodeHash         []byte
	RootHash         []byte
	RegistrationData []RegistrationData

	TrackableDataTrie
	addressContainer AddressContainer
	code             []byte
	dataTrie         trie.PatriciaMerkelTree
}

func (acc *Account) IntegrityAndValidity() error {
	if acc.CodeHash == nil {
		return data.ErrNilCodeHash
	}
	if acc.RootHash == nil {
		return data.ErrNilRootHash
	}
	if acc.RegistrationData == nil {
		return data.ErrNilRegistrationData
	}
	if acc.Balance == nil {
		return data.ErrNilBalance
	}
	if acc.Balance.Sign() < 0 {
		return data.ErrNegativeBalance
	}

	return nil
}

func (acc *Account) SetNonce(nonce uint64) {
	acc.Nonce = nonce
}

func (acc *Account) GetRootHash() []byte {
	return acc.RootHash
}

func (acc *Account) SetRootHash(roothash []byte) {
	if roothash == nil {
		return
	}
	acc.RootHash = roothash
}

func (acc *Account) GetCodeHash() []byte {
	return acc.CodeHash
}

func (acc *Account) SetCodeHash(roothash []byte) {
	if roothash == nil {
		return
	}
	acc.CodeHash = roothash
}

// NewAccount creates new simple account wrapper for an AccountContainer (that has just been initialized)
func NewAccount(addressContainer AddressContainer) (*Account, error) {
	if addressContainer == nil {
		return nil, ErrNilAddressContainer
	}

	return &Account{
		Balance:          big.NewInt(0),
		addressContainer: addressContainer,
	}, nil
}

// AddressContainer returns the address associated with the account
func (a *Account) AddressContainer() AddressContainer {
	return a.addressContainer
}

// Code gets the actual code that needs to be run in the VM
func (a *Account) Code() []byte {
	return a.code
}

// SetCode sets the actual code that needs to be run in the VM
func (a *Account) SetCode(code []byte) {
	a.code = code
}

// DataTrie returns the trie that holds the current account's data
func (a *Account) DataTrie() trie.PatriciaMerkelTree {
	return a.dataTrie
}

// SetDataTrie sets the trie that holds the current account's data
func (a *Account) SetDataTrie(trie trie.PatriciaMerkelTree) {
	a.dataTrie = trie
}

// AppendRegistrationData will add a new RegistrationData to the slice
func (a *Account) AppendRegistrationData(data *RegistrationData) error {
	if !bytes.Equal(a.addressContainer.Bytes(), RegistrationAddress.Bytes()) {
		return ErrNotSupportedAccountsRegistration
	}

	a.RegistrationData = append(a.RegistrationData, *data)
	return nil
}

// TrimLastRegistrationData will remove the last item added in the slice
func (a *Account) TrimLastRegistrationData() error {
	if !bytes.Equal(a.addressContainer.Bytes(), RegistrationAddress.Bytes()) {
		return ErrNotSupportedAccountsRegistration
	}

	if len(a.RegistrationData) == 0 {
		return ErrTrimOperationNotSupported
	}

	a.RegistrationData = a.RegistrationData[:len(a.RegistrationData)-1]
	return nil
}

// CleanRegistrationData will clean-up the RegistrationData slice
func (a *Account) CleanRegistrationData() error {
	//TODO when it has been establied how and when to call this method
	return nil
}

//TODO add Cap'N'Proto converter funcs
