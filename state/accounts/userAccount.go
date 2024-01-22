//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. userAccountData.proto
package accounts

import (
	"bytes"
	"context"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var _ state.UserAccountHandler = (*userAccount)(nil)

// userAccount holds all the information about a user account
type userAccount struct {
	UserAccountData

	dataTrieInteractor
	dataTrieLeafParser common.TrieLeafParser
	code               []byte
	hasNewCode         bool
	version            uint8
}

var zero = big.NewInt(0)

// NewUserAccount creates a new instance of user account
func NewUserAccount(
	address []byte,
	trackableDataTrie state.DataTrieTracker,
	trieLeafParser common.TrieLeafParser,
) (*userAccount, error) {
	if len(address) == 0 {
		return nil, errors.ErrNilAddress
	}
	if check.IfNil(trackableDataTrie) {
		return nil, errors.ErrNilTrackableDataTrie
	}
	if check.IfNil(trieLeafParser) {
		return nil, errors.ErrNilTrieLeafParser
	}

	return &userAccount{
		UserAccountData: UserAccountData{
			DeveloperReward: big.NewInt(0),
			Balance:         big.NewInt(0),
			Address:         address,
		},
		dataTrieInteractor: trackableDataTrie,
		dataTrieLeafParser: trieLeafParser,
	}, nil
}

// AddressBytes returns the address associated with the account as byte slice
func (a *userAccount) AddressBytes() []byte {
	return a.Address
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
		return errors.ErrInsufficientFunds
	}

	a.Balance = newBalance
	return nil
}

// SubFromBalance subtracts new value from balance
func (a *userAccount) SubFromBalance(value *big.Int) error {
	newBalance := big.NewInt(0).Sub(a.Balance, value)
	if newBalance.Cmp(zero) < 0 {
		return errors.ErrInsufficientFunds
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
		return nil, errors.ErrOperationNotPermitted
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
		return errors.ErrOperationNotPermitted
	}
	if len(newAddress) != len(a.Address) {
		return errors.ErrInvalidAddressLength
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

// GetAllLeaves returns all the leaves of the account's data trie
func (a *userAccount) GetAllLeaves(
	leavesChannels *common.TrieIteratorChannels,
	ctx context.Context,
) error {
	dt := a.dataTrieInteractor.DataTrie()
	if check.IfNil(dt) {
		return errors.ErrNilTrie
	}

	rootHash, err := dt.RootHash()
	if err != nil {
		return err
	}

	return dt.GetAllLeavesOnChannel(leavesChannels, ctx, rootHash, keyBuilder.NewKeyBuilder(), a.dataTrieLeafParser)
}

// IsDataTrieMigrated returns true if the data trie is migrated to the latest version
func (a *userAccount) IsDataTrieMigrated() (bool, error) {
	dt := a.dataTrieInteractor.DataTrie()
	if check.IfNil(dt) {
		return false, errors.ErrNilTrie
	}

	return dt.IsMigratedToLatestVersion()
}

// SetCode sets the actual code that needs to be run in the VM
func (a *userAccount) SetCode(code []byte) {
	a.hasNewCode = true
	a.code = code
}

// GetCode returns the code that needs to be run in the VM
func (a *userAccount) GetCode() []byte {
	return a.code
}

// SetVersion sets the account version
func (a *userAccount) SetVersion(version uint8) {
	a.version = version
}

// GetVersion returns the account version
func (a *userAccount) GetVersion() uint8 {
	return a.version
}

// HasNewCode returns true if there was a code change for the account
func (a *userAccount) HasNewCode() bool {
	return a.hasNewCode
}

// AccountDataHandler returns the account data handler
func (a *userAccount) AccountDataHandler() vmcommon.AccountDataHandler {
	return a.dataTrieInteractor
}

// IsInterfaceNil returns true if there is no value under the interface
func (a *userAccount) IsInterfaceNil() bool {
	return a == nil
}
