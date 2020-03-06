package accounts

import (
	"bytes"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// Account is the struct used in serialization/deserialization
type userAccount struct {
	*baseAccount
	Balance *big.Int
	Address []byte

	DeveloperReward *big.Int
	OwnerAddress    []byte
}

func NewEmptyUserAccount() *userAccount {
	return &userAccount{
		baseAccount:     &baseAccount{},
		Balance:         big.NewInt(0),
		DeveloperReward: big.NewInt(0),
	}
}

var zero = big.NewInt(0)

// NewUserAccount creates new simple account wrapper for an AccountContainer (that has just been initialized)
func NewUserAccount(addressContainer state.AddressContainer) (*userAccount, error) {
	if check.IfNil(addressContainer) {
		return nil, state.ErrNilAddressContainer
	}

	addressBytes := addressContainer.Bytes()

	return &userAccount{
		baseAccount: &baseAccount{
			addressContainer: addressContainer,
			dataTrieTracker:  state.NewTrackableDataTrie(addressBytes, nil),
		},
		DeveloperReward: big.NewInt(0),
		Balance:         big.NewInt(0),
		Address:         addressBytes,
	}, nil
}

// AddToBalance adds new value to balance
func (a *userAccount) AddToBalance(value *big.Int) error {
	newBalance := big.NewInt(0).Add(a.Balance, value)
	if newBalance.Cmp(zero) < 0 {
		return state.ErrInsufficientFunds
	}

	a.Balance = newBalance
	return nil
}

// SubFromBalance subtracts new value from balance
func (a *userAccount) SubFromBalance(value *big.Int) error {
	newBalance := big.NewInt(0).Sub(a.Balance, value)
	if newBalance.Cmp(zero) < 0 {
		return state.ErrInsufficientFunds
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
		return nil, state.ErrOperationNotPermitted
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
		return state.ErrOperationNotPermitted
	}
	if len(newAddress) != len(a.addressContainer.Bytes()) {
		return state.ErrInvalidAddressLength
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

// IsInterfaceNil returns true if there is no value under the interface
func (a *userAccount) IsInterfaceNil() bool {
	return a == nil
}
