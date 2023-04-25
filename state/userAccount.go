//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. userAccountData.proto
package state

import (
	"bytes"
	"math/big"

	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var _ UserAccountHandler = (*userAccount)(nil)

// Account is the struct used in serialization/deserialization
type userAccount struct {
	*baseAccount
	UserAccountData
}

var zero = big.NewInt(0)

// NewEmptyUserAccount creates new simple account wrapper for an AccountContainer (that has just been initialized)
func NewEmptyUserAccount() *userAccount {
	return &userAccount{
		baseAccount: &baseAccount{},
		UserAccountData: UserAccountData{
			DeveloperReward: big.NewInt(0),
			Balance:         big.NewInt(0),
		},
	}
}

// NewUserAccount creates new simple account wrapper for an AccountContainer (that has just been initialized)
func NewUserAccount(address []byte) (*userAccount, error) {
	if len(address) == 0 {
		return nil, ErrNilAddress
	}

	return &userAccount{
		baseAccount: &baseAccount{
			address:         address,
			dataTrieTracker: NewTrackableDataTrie(address, nil),
		},
		UserAccountData: UserAccountData{
			DeveloperReward: big.NewInt(0),
			Balance:         big.NewInt(0),
			Address:         address,
		},
	}, nil
}

// SetUserName sets the users name
func (a *userAccount) SetUserName(userName []byte) {
	a.UserName = make([]byte, 0, len(userName))
	a.UserName = append(a.UserName, userName...)
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

// ClaimDeveloperRewards returns the accumulated developer rewards and sets it to 0 in the account
func (a *userAccount) ClaimDeveloperRewards(sndAddress []byte) (*big.Int, error) {
	if !bytes.Equal(sndAddress, a.OwnerAddress) {
		return nil, ErrOperationNotPermitted
	}

	oldValue := big.NewInt(0).Set(a.DeveloperReward)
	a.DeveloperReward = big.NewInt(0)

	return oldValue, nil
}

// AddToDeveloperReward adds new value to developer reward
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
	if len(newAddress) != len(a.address) {
		return ErrInvalidAddressLength
	}

	a.OwnerAddress = newAddress

	return nil
}

// SetOwnerAddress sets the owner address of an account,
func (a *userAccount) SetOwnerAddress(address []byte) {
	a.OwnerAddress = address
}

// IncreaseNonce adds the given value to the current nonce
func (a *userAccount) IncreaseNonce(value uint64) {
	a.Nonce = a.Nonce + value
}

// SetCodeHash sets the code hash associated with the account
func (a *userAccount) SetCodeHash(codeHash []byte) {
	a.CodeHash = codeHash
}

// SetRootHash sets the root hash associated with the account
func (a *userAccount) SetRootHash(roothash []byte) {
	a.RootHash = roothash
}

// SetCodeMetadata sets the code metadata
func (a *userAccount) SetCodeMetadata(codeMetadata []byte) {
	a.CodeMetadata = codeMetadata
}

// IsGuarded returns true if the account is in guarded state
func (a *userAccount) IsGuarded() bool {
	codeMetaDataBytes := a.GetCodeMetadata()
	codeMetaData := vmcommon.CodeMetadataFromBytes(codeMetaDataBytes)
	return codeMetaData.Guarded
}

// IsInterfaceNil returns true if there is no value under the interface
func (a *userAccount) IsInterfaceNil() bool {
	return a == nil
}
