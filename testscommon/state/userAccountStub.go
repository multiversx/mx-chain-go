//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. accountWrapperMock.proto
package state

import (
	"math/big"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var _ state.UserAccountHandler = (*UserAccountStub)(nil)

// UserAccountStub -
type UserAccountStub struct {
	Balance                  *big.Int
	DeveloperRewards         *big.Int
	UserName                 []byte
	Owner                    []byte
	Address                  []byte
	Nonce               uint64
	AddToBalanceCalled       func(value *big.Int) error
	DataTrieTrackerCalled    func() state.DataTrieTracker
	IsGuardedCalled          func() bool
	AccountDataHandlerCalled func() vmcommon.AccountDataHandler
	RetrieveValueCalled      func(_ []byte) ([]byte, uint32, error)
}

// HasNewCode -
func (u *UserAccountStub) HasNewCode() bool {
	return false
}

// SetUserName -
func (u *UserAccountStub) SetUserName(_ []byte) {
}

// GetUserName -
func (u *UserAccountStub) GetUserName() []byte {
	return u.UserName
}

// AddToBalance -
func (u *UserAccountStub) AddToBalance(value *big.Int) error {
	if u.AddToBalanceCalled != nil {
		return u.AddToBalanceCalled(value)
	}
	return nil
}

// SubFromBalance -
func (u *UserAccountStub) SubFromBalance(_ *big.Int) error {
	return nil
}

// GetBalance -
func (u *UserAccountStub) GetBalance() *big.Int {
	return u.Balance
}

// ClaimDeveloperRewards -
func (u *UserAccountStub) ClaimDeveloperRewards([]byte) (*big.Int, error) {
	return nil, nil
}

// AddToDeveloperReward -
func (u *UserAccountStub) AddToDeveloperReward(*big.Int) {

}

// GetDeveloperReward -
func (u *UserAccountStub) GetDeveloperReward() *big.Int {
	return u.DeveloperRewards
}

// ChangeOwnerAddress -
func (u *UserAccountStub) ChangeOwnerAddress([]byte, []byte) error {
	return nil
}

// SetOwnerAddress -
func (u *UserAccountStub) SetOwnerAddress([]byte) {

}

// GetOwnerAddress -
func (u *UserAccountStub) GetOwnerAddress() []byte {
	return u.Owner
}

// AddressBytes -
func (u *UserAccountStub) AddressBytes() []byte {
	return u.Address
}

//IncreaseNonce -
func (u *UserAccountStub) IncreaseNonce(_ uint64) {
}

// GetNonce -
func (u *UserAccountStub) GetNonce() uint64 {
	return u.Nonce
}

// SetCode -
func (u *UserAccountStub) SetCode(_ []byte) {

}

// SetCodeMetadata -
func (u *UserAccountStub) SetCodeMetadata(_ []byte) {
}

// GetCodeMetadata -
func (u *UserAccountStub) GetCodeMetadata() []byte {
	return nil
}

// SetCodeHash -
func (u *UserAccountStub) SetCodeHash([]byte) {

}

// GetCodeHash -
func (u *UserAccountStub) GetCodeHash() []byte {
	return nil
}

// SetRootHash -
func (u *UserAccountStub) SetRootHash([]byte) {

}

// GetRootHash -
func (u *UserAccountStub) GetRootHash() []byte {
	return nil
}

// SetDataTrie -
func (u *UserAccountStub) SetDataTrie(_ common.Trie) {

}

// DataTrie -
func (u *UserAccountStub) DataTrie() common.DataTrieHandler {
	return nil
}

// RetrieveValue -
func (u *UserAccountStub) RetrieveValue(key []byte) ([]byte, uint32, error) {
	if u.RetrieveValueCalled != nil {
		return u.RetrieveValueCalled(key)
	}

	return nil, 0, nil
}

// SaveKeyValue -
func (u *UserAccountStub) SaveKeyValue(_ []byte, _ []byte) error {
	return nil
}

// IsGuarded -
func (u *UserAccountStub) IsGuarded() bool {
	if u.IsGuardedCalled != nil {
		return u.IsGuardedCalled()
	}
	return false
}

// SaveDirtyData -
func (u *UserAccountStub) SaveDirtyData(_ common.Trie) (map[string][]byte, error) {
	return nil, nil
}

// IsInterfaceNil -
func (u *UserAccountStub) IsInterfaceNil() bool {
	return u == nil
}

// AccountDataHandler -
func (u *UserAccountStub) AccountDataHandler() vmcommon.AccountDataHandler {
	if u.AccountDataHandlerCalled != nil {
		return u.AccountDataHandlerCalled()
	}
	return nil
}
