package accounts

import (
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

// SetBalance sets the given balance to the account
func (a *userAccount) SetBalance(value *big.Int) {
	a.Balance = value
}

// GetBalance returns the actual balance from the account
func (a *userAccount) GetBalance() *big.Int {
	return big.NewInt(0).Set(a.Balance)
}

// SetDeveloperReward sets the given developer reward to the account
func (a *userAccount) SetDeveloperReward(value *big.Int) {
	a.DeveloperReward = value
}

// GetDeveloperReward returns the actual developer reward from the account
func (a *userAccount) GetDeveloperReward() *big.Int {
	return big.NewInt(0).Set(a.DeveloperReward)
}

// SetOwnerAddress sets the given owner address to the account
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
