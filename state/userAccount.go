//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. userAccountData.proto
package state

import (
	"bytes"
	"context"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state/parsers"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var _ UserAccountHandler = (*userAccount)(nil)

// Account is the struct used in serialization/deserialization
type userAccount struct {
	*baseAccount
	UserAccountData
	marshaller          marshal.Marshalizer
	enableEpochsHandler common.EnableEpochsHandler
}

var zero = big.NewInt(0)

// ArgsAccountCreation holds the arguments needed to create a new instance of userAccount
type ArgsAccountCreation struct {
	Hasher              hashing.Hasher
	Marshaller          marshal.Marshalizer
	EnableEpochsHandler common.EnableEpochsHandler
}

// NewUserAccount creates a new instance of userAccount
func NewUserAccount(
	address []byte,
	args ArgsAccountCreation,
) (*userAccount, error) {
	if len(address) == 0 {
		return nil, ErrNilAddress
	}
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	tdt, err := NewTrackableDataTrie(address, nil, args.Hasher, args.Marshaller, args.EnableEpochsHandler)
	if err != nil {
		return nil, err
	}

	return &userAccount{
		baseAccount: &baseAccount{
			address:         address,
			dataTrieTracker: tdt,
		},
		UserAccountData: UserAccountData{
			DeveloperReward: big.NewInt(0),
			Balance:         big.NewInt(0),
			Address:         address,
		},
		marshaller:          args.Marshaller,
		enableEpochsHandler: args.EnableEpochsHandler,
	}, nil
}

// NewUserAccountFromBytes creates a new instance of userAccount from the given bytes
func NewUserAccountFromBytes(
	accountBytes []byte,
	args ArgsAccountCreation,
) (*userAccount, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	acc := &userAccount{}
	err = args.Marshaller.Unmarshal(acc, accountBytes)
	if err != nil {
		return nil, err
	}

	tdt, err := NewTrackableDataTrie(acc.Address, nil, args.Hasher, args.Marshaller, args.EnableEpochsHandler)
	if err != nil {
		return nil, err
	}

	acc.baseAccount = &baseAccount{
		address:         acc.Address,
		dataTrieTracker: tdt,
	}
	acc.marshaller = args.Marshaller
	acc.enableEpochsHandler = args.EnableEpochsHandler

	return acc, nil
}

func checkArgs(args ArgsAccountCreation) error {
	if check.IfNil(args.Marshaller) {
		return ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return ErrNilHasher
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return ErrNilEnableEpochsHandler
	}

	return nil
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

// GetAllLeaves returns all the leaves of the account's data trie
func (a *userAccount) GetAllLeaves(
	leavesChannels *common.TrieIteratorChannels,
	ctx context.Context,
) error {
	dataTrie := a.dataTrieTracker.DataTrie()
	if check.IfNil(dataTrie) {
		return ErrNilTrie
	}

	rootHash, err := dataTrie.RootHash()
	if err != nil {
		return err
	}

	tlp, err := parsers.NewDataTrieLeafParser(a.Address, a.marshaller, a.enableEpochsHandler)
	if err != nil {
		return err
	}

	return dataTrie.GetAllLeavesOnChannel(leavesChannels, ctx, rootHash, keyBuilder.NewKeyBuilder(), tlp)
}

// IsDataTrieMigrated returns true if the data trie is migrated to the latest version
func (a *userAccount) IsDataTrieMigrated() (bool, error) {
	dt := a.dataTrieTracker.DataTrie()
	if check.IfNil(dt) {
		return false, ErrNilTrie
	}

	return dt.IsMigratedToLatestVersion()
}

// IsInterfaceNil returns true if there is no value under the interface
func (a *userAccount) IsInterfaceNil() bool {
	return a == nil
}
