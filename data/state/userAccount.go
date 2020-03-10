//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogoslick_out=. userAccountData.proto
package state

import (
	"bytes"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/check"
)

// Account is the struct used in serialization/deserialization
type userAccount struct {
	*baseAccount
	UserAccountData
}

func NewEmptyUserAccount() *userAccount {
	return &userAccount{
		baseAccount: &baseAccount{},
		UserAccountData: UserAccountData{
			Balance:         big.NewInt(0),
			DeveloperReward: big.NewInt(0),
		},
	}
}

var zero = big.NewInt(0)

// NewUserAccount creates new simple account wrapper for an AccountContainer (that has just been initialized)
func NewUserAccount(addressContainer AddressContainer) (*userAccount, error) {
	if check.IfNil(addressContainer) {
		return nil, ErrNilAddressContainer
	}

	addressBytes := addressContainer.Bytes()

	return &userAccount{
		baseAccount: &baseAccount{
			addressContainer: addressContainer,
			dataTrieTracker:  NewTrackableDataTrie(addressBytes, nil),
		},
		UserAccountData: UserAccountData{
			DeveloperReward: big.NewInt(0),
			Balance:         big.NewInt(0),
			Address:         addressBytes,
		},
	}, nil
}

// AddToBalance adds new value to balance
func (a *userAccount) AddToBalance(value *big.Int) error {
	newBalance := big.NewInt(0).Add(a.Balance, value)
	if newBalance.Cmp(zero) < 0 {
		return ErrInsufficientFunds
	}

	a.Balance = newBalance
	return nil
}

// SubFromBalance subtracts new value from balance
func (a *userAccount) SubFromBalance(value *big.Int) error {
	newBalance := big.NewInt(0).Sub(a.Balance, value)
	if newBalance.Cmp(zero) < 0 {
		return ErrInsufficientFunds
	}

	a.Balance = newBalance
	return nil
}

// GetBalance returns the actual balance from the account
func (a *userAccount) GetBalance() *big.Int {
	return big.NewInt(0).Set(a.Balance)
}

func (a *userAccount) ClaimDeveloperRewards(sndAddress []byte) (*big.Int, error) {
	if !bytes.Equal(sndAddress, a.OwnerAddress) {
		return nil, ErrOperationNotPermitted
	}

	oldValue := big.NewInt(0).Set(a.DeveloperReward)
	a.DeveloperReward = big.NewInt(0)

	return oldValue, nil
}

func (a *userAccount) AddToDeveloperReward(value *big.Int) {
	a.DeveloperReward = big.NewInt(0).Add(a.DeveloperReward, value)
}

// GetDeveloperReward returns the actual developer reward from the account
func (a *userAccount) GetDeveloperReward() *big.Int {
	return big.NewInt(0).Set(a.DeveloperReward)
}

// ChangeOwnerAddress changes the owner account if operation is permitted
func (a *userAccount) ChangeOwnerAddress(sndAddress []byte, newAddress []byte) error {
	if !bytes.Equal(sndAddress, a.OwnerAddress) {
		return ErrOperationNotPermitted
	}
	if len(newAddress) != len(a.addressContainer.Bytes()) {
		return ErrInvalidAddressLength
	}

	a.OwnerAddress = newAddress

	return nil
}

func (a *userAccount) SetOwnerAddress(address []byte) {
	a.OwnerAddress = address
}

// GetOwnerAddress returns the actual owner address from the account
func (a *userAccount) GetOwnerAddress() []byte {
	return a.OwnerAddress
}

//SetNonce saves the nonce to the account
func (a *userAccount) SetNonce(nonce uint64) {
	a.Nonce = nonce
}

// GetNonce gets the nonce of the account
func (a *userAccount) GetNonce() uint64 {
	return a.Nonce
}

// GetCodeHash returns the code hash associated with this account
func (a *userAccount) GetCodeHash() []byte {
	return a.CodeHash
}

// SetCodeHash sets the code hash associated with the account
func (a *userAccount) SetCodeHash(codeHash []byte) {
	a.CodeHash = codeHash
}

// GetRootHash returns the root hash associated with this account
func (a *userAccount) GetRootHash() []byte {
	return a.RootHash
}

// SetRootHash sets the root hash associated with the account
func (a *userAccount) SetRootHash(roothash []byte) {
	a.RootHash = roothash
}

// IsInterfaceNil returns true if there is no value under the interface
func (a *userAccount) IsInterfaceNil() bool {
	return a == nil
}
